package delivery

import (
	"io/fs"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	sessions "github.com/maxshend/grader/pkg/sessions"
	"github.com/maxshend/grader/pkg/users"
	"github.com/maxshend/grader/pkg/users/services"
	"github.com/maxshend/grader/pkg/utils"
)

type UsersHttpHandler struct {
	Service        services.UsersServiceInterface
	SessionManager sessions.HttpSessionManager
	Views          map[string]*utils.View
}

type userFormData struct {
	User   *users.User
	Errors []string
}

func NewUsersHttpHandler(
	service services.UsersServiceInterface,
	sessionManager sessions.HttpSessionManager,
	templatesFS fs.FS,
) (*UsersHttpHandler, error) {
	views := make(map[string]*utils.View)
	var err error

	views["Signup"], err = utils.NewView(templatesFS, "templates/users/signup.gohtml")
	if err != nil {
		return nil, err
	}

	views["GetAll"], err = utils.NewView(templatesFS, "templates/users/admin/list.gohtml")
	if err != nil {
		return nil, err
	}

	views["UserForm"], err = utils.NewView(templatesFS, "templates/users/admin/user_form.gohtml")
	if err != nil {
		return nil, err
	}

	views["ProfileForm"], err = utils.NewView(templatesFS, "templates/users/profile_form.gohtml")
	if err != nil {
		return nil, err
	}

	return &UsersHttpHandler{
		Service:        service,
		Views:          views,
		SessionManager: sessionManager,
	}, nil
}

func (h UsersHttpHandler) EditProfile(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = h.Views["ProfileForm"].RenderView(
		w,
		&userFormData{User: currentUser},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}

func (h UsersHttpHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := h.Service.CheckCredentials(currentUser.Username, r.FormValue("current_password"))
	if err != nil {
		err := h.Views["ProfileForm"].RenderView(
			w,
			&userFormData{User: currentUser, Errors: []string{services.MsgInvalidCurrentPassword}},
			currentUser,
		)
		if err != nil {
			utils.RenderInternalError(w, r, err)
		}

		return
	}

	user.Username = r.FormValue("username")
	_, err = h.Service.UpdateProfile(
		user,
		r.FormValue("new_password"),
		r.FormValue("new_password_confirmation"),
	)
	if err != nil {
		if _, ok := err.(*services.UserValidationError); ok {
			err := h.Views["ProfileForm"].RenderView(
				w,
				&userFormData{User: user, Errors: []string{err.Error()}},
				currentUser,
			)
			if err != nil {
				utils.RenderInternalError(w, r, err)
			}
		} else {
			utils.RenderInternalError(w, r, err)
		}

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h UsersHttpHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	result, err := h.Service.GetAll()
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	err = h.Views["GetAll"].RenderView(
		w,
		&struct{ Users []*users.User }{result},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}

func (h UsersHttpHandler) Edit(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	user, err := h.Service.GetByID(userID(params["id"]))
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
	if user == nil || user.IsAdmin {
		http.NotFound(w, r)
		return
	}

	err = h.Views["UserForm"].RenderView(
		w,
		&userFormData{User: user},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}

func (h UsersHttpHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	user, err := h.Service.GetByID(userID(params["id"]))
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
	if user == nil || user.IsAdmin {
		http.NotFound(w, r)
		return
	}

	user.IsAdmin = utils.BoolFromParam(r.FormValue("is_admin"))

	_, err = h.Service.Update(user)
	if err != nil {
		if _, ok := err.(*services.UserValidationError); ok {
			err = h.Views["AssignmentForm"].RenderView(
				w,
				userFormData{
					User:   user,
					Errors: []string{err.Error()},
				},
				currentUser,
			)
			if err != nil {
				utils.RenderInternalError(w, r, err)
			}
		} else {
			utils.RenderInternalError(w, r, err)
		}
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h UsersHttpHandler) New(w http.ResponseWriter, r *http.Request) {
	err := h.Views["Signup"].RenderView(
		w,
		&userFormData{User: &users.User{}},
		nil,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
}

func (h UsersHttpHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, err := h.Service.Create(
		r.FormValue("username"),
		r.FormValue("password"),
		r.FormValue("password_confirmation"),
	)
	if err != nil {
		if _, ok := err.(*services.UserValidationError); ok {
			err := h.Views["Signup"].RenderView(
				w,
				&userFormData{User: user, Errors: []string{err.Error()}},
				nil,
			)
			if err != nil {
				utils.RenderInternalError(w, r, err)
			}
		} else {
			utils.RenderInternalError(w, r, err)
		}

		return
	}

	_, err = h.SessionManager.Create(w, user)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}

	http.Redirect(w, r, "/assignments", http.StatusSeeOther)
}

func userID(param string) int64 {
	id, _ := strconv.ParseInt(param, 10, 64)

	return id
}

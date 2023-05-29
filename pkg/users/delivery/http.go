package delivery

import (
	"io/fs"
	"net/http"

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

type signupData struct {
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

	return &UsersHttpHandler{
		Service:        service,
		Views:          views,
		SessionManager: sessionManager,
	}, nil
}

func (h UsersHttpHandler) New(w http.ResponseWriter, r *http.Request) {
	err := h.Views["Signup"].RenderView(
		w,
		&signupData{User: &users.User{}},
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
				&signupData{User: user, Errors: []string{err.Error()}},
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

	// TODO: Remove user if cannot create session
	_, err = h.SessionManager.Create(w, user)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}

	http.Redirect(w, r, "/assignments", http.StatusSeeOther)
}

package delivery

import (
	"io/fs"
	"net/http"

	"github.com/maxshend/grader/pkg/sessions"
	"github.com/maxshend/grader/pkg/users"
	"github.com/maxshend/grader/pkg/users/services"
	"github.com/maxshend/grader/pkg/utils"
)

type SessionsHttpHandler struct {
	SessionManager sessions.HttpSessionManager
	UsersService   services.UsersServiceInterface
	Views          map[string]*utils.View
}

type signinData struct {
	User   *users.User
	Errors []string
}

const (
	MsgInvalidCredentials = "Invalid username or password"
)

func NewSessionsHttpHandler(
	sessionManager sessions.HttpSessionManager,
	usersService services.UsersServiceInterface,
	templatesFS fs.FS,
) (*SessionsHttpHandler, error) {
	views := make(map[string]*utils.View)
	var err error

	views["Signin"], err = utils.NewView(templatesFS, "templates/sessions/signin.gohtml")
	if err != nil {
		return nil, err
	}

	return &SessionsHttpHandler{
		Views:          views,
		SessionManager: sessionManager,
		UsersService:   usersService,
	}, nil
}

func (h SessionsHttpHandler) New(w http.ResponseWriter, r *http.Request) {
	err := h.Views["Signin"].RenderView(
		w,
		&signinData{User: &users.User{}},
		nil,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
}

func (h SessionsHttpHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.UsersService.CheckCredentials(r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		if _, ok := err.(*services.UserCredentialsError); ok {
			err := h.Views["Signin"].RenderView(
				w,
				&signinData{User: &users.User{}, Errors: []string{err.Error()}},
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

	_, err = h.SessionManager.Create(w, currentUser)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	http.Redirect(w, r, "/assignments", http.StatusSeeOther)
}

func (h SessionsHttpHandler) Destroy(w http.ResponseWriter, r *http.Request) {
	err := h.SessionManager.Destroy(w, r)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	utils.RedirectUnauthenticated(w, r)
}

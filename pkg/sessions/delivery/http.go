package delivery

import (
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/maxshend/grader/pkg/sessions"
	"github.com/maxshend/grader/pkg/users"
	"github.com/maxshend/grader/pkg/users/services"
	"github.com/maxshend/grader/pkg/utils"
)

type SessionsHttpHandler struct {
	SessionManager sessions.HttpSessionManager
	UsersService   services.UsersServiceInterface
	Views          map[string]*utils.View
	OauthCreds     map[string]*sessions.OauthCred
}

type signinData struct {
	OauthLinks map[string]string
	Errors     []string
}

const (
	MsgInvalidCredentials  = "Invalid username or password"
	MsgCannotGetOauthToken = "Can't get oauth token from code"
)

func NewSessionsHttpHandler(
	sessionManager sessions.HttpSessionManager,
	usersService services.UsersServiceInterface,
	templatesFS fs.FS,
	oauthCreds map[string]*sessions.OauthCred,
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
		OauthCreds:     oauthCreds,
	}, nil
}

func (h SessionsHttpHandler) New(w http.ResponseWriter, r *http.Request) {
	err := h.Views["Signin"].RenderView(
		w,
		&signinData{OauthLinks: oauthLinks(h.OauthCreds)},
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
				&signinData{OauthLinks: oauthLinks(h.OauthCreds), Errors: []string{err.Error()}},
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

func (h SessionsHttpHandler) OauthVK(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		utils.RedirectUnauthenticated(w, r)
		return
	}

	oauthCred := h.OauthCreds["vk"]
	token, err := h.SessionManager.CreateOauthToken(r, code, oauthCred)
	if err != nil {
		log.Printf("Error while converting code into token: %v\n", err)
		err := h.Views["Signin"].RenderView(
			w,
			&signinData{OauthLinks: oauthLinks(h.OauthCreds), Errors: []string{MsgCannotGetOauthToken}},
			nil,
		)
		if err != nil {
			utils.RenderInternalError(w, r, err)
		}
		return
	}

	user, err := h.UsersService.CreateOauth(
		token,
		users.VkProvider,
	)

	if err != nil {
		if _, ok := err.(*services.UserValidationError); ok {
			err := h.Views["Signin"].RenderView(
				w,
				&signinData{OauthLinks: oauthLinks(h.OauthCreds), Errors: []string{err.Error()}},
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

func oauthLinks(creds map[string]*sessions.OauthCred) map[string]string {
	result := make(map[string]string)

	for k, v := range creds {
		result[strings.ToUpper(k)] = v.AuthURL
	}

	return result
}

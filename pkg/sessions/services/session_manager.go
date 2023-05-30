package services

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/maxshend/grader/pkg/sessions"
	"github.com/maxshend/grader/pkg/users"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

type HttpSession struct {
	Repo sessions.RepositoryInterface
}

func NewHttpSession(repo sessions.RepositoryInterface) *HttpSession {
	return &HttpSession{
		Repo: repo,
	}
}

const cookieTokenKey = "session_token"

func (sm *HttpSession) Check(r *http.Request) (*sessions.Session, error) {
	cookie, err := r.Cookie(cookieTokenKey)
	if err != nil {
		return nil, err
	}
	token := cookie.Value

	session, err := sm.Repo.GetByToken(token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, sessions.ErrUnauthenticatedUser
	}

	return session, nil
}

func (sm *HttpSession) Create(w http.ResponseWriter, user *users.User) (*sessions.Session, error) {
	token := uuid.NewString()

	session, err := sm.Repo.Create(user.ID, token)
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:    cookieTokenKey,
		Value:   token,
		Expires: time.Now().Add(7 * 24 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, cookie)

	return session, nil
}

func (sm *HttpSession) CurrentUser(r *http.Request) (*users.User, error) {
	user, ok := (r.Context().Value(sessions.CurrentUserKey)).(*users.User)
	if !ok {
		return nil, sessions.ErrUnauthenticatedUser
	}

	return user, nil
}

func (sm *HttpSession) CurrentSession(r *http.Request) (*sessions.Session, error) {
	session, ok := (r.Context().Value(sessions.SessionKey)).(*sessions.Session)
	if !ok {
		return nil, sessions.ErrUnauthenticatedUser
	}

	return session, nil
}

func (sm *HttpSession) Destroy(w http.ResponseWriter, r *http.Request) error {
	session, err := sm.CurrentSession(r)
	if err != nil {
		return err
	}
	err = sm.Repo.Destroy(session)
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:    cookieTokenKey,
		Expires: time.Now().AddDate(0, 0, -1),
		Path:    "/",
	}
	http.SetCookie(w, &cookie)

	return nil
}

func (sm *HttpSession) CreateOauthToken(r *http.Request, code string, cred *sessions.OauthCred) (*oauth2.Token, error) {
	conf := oauth2.Config{
		ClientID:     cred.ClientID,
		ClientSecret: cred.ClientSecret,
		RedirectURL:  cred.RedirectURL,
		Endpoint:     vk.Endpoint,
	}

	token, err := conf.Exchange(r.Context(), code)
	if err != nil {
		return nil, err
	}

	return token, nil
}

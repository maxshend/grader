package services

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/maxshend/grader/pkg/sessions"
	"github.com/maxshend/grader/pkg/users"
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

func (sm *HttpSession) CurrentUser(r *http.Request) *users.User {
	user, ok := (r.Context().Value(sessions.CurrentUserKey)).(*users.User)
	if ok {
		return user
	}

	return nil
}

package sessions

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maxshend/grader/pkg/users"
	"github.com/maxshend/grader/pkg/utils"
)

type Session struct {
	ID     int64
	UserID int64
	Token  string
}

type RepositoryInterface interface {
	GetByToken(string) (*Session, error)
	Create(userID int64, token string) (*Session, error)
	Destroy(*Session) error
}

type HttpSessionManager interface {
	Check(*http.Request) (*Session, error)
	Create(http.ResponseWriter, *users.User) (*Session, error)
	CurrentUser(*http.Request) (*users.User, error)
	CurrentSession(*http.Request) (*Session, error)
	Destroy(http.ResponseWriter, *http.Request) error
}

type ctxKey int

const (
	SessionKey     ctxKey = 1
	CurrentUserKey ctxKey = 2
)

const (
	MsgUnauthenticatedUser = "unauthenticated user"
	MsgForbiddenUser       = "not enough access rights"
)

var (
	ErrUnauthenticatedUser = errors.New(MsgUnauthenticatedUser)
)

func AuthMiddleware(sm HttpSessionManager, repo users.RepositoryInterface) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sm.Check(r)
			if err != nil {
				if err == http.ErrNoCookie || err == ErrUnauthenticatedUser {
					utils.RedirectUnauthenticated(w, r)
					return
				}

				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err != nil {
				utils.RedirectUnauthenticated(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), SessionKey, session)

			user, err := repo.GetByID(session.UserID)
			if err != nil {
				utils.RenderInternalError(w, r, err)
				return
			}
			if user == nil {
				utils.RedirectUnauthenticated(w, r)
				return
			}

			ctx = context.WithValue(ctx, CurrentUserKey, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminPolicyMiddleware(sm HttpSessionManager) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := sm.CurrentUser(r)
			if err != nil {
				http.Error(w, MsgForbiddenUser, http.StatusForbidden)
				return
			}

			if !user.IsAdmin {
				http.Error(w, MsgForbiddenUser, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

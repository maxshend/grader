package sessions

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maxshend/grader/pkg/users"
)

type Session struct {
	ID     int64
	UserID int64
	Token  string
}

type RepositoryInterface interface {
	GetByToken(string) (*Session, error)
	Create(userID int64, token string) (*Session, error)
}

type HttpSessionManager interface {
	Check(*http.Request) (*Session, error)
	Create(http.ResponseWriter, *users.User) (*Session, error)
	CurrentUser(*http.Request) *users.User
}

type ctxKey int

const (
	SessionKey     ctxKey = 1
	CurrentUserKey ctxKey = 2
)

const (
	ErrUnauthenticatedUser = "Unauthenticated user"
)

func AuthMiddleware(sm HttpSessionManager, repo users.RepositoryInterface) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sm.Check(r)
			if err != nil {
				if err == http.ErrNoCookie {
					http.Error(w, ErrUnauthenticatedUser, http.StatusUnauthorized)
					return
				}

				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err != nil {
				http.Error(w, ErrUnauthenticatedUser, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), SessionKey, session)

			user, err := repo.GetByID(session.UserID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if user == nil {
				http.Error(w, ErrUnauthenticatedUser, http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, CurrentUserKey, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

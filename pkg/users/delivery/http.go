package delivery

import (
	"net/http"

	"github.com/maxshend/grader/pkg/users"
	"github.com/maxshend/grader/pkg/users/services"
	"github.com/maxshend/grader/pkg/utils"
)

type UsersHttpHandler struct {
	Service services.UsersServiceInterface
	Views   map[string]*utils.View
}

type signupData struct {
	User   *users.User
	Errors []string
}

func NewUsersHttpHandler(
	service services.UsersServiceInterface,
) (*UsersHttpHandler, error) {
	views := make(map[string]*utils.View)
	var err error

	views["Signup"], err = utils.NewView("./web/templates/users/signup.gohtml")
	if err != nil {
		return nil, err
	}

	return &UsersHttpHandler{
		Service: service,
		Views:   views,
	}, nil
}

func (h UsersHttpHandler) New(w http.ResponseWriter, r *http.Request) {
	err := h.Views["Signup"].RenderView(w, &signupData{User: &users.User{}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
			err := h.Views["Signup"].RenderView(w, &signupData{User: user, Errors: []string{err.Error()}})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Redirect(w, r, "/assignments", http.StatusSeeOther)
}

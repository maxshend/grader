package delivery

import (
	"net/http"

	"github.com/maxshend/grader/pkg/assignments/services"
	"github.com/maxshend/grader/pkg/utils"
)

type AssignmentsHttpHandler struct {
	Service services.AssignmentsServiceInterface
	Views   map[string]*utils.View
}

func NewAssignmentsHttpHandler(
	service services.AssignmentsServiceInterface,
) (*AssignmentsHttpHandler, error) {
	views := make(map[string]*utils.View)
	var err error

	views["GetAll"], err = utils.NewView("./web/templates/assignments/list.gohtml")
	if err != nil {
		return nil, err
	}

	return &AssignmentsHttpHandler{
		Service: service,
		Views:   views,
	}, nil
}

func (h AssignmentsHttpHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	err := h.Views["GetAll"].RenderView(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

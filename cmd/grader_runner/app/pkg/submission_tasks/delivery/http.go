package delivery

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks"
	"github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks/services"
)

type SubmissionTasksHandler struct {
	Service services.SubmissionTaskServiceInterface
}

func NewSubmissionTasksHandler(service services.SubmissionTaskServiceInterface) *SubmissionTasksHandler {
	return &SubmissionTasksHandler{
		Service: service,
	}
}

func (h *SubmissionTasksHandler) Grade(w http.ResponseWriter, r *http.Request) {
	task := &submission_tasks.SubmissionTask{}
	err := json.NewDecoder(r.Body).Decode(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Service.RunSubmission(context.Background(), task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

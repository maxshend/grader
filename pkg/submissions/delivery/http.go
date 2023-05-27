package delivery

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/maxshend/grader/pkg/submissions/services"
	"github.com/maxshend/grader/pkg/utils"
)

type SubmissionsHttpHandler struct {
	Service services.SubmissionsServiceInterface
}

type RunnerResponse struct {
	Pass bool   `json:"pass"`
	Text string `json:"text"`
}

func NewSubmissionsHttpHandler(service services.SubmissionsServiceInterface) *SubmissionsHttpHandler {
	return &SubmissionsHttpHandler{
		Service: service,
	}
}

func (h *SubmissionsHttpHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if len(token) == 0 {
		http.Error(w, utils.ErrInvalidAccessToken.Error(), http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	runnerResponse := &RunnerResponse{}
	err := json.NewDecoder(r.Body).Decode(runnerResponse)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	log.Printf("Webhook: Submission #%s: \n%+v\n", params["id"], runnerResponse.Text)

	submissionID, err := strconv.Atoi(params["id"])
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	err = h.Service.HandleWebhook(token, int64(submissionID), runnerResponse.Pass, runnerResponse.Text)
	if err != nil {
		if err == services.ErrSubmissionNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err == utils.ErrInvalidAccessToken {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			utils.RenderInternalError(w, r, err)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

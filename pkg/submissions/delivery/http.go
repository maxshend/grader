package delivery

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maxshend/grader/pkg/submissions/services"
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
	params := mux.Vars(r)
	runnerResponse := &RunnerResponse{}
	err := json.NewDecoder(r.Body).Decode(runnerResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Runner Response for submission #%s: \n%+v\n", params["id"], runnerResponse)

	w.WriteHeader(http.StatusOK)
}

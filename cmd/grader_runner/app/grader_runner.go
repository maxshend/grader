package app

import (
	"log"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks/delivery"
	"github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks/services"
)

var dockerClient *client.Client

func Run() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	service := services.NewSubmissionTaskService(dockerClient)
	handler := delivery.NewSubmissionTasksHandler(service)

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/grader", handler.Grade).Methods("POST")

	log.Fatal(http.ListenAndServe(":8021", router))
}

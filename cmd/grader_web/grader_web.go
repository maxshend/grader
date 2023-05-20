package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	assignmentsDelivery "github.com/maxshend/grader/pkg/assignments/delivery"
	"github.com/maxshend/grader/pkg/assignments/repo"
	assignmentsServices "github.com/maxshend/grader/pkg/assignments/services"
	attachmentsRepo "github.com/maxshend/grader/pkg/attachments/repo"
	submissionsDelivery "github.com/maxshend/grader/pkg/submissions/delivery"
	submissionsRepo "github.com/maxshend/grader/pkg/submissions/repo"
	submissionsServices "github.com/maxshend/grader/pkg/submissions/services"
	amqp "github.com/rabbitmq/amqp091-go"

	_ "github.com/lib/pq"
)

func main() {
	hostURL := os.Getenv("HOST")
	if len(hostURL) == 0 {
		log.Fatal("HOST should be set")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if len(dbURL) == 0 {
		log.Fatal("DATABASE_URL should be set")
	}
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbConn.SetMaxOpenConns(10)
	err = dbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if len(rabbitURL) == 0 {
		log.Fatal("RABBITMQ_URL should be set")
	}
	rabbitQueueName := os.Getenv("RABBITMQ_QUEUE")
	if len(rabbitQueueName) == 0 {
		log.Fatal("RABBITMQ_QUEUE should be set")
	}
	rabbitConn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()
	rabbitCh, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitCh.Close()
	_, err = rabbitCh.QueueDeclare(
		rabbitQueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	webhookURL := "/webhooks/submissions/"
	webhookFullURL, err := url.JoinPath(hostURL, webhookURL)
	if err != nil {
		log.Fatal(err)
	}

	assignmentsRepo := repo.NewAssignmentsSQLRepo(dbConn)
	submRepo := submissionsRepo.NewSubmissionsSQLRepo(dbConn)
	attachRepo := attachmentsRepo.NewAttachmentsInmemRepo("./uploads", os.Getenv("HOST"))
	assignmentsService := assignmentsServices.NewAssignmentsService(webhookFullURL, assignmentsRepo, attachRepo, submRepo, rabbitCh, rabbitQueueName)
	assignmentsHandler, err := assignmentsDelivery.NewAssignmentsHttpHandler(assignmentsService)
	if err != nil {
		log.Fatal(err)
	}
	submissionsService := submissionsServices.NewSubmissionsService(submRepo)
	submissionsHandler := submissionsDelivery.NewSubmissionsHttpHandler(submissionsService)
	router := mux.NewRouter()

	adminPages := router.PathPrefix("/admin").Subrouter()
	adminPages.HandleFunc("/assignments", assignmentsHandler.GetAll).Methods("GET")

	router.HandleFunc("/assignments/{id}/submissions/new", assignmentsHandler.NewSubmission).Methods("GET")
	router.HandleFunc("/assignments/{id}/submissions", assignmentsHandler.Submit).Methods("POST")

	router.HandleFunc(webhookURL+"{id}", submissionsHandler.Webhook).Methods("POST")

	staticFs := http.FileServer(http.Dir("./web/static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFs))

	uploadsFs := http.FileServer(http.Dir("./uploads"))
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", uploadsFs))

	log.Fatal(http.ListenAndServe(":8080", router))
}

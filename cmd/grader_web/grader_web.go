package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/maxshend/grader/pkg/assignments/delivery"
	"github.com/maxshend/grader/pkg/assignments/repo"
	"github.com/maxshend/grader/pkg/assignments/services"
	attachmentsRepo "github.com/maxshend/grader/pkg/attachments/repo"
	submissionsRepo "github.com/maxshend/grader/pkg/submissions/repo"
	amqp "github.com/rabbitmq/amqp091-go"

	_ "github.com/lib/pq"
)

func main() {
	router := mux.NewRouter()

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

	assignmentsRepo := repo.NewAssignmentsSQLRepo(dbConn)
	submRepo := submissionsRepo.NewSubmissionsSQLRepo(dbConn)
	attachRepo := attachmentsRepo.NewAttachmentsInmemRepo("./uploads")
	assignmentsService := services.NewAssignmentsService(assignmentsRepo, attachRepo, submRepo, rabbitCh, rabbitQueueName)
	assignmentsHandler, err := delivery.NewAssignmentsHttpHandler(assignmentsService)
	if err != nil {
		log.Fatal(err)
	}

	adminPages := router.PathPrefix("/admin").Subrouter()
	adminPages.HandleFunc("/assignments", assignmentsHandler.GetAll).Methods("GET")

	router.HandleFunc("/assignments/{id}/submissions/new", assignmentsHandler.NewSubmission).Methods("GET")
	router.HandleFunc("/assignments/{id}/submissions", assignmentsHandler.Submit).Methods("POST")

	staticFs := http.FileServer(http.Dir("./web/static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFs))

	uploadsFs := http.FileServer(http.Dir("./uploads"))
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", uploadsFs))

	log.Fatal(http.ListenAndServe(":8080", router))
}

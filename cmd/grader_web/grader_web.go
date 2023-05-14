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
	assignmentsRepo := repo.NewAssignmentsSQLRepo(dbConn)
	assignmentsService := services.NewAssignmentsService(assignmentsRepo)
	assignmentsHandlers, err := delivery.NewAssignmentsHttpHandler(assignmentsService)
	if err != nil {
		log.Fatal(err)
	}

	adminPages := router.PathPrefix("/admin").Subrouter()
	adminPages.HandleFunc("/assignments", assignmentsHandlers.GetAll).Methods("GET")

	fs := http.FileServer(http.Dir("./web/static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	log.Fatal(http.ListenAndServe(":8080", router))
}

package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"

	assignmentsDelivery "github.com/maxshend/grader/pkg/assignments/delivery"
	"github.com/maxshend/grader/pkg/assignments/repo"
	assignmentsServices "github.com/maxshend/grader/pkg/assignments/services"
	attachmentsRepo "github.com/maxshend/grader/pkg/attachments/repo"
	"github.com/maxshend/grader/pkg/sessions"

	submissionsDelivery "github.com/maxshend/grader/pkg/submissions/delivery"
	submissionsRepo "github.com/maxshend/grader/pkg/submissions/repo"
	submissionsServices "github.com/maxshend/grader/pkg/submissions/services"

	usersDelivery "github.com/maxshend/grader/pkg/users/delivery"
	usersRepo "github.com/maxshend/grader/pkg/users/repo"
	usersServices "github.com/maxshend/grader/pkg/users/services"

	sessionsDelivery "github.com/maxshend/grader/pkg/sessions/delivery"
	sessionsRepo "github.com/maxshend/grader/pkg/sessions/repo"
	sessionsServices "github.com/maxshend/grader/pkg/sessions/services"

	amqp "github.com/rabbitmq/amqp091-go"

	_ "embed"

	_ "github.com/lib/pq"
)

//go:embed all:static/*
var staticFS embed.FS

//go:embed all:templates/*
var templatesFS embed.FS

const oauthVkPath = "/sessions/oauth/vk"

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET should be set")
	}
	hostURL := os.Getenv("HOST")
	if len(hostURL) == 0 {
		log.Fatal("HOST should be set")
	}
	externalHostURL := os.Getenv("EXTERNAL_HOST")
	if len(externalHostURL) == 0 {
		log.Fatal("EXTERNAL_HOST should be set")
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
	userRepo := usersRepo.NewUsersSQLRepo(dbConn)
	sessionRepo := sessionsRepo.NewSessionsSQLRepo(dbConn)

	assignmentsService := assignmentsServices.NewAssignmentsService(
		webhookFullURL,
		assignmentsRepo,
		attachRepo,
		submRepo,
		rabbitCh,
		rabbitQueueName,
		jwtSecret,
	)
	submissionsService := submissionsServices.NewSubmissionsService(submRepo, jwtSecret)
	usersService := usersServices.NewUsersService(userRepo)

	sessionManager := sessionsServices.NewHttpSession(sessionRepo)

	assignmentsHandler, err := assignmentsDelivery.NewAssignmentsHttpHandler(
		assignmentsService,
		sessionManager,
		submissionsService,
		templatesFS,
	)
	if err != nil {
		log.Fatal(err)
	}
	usersHandler, err := usersDelivery.NewUsersHttpHandler(usersService, sessionManager, templatesFS)
	if err != nil {
		log.Fatal(err)
	}
	submissionsHandler := submissionsDelivery.NewSubmissionsHttpHandler(submissionsService)

	oauthCreds := map[string]*sessions.OauthCred{
		"vk": {
			ClientID:     os.Getenv("OAUTH_VK_APP_ID"),
			ClientSecret: os.Getenv("OAUTH_VK_APP_KEY"),
			RedirectURL:  externalHostURL + oauthVkPath,
			AuthURL: fmt.Sprintf(
				"https://oauth.vk.com/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=email",
				os.Getenv("OAUTH_VK_APP_ID"),
				externalHostURL+oauthVkPath,
			),
		},
	}
	sessionsHandler, err := sessionsDelivery.NewSessionsHttpHandler(
		sessionManager,
		usersService,
		templatesFS,
		oauthCreds,
	)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	adminPages := router.PathPrefix("/admin").Subrouter()
	adminPages.HandleFunc("/assignments", assignmentsHandler.GetAll).Methods("GET")
	adminPages.HandleFunc("/assignments/", assignmentsHandler.Create).Methods("POST")
	adminPages.HandleFunc("/assignments/new", assignmentsHandler.New).Methods("GET")
	adminPages.HandleFunc("/assignments/{id}/edit", assignmentsHandler.Edit).Methods("GET")
	adminPages.HandleFunc("/assignments/{id}", assignmentsHandler.Update).Methods("POST")
	adminPages.HandleFunc("/assignments/{id}", assignmentsHandler.Show).Methods("GET")
	adminPages.Use(
		sessions.AuthMiddleware(sessionManager, userRepo),
		sessions.AdminPolicyMiddleware(sessionManager),
	)

	router.HandleFunc("/signup", usersHandler.New).Methods("GET")
	router.HandleFunc("/users", usersHandler.Create).Methods("POST")

	router.HandleFunc("/signin", sessionsHandler.New).Methods("GET")
	router.HandleFunc("/sessions", sessionsHandler.Create).Methods("POST")
	router.HandleFunc(oauthVkPath, sessionsHandler.OauthVK).Methods("GET")

	authPages := router.NewRoute().Subrouter()

	authPages.HandleFunc("/assignments", assignmentsHandler.PersonalAssignments).Methods("GET")
	authPages.HandleFunc("/", assignmentsHandler.PersonalAssignments).Methods("GET")
	authPages.HandleFunc("/assignments/{id}/submissions/new", assignmentsHandler.NewSubmission).Methods("GET")
	authPages.HandleFunc("/assignments/{id}/submissions", assignmentsHandler.CreateSubmission).Methods("POST")
	authPages.HandleFunc("/assignments/{id}", assignmentsHandler.ShowPersonal).Methods("GET")

	authPages.HandleFunc("/logout", sessionsHandler.Destroy).Methods("POST")

	authPages.Use(sessions.AuthMiddleware(sessionManager, userRepo))

	router.HandleFunc(webhookURL+"{id}", submissionsHandler.Webhook).Methods("POST")

	router.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFS)))

	uploadsFs := http.FileServer(http.Dir("./uploads"))
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", uploadsFs))

	log.Fatal(http.ListenAndServe(":8080", router))
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/maxshend/grader/pkg/assignments"
	"github.com/maxshend/grader/pkg/attachments"
	"github.com/maxshend/grader/pkg/submissions"
	"github.com/maxshend/grader/pkg/users"
	"github.com/maxshend/grader/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AssignmentsService struct {
	WebhookFullURL  string
	Repo            assignments.RepositoryInterface
	AttachRepo      attachments.RepositoryInterface
	SubmissionsRepo submissions.RepositoryInterface
	QueueCh         *amqp.Channel
	QueueName       string
	JwtSecret       string
}

type SubmissionFile struct {
	Content io.Reader
	Name    string
}

type SubmitAssignmentTask struct {
	GraderURL    string                    `json:"grader_url"`
	AccessToken  string                    `json:"access_token"`
	WebhookURL   string                    `json:"webhook_url"`
	Container    string                    `json:"container"`
	SubmissionID int64                     `json:"submission_id"`
	PartID       string                    `json:"part_id"`
	Files        []*submissions.Attachment `json:"files"`
}

const (
	MsgSubmissionFilesError  = "required submission file not present or has a wrong name"
	MsgBlankTitleError       = "title can't be blank"
	MsgBlankDescriptionError = "description can't be blank"
	MsgInvalidGraderURLError = "grader url is not a valid url"
	MsgBlankContainerError   = "container can't be blank"
	MsgBlankPartIDError      = "part id can't be blank"
	MsgUniqueTitleError      = "title already exists"
	MsgInvalidFilesError     = "files have invalid format"
)

type AssignmentsServiceInterface interface {
	GetAll() ([]*assignments.Assignment, error)
	GetByID(int64) (*assignments.Assignment, error)
	GetByUserID(int64) ([]*assignments.Assignment, error)
	Submit(*users.User, *assignments.Assignment, []*SubmissionFile) (*submissions.Submission, error)
	Create(*assignments.Assignment) (*assignments.Assignment, error)
	Update(*assignments.Assignment) (*assignments.Assignment, error)
	ValidateAssignment(*assignments.Assignment) error
}

func NewAssignmentsService(
	webhookFullURL string,
	repo assignments.RepositoryInterface,
	attachRepo attachments.RepositoryInterface,
	submissionsRepo submissions.RepositoryInterface,
	queueCh *amqp.Channel,
	queueName string,
	jwtSecret string,
) AssignmentsServiceInterface {
	return &AssignmentsService{
		WebhookFullURL:  webhookFullURL,
		Repo:            repo,
		AttachRepo:      attachRepo,
		SubmissionsRepo: submissionsRepo,
		QueueCh:         queueCh,
		QueueName:       queueName,
		JwtSecret:       jwtSecret,
	}
}

func (s *AssignmentsService) GetAll() ([]*assignments.Assignment, error) {
	// TODO: Pagination handling
	return s.Repo.GetAll(100, 0)
}

func (s *AssignmentsService) GetByID(id int64) (*assignments.Assignment, error) {
	return s.Repo.GetByID(id)
}

func (s *AssignmentsService) GetByUserID(userID int64) ([]*assignments.Assignment, error) {
	return s.Repo.GetByUserID(userID, 100, 0)
}

func (s *AssignmentsService) Submit(
	user *users.User,
	assignment *assignments.Assignment,
	files []*SubmissionFile,
) (submission *submissions.Submission, err error) {
	err = checkSubmissionFiles(assignment.Files, files)
	if err != nil {
		return nil, err
	}
	newAttachments := []*attachments.Attachment{}

	txn, err := s.SubmissionsRepo.CreateTxn()
	if err != nil {
		return nil, err
	}
	defer func(newAttachments []*attachments.Attachment) {
		p := recover()

		rollbackErr := txn.Rollback()
		if rollbackErr != nil {
			log.Printf("Error while reverting db changes: %v", rollbackErr)
		}

		for _, att := range newAttachments {
			attErr := s.AttachRepo.Destroy(att.URL)
			if attErr != nil {
				log.Printf("Error while reverting attachment creation: %v", err)
			}
		}

		if p != nil {
			panic(p)
		}
	}(newAttachments)

	submission, err = s.SubmissionsRepo.Create(txn, user.ID, assignment.ID)
	if err != nil {
		return nil, err
	}

	pathPrefix := fmt.Sprintf("submissions/%d", submission.ID)
	for _, file := range files {
		attachment, err := s.AttachRepo.Create(pathPrefix, file.Name, file.Content)
		if err != nil {
			return nil, err
		}
		newAttachments = append(newAttachments, attachment)
	}

	submissionAttachments, err := s.SubmissionsRepo.CreateSubmissionAttachments(
		txn,
		submission.ID,
		newAttachments,
	)
	if err != nil {
		return nil, err
	}

	token, err := utils.AccessToken(s.JwtSecret, strconv.FormatInt(submission.ID, 10))
	if err != nil {
		return nil, err
	}
	submission.Attachments = submissionAttachments
	task := &SubmitAssignmentTask{
		GraderURL:    assignment.GraderURL,
		Container:    assignment.Container,
		PartID:       assignment.PartID,
		Files:        submission.Attachments,
		SubmissionID: submission.ID,
		AccessToken:  token,
		WebhookURL:   fmt.Sprint(s.WebhookFullURL, submission.ID),
	}
	data, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return
	}

	err = s.QueueCh.PublishWithContext(
		context.Background(),
		"",
		s.QueueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         data,
		},
	)
	if err != nil {
		return nil, err
	}

	return
}

func (s *AssignmentsService) Create(assignment *assignments.Assignment) (*assignments.Assignment, error) {
	foundAssignment, err := s.Repo.GetByTitle(assignment.Title)
	if err != nil {
		return nil, err
	}
	if foundAssignment != nil {
		return nil, &AssignmentValidationError{MsgUniqueTitleError}
	}

	for i, file := range assignment.Files {
		assignment.Files[i] = strings.TrimSpace(file)
	}

	err = s.ValidateAssignment(assignment)
	if err != nil {
		return nil, err
	}

	return s.Repo.Create(
		assignment.Title,
		assignment.Description,
		assignment.GraderURL,
		assignment.Container,
		assignment.PartID,
		assignment.Files,
	)
}

func (s *AssignmentsService) Update(assignment *assignments.Assignment) (*assignments.Assignment, error) {
	for i, file := range assignment.Files {
		assignment.Files[i] = strings.TrimSpace(file)
	}

	err := s.ValidateAssignment(assignment)
	if err != nil {
		return nil, err
	}

	return s.Repo.Update(assignment)
}

func (s *AssignmentsService) ValidateAssignment(assignment *assignments.Assignment) error {
	if len(assignment.Title) == 0 {
		return &AssignmentValidationError{MsgBlankTitleError}
	}
	if len(assignment.Description) == 0 {
		return &AssignmentValidationError{MsgBlankDescriptionError}
	}
	if _, err := url.ParseRequestURI(assignment.GraderURL); err != nil {
		return &AssignmentValidationError{MsgInvalidGraderURLError}
	}
	if len(assignment.Container) == 0 {
		return &AssignmentValidationError{MsgBlankContainerError}
	}
	if len(assignment.PartID) == 0 {
		return &AssignmentValidationError{MsgBlankPartIDError}
	}
	for _, file := range assignment.Files {
		if len(file) == 0 {
			return &AssignmentValidationError{MsgInvalidFilesError}
		}
	}

	return nil
}

func checkSubmissionFiles(requiredFiles []string, requestFiles []*SubmissionFile) error {
	for _, file := range requestFiles {
		for i, requiredFile := range requiredFiles {
			if requiredFile == file.Name {
				break
			}

			if i == len(requiredFiles)-1 {
				return &AssignmentValidationError{MsgSubmissionFilesError}
			}
		}
	}

	return nil
}

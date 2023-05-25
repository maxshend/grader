package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/maxshend/grader/pkg/assignments"
	"github.com/maxshend/grader/pkg/attachments"
	"github.com/maxshend/grader/pkg/submissions"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AssignmentsService struct {
	WebhookFullURL  string
	Repo            assignments.RepositoryInterface
	AttachRepo      attachments.RepositoryInterface
	SubmissionsRepo submissions.RepositoryInterface
	QueueCh         *amqp.Channel
	QueueName       string
}

type SubmissionFile struct {
	Content io.Reader
	Name    string
}

type SubmitAssignmentTask struct {
	GraderURL    string                    `json:"grader_url"`
	WebhookURL   string                    `json:"webhook_url"`
	Container    string                    `json:"container"`
	SubmissionID int64                     `json:"submission_id"`
	PartID       string                    `json:"part_id"`
	Files        []*submissions.Attachment `json:"files"`
}

const (
	ErrSubmissionFiles = "Required submission file not present or has a wrong name"
)

type AssignmentsServiceInterface interface {
	GetAll() ([]*assignments.Assignment, error)
	GetByID(string) (*assignments.Assignment, error)
	GetByUserID(int64) ([]*assignments.Assignment, error)
	Submit(*assignments.Assignment, []*SubmissionFile) (*submissions.Submission, error)
}

func NewAssignmentsService(
	webhookFullURL string,
	repo assignments.RepositoryInterface,
	attachRepo attachments.RepositoryInterface,
	submissionsRepo submissions.RepositoryInterface,
	queueCh *amqp.Channel,
	queueName string,
) AssignmentsServiceInterface {
	return &AssignmentsService{
		WebhookFullURL:  webhookFullURL,
		Repo:            repo,
		AttachRepo:      attachRepo,
		SubmissionsRepo: submissionsRepo,
		QueueCh:         queueCh,
		QueueName:       queueName,
	}
}

func (s *AssignmentsService) GetAll() ([]*assignments.Assignment, error) {
	// TODO: Pagination handling
	return s.Repo.GetAll(100, 0)
}

func (s *AssignmentsService) GetByID(id string) (*assignments.Assignment, error) {
	return s.Repo.GetByID(id)
}

func (s *AssignmentsService) GetByUserID(userID int64) ([]*assignments.Assignment, error) {
	return s.Repo.GetByUserID(userID)
}

func (s *AssignmentsService) Submit(assignment *assignments.Assignment, files []*SubmissionFile) (*submissions.Submission, error) {
	for _, file := range files {
		for i, requiredFile := range assignment.Files {
			if requiredFile == file.Name {
				break
			}

			if i == len(assignment.Files)-1 {
				return nil, &AssignmentValidationError{ErrSubmissionFiles}
			}
		}
	}

	// TODO: Use real user ID
	var userID int64 = 1

	submission, err := s.SubmissionsRepo.Create(userID, assignment.ID)
	if err != nil {
		return nil, err
	}
	// TODO: Remove submission from DB in case of errors

	pathPrefix := fmt.Sprintf("submissions/%d", submission.ID)
	attachments := []*attachments.Attachment{}
	for _, file := range files {
		attachment, err := s.AttachRepo.Create(pathPrefix, file.Name, file.Content)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}

	submissionAttachments, err := s.SubmissionsRepo.CreateSubmissionAttachments(
		submission.ID,
		attachments,
	)
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
		WebhookURL:   fmt.Sprint(s.WebhookFullURL, submission.ID),
	}
	data, err := json.Marshal(task)
	if err != nil {
		return nil, err
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

	return submission, nil
}

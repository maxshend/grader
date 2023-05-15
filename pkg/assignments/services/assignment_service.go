package services

import (
	"github.com/maxshend/grader/pkg/assignments"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AssignmentsService struct {
	Repo     assignments.RepositoryInterface
	RabbitCh amqp.Queue
}

type AssignmentsServiceInterface interface {
	GetAll() ([]*assignments.Assignment, error)
}

func NewAssignmentsService(repo assignments.RepositoryInterface, rabbitCh amqp.Queue) AssignmentsServiceInterface {
	return &AssignmentsService{Repo: repo, RabbitCh: rabbitCh}
}

func (s *AssignmentsService) GetAll() ([]*assignments.Assignment, error) {
	// TODO: Pagination handling
	return s.Repo.GetAll(100, 0)
}

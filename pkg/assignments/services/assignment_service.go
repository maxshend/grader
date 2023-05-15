package services

import (
	"github.com/maxshend/grader/pkg/assignments"
)

type AssignmentsService struct {
	Repo assignments.RepositoryInterface
}

type AssignmentsServiceInterface interface {
	GetAll() ([]*assignments.Assignment, error)
}

func NewAssignmentsService(repo assignments.RepositoryInterface) AssignmentsServiceInterface {
	return &AssignmentsService{Repo: repo}
}

func (s *AssignmentsService) GetAll() ([]*assignments.Assignment, error) {
	// TODO: Pagination handling
	return s.Repo.GetAll(100, 0)
}

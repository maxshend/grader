package services

import "github.com/maxshend/grader/pkg/submissions"

type SubmissionsService struct {
	Repo submissions.RepositoryInterface
}

type SubmissionsServiceInterface interface {
	HandleWebhook()
}

func NewSubmissionsService(repo submissions.RepositoryInterface) SubmissionsServiceInterface {
	return &SubmissionsService{
		Repo: repo,
	}
}

func (s *SubmissionsService) HandleWebhook() {}

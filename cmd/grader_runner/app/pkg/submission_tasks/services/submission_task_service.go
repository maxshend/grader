package services

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks"
)

type DockerClientInterface interface {
	client.CommonAPIClient
}

type SubmissionTaskService struct {
	DockerClient DockerClientInterface
}

type SubmissionTaskServiceInterface interface {
	RunSubmission(context.Context, *submission_tasks.SubmissionTask) error
}

func NewSubmissionTaskService(dockerClient DockerClientInterface) *SubmissionTaskService {
	return &SubmissionTaskService{
		DockerClient: dockerClient,
	}
}

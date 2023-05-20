package services

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks"
)

type SubmissionTaskService struct {
	DockerClient *client.Client
}

type SubmissionTaskServiceInterface interface {
	RunSubmission(context.Context, *submission_tasks.SubmissionTask) error
}

func NewSubmissionTaskService(dockerClient *client.Client) *SubmissionTaskService {
	return &SubmissionTaskService{
		DockerClient: dockerClient,
	}
}

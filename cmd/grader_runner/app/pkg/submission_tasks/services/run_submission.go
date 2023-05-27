package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks"
)

type ContainerResponse struct {
	Pass bool   `json:"pass"`
	Text string `json:"text"`
}

const (
	successMsg         = "Поздравляем! Вы успешно сделали задание"
	submissionFilesDir = "/app/src"
)

// TODO: Refactor RunSubmission
func (s *SubmissionTaskService) RunSubmission(ctx context.Context, task *submission_tasks.SubmissionTask) error {
	dir := fmt.Sprintf("/tmp/submission_%d", task.SubmissionID)
	err := os.Mkdir(dir, 0755)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	for _, file := range task.Files {
		fullpath := filepath.Join(dir, file.Name)
		newFile, err := os.Create(fullpath)
		if err != nil {
			return err
		}
		defer newFile.Close()

		resp, err := http.Get(file.URL)
		if err != nil {
			return err
		}

		_, err = io.Copy(newFile, resp.Body)
		if err != nil {
			return err
		}
	}

	containerName := fmt.Sprintf("run_submission_%d", task.SubmissionID)
	resp, err := s.DockerClient.ContainerCreate(
		ctx,
		&container.Config{
			User:            "1000",
			NetworkDisabled: true,
			Image:           task.Container,
			Cmd:             []string{"sh", fmt.Sprintf("%s.sh", task.PartID)},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: dir,
					Target: submissionFilesDir,
				},
			},
		},
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return err
	}

	if err := s.DockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// TODO: Restrict wait time

	containerResponse := &ContainerResponse{}
	statusCh, errCh := s.DockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return err
	case status := <-statusCh:
		if status.StatusCode == 0 {
			containerResponse.Pass = true
			containerResponse.Text = successMsg
		} else {
			containerOut, err := s.DockerClient.ContainerLogs(
				ctx,
				resp.ID,
				types.ContainerLogsOptions{ShowStdout: true},
			)
			if err != nil {
				return err
			}

			out, err := io.ReadAll(containerOut)
			if err != nil {
				return err
			}

			containerResponse.Pass = false
			containerResponse.Text = string(out)
		}
	}

	noWaitTimeout := 0 // to not wait for the container to exit gracefully
	if err := s.DockerClient.ContainerStop(ctx, resp.ID, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
		return err
	}

	httpResponse, err := sendResults(task.WebhookURL, task.AccessToken, containerResponse)
	if err != nil {
		return err
	}
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}
	log.Printf("Webhook %q Response %d\n%s\n", task.WebhookURL, httpResponse.StatusCode, responseBody)

	return nil
}

func sendResults(graderURL string, authorization string, containerResponse *ContainerResponse) (*http.Response, error) {
	requestBody, err := json.Marshal(containerResponse)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", graderURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", authorization)

	client := &http.Client{
		Timeout: time.Minute * 1,
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

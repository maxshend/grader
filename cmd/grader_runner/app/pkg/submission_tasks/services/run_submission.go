package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"golang.org/x/sync/errgroup"
)

type ContainerResponse struct {
	Pass bool   `json:"pass"`
	Text string `json:"text"`
}

const (
	successMsg         = "Поздравляем! Вы успешно сделали задание"
	submissionFilesDir = "/app/src"
	MaxWaitMinutes     = 5
	TimeoutMsg         = "Timeout"
)

var (
	ErrSubmissionFileDonwload = errors.New("can't download submission file")
	ErrSendResults            = errors.New("can't send submission results")
)

func (s *SubmissionTaskService) RunSubmission(ctx context.Context, task *submission_tasks.SubmissionTask) error {
	dir, rmDir, err := tmpSaveAttachments(ctx, task)
	if err != nil {
		return err
	}
	defer func() {
		if err := rmDir(); err != nil {
			log.Printf("Error while removing dir: %q\n", err)
		}
	}()

	resp, err := s.createContainer(ctx, task, dir)
	if err != nil {
		return err
	}

	if err := s.DockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	defer func() {
		noWaitTimeout := 0 // to not wait for the container to exit gracefully
		if err := s.DockerClient.ContainerStop(ctx, resp.ID, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
			log.Printf("Can't stop docker container: %v", err)
		}
	}()

	containerResponse := &ContainerResponse{}
	statusCh, errCh := s.DockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	containerTimer := time.NewTimer(MaxWaitMinutes * time.Minute)
	defer containerTimer.Stop()

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
	case <-containerTimer.C:
		containerResponse.Pass = false
		containerResponse.Text = TimeoutMsg
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
	if response.StatusCode != http.StatusOK {
		return nil, ErrSendResults
	}

	return response, nil
}

func (s *SubmissionTaskService) createContainer(
	ctx context.Context,
	task *submission_tasks.SubmissionTask,
	mountDir string,
) (resp container.CreateResponse, err error) {
	containerName := fmt.Sprintf("run_submission_%d", task.SubmissionID)
	resp, err = s.DockerClient.ContainerCreate(
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
					Source: mountDir,
					Target: submissionFilesDir,
				},
			},
		},
		nil,
		nil,
		containerName,
	)

	return
}

func tmpSaveAttachments(ctx context.Context, task *submission_tasks.SubmissionTask) (dir string, rmDir func() error, err error) {
	dir = fmt.Sprintf("/tmp/submission_%d", task.SubmissionID)
	err = os.Mkdir(dir, 0755)
	if err != nil {
		return
	}
	rmDir = func() error {
		return os.RemoveAll(dir)
	}
	defer func() {
		p := recover()

		if p != nil || err != nil {
			rmErr := rmDir()
			if rmErr != nil {
				log.Printf("Error while removing dir: %v", rmErr)
			}
		}

		if p != nil {
			panic(p)
		}
	}()

	errs, ctx := errgroup.WithContext(ctx)

	for _, file := range task.Files {
		func(file *submission_tasks.SubmissionFile) {
			errs.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					var newFile *os.File
					fullpath := filepath.Join(dir, file.Name)

					newFile, err := os.Create(fullpath)
					if err != nil {
						return err
					}
					defer newFile.Close()

					var resp *http.Response

					resp, err = http.Get(file.URL)
					if err != nil {
						return err
					}
					if resp.StatusCode != http.StatusOK {
						log.Printf("Cannot get submission file: %q (%d)", file.URL, resp.StatusCode)
						return ErrSubmissionFileDonwload
					}

					_, err = io.Copy(newFile, resp.Body)
					if err != nil {
						return err
					}

					return nil
				}
			})
		}(file)
	}

	err = errs.Wait()

	return
}

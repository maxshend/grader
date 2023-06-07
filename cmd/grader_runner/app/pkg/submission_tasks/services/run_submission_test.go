package services

import (
	"context"
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"

	types "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	gomock "github.com/golang/mock/gomock"
	submission_tasks "github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks"
)

func TestRunSubmission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerCli := NewMockDockerClientInterface(ctrl)
	service := NewSubmissionTaskService(dockerCli)
	ctx := context.Background()
	dockerErr := errors.New("docker err")

	type testCase struct {
		Title            string
		Success          bool
		FileServerStatus int
		WebServerStatus  int
		DockerSetup      func(*testing.T)
		Check            func(*testing.T, error)
	}

	testCases := []*testCase{
		{
			Title:            "success",
			Success:          true,
			FileServerStatus: http.StatusOK,
			WebServerStatus:  http.StatusOK,
			DockerSetup: func(*testing.T) {
				t.Helper()

				statusCh, _ := setupDockerExpectations(ctx, dockerCli)
				statusCh <- container.WaitResponse{StatusCode: 0}
			},
		},
		{
			Title:            "docker error",
			Success:          false,
			FileServerStatus: http.StatusOK,
			WebServerStatus:  http.StatusOK,
			DockerSetup: func(*testing.T) {
				t.Helper()

				_, errCh := setupDockerExpectations(ctx, dockerCli)
				errCh <- dockerErr
			},
			Check: func(t *testing.T, err error) {
				t.Helper()

				if err != dockerErr {
					t.Errorf("expected to have %v, got %v", dockerErr, err)
				}
			},
		},
		{
			Title:            "file download error",
			Success:          false,
			FileServerStatus: http.StatusInternalServerError,
			WebServerStatus:  http.StatusOK,
		},
		{
			Title:            "send submission results error",
			Success:          false,
			FileServerStatus: http.StatusOK,
			WebServerStatus:  http.StatusInternalServerError,
			DockerSetup: func(*testing.T) {
				t.Helper()

				statusCh, _ := setupDockerExpectations(ctx, dockerCli)
				statusCh <- container.WaitResponse{StatusCode: 0}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			webServer := createTestServer(testCase.WebServerStatus)
			defer webServer.Close()

			fileServer := createTestServer(testCase.FileServerStatus)
			defer fileServer.Close()

			submissionFiles := []*submission_tasks.SubmissionFile{
				{URL: fileServer.URL, Name: "main.go"},
			}
			task := &submission_tasks.SubmissionTask{
				WebhookURL:  webServer.URL,
				Container:   "container_name",
				PartID:      "part_id",
				Files:       submissionFiles,
				AccessToken: "foobar123",
			}

			if testCase.DockerSetup != nil {
				testCase.DockerSetup(t)
			}

			err := service.RunSubmission(ctx, task)
			if testCase.Success && err != nil {
				t.Errorf("expected to not have errors, got %v", err)
			} else if !testCase.Success && err == nil {
				t.Errorf("expected to have errors")
			}
			if testCase.Check != nil {
				testCase.Check(t, err)
			}
		})
	}
}

func setupDockerExpectations(ctx context.Context, dockerCli *MockDockerClientInterface) (chan container.WaitResponse, chan error) {
	statusCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)

	createResponse := container.CreateResponse{}
	dockerCli.
		EXPECT().
		ContainerCreate(ctx, gomock.Any(), gomock.Any(), nil, nil, gomock.Any()).
		Return(createResponse, nil)

	dockerCli.
		EXPECT().
		ContainerStart(ctx, createResponse.ID, types.ContainerStartOptions{}).
		Return(nil)

	dockerCli.
		EXPECT().
		ContainerWait(ctx, createResponse.ID, container.WaitConditionNotRunning).
		Return(statusCh, errCh)

	noWaitTimeout := 0
	dockerCli.
		EXPECT().
		ContainerStop(ctx, createResponse.ID, container.StopOptions{Timeout: &noWaitTimeout}).
		Return(nil)

	return statusCh, errCh
}

func createTestServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

package delivery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks"
	"github.com/maxshend/grader/cmd/grader_runner/app/pkg/submission_tasks/services"
)

func TestGrade(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := services.NewMockSubmissionTaskServiceInterface(ctrl)
	handler := NewSubmissionTasksHandler(service)

	type testCase struct {
		Title              string
		SetupService       func(*testing.T, *submission_tasks.SubmissionTask)
		ExpectedStatusCode int
		SubmissionTask     *submission_tasks.SubmissionTask
	}

	testCases := []*testCase{
		{
			Title:              "success",
			ExpectedStatusCode: http.StatusOK,
			SubmissionTask:     &submission_tasks.SubmissionTask{},
			SetupService: func(t *testing.T, task *submission_tasks.SubmissionTask) {
				t.Helper()

				service.EXPECT().RunSubmission(gomock.Any(), task).Return(nil)
			},
		},
		{
			Title:              "error",
			ExpectedStatusCode: http.StatusInternalServerError,
			SubmissionTask:     &submission_tasks.SubmissionTask{},
			SetupService: func(t *testing.T, task *submission_tasks.SubmissionTask) {
				t.Helper()

				service.EXPECT().RunSubmission(gomock.Any(), task).Return(fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			reqBody, _ := json.Marshal(testCase.SubmissionTask)
			req := httptest.NewRequest("POST", "/api/posts", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			if testCase.SetupService != nil {
				testCase.SetupService(t, testCase.SubmissionTask)
			}

			handler.Grade(w, req)

			statusCode := w.Result().StatusCode
			if statusCode != testCase.ExpectedStatusCode {
				t.Errorf("expected to have status code %d, got %d", testCase.ExpectedStatusCode, statusCode)
			}
		})
	}
}

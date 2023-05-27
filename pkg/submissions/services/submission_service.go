package services

import (
	"strconv"
	"strings"

	"github.com/maxshend/grader/pkg/submissions"
	"github.com/maxshend/grader/pkg/utils"
)

type SubmissionsService struct {
	Repo      submissions.RepositoryInterface
	JwtSecret string
}

type SubmissionsServiceInterface interface {
	HandleWebhook(token string, submissionID int64, pass bool, text string) error
	GetByID(int64) (*submissions.Submission, error)
	Update(*submissions.Submission) error
	GetByUserAssignment(assignmentID, userID int64) ([]*submissions.Submission, error)
}

func NewSubmissionsService(repo submissions.RepositoryInterface, jwtSeret string) SubmissionsServiceInterface {
	return &SubmissionsService{
		Repo:      repo,
		JwtSecret: jwtSeret,
	}
}

func (s *SubmissionsService) HandleWebhook(token string, submissionID int64, pass bool, text string) error {
	submission, err := s.GetByID(submissionID)
	if err != nil {
		return err
	}
	if submission == nil {
		return ErrSubmissionNotFound
	}

	err = utils.CheckAccessToken(s.JwtSecret, token, strconv.FormatInt(submission.ID, 10))
	if err != nil {
		return err
	}

	var newStatus int
	if pass {
		newStatus = submissions.Success
	} else {
		newStatus = submissions.Fail
	}
	submission.Status = newStatus
	// strings.Replace fixes: pq: invalid byte sequence for encoding "UTF8": 0x00
	submission.Details = strings.Replace(text, "\u0000", "", -1)

	err = s.Update(submission)
	if err != nil {
		return err
	}

	return nil
}

func (s *SubmissionsService) GetByID(id int64) (*submissions.Submission, error) {
	return s.Repo.GetByID(id)
}

func (s *SubmissionsService) Update(submission *submissions.Submission) error {
	return s.Repo.Update(submission)
}

func (s *SubmissionsService) GetByUserAssignment(assignmentID, userID int64) ([]*submissions.Submission, error) {
	// TODO: Pagination handling
	return s.Repo.GetByUserAssignment(assignmentID, userID, 100, 0)
}

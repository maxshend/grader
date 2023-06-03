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

const DefaultPageSize = 25

type SubmissionsServiceInterface interface {
	HandleWebhook(token string, submissionID int64, pass bool, text string) error
	GetByID(int64) (*submissions.Submission, error)
	Update(*submissions.Submission) error
	GetByUserAssignment(
		assignmentID, userID int64,
		page int,
	) ([]*submissions.Submission, *utils.PaginationData, error)
	GetByAssignment(assignmentID int64) ([]*submissions.Submission, error)
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
	// strings.Replace is used to fix: pq: invalid byte sequence for encoding "UTF8": 0x00
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

func (s *SubmissionsService) GetByUserAssignment(
	assignmentID, userID int64,
	page int,
) ([]*submissions.Submission, *utils.PaginationData, error) {
	if page <= 0 {
		page = 1
	}

	totalCount, err := s.Repo.GetByUserAssignmentCount(assignmentID, userID)
	if err != nil {
		return nil, nil, err
	}
	maxPage := utils.GetMaxPage(DefaultPageSize, totalCount)
	if page > maxPage {
		page = maxPage
	}
	offset := utils.GetPageOffset(page, DefaultPageSize)
	paginationData := &utils.PaginationData{
		CurrentPage: page,
		MaxPage:     maxPage,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		LastPage:    page == maxPage,
		FirstPage:   page == 1,
	}
	assignments, err := s.Repo.GetByUserAssignment(assignmentID, userID, DefaultPageSize, offset)
	if err != nil {
		return nil, nil, err
	}

	return assignments, paginationData, nil
}

func (s *SubmissionsService) GetByAssignment(assignmentID int64) ([]*submissions.Submission, error) {
	// TODO: Pagination handling
	return s.Repo.GetByAssignment(assignmentID, 100, 0)
}

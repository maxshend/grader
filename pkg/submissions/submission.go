package submissions

import (
	"time"

	"github.com/maxshend/grader/pkg/attachments"
)

const (
	InProgress int = iota
	Success
	Fail
)

type Submission struct {
	ID           int64
	Status       int
	AssignmentID int64
	UserID       int64
	Username     string
	Details      string
	Attachments  []*Attachment
	CreatedAt    time.Time
}

type Attachment struct {
	ID   int64  `json:"-"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

type RepositoryInterface interface {
	Create(userID int64, assignmentID int64) (*Submission, error)
	CreateSubmissionAttachments(int64, []*attachments.Attachment) ([]*Attachment, error)
	GetSubmissionAttachments(int64) ([]*Attachment, error)
	GetByID(int64) (*Submission, error)
	Update(*Submission) error
	GetByUserAssignment(assignmentID int64, userID int64, limit, offset int) ([]*Submission, error)
	GetByAssignment(assignmentID int64, limit, offset int) ([]*Submission, error)
}

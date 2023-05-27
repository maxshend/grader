package submissions

import (
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
	Details      string
	Attachments  []*Attachment
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
}

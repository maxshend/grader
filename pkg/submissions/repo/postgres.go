package repo

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
	"github.com/maxshend/grader/pkg/attachments"
	"github.com/maxshend/grader/pkg/submissions"
)

type SubmissionsSQLRepo struct {
	DB *sql.DB
}

func NewSubmissionsSQLRepo(db *sql.DB) *SubmissionsSQLRepo {
	return &SubmissionsSQLRepo{DB: db}
}

func (r *SubmissionsSQLRepo) Create(userID int64, assignmentID int64) (*submissions.Submission, error) {
	submission := &submissions.Submission{
		UserID:       userID,
		AssignmentID: assignmentID,
		Status:       submissions.InProgress,
	}
	err := r.DB.QueryRow(
		"INSERT INTO submissions (user_id, assignment_id, status) VALUES ($1, $2, $3) RETURNING id",
		userID,
		assignmentID,
		submission.Status,
	).Scan(&submission.ID)
	if err != nil {
		return nil, err
	}

	return submission, nil
}

func (r *SubmissionsSQLRepo) GetSubmissionAttachments(submissionID int64) ([]*submissions.Attachment, error) {
	rows, err := r.DB.Query("SELECT id, url, name FROM submission_attachments WHERE submission_id = $1", submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*submissions.Attachment{}
	for rows.Next() {
		attach := &submissions.Attachment{}

		if err := rows.Scan(&attach.ID, &attach.URL, &attach.Name); err != nil {
			return nil, err
		}
		result = append(result, attach)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *SubmissionsSQLRepo) CreateSubmissionAttachments(
	submissionID int64,
	attachments []*attachments.Attachment,
) ([]*submissions.Attachment, error) {
	err := r.insertMultipleAttachments(submissionID, attachments)
	if err != nil {
		return nil, err
	}

	return r.GetSubmissionAttachments(submissionID)
}

func (r *SubmissionsSQLRepo) insertMultipleAttachments(submissionID int64, attachments []*attachments.Attachment) error {
	txn, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer txn.Commit()

	stm, err := txn.Prepare(pq.CopyIn("submission_attachments", "url", "name", "submission_id"))
	if err != nil {
		return err
	}
	defer stm.Close()

	for _, attachment := range attachments {
		_, err := stm.Exec(attachment.URL, attachment.Name, submissionID)
		if err != nil {
			return err
		}
	}

	_, err = stm.Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r *SubmissionsSQLRepo) GetByID(id int64) (*submissions.Submission, error) {
	submission := &submissions.Submission{}
	detailsString := sql.NullString{}
	err := r.DB.QueryRow(
		"SELECT id, user_id, assignment_id, status, details FROM submissions WHERE id = $1 LIMIT 1",
		id,
	).Scan(
		&submission.ID, &submission.UserID, &submission.AssignmentID,
		&submission.Status, &detailsString,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}
	if detailsString.Valid {
		submission.Details = detailsString.String
	}

	return submission, nil
}

func (r *SubmissionsSQLRepo) Update(submission *submissions.Submission) error {
	log.Printf("%+v\n", submission)

	_, err := r.DB.Exec(
		"UPDATE submissions SET status = $1, details = $2 WHERE id = $3",
		submission.Status, submission.Details, submission.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

package repo

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/maxshend/grader/pkg/attachments"
	"github.com/maxshend/grader/pkg/repo"
	"github.com/maxshend/grader/pkg/submissions"
)

type SubmissionsSQLRepo struct {
	DB *sql.DB
}

func NewSubmissionsSQLRepo(db *sql.DB) *SubmissionsSQLRepo {
	return &SubmissionsSQLRepo{DB: db}
}

func (r *SubmissionsSQLRepo) Create(sqlExec repo.SqlQueryable, userID int64, assignmentID int64) (*submissions.Submission, error) {
	submission := &submissions.Submission{
		UserID:       userID,
		AssignmentID: assignmentID,
		Status:       submissions.InProgress,
	}
	err := sqlExec.QueryRow(
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
	sqlExec repo.SqlQueryable,
	submissionID int64,
	attachments []*attachments.Attachment,
) ([]*submissions.Attachment, error) {
	stm, err := sqlExec.Prepare(pq.CopyIn("submission_attachments", "url", "name", "submission_id"))
	if err != nil {
		return nil, err
	}
	defer stm.Close()

	submissionAttachments := []*submissions.Attachment{}

	for _, attachment := range attachments {
		_, err = stm.Exec(attachment.URL, attachment.Name, submissionID)
		if err != nil {
			return nil, err
		}

		submissionAttachments = append(
			submissionAttachments,
			&submissions.Attachment{URL: attachment.URL, Name: attachment.Name},
		)
	}

	_, err = stm.Exec()
	if err != nil {
		return nil, err
	}

	return submissionAttachments, nil
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
	_, err := r.DB.Exec(
		"UPDATE submissions SET status = $1, details = $2 WHERE id = $3",
		submission.Status, submission.Details, submission.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *SubmissionsSQLRepo) GetByUserAssignment(
	assignmentID int64,
	userID int64,
	limit int,
	offset int,
) ([]*submissions.Submission, error) {
	rows, err := r.DB.Query(
		"SELECT id, status, details, created_at "+
			"FROM submissions WHERE user_id = $1 AND assignment_id = $2 "+
			"ORDER BY id DESC LIMIT $3 OFFSET $4",
		userID, assignmentID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*submissions.Submission{}
	for rows.Next() {
		detailsString := sql.NullString{}
		submission := &submissions.Submission{UserID: userID, AssignmentID: assignmentID}
		err = rows.Scan(
			&submission.ID, &submission.Status, &detailsString, &submission.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if detailsString.Valid {
			submission.Details = detailsString.String
		}

		result = append(result, submission)
	}
	if err = rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

func (r *SubmissionsSQLRepo) GetByAssignment(
	assignmentID int64,
	limit, offset int,
) ([]*submissions.Submission, error) {
	rows, err := r.DB.Query(
		"SELECT submissions.id, submissions.status, submissions.details, "+
			"submissions.created_at, users.username AS username "+
			"FROM submissions JOIN users ON submissions.user_id = users.id WHERE assignment_id = $1 "+
			"ORDER BY id DESC LIMIT $2 OFFSET $3",
		assignmentID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*submissions.Submission{}
	for rows.Next() {
		detailsString := sql.NullString{}
		submission := &submissions.Submission{AssignmentID: assignmentID}
		err = rows.Scan(
			&submission.ID, &submission.Status, &detailsString, &submission.CreatedAt, &submission.Username,
		)
		if err != nil {
			return nil, err
		}
		if detailsString.Valid {
			submission.Details = detailsString.String
		}

		result = append(result, submission)
	}
	if err = rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

func (r *SubmissionsSQLRepo) CreateTxn() (*sql.Tx, error) {
	return r.DB.Begin()
}

package repo

import (
	"database/sql"

	"github.com/maxshend/grader/pkg/assignments"
)

type AssignmentsSQLRepo struct {
	DB *sql.DB
}

func NewAssignmentsSQLRepo(db *sql.DB) *AssignmentsSQLRepo {
	return &AssignmentsSQLRepo{DB: db}
}

func (r *AssignmentsSQLRepo) GetAll(limit int, offset int) ([]*assignments.Assignment, error) {
	rows, err := r.DB.Query(
		"SELECT id, title, grader_url "+
			"FROM assignments LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*assignments.Assignment{}
	for rows.Next() {
		assignment := &assignments.Assignment{}
		err = rows.Scan(
			&assignment.ID, &assignment.Title, &assignment.GraderURL,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, assignment)
	}
	if err = rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

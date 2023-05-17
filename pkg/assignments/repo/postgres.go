package repo

import (
	"database/sql"
	"strconv"

	"github.com/lib/pq"
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

func (r *AssignmentsSQLRepo) GetByID(id string) (*assignments.Assignment, error) {
	assignmentID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	assignment := &assignments.Assignment{}
	err = r.DB.QueryRow(
		"SELECT id, title, description, grader_url, container, part_id, files "+
			"FROM assignments WHERE id = $1 LIMIT 1",
		assignmentID,
	).Scan(
		&assignment.ID, &assignment.Title, &assignment.Description,
		&assignment.GraderURL, &assignment.Container, &assignment.PartID, pq.Array(&assignment.Files),
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return assignment, nil
}

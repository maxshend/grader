package repo

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/maxshend/grader/pkg/assignments"
)

type AssignmentsSQLRepo struct {
	DB *sql.DB
}

func NewAssignmentsSQLRepo(db *sql.DB) *AssignmentsSQLRepo {
	return &AssignmentsSQLRepo{DB: db}
}

func (r *AssignmentsSQLRepo) GetAllByCreator(creatorID int64, limit int, offset int) ([]*assignments.Assignment, error) {
	rows, err := r.DB.Query(
		"SELECT id, title, grader_url "+
			"FROM assignments WHERE (creator_id = $3 OR creator_id IS NULL) "+
			"ORDER BY id DESC LIMIT $1 OFFSET $2",
		limit, offset, creatorID,
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

func (r *AssignmentsSQLRepo) GetByID(id int64) (*assignments.Assignment, error) {
	assignment := &assignments.Assignment{}
	err := r.DB.QueryRow(
		"SELECT id, title, description, grader_url, container, part_id, files, creator_id "+
			"FROM assignments WHERE id = $1 LIMIT 1",
		id,
	).Scan(
		&assignment.ID, &assignment.Title, &assignment.Description,
		&assignment.GraderURL, &assignment.Container, &assignment.PartID, pq.Array(&assignment.Files),
		&assignment.CreatorID,
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

func (r *AssignmentsSQLRepo) GetByIDByCreator(id int64, creatorID int64) (*assignments.Assignment, error) {
	assignment := &assignments.Assignment{}
	err := r.DB.QueryRow(
		"SELECT id, title, description, grader_url, container, part_id, files, creator_id "+
			"FROM assignments WHERE id = $1 AND (creator_id = $2 OR creator_id IS NULL) LIMIT 1",
		id, creatorID,
	).Scan(
		&assignment.ID, &assignment.Title, &assignment.Description,
		&assignment.GraderURL, &assignment.Container, &assignment.PartID, pq.Array(&assignment.Files),
		&assignment.CreatorID,
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

func (r *AssignmentsSQLRepo) GetByUserID(userID int64, limit, offset int) ([]*assignments.Assignment, error) {
	rows, err := r.DB.Query(
		"SELECT assignments.id, assignments.title "+
			"FROM assignments JOIN submissions ON assignments.id = submissions.assignment_id "+
			"WHERE submissions.user_id = $1 GROUP BY assignments.id, assignments.title "+
			"ORDER BY assignments.id DESC LIMIT $2 OFFSET $3",
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*assignments.Assignment{}
	for rows.Next() {
		assignment := &assignments.Assignment{}
		err = rows.Scan(
			&assignment.ID, &assignment.Title,
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

func (r *AssignmentsSQLRepo) Create(
	creatorID int64,
	title, description, graderURL,
	container, partID string, files []string,
) (*assignments.Assignment, error) {
	assignment := &assignments.Assignment{
		CreatorID:   creatorID,
		Title:       title,
		Description: description,
		GraderURL:   graderURL,
		Container:   container,
		PartID:      partID,
		Files:       files,
	}

	err := r.DB.QueryRow(
		"INSERT INTO assignments (title, description, grader_url, container, part_id, files, creator_id) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		title, description, graderURL, container, partID, pq.Array(files), creatorID,
	).Scan(&assignment.ID)
	if err != nil {
		return nil, err
	}

	return assignment, nil
}

func (r *AssignmentsSQLRepo) Update(assignment *assignments.Assignment) (*assignments.Assignment, error) {
	_, err := r.DB.Exec(
		"UPDATE assignments SET title = $1, description = $2, grader_url = $3, container = $4, "+
			"part_id = $5, files = $6 WHERE id = $7",
		assignment.Title, assignment.Description, assignment.GraderURL, assignment.Container,
		assignment.PartID, pq.Array(assignment.Files), assignment.ID,
	)
	if err != nil {
		return nil, err
	}

	return assignment, nil
}

func (r *AssignmentsSQLRepo) GetByTitle(title string) (*assignments.Assignment, error) {
	assignment := &assignments.Assignment{}
	err := r.DB.QueryRow(
		"SELECT id, title, description, grader_url, container, part_id, files, creator_id "+
			"FROM assignments WHERE title = $1 LIMIT 1",
		title,
	).Scan(
		&assignment.ID, &assignment.Title, &assignment.Description,
		&assignment.GraderURL, &assignment.Container, &assignment.PartID, pq.Array(&assignment.Files),
		&assignment.CreatorID,
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

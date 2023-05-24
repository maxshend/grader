package repo

import (
	"database/sql"

	"github.com/maxshend/grader/pkg/users"
)

type UsersSQLRepo struct {
	DB *sql.DB
}

func NewUsersSQLRepo(db *sql.DB) *UsersSQLRepo {
	return &UsersSQLRepo{DB: db}
}

func (r *UsersSQLRepo) Create(username, password string, role int) (*users.User, error) {
	user := &users.User{Username: username, Role: role}

	err := r.DB.QueryRow(
		"INSERT INTO users (username, password, role) VALUES ($1, $2, $3) RETURNING id",
		username,
		password,
		role,
	).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UsersSQLRepo) GetByID(id int64) (*users.User, error) {
	user := &users.User{}

	err := r.DB.QueryRow("SELECT id, username, role FROM users WHERE id = $1 LIMIT 1", id).Scan(
		&user.ID, &user.Username, &user.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return user, nil
}

func (r *UsersSQLRepo) GetByUsername(username string) (*users.User, error) {
	user := &users.User{}

	err := r.DB.QueryRow("SELECT id, username, role FROM users WHERE username = $1 LIMIT 1", username).Scan(
		&user.ID, &user.Username, &user.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return user, nil
}

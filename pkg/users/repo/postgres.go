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

func (r *UsersSQLRepo) Create(username, password string, provider, role int) (*users.User, error) {
	user := &users.User{Username: username, Role: role}

	err := r.DB.QueryRow(
		"INSERT INTO users (username, password, role, provider) VALUES ($1, $2, $3, $4) RETURNING id",
		username,
		password,
		role,
		provider,
	).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UsersSQLRepo) GetByID(id int64) (*users.User, error) {
	user := &users.User{}

	err := r.DB.QueryRow(
		"SELECT id, username, role, is_admin, password, provider FROM users WHERE id = $1 LIMIT 1",
		id,
	).Scan(
		&user.ID, &user.Username, &user.Role, &user.IsAdmin, &user.Password, &user.Provider,
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

	err := r.DB.QueryRow(
		"SELECT id, username, role, is_admin, password, provider FROM users WHERE username = $1 LIMIT 1",
		username,
	).Scan(
		&user.ID, &user.Username, &user.Role, &user.IsAdmin, &user.Password, &user.Provider,
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

func (r *UsersSQLRepo) GetByUsernameProvider(username string, provider int) (*users.User, error) {
	user := &users.User{}

	err := r.DB.QueryRow(
		"SELECT id, username, role, is_admin, password, provider FROM users "+
			"WHERE username = $1 AND provider = $2 LIMIT 1",
		username, provider,
	).Scan(
		&user.ID, &user.Username, &user.Role, &user.IsAdmin, &user.Password, &user.Provider,
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

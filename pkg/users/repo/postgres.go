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

func (r *UsersSQLRepo) GetAll(limit int, offset int) ([]*users.User, error) {
	rows, err := r.DB.Query(
		"SELECT id, username, password, is_admin, provider "+
			"FROM users ORDER BY id DESC LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*users.User{}
	for rows.Next() {
		user := &users.User{}
		err = rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.IsAdmin, &user.Provider,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, user)
	}
	if err = rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

func (r *UsersSQLRepo) Create(username, password string, provider int, isAdmin bool) (*users.User, error) {
	user := &users.User{Username: username, IsAdmin: isAdmin}

	err := r.DB.QueryRow(
		"INSERT INTO users (username, password, is_admin, provider) VALUES ($1, $2, $3, $4) RETURNING id",
		username,
		password,
		isAdmin,
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
		"SELECT id, username, is_admin, password, provider FROM users WHERE id = $1 LIMIT 1",
		id,
	).Scan(
		&user.ID, &user.Username, &user.IsAdmin, &user.Password, &user.Provider,
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
		"SELECT id, username, is_admin, password, provider FROM users WHERE username = $1 LIMIT 1",
		username,
	).Scan(
		&user.ID, &user.Username, &user.IsAdmin, &user.Password, &user.Provider,
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
		"SELECT id, username, is_admin, password, provider FROM users "+
			"WHERE username = $1 AND provider = $2 LIMIT 1",
		username, provider,
	).Scan(
		&user.ID, &user.Username, &user.IsAdmin, &user.Password, &user.Provider,
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

func (r *UsersSQLRepo) Update(user *users.User) (*users.User, error) {
	_, err := r.DB.Exec(
		"UPDATE users SET is_admin = $2 WHERE id = $1",
		user.ID, user.IsAdmin,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

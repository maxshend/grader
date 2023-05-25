package repo

import (
	"database/sql"

	"github.com/maxshend/grader/pkg/sessions"
)

type SessionsSQLRepo struct {
	DB *sql.DB
}

func NewSessionsSQLRepo(db *sql.DB) *SessionsSQLRepo {
	return &SessionsSQLRepo{
		DB: db,
	}
}

func (sm *SessionsSQLRepo) GetByToken(token string) (*sessions.Session, error) {
	session := &sessions.Session{Token: token}
	err := sm.DB.QueryRow("SELECT id, user_id FROM sessions WHERE token = $1", token).Scan(&session.ID, &session.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return session, nil
}

func (sm *SessionsSQLRepo) Create(userID int64, token string) (*sessions.Session, error) {
	session := &sessions.Session{UserID: userID, Token: token}

	err := sm.DB.QueryRow("INSERT INTO sessions (user_id, token) VALUES ($1, $2) RETURNING id", userID, token).Scan(&session.ID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

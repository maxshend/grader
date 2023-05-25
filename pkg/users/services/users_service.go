package services

import (
	"github.com/maxshend/grader/pkg/users"
	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	Repo users.RepositoryInterface
}

type UsersServiceInterface interface {
	Create(username, password, password_confirmation string) (*users.User, error)
	GetByID(int64) (*users.User, error)
	GetByUsername(string) (*users.User, error)
}

func NewUsersService(repo users.RepositoryInterface) UsersServiceInterface {
	return &UsersService{Repo: repo}
}

const (
	ErrPasswordConfirmation  = "Password should match password confirmation"
	ErrUsernameBlank         = "Username should be present"
	ErrUsernameAlreadyExists = "This username already exists"
	ErrPasswordTooShort      = "Password is too short"
	MinPasswordLength        = 8
)

func (s *UsersService) Create(username, password, password_confirmation string) (user *users.User, err error) {
	user = &users.User{Username: username}

	if len(username) == 0 {
		err = &UserValidationError{ErrUsernameBlank}
		return
	}

	if len(password) < MinPasswordLength {
		err = &UserValidationError{ErrPasswordTooShort}
		return
	}

	if password != password_confirmation {
		err = &UserValidationError{ErrPasswordConfirmation}
		return
	}

	foundUser, err := s.Repo.GetByUsername(username)
	if err != nil {
		return
	}
	if foundUser != nil {
		err = &UserValidationError{ErrUsernameAlreadyExists}
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.Repo.Create(username, string(hash), users.RegularUser)
}

func (s *UsersService) GetByID(id int64) (*users.User, error) {
	return s.Repo.GetByID(id)
}

func (s *UsersService) GetByUsername(username string) (*users.User, error) {
	return s.Repo.GetByUsername(username)
}

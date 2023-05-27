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
	CheckCredentials(username, password string) (*users.User, error)
}

func NewUsersService(repo users.RepositoryInterface) UsersServiceInterface {
	return &UsersService{Repo: repo}
}

const (
	MsgPasswordConfirmation  = "Password should match password confirmation"
	MsgUsernameBlank         = "Username should be present"
	MsgUsernameAlreadyExists = "This username already exists"
	MsgPasswordTooShort      = "Password is too short"
	MinPasswordLength        = 8
)

var (
	MsgInvalidUserCredentials = "Invalid email or password"
)

func (s *UsersService) Create(username, password, password_confirmation string) (user *users.User, err error) {
	user = &users.User{Username: username}

	if len(username) == 0 {
		err = &UserValidationError{MsgUsernameBlank}
		return
	}

	if len(password) < MinPasswordLength {
		err = &UserValidationError{MsgPasswordTooShort}
		return
	}

	if password != password_confirmation {
		err = &UserValidationError{MsgPasswordConfirmation}
		return
	}

	foundUser, err := s.Repo.GetByUsername(username)
	if err != nil {
		return
	}
	if foundUser != nil {
		err = &UserValidationError{MsgUsernameAlreadyExists}
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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

func (s *UsersService) CheckCredentials(username, password string) (*users.User, error) {
	if len(password) == 0 {
		return nil, &UserCredentialsError{MsgInvalidUserCredentials}
	}

	foundUser, err := s.Repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if foundUser == nil {
		return nil, &UserCredentialsError{MsgInvalidUserCredentials}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password)); err != nil {
		return nil, &UserCredentialsError{MsgInvalidUserCredentials}
	}

	return foundUser, nil
}

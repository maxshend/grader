package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/maxshend/grader/pkg/users"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

type UsersService struct {
	Repo users.RepositoryInterface
}

type UsersServiceInterface interface {
	Create(username, password, password_confirmation string) (*users.User, error)
	CreateOauth(token *oauth2.Token, provider int) (*users.User, error)
	GetByID(int64) (*users.User, error)
	GetAll() ([]*users.User, error)
	GetByUsername(string) (*users.User, error)
	CheckCredentials(username, password string) (*users.User, error)
	Update(*users.User) (*users.User, error)
	UpdateProfile(
		user *users.User,
		new_password, new_password_confirmation string,
	) (*users.User, error)
}

func NewUsersService(repo users.RepositoryInterface) UsersServiceInterface {
	return &UsersService{Repo: repo}
}

const (
	MsgPasswordConfirmation = "Password should match password confirmation"
	MsgUsernameBlank        = "Username should be present"
	MsgPasswordTooShort     = "Password is too short"
	MinPasswordLength       = 8
)

var (
	MsgInvalidUserCredentials = "Invalid username or password"
	MsgInvalidCurrentPassword = "Invalid current password"
)

func (s *UsersService) GetAll() ([]*users.User, error) {
	// TODO: Pagination handling
	return s.Repo.GetAll(100, 0)
}

func (s *UsersService) Create(username, password, password_confirmation string) (user *users.User, err error) {
	return s.create(username, password, password_confirmation, users.DefaultProvider)
}

func (s *UsersService) CreateOauth(token *oauth2.Token, provider int) (*users.User, error) {
	rawEmail := token.Extra("email")
	email := ""
	okEmail := true
	if rawEmail != nil {
		email, okEmail = rawEmail.(string)
	}
	rawID, okID := token.Extra("user_id").(float64)
	if !okEmail || !okID {
		return nil, OauthDataConversionError
	}

	username := email
	if len(username) == 0 {
		username = fmt.Sprintf("vk_%f", rawID)
	}
	password := uuid.NewString()
	user, err := s.create(username, password, password, provider)
	isDup := false
	if err != nil {
		_, isDup = err.(*UserAlreadyExistsError)
	}
	if isDup {
		return user, nil
	}

	return user, err
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

func (s *UsersService) Update(user *users.User) (*users.User, error) {
	return s.Repo.Update(user)
}

func (s *UsersService) UpdateProfile(
	user *users.User,
	new_password, new_password_confirmation string,
) (*users.User, error) {
	err := s.validateUser(user)
	if err != nil {
		return nil, err
	}
	if new_password != new_password_confirmation {
		return nil, &UserValidationError{MsgPasswordConfirmation}
	}

	foundUser, err := s.Repo.GetByUsername(user.Username)
	if err != nil {
		return nil, err
	}
	if foundUser != nil && user.ID != foundUser.ID {
		return nil, &UserAlreadyExistsError{user.Username}
	}

	if len(new_password) > 0 {
		hash, err := s.generatePasswordHash(new_password)
		if err != nil {
			return nil, err
		}
		user.Password = hash
	}

	return s.Repo.Update(user)
}

func (s *UsersService) create(username, password, password_confirmation string, provider int) (user *users.User, err error) {
	user = &users.User{Username: username}

	err = s.validateUser(user)
	if err != nil {
		return
	}
	err = s.validatePassword(password, password_confirmation)
	if err != nil {
		return
	}

	foundUser, err := s.Repo.GetByUsername(username)
	if err != nil {
		return
	}
	if foundUser != nil {
		err = &UserAlreadyExistsError{username}
		return
	}

	hash, err := s.generatePasswordHash(password)
	if err != nil {
		return nil, err
	}

	return s.Repo.Create(username, hash, provider, false)
}

func (s *UsersService) generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (s *UsersService) validateUser(user *users.User) error {
	if len(user.Username) == 0 {
		return &UserValidationError{MsgUsernameBlank}
	}

	return nil
}

func (s *UsersService) validatePassword(password, password_confirmation string) error {
	if len(password) < MinPasswordLength {
		return &UserValidationError{MsgPasswordTooShort}
	}

	if password != password_confirmation {
		return &UserValidationError{MsgPasswordConfirmation}
	}

	return nil
}

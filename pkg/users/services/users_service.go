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
	GetByUsername(string) (*users.User, error)
	CheckCredentials(username, password string) (*users.User, error)
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
	MsgInvalidUserCredentials = "Invalid email or password"
)

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

func (s *UsersService) create(username, password, password_confirmation string, provider int) (user *users.User, err error) {
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

	foundUser, err := s.Repo.GetByUsernameProvider(username, provider)
	if err != nil {
		return
	}
	if foundUser != nil {
		user = foundUser
		err = &UserAlreadyExistsError{username}
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.Repo.Create(username, string(hash), provider, users.RegularUser)
}

package services

import (
	"errors"
	"fmt"
)

type UserValidationError struct {
	Message string
}

func (e *UserValidationError) Error() string {
	return e.Message
}

type UserCredentialsError struct {
	Message string
}

func (e *UserCredentialsError) Error() string {
	return e.Message
}

type UserAlreadyExistsError struct {
	Username string
}

func (e *UserAlreadyExistsError) Error() string {
	return fmt.Sprintf("User with username %q already exists", e.Username)
}

var OauthDataConversionError = errors.New("Can't convert oauth data")

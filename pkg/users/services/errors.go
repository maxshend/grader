package services

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

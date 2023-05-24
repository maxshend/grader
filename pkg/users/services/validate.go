package services

type UserValidationError struct {
	Message string
}

func (e *UserValidationError) Error() string {
	return e.Message
}

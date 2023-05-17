package services

type AssignmentValidationError struct {
	Message string
}

func (e *AssignmentValidationError) Error() string {
	return e.Message
}

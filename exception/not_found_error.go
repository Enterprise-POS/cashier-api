package exception

type NotFoundError struct {
	Code    int
	Status  string
	Message string
}

func NewNotFoundError(error string) *NotFoundError {
	return &NotFoundError{
		Code:    404,
		Status:  "Not Found",
		Message: error,
	}
}

func (notFoundError *NotFoundError) Error() string {
	return notFoundError.Message
}

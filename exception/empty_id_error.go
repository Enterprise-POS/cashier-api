package exception

import "fmt"

type EmptyUidError struct {
	Code    int
	Status  string
	Message string
}

func NewEmptyUidError(idNotProvided string) *EmptyUidError {
	return &EmptyUidError{
		Code:    400,
		Status:  "Bad Request",
		Message: fmt.Sprintf("%s is not provided !", idNotProvided),
	}
}

func (notFoundError *EmptyUidError) Error() string {
	return notFoundError.Message
}

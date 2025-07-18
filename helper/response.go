package common

type WebResponse struct {
	Code   int            `json:"code"`
	Status StatusResponse `json:"status"`
	Data   interface{}    `json:"data"`
}

type StatusResponse = string

const StatusSuccess StatusResponse = "success"
const StatusError StatusResponse = "error"
const StatusInternalServerError = "internal server error"

func NewWebResponse(code int, status StatusResponse, data interface{}) *WebResponse {
	return &WebResponse{
		Code:   code,
		Status: status,
		Data:   data,
	}
}

type WebResponseError struct {
	Code    int            `json:"code"`
	Status  StatusResponse `json:"status"`
	Message string         `json:"message"`
}

func NewWebResponseError(code int, status StatusResponse, message string) *WebResponseError {
	return &WebResponseError{
		Code:    code,
		Status:  status,
		Message: message,
	}
}

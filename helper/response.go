package common

type WebResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

func NewWebResponse(code int, status string, data interface{}) *WebResponse {
	return &WebResponse{
		Code:   code,
		Status: status,
		Data:   data,
	}
}

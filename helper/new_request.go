package common

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type NewRequest struct {
	Url       string
	Method    string
	Body      *strings.Reader
	TimeoutMs int
}

func (nr *NewRequest) RunRequest(app *fiber.App) (*http.Request, *http.Response, error) {
	timeoutMs := 0
	if nr.TimeoutMs == 0 {
		timeoutMs = int(time.Second * 10)
	}

	request := httptest.NewRequest(nr.Method, nr.Url, nr.Body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request, fiber.TestConfig{Timeout: time.Duration(timeoutMs) * time.Millisecond})

	return request, response, err
}

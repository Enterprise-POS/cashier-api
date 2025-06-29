package common

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebResponse_SuccessBody(t *testing.T) {
	webResponse := NewWebResponse(http.StatusOK, "success", nil)
	assert.Equal(t, 200, webResponse.Code, "Check response code property")
	assert.Equal(t, "success", webResponse.Status, "Check response status property")
	assert.Nil(t, webResponse.Data)
}

func TestWebResponse_SuccessBodyWithData(t *testing.T) {
	webResponse := NewWebResponse(http.StatusOK, "success", map[string]interface{}{"message": "Ok"})
	assert.Equal(t, 200, webResponse.Code, "Check response code property")
	assert.Equal(t, "success", webResponse.Status, "Check response status property")
	assert.NotNil(t, webResponse.Data)

	data, ok := webResponse.Data.(map[string]interface{})
	assert.True(t, ok, "Data should be of type map[string]interface{}")
	assert.Equal(t, "Ok", data["message"])
}

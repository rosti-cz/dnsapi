package main

import (
	"testing"
	"net/http/httptest"
	"github.com/labstack/echo"
	"strings"
	"github.com/stretchr/testify/assert"
	"net/http"
)

func TestGetZonesHandler(t *testing.T) {
	// Setup
	e := echo.New()
	request := httptest.NewRequest(echo.GET, "/zones/", strings.NewReader(""))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	// Assertions
	if assert.NoError(t, GetZonesHandler(context)) {
		assert.Equal(t, http.StatusOK, recorder.Code)
		// TODO: test content
	}
}
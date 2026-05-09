package dnsapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
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

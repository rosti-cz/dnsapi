package main

import (
	"net/http"
	"github.com/labstack/echo"
)

func GetZonesHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
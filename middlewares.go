package main

import (
	"strings"

	"github.com/labstack/echo"
)

// Process is the middleware function.
func TokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenHeader := c.Request().Header.Get("Authorization")
		token := strings.Replace(tokenHeader, "Token ", "", -1)

		if (token != config.APIToken || config.APIToken == "") && c.Request().URL.Path != "/metrics" {
			return c.JSONPretty(403, map[string]string{"message": "access denied"}, " ")
		}

		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}

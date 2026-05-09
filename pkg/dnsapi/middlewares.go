package dnsapi

import (
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo"
)

var skipPaths = []string{
	"/",
	"/metrics",
	"/ui",
	"/swagger",
	"/swagger/index.html",
	"/swagger/doc.json",
	"/swagger/swagger-ui.css",
	"/swagger/swagger-ui-bundle.js",
	"/swagger/swagger-ui-standalone-preset.js",
	"/swagger/favicon-16x16.png",
	"/swagger/favicon-32x32.png",
}

func shouldSkipAuth(path string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath || strings.HasPrefix(path, skipPath+"/") {
			return true
		}
	}

	return false
}

func SentryMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				captureRecoveredPanic(recovered)
				panic(recovered)
			}

			if err != nil {
				sentry.WithScope(func(scope *sentry.Scope) {
					scope.SetRequest(c.Request())
					scope.SetTag("path", c.Path())
					sentry.CaptureException(err)
				})
			}
		}()

		return next(c)
	}
}

// Process is the middleware function.
func TokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenHeader := c.Request().Header.Get("Authorization")
		token := strings.TrimPrefix(tokenHeader, "Bearer ")
		token = strings.TrimPrefix(token, "Token ")

		if shouldSkipAuth(c.Request().URL.Path) {
			return next(c)
		}

		if token != config.APIToken || config.APIToken == "" {
			return c.JSONPretty(403, map[string]string{"message": "access denied"}, " ")
		}

		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}

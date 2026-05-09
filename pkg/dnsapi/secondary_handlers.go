package dnsapi

import (
	"io"
	"net/http"

	"github.com/labstack/echo"
)

// @Summary Write bind config (secondary mode)
// @Description Receives a rendered named.conf include file, writes it to disk, and reloads bind9.
// @ID put-bind-config
// @Tags secondary
// @Security ApiKeyAuth
// @Accept plain
// @Produce json
// @Success 200 {object} MessageResponse
// @Failure 500 {object} ErrorResponse
// @Router /bind/config [put]
func PutBindConfigHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	if err := WriteBindConfig(SecondaryBindConfigPath, string(body)); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	if err := ReloadBind(); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

// @Summary Write zone file (secondary mode)
// @Description Receives a zone file and writes it to the bind zone directory.
// @ID put-zone-file
// @Tags secondary
// @Security ApiKeyAuth
// @Accept plain
// @Param domain path string true "Zone domain name"
// @Produce json
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bind/zones/{domain} [put]
func PutZoneFileHandler(c echo.Context) error {
	domain := c.Param("domain")
	if domain == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "domain is required"})
	}
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	if err := WriteZoneFile(domain, string(body)); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

// @Summary Delete zone file (secondary mode)
// @Description Removes a zone file from the bind zone directory.
// @ID delete-zone-file
// @Tags secondary
// @Security ApiKeyAuth
// @Param domain path string true "Zone domain name"
// @Produce json
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bind/zones/{domain} [delete]
func DeleteZoneFileHandler(c echo.Context) error {
	domain := c.Param("domain")
	if domain == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "domain is required"})
	}
	if err := DeleteZoneFile(domain); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

// @Summary Reload bind9 (secondary mode)
// @Description Reloads bind9 via the configured reload command.
// @ID post-bind-reload
// @Tags secondary
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} MessageResponse
// @Failure 500 {object} ErrorResponse
// @Router /bind/reload [post]
func PostBindReloadHandler(c echo.Context) error {
	if err := ReloadBind(); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

// @Summary Refresh zone (secondary mode)
// @Description Forces a zone refresh using the configured refresh command.
// @ID post-zone-refresh
// @Tags secondary
// @Security ApiKeyAuth
// @Param domain path string true "Zone domain name"
// @Produce json
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bind/refresh/{domain} [post]
func PostZoneRefreshHandler(c echo.Context) error {
	domain := c.Param("domain")
	if domain == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "domain is required"})
	}
	if err := RefreshZone(domain); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

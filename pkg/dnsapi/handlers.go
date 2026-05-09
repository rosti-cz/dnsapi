package dnsapi

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

func parseUintParam(c echo.Context, key string) (uint, error) {
	value := c.Param(key)
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "invalid " + key,
		}
	}

	if parsed < 0 {
		return 0, &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "invalid " + key,
		}
	}

	return uint(parsed), nil
}

// ##############
// Zones handlers
// ##############

// @Summary List zones
// @Description Return all zones including records.
// @ID get-zones
// @Tags zones
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} Zone
// @Failure 500 {object} ErrorResponse
// @Router /zones/ [get]
func GetZonesHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	var zones []Zone

	err := db.Model(&Zone{}).Preload("Records").Find(&zones).Error
	if err != nil {
		return c.JSONPretty(http.StatusInternalServerError, ErrorResponse{Message: err.Error()}, "  ")
	}

	return c.JSONPretty(http.StatusOK, zones, "  ")
}

// @Summary Get zone
// @Description Return one zone including records.
// @ID get-zone
// @Tags zones
// @Security ApiKeyAuth
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Success 200 {object} Zone
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /zones/{zone_id} [get]
func GetZoneHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	var zoneId = c.Param("zone_id")

	var zone Zone

	err := db.Where("id = ?", zoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		if strings.Trim(err.Error(), "\n") == RECORD_NOT_FOUND_MESSAGE {
			return c.JSONPretty(http.StatusNotFound, ErrorResponse{Message: "zone not found"}, "  ")
		}
		return c.JSONPretty(http.StatusInternalServerError, ErrorResponse{Message: err.Error()}, "  ")
	}

	return c.JSONPretty(http.StatusOK, zone, "  ")
}

// @Summary Create zone
// @Description Create a new zone.
// @ID create-zone
// @Tags zones
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param zone body CreateZoneRequest true "Zone payload"
// @Success 201 {object} Zone
// @Failure 400 {object} ErrorResponse
// @Router /zones/ [post]
func NewZoneHandler(c echo.Context) error {
	var body CreateZoneRequest

	err := c.Bind(&body)
	if err != nil {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	pzone, errs := NewZone(body.Domain, strings.Split(body.Tags, ","), body.AbuseEmail, body.Owner)
	if len(errs) != 0 {
		message := ""
		for _, err := range errs {
			message += "\n" + err.Error()
		}

		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: message,
		}
	}

	return c.JSONPretty(http.StatusCreated, *pzone, "  ")
}

// @Summary Delete zone
// @Description Delete a zone and all its records.
// @ID delete-zone
// @Tags zones
// @Security ApiKeyAuth
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Success 200 {object} MessageResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /zones/{zone_id} [delete]
func DeleteZoneHandler(c echo.Context) error {
	zoneId, err := parseUintParam(c, "zone_id")
	if err != nil {
		return err
	}

	err = DeleteZone(zoneId)
	if err != nil {
		if strings.Trim(err.Error(), "\n") == RECORD_NOT_FOUND_MESSAGE {
			return &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: strings.Trim(err.Error(), "\n"),
			}
		}

		return &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return c.JSONPretty(http.StatusOK, map[string]string{"message": "deleted"}, "  ")
}

// @Summary Update zone
// @Description Update zone metadata (tags and abuse email only). Use the records endpoints to manage records.
// @ID update-zone
// @Tags zones
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Param zone body UpdateZoneRequest true "Zone metadata to update"
// @Success 200 {object} Zone
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /zones/{zone_id} [put]
func UpdateZoneHandler(c echo.Context) error {
	var zoneBody UpdateZoneRequest

	err := c.Bind(&zoneBody)
	if err != nil {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	zoneId, err := parseUintParam(c, "zone_id")
	if err != nil {
		return err
	}

	zone, errs := UpdateZone(zoneId, strings.Split(zoneBody.Tags, ","), zoneBody.AbuseEmail, zoneBody.Owner)
	if len(errs) != 0 {
		message := ""
		for _, err := range errs {
			message += "\n" + err.Error()
		}

		if strings.Trim(message, "\n") == RECORD_NOT_FOUND_MESSAGE {
			return &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: strings.Trim(message, "\n"),
			}
		}

		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: message,
		}
	}

	return c.JSONPretty(http.StatusOK, zone, "  ")
}

// @Summary Commit zone
// @Description Write zone changes into DNS servers.
// @ID commit-zone
// @Tags zones
// @Security ApiKeyAuth
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Success 200 {object} MessageResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /zones/{zone_id}/commit [put]
func CommitHandler(c echo.Context) error {
	zoneId, err := parseUintParam(c, "zone_id")
	if err != nil {
		return err
	}

	err = Commit(zoneId)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err.Error() == "Zone not found" {
			return c.JSONPretty(http.StatusNotFound, ErrorResponse{Message: "zone not found"}, "  ")
		}
		return c.JSONPretty(http.StatusInternalServerError, ErrorResponse{Message: err.Error()}, "  ")
	}

	return c.JSONPretty(http.StatusOK, map[string]string{"message": "committed"}, "  ")
}

// ################
// Records handlers
// ################

// @Summary List records
// @Description Return all records for a zone.
// @ID get-records
// @Tags records
// @Security ApiKeyAuth
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Success 200 {array} Record
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /zones/{zone_id}/records/ [get]
func GetRecordsHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	zoneId := c.Param("zone_id")

	var records []Record

	err := db.Model(&Record{}).Where("zone_id = ?", zoneId).Find(&records).Error
	if err != nil {
		if strings.Trim(err.Error(), "\n") == RECORD_NOT_FOUND_MESSAGE {
			return c.JSONPretty(http.StatusNotFound, ErrorResponse{Message: "zone not found"}, "  ")
		}
		return c.JSONPretty(http.StatusInternalServerError, ErrorResponse{Message: err.Error()}, "  ")
	}

	return c.JSONPretty(http.StatusOK, records, "  ")
}

// @Summary Get record
// @Description Return one record within a zone.
// @ID get-record
// @Tags records
// @Security ApiKeyAuth
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Param record_id path int true "Record ID"
// @Success 200 {object} Record
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /zones/{zone_id}/records/{record_id} [get]
func GetRecordHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	var record Record
	zoneId, err := parseUintParam(c, "zone_id")
	if err != nil {
		return err
	}
	recordId, err := parseUintParam(c, "record_id")
	if err != nil {
		return err
	}

	err = db.Where("id = ? AND zone_id = ?", recordId, zoneId).Find(&record).Error
	if err != nil {
		if strings.Trim(err.Error(), "\n") == RECORD_NOT_FOUND_MESSAGE {
			return &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: strings.Trim(err.Error(), "\n"),
			}
		}
		return &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return c.JSONPretty(http.StatusOK, record, "  ")
}

// @Summary Create record
// @Description Create a new record in a zone. Call commit afterwards to write it to bind.
// @ID create-record
// @Tags records
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Param record body CreateRecordRequest true "Record payload"
// @Success 201 {object} Record
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /zones/{zone_id}/records/ [post]
func NewRecordHandler(c echo.Context) error {
	var recordBody Record

	err := c.Bind(&recordBody)
	if err != nil {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	zoneId, err := parseUintParam(c, "zone_id")
	if err != nil {
		return err
	}

	record, errs := NewRecord(
		zoneId,
		recordBody.Name,
		recordBody.TTL,
		recordBody.Type,
		recordBody.Prio,
		recordBody.Value,
	)
	if len(errs) != 0 {
		message := ""
		for _, err := range errs {
			message += "\n" + err.Error()
		}

		if strings.Trim(message, "\n") == RECORD_NOT_FOUND_MESSAGE {
			return &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: strings.Trim(message, "\n"),
			}
		}

		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: strings.Trim(message, "\n"),
		}
	}

	return c.JSONPretty(http.StatusCreated, record, "  ")
}

// @Summary Delete record
// @Description Delete a record within a zone.
// @ID delete-record
// @Tags records
// @Security ApiKeyAuth
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Param record_id path int true "Record ID"
// @Success 200 {object} MessageResponse
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /zones/{zone_id}/records/{record_id} [delete]
func DeleteRecordHandler(c echo.Context) error {
	zoneId, err := parseUintParam(c, "zone_id")
	if err != nil {
		return err
	}
	recordId, err := parseUintParam(c, "record_id")
	if err != nil {
		return err
	}

	err = DeleteRecord(zoneId, recordId)
	if err != nil {
		if strings.Trim(err.Error(), "\n") == RECORD_NOT_FOUND_MESSAGE {
			return &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: strings.Trim(err.Error(), "\n"),
			}
		}

		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	return c.JSONPretty(http.StatusOK, map[string]string{"message": "deleted"}, "  ")
}

// @Summary Update record
// @Description Update an existing record within a zone. Type cannot be changed; delete and recreate instead. Call commit afterwards.
// @ID update-record
// @Tags records
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Param record_id path int true "Record ID"
// @Param record body UpdateRecordRequest true "Record fields to update"
// @Success 200 {object} Record
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /zones/{zone_id}/records/{record_id} [put]
func UpdateRecordHandler(c echo.Context) error {
	var recordBody Record

	err := c.Bind(&recordBody)
	if err != nil {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	zoneId, err := parseUintParam(c, "zone_id")
	if err != nil {
		return err
	}
	recordId, err := parseUintParam(c, "record_id")
	if err != nil {
		return err
	}

	zone, errs := UpdateRecord(
		zoneId,
		recordId,
		recordBody.Name,
		recordBody.TTL,
		recordBody.Prio,
		recordBody.Value,
	)
	if len(errs) != 0 {
		message := ""
		for _, err := range errs {
			message += "\n" + err.Error()
		}

		if strings.Trim(message, "\n") == RECORD_NOT_FOUND_MESSAGE {
			return &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: strings.Trim(message, "\n"),
			}
		}

		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: strings.Trim(message, "\n"),
		}
	}

	return c.JSONPretty(http.StatusOK, zone, "  ")
}

// @Summary Metrics
// @Description Return prometheus metrics.
// @ID metrics
// @Tags default
// @Produce plain
// @Success 200 {string} string
// @Failure 500 {object} ErrorResponse
// @Router /metrics [get]
func GetMetricsHandler(c echo.Context) error {
	db := GetDatabaseConnection()
	responseBody := ""
	var count int64

	responseBody += fmt.Sprintf("dnsapi_last %d\n", time.Now().Unix())

	err := db.Model(&Zone{}).Preload("Records").Count(&count).Error
	if err != nil {
		log.Println("ERROR:", err)
		return &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: strings.Trim(err.Error(), "\n"),
		}
	}

	responseBody += fmt.Sprintf("dnsapi_zones_count %d\n", count)

	return c.String(200, responseBody)
}

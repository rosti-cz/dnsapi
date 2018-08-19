package main

import (
	"net/http"
	"github.com/labstack/echo"
	"strings"
	"strconv"
	"github.com/jinzhu/gorm"
)

// ##############
// Zones handlers
// ##############

func GetZonesHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	var zones []Zone

	err := db.Model(&Zone{}).Preload("Records").Find(&zones).Error
	if err != nil {
		panic(err)
	}

	return c.JSONPretty(http.StatusOK, zones, "  ")
}

func GetZoneHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	var zoneId = c.Param("zone_id")

	var zone Zone

	err := db.Where("id = ?", zoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		panic(err)
	}

	return c.JSONPretty(http.StatusOK, zone, "  ")
}

func NewZoneHandler(c echo.Context) error {
	var zone Zone
	var pzone *Zone

	err := c.Bind(&zone)
	if err != nil {
		return &echo.HTTPError{
			Code: http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	pzone, errs := NewZone(zone.Domain, strings.Split(zone.Tags, ","), zone.AbuseEmail)
	if len(errs) != 0 {
		message := ""
		for _, err := range errs {
			message += "\n" + err.Error()
		}

		return &echo.HTTPError{
			Code: http.StatusBadRequest,
			Message: message,
		}
	}

	return c.JSONPretty(http.StatusCreated, *pzone, "  ")
}

func DeleteZoneHandler(c echo.Context) error {
	var zoneId = c.Param("zone_id")

	zoneIdInt, err := strconv.Atoi(zoneId)
	if err != nil {
		panic(err)
	}

	err = DeleteZone(uint(zoneIdInt))
	if err != nil {
		return &echo.HTTPError{
			Code: http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return c.JSONPretty(http.StatusOK, map[string]string{"message": "deleted"}, "  ")
}

func UpdateZoneHandler(c echo.Context) error {
	var zoneId = c.Param("zone_id")
	var zoneBody Zone

	err := c.Bind(&zoneBody)
	if err != nil {
		return &echo.HTTPError{
			Code: http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	zoneIdInt, err := strconv.Atoi(zoneId)
	if err != nil {
		panic(err)
	}

	zone, errs := UpdateZone(uint(zoneIdInt), strings.Split(zoneBody.Tags, ","), zoneBody.AbuseEmail)
	if len(errs) != 0 {
		message := ""
		for _, err := range errs {
			message += "\n" + err.Error()
		}

		return &echo.HTTPError{
			Code: http.StatusBadRequest,
			Message: message,
		}
	}

	return c.JSONPretty(http.StatusOK, zone, "  ")
}


func CommitHandler(c echo.Context) error {
	var zoneId = c.Param("zone_id")

	zoneIdInt, err := strconv.Atoi(zoneId)
	if err != nil {
		panic(err)
	}

	err = Commit(uint(zoneIdInt))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSONPretty(http.StatusNotFound, map[string]string{"message": "Zone not found"}, "  ")
		}
		panic(err)
	}

	return c.JSONPretty(http.StatusOK, map[string]string{"message": "committed"}, "  ")
}

// ################
// Records handlers
// ################

func GetRecordsHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	var records []Record

	err := db.Model(&Record{}).Find(&records).Error
	if err != nil {
		panic(err)
	}

	return c.JSONPretty(http.StatusOK, records, "  ")
}

func GetRecordHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	var record Record

	recordId := c.Param("record_id")

	err := db.Where("id = ?", recordId).Find(&record).Error
	if err != nil {
		panic(err)
	}

	return c.JSONPretty(http.StatusOK, record, "  ")
}

func NewRecordHandler(c echo.Context) error {
	var recordBody Record

	err := c.Bind(&recordBody)
	if err != nil {
		return &echo.HTTPError{
			Code: http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	zoneId := c.Param("zone_id")

	zoneIdInt, err := strconv.Atoi(zoneId)
	if err != nil {
		panic(err)
	}

	record, errs := NewRecord(
		uint(zoneIdInt),
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

		return &echo.HTTPError{
			Code: http.StatusBadRequest,
			Message: message,
		}
	}

	return c.JSONPretty(http.StatusCreated, record, "  ")
}

func DeleteRecordHandler(c echo.Context) error {
	recordId := c.Param("record_id")

	recordIdInt, err := strconv.Atoi(recordId)
	if err != nil {
		panic(err)
	}

	err = DeleteRecord(uint(recordIdInt))
	if err != nil {
		return &echo.HTTPError{
			Code: http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	return c.JSONPretty(http.StatusOK, map[string]string{"message": "deleted"}, "  ")
}

func UpdateRecordHandler(c echo.Context) error {
	var recordId = c.Param("record_id")
	var recordBody Record

	err := c.Bind(&recordBody)
	if err != nil {
		return &echo.HTTPError{
			Code: http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	recordIdInt, err := strconv.Atoi(recordId)
	if err != nil {
		panic(err)
	}

	zone, errs := UpdateRecord(
		uint(recordIdInt),
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

		return &echo.HTTPError{
			Code: http.StatusBadRequest,
			Message: message,
		}
	}

	return c.JSONPretty(http.StatusOK, zone, "  ")
}

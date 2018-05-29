package main

import (
	"net/http"
	"github.com/labstack/echo"
	"strings"
	"strconv"
)

func GetZonesHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	var zones []Zone

	err := db.Model(&Zone{}).Find(&zones).Error
	if err != nil {
		panic(err)
	}

	return c.JSONPretty(http.StatusOK, zones, "  ")
}

func GetZoneHandler(c echo.Context) error {
	db := GetDatabaseConnection()

	var zoneId = c.Param("zone_id")

	var zone Zone

	err := db.Where("id = ?", zoneId).Find(&zone).Error
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
		panic(err)
	}

	pzone, errs := NewZone(zone.Domain, strings.Split(zone.Tags, ","), zone.AbuseEmail)
	if len(errs) != 0 {
		panic(errs)
	}

	return c.JSONPretty(http.StatusOK, *pzone, "  ")
}

func DeleteZoneHandler(c echo.Context) error {
	var zoneId = c.Param("zone_id")

	zoneIdInt, err := strconv.Atoi(zoneId)
	if err != nil {
		panic(err)
	}

	err = DeleteZone(uint(zoneIdInt))
	if err != nil {
		panic(err)
	}

	return c.JSONPretty(http.StatusOK, map[string]string{"status": "deleted"}, "  ")
}

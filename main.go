package main

import (
	"github.com/kelseyhightower/envconfig"
	"log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var config Config

func init() {
	err := envconfig.Process("DNSAPI", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = config.Validate()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	log.Println(config)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())

	// Routes
	e.GET("/zones/", GetZonesHandler) // List of zone
	e.GET("/zones/:zone_id", nil) // Get one zone
	e.POST("/zones/", nil) // New zone
	e.DELETE("/zones/:zone_id", nil) // Delete zone
	e.PUT("/zones/:zone_id", nil) // Update zone

	e.GET("/zones/:zone_id/records/", nil) // List of records
	e.GET("/zones/:zone_id/records/:record_id", nil) // Get record
	e.POST("/zones/:zone_id/records/", nil) // New record
	e.DELETE("/zones/:zone_id/records/:record_id", nil) // Delete record
	e.PUT("/zones/:zone_id/records/:record_id", nil) // Update record

	e.GET("/export/", nil) // Export all data
	e.POST("/import/", nil) // Import all data

	// Start server
	e.Logger.Print("http://localhost:1323")
	e.Logger.Fatal(e.Start(":1323"))
}

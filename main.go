package main

import (
	"github.com/kelseyhightower/envconfig"
	"log"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"net"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"errors"
)

var config Config

var dbConnection *gorm.DB

func GetDatabaseConnection() *gorm.DB {
	if dbConnection == nil {
		db, err := gorm.Open("sqlite3", config.DatabasePath)

		if err != nil {
			log.Fatalln(err)
		}

		db.AutoMigrate(&Zone{})
		db.AutoMigrate(&Record{})

		dbConnection = db
	}

	return dbConnection
}

func FetchConfigData() {
	err := envconfig.Process("DNSAPI", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = config.Validate()
	if err != nil {
		log.Fatal(err.Error())
	}
}

// If necessary, this takes PrimaryNameServer and NameServers domain names and resolves IP addresses for
// PrimaryNameServerIP and SecondaryNameServerIPs. Panic in case of fire!
func SetNameServerIPs() {
	if config.PrimaryNameServerIP == "" {
		ips, err := net.LookupIP(config.PrimaryNameServer)
		if err != nil {
			panic(err)
		}

		for _, ip := range ips {
			config.PrimaryNameServerIP = ip.String()
			break
		}
	}
	if len(config.SecondaryNameServerIPs) == 0 {
		for _, secondaryNameServer := range config.NameServers {
			if secondaryNameServer == config.PrimaryNameServer {
				continue
			}

			ips, err := net.LookupIP(config.PrimaryNameServer)
			if err != nil {
				panic(err)
			}

			for _, ip := range ips {
				config.SecondaryNameServerIPs = append(config.SecondaryNameServerIPs, ip.String())
			}
		}
	}

	if config.PrimaryNameServerIP == "" {
		panic(errors.New("PrimaryNameServerIP is not set and can't be resolved"))
	}
	if len(config.SecondaryNameServerIPs) == 0 {
		panic(errors.New("SecondaryNameServerIPs is not set and can't be resolved"))
	}
}

func main() {
	FetchConfigData()
	SetNameServerIPs()

	// Database stuff
	db := GetDatabaseConnection()
	defer db.Close()

	log.Println(config)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())

	// Routes
	e.GET("/zones/", GetZonesHandler) // List of zone
	e.GET("/zones/:zone_id", GetZoneHandler) // Get one zone
	e.POST("/zones/", NewZoneHandler) // New zone
	e.DELETE("/zones/:zone_id", DeleteZoneHandler) // Delete zone
	e.PUT("/zones/:zone_id", UpdateZoneHandler) // Update zone

	e.GET("/zones/:zone_id/records/", GetRecordsHandler) // List of records
	e.GET("/zones/:zone_id/records/:record_id", GetRecordHandler) // Get record
	e.POST("/zones/:zone_id/records/", NewRecordHandler) // New record
	e.DELETE("/zones/:zone_id/records/:record_id", DeleteRecordHandler) // Delete record
	e.PUT("/zones/:zone_id/records/:record_id", UpdateRecordHandler) // Update record

	e.GET("/export/", nil) // Export all data
	e.POST("/import/", nil) // Import all data

	// Start server
	e.Logger.Print("http://localhost:1323")
	e.Logger.Fatal(e.Start(":1323"))
}

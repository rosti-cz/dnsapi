package dnsapi

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	_ "github.com/by-cx/dnsapi/docs"
	"github.com/getsentry/sentry-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"golang.org/x/crypto/acme/autocert"
)

//go:embed ui/index.html
var uiFS embed.FS

var config Config

var dbConnection *gorm.DB

func GetDatabaseConnection() *gorm.DB {
	if dbConnection == nil {
		db, err := gorm.Open("sqlite3", config.DatabasePath)

		if err != nil {
			log.Fatalln(err)
		}

		err = runMigrations(db)
		if err != nil {
			log.Fatalln(err)
		}

		dbConnection = db
	}

	return dbConnection
}

func FetchConfigData() {
	err := envconfig.Process("", &config)
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

			ips, err := net.LookupIP(secondaryNameServer)
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

// @title DNS API
// @version 1.1
// @description Simple DNS API for managing Bind zones and records.
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func Run() {
	FetchConfigData()

	if config.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              config.SentryDSN,
			AttachStacktrace: true,
			Environment:      config.SentryENV,
			TracesSampleRate: 1,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
		defer sentry.Flush(10 * time.Second)
	}

	log.Printf("dnsapi mode=%s secondary_instances=%d https=%t public_domain=%s", config.Mode, len(config.SecondaryInstanceList()), config.HTTPS, config.PublicDomain)

	if config.IsSecondaryMode() {
		runSecondary()
		return
	}

	runPrimary()
}

func runPrimary() {
	SetNameServerIPs()

	// Database stuff
	db := GetDatabaseConnection()
	defer db.Close()

	log.Println("Loaded configuration:")
	log.Printf("%+v\n", config)

	e := newServer(true)
	startServer(e)
}

func runSecondary() {
	log.Println("Loaded configuration:")
	log.Printf("%+v\n", config)

	e := newServer(false)
	startServer(e)
}

func newServer(includePrimaryRoutes bool) *echo.Echo {
	e := echo.New()

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 4 << 10, // 1 KB
	}))
	e.Use(SentryMiddleware)
	e.Use(TokenMiddleware)
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		if includePrimaryRoutes {
			return c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
		}
		return c.JSON(http.StatusOK, map[string]string{"mode": RuntimeModeSecondary})
	})
	e.GET("/swagger/*", echo.WrapHandler(httpSwagger.Handler()))
	e.GET("/metrics", GetMetricsHandler)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "mode": config.Mode})
	})

	if includePrimaryRoutes {
		e.GET("/ui", func(c echo.Context) error {
			content, err := uiFS.ReadFile("ui/index.html")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "ui not found"})
			}
			return c.HTMLBlob(http.StatusOK, content)
		})

		e.GET("/zones/", GetZonesHandler)                  // List of zone
		e.GET("/zones/:zone_id", GetZoneHandler)           // Get one zone
		e.POST("/zones/", NewZoneHandler)                  // New zone
		e.DELETE("/zones/:zone_id", DeleteZoneHandler)     // Delete the zone
		e.PUT("/zones/:zone_id", UpdateZoneHandler)        // Update the zone
		e.PUT("/zones/:zone_id/commit", CommitHandler)     // Commit the zone
		e.GET("/zones/:zone_id/test", TestZoneHandler)     // Test zone DNS
		e.GET("/zones/:zone_id/status", ZoneStatusHandler) // Zone commit/NS status

		e.GET("/zones/:zone_id/records/", GetRecordsHandler)                // List of records
		e.GET("/zones/:zone_id/records/:record_id", GetRecordHandler)       // Get record
		e.POST("/zones/:zone_id/records/", NewRecordHandler)                // New record
		e.DELETE("/zones/:zone_id/records/:record_id", DeleteRecordHandler) // Delete record
		e.PUT("/zones/:zone_id/records/:record_id", UpdateRecordHandler)    // Update record

		e.GET("/export/", nil)  // Export all data
		e.POST("/import/", nil) // Import all data
	}

	if !includePrimaryRoutes {
		// Secondary-mode management endpoints (called by the primary to sync bind config/zones)
		e.PUT("/bind/config", PutBindConfigHandler)
		e.PUT("/bind/zones/:domain", PutZoneFileHandler)
		e.DELETE("/bind/zones/:domain", DeleteZoneFileHandler)
		e.POST("/bind/reload", PostBindReloadHandler)
		e.POST("/bind/refresh/:domain", PostZoneRefreshHandler)
	}

	return e
}

func startServer(e *echo.Echo) {
	if config.HTTPS {
		e.AutoTLSManager.Cache = autocert.DirCache("/var/cache/dnsapi/autocert")
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(config.PublicDomain)
		e.AutoTLSManager.Prompt = autocert.AcceptTOS
		e.AutoTLSManager.Email = config.ACMEEmail
		e.Logger.Print(fmt.Sprintf("https://%s", config.PublicDomain))
		// HTTP-01 challenge listener on :80
		go func() {
			if err := http.ListenAndServe(":80", e.AutoTLSManager.HTTPHandler(nil)); err != nil {
				e.Logger.Errorf("http challenge server: %v", err)
			}
		}()
		e.Logger.Fatal(e.StartAutoTLS(":443"))
		return
	}
	addr := ":" + strconv.Itoa(int(config.Port))
	e.Logger.Print(fmt.Sprintf("http://localhost%s", addr))
	e.Logger.Fatal(e.Start(addr))
}

package dnsapi

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

// ZoneNSStatus is the NS-reachability result for one nameserver.
type ZoneNSStatus struct {
	IP         string `json:"ip"`
	Configured bool   `json:"configured"`
	Error      string `json:"error,omitempty"`
}

// ZoneStatusResponse is returned by GET /zones/:zone_id/status.
type ZoneStatusResponse struct {
	ZoneID          uint           `json:"zone_id"`
	Domain          string         `json:"domain"`
	DBSerial        string         `json:"db_serial"`
	CommittedSerial string         `json:"committed_serial"`
	NeedsCommit     bool           `json:"needs_commit"`
	Primary         ZoneNSStatus   `json:"primary"`
	Secondary       []ZoneNSStatus `json:"secondary"`
	AllConfigured   bool           `json:"all_configured"`
}

// checkNSConfigured returns true if nsIP serves authoritative NS records for domain.
// We reuse queryNS with type "NS" — if the zone is configured on that server,
// it will return its own NS records (or at least not error with NXDOMAIN).
func checkNSConfigured(ctx context.Context, nsIP, domain string) (bool, error) {
	vals, supported, err := queryNS(ctx, nsIP, domain, "NS")
	if !supported || err != nil {
		return false, err
	}
	return len(vals) > 0, nil
}

// @Summary Zone status
// @Description Check whether the zone has uncommitted changes and whether it is configured on all nameservers.
// @ID zone-status
// @Tags zones
// @Security ApiKeyAuth
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Success 200 {object} ZoneStatusResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /zones/{zone_id}/status [get]
func ZoneStatusHandler(c echo.Context) error {
	zoneId, err := parseUintParam(c, "zone_id")
	if err != nil {
		return err
	}

	db := GetDatabaseConnection()
	var zone Zone
	err = db.Where("id = ?", zoneId).Find(&zone).Error
	if err != nil || zone.ID == 0 {
		return c.JSONPretty(http.StatusNotFound, ErrorResponse{Message: "zone not found"}, "  ")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp := ZoneStatusResponse{
		ZoneID:          zone.ID,
		Domain:          zone.Domain,
		DBSerial:        zone.Serial,
		CommittedSerial: zone.CommittedSerial,
		NeedsCommit:     zone.Serial != zone.CommittedSerial,
		Secondary:       []ZoneNSStatus{},
	}

	// Check primary NS
	primaryOK, primaryErr := checkNSConfigured(ctx, config.PrimaryNameServerIP, zone.Domain)
	resp.Primary = ZoneNSStatus{
		IP:         config.PrimaryNameServerIP,
		Configured: primaryOK,
	}
	if primaryErr != nil {
		resp.Primary.Error = primaryErr.Error()
	}

	// Check each secondary NS
	allConfigured := primaryOK
	for _, secIP := range config.SecondaryNameServerIPs {
		secOK, secErr := checkNSConfigured(ctx, secIP, zone.Domain)
		ns := ZoneNSStatus{
			IP:         secIP,
			Configured: secOK,
		}
		if secErr != nil {
			ns.Error = secErr.Error()
			allConfigured = false
		} else if !secOK {
			allConfigured = false
		}
		resp.Secondary = append(resp.Secondary, ns)
	}
	resp.AllConfigured = allConfigured

	return c.JSONPretty(http.StatusOK, resp, "  ")
}

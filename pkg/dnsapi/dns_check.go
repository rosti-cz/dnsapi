package dnsapi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
)

// DNSCheckRecord holds the comparison result for one zone record.
type DNSCheckRecord struct {
	RecordID        uint     `json:"record_id"`
	Name            string   `json:"name"`
	FQDN            string   `json:"fqdn"`
	Type            string   `json:"type"`
	DBValue         string   `json:"db_value"`
	PrimaryValues   []string `json:"primary_values"`
	SecondaryValues []string `json:"secondary_values"`
	PrimaryOK       bool     `json:"primary_ok"`
	SecondaryOK     bool     `json:"secondary_ok"`
	Supported       bool     `json:"supported"`
	Error           string   `json:"error,omitempty"`
}

// DNSCheckResponse is the response body for TestZoneHandler.
type DNSCheckResponse struct {
	ZoneID      uint             `json:"zone_id"`
	Domain      string           `json:"domain"`
	PrimaryNS   string           `json:"primary_ns"`
	SecondaryNS []string         `json:"secondary_ns"`
	Records     []DNSCheckRecord `json:"records"`
}

func buildFQDN(name, domain string) string {
	if name == "@" || name == "" {
		return domain
	}
	return name + "." + domain
}

// queryNS queries recordType records for fqdn from the given nameserver IP.
// Returns (values, supported, error). supported=false means the type is not handled.
func queryNS(ctx context.Context, nsIP, fqdn, recordType string) ([]string, bool, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", nsIP+":53")
		},
	}

	switch strings.ToUpper(recordType) {
	case "A":
		addrs, err := r.LookupIPAddr(ctx, fqdn)
		if err != nil {
			return nil, true, err
		}
		var out []string
		for _, a := range addrs {
			if a.IP.To4() != nil {
				out = append(out, a.IP.String())
			}
		}
		return out, true, nil

	case "AAAA":
		addrs, err := r.LookupIPAddr(ctx, fqdn)
		if err != nil {
			return nil, true, err
		}
		var out []string
		for _, a := range addrs {
			if a.IP.To4() == nil {
				out = append(out, a.IP.String())
			}
		}
		return out, true, nil

	case "MX":
		mxs, err := r.LookupMX(ctx, fqdn)
		if err != nil {
			return nil, true, err
		}
		var out []string
		for _, mx := range mxs {
			out = append(out, fmt.Sprintf("%d %s", mx.Pref, strings.TrimRight(mx.Host, ".")))
		}
		return out, true, nil

	case "TXT":
		txts, err := r.LookupTXT(ctx, fqdn)
		if err != nil {
			return nil, true, err
		}
		return txts, true, nil

	case "NS":
		nss, err := r.LookupNS(ctx, fqdn)
		if err != nil {
			return nil, true, err
		}
		var out []string
		for _, ns := range nss {
			out = append(out, strings.TrimRight(ns.Host, "."))
		}
		return out, true, nil

	case "CNAME":
		cname, err := r.LookupCNAME(ctx, fqdn)
		if err != nil {
			return nil, true, err
		}
		return []string{strings.TrimRight(cname, ".")}, true, nil

	default:
		return nil, false, nil
	}
}

func dnsContains(slice []string, val string) bool {
	val = strings.ToLower(strings.TrimRight(val, "."))
	for _, s := range slice {
		if strings.ToLower(strings.TrimRight(s, ".")) == val {
			return true
		}
	}
	return false
}

// @Summary Test zone DNS
// @Description Compare zone records in the database against what the primary and secondary nameservers currently serve.
// @ID test-zone
// @Tags zones
// @Security ApiKeyAuth
// @Produce json
// @Param zone_id path int true "Zone ID"
// @Success 200 {object} DNSCheckResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /zones/{zone_id}/test [get]
func TestZoneHandler(c echo.Context) error {
	zoneId, err := parseUintParam(c, "zone_id")
	if err != nil {
		return err
	}

	db := GetDatabaseConnection()
	var zone Zone
	err = db.Where("id = ?", zoneId).Preload("Records").Find(&zone).Error
	if err != nil || zone.ID == 0 {
		return c.JSONPretty(http.StatusNotFound, ErrorResponse{Message: "zone not found"}, "  ")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp := DNSCheckResponse{
		ZoneID:      zone.ID,
		Domain:      zone.Domain,
		PrimaryNS:   config.PrimaryNameServerIP,
		SecondaryNS: config.SecondaryNameServerIPs,
		Records:     []DNSCheckRecord{},
	}

	for _, record := range zone.Records {
		fqdn := buildFQDN(record.Name, zone.Domain)

		dbValue := record.Value
		if strings.ToUpper(record.Type) == "MX" {
			dbValue = fmt.Sprintf("%d %s", record.Prio, record.Value)
		}

		r := DNSCheckRecord{
			RecordID:        record.ID,
			Name:            record.Name,
			FQDN:            fqdn,
			Type:            record.Type,
			DBValue:         dbValue,
			PrimaryValues:   []string{},
			SecondaryValues: []string{},
		}

		primaryVals, supported, primaryErr := queryNS(ctx, config.PrimaryNameServerIP, fqdn, record.Type)
		r.Supported = supported
		if !supported {
			resp.Records = append(resp.Records, r)
			continue
		}
		if primaryErr != nil {
			r.Error = "primary: " + primaryErr.Error()
		} else {
			r.PrimaryValues = primaryVals
			r.PrimaryOK = dnsContains(primaryVals, dbValue)
		}

		allSecOK := true
		for _, secIP := range config.SecondaryNameServerIPs {
			vals, _, secErr := queryNS(ctx, secIP, fqdn, record.Type)
			if secErr != nil {
				allSecOK = false
				if r.Error == "" {
					r.Error = "secondary " + secIP + ": " + secErr.Error()
				}
			} else {
				r.SecondaryValues = append(r.SecondaryValues, vals...)
				if !dnsContains(vals, dbValue) {
					allSecOK = false
				}
			}
		}
		if len(config.SecondaryNameServerIPs) > 0 {
			r.SecondaryOK = allSecOK
		} else {
			r.SecondaryOK = true
		}

		resp.Records = append(resp.Records, r)
	}

	return c.JSONPretty(http.StatusOK, resp, "  ")
}

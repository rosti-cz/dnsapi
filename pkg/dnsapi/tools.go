package dnsapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// --- Local file operations ---

// WriteZoneFile writes a zone file to the bind zone directory.
func WriteZoneFile(domain, content string) error {
	return os.WriteFile(path.Join(PrimaryZonePath, domain+".zone"), []byte(content), 0644)
}

// DeleteZoneFile removes a zone file. Returns nil if the file does not exist.
func DeleteZoneFile(domain string) error {
	err := os.Remove(path.Join(PrimaryZonePath, domain+".zone"))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// WriteBindConfig writes content to a named.conf include file at configPath.
func WriteBindConfig(configPath, content string) error {
	return os.WriteFile(configPath, []byte(content), 0644)
}

// ReloadBind reloads bind9 using the configured reload command.
func ReloadBind() error {
	parts := strings.Fields(config.BindReloadCommand)
	if len(parts) == 0 {
		return fmt.Errorf("DNSAPI_BIND_RELOAD_COMMAND is empty")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %w: %s", config.BindReloadCommand, err, out)
	}
	return nil
}

// RefreshZone forces a zone refresh using the configured refresh command.
func RefreshZone(domain string) error {
	parts := strings.Fields(config.BindRefreshCommand)
	if len(parts) == 0 {
		return fmt.Errorf("DNSAPI_BIND_REFRESH_COMMAND is empty")
	}
	parts = append(parts, domain)
	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w: %s", config.BindRefreshCommand, domain, err, out)
	}
	return nil
}

// --- HTTP helpers for primary→secondary sync ---

// callSecondary performs an HTTP request against a secondary dnsapi instance.
func callSecondary(method, secondaryURL, urlPath, token string, body []byte) error {
	fullURL := strings.TrimRight(secondaryURL, "/") + urlPath
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	if body != nil {
		req.Header.Set("Content-Type", "text/plain")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("secondary %s %s returned %d: %s", method, fullURL, resp.StatusCode, string(b))
	}
	return nil
}

// SyncBindConfigToSecondary pushes a rendered bind config to a secondary instance.
func SyncBindConfigToSecondary(secondaryURL, token, content string) error {
	return callSecondary("PUT", secondaryURL, "/bind/config", token, []byte(content))
}

// SyncZoneToSecondary pushes a zone file to a secondary instance.
func SyncZoneToSecondary(secondaryURL, token, domain, content string) error {
	return callSecondary("PUT", secondaryURL, "/bind/zones/"+domain, token, []byte(content))
}

// DeleteZoneOnSecondary removes a zone file on a secondary instance.
func DeleteZoneOnSecondary(secondaryURL, token, domain string) error {
	return callSecondary("DELETE", secondaryURL, "/bind/zones/"+domain, token, nil)
}

// RefreshZoneOnSecondary asks a secondary instance to run the refresh command for the given zone.
func RefreshZoneOnSecondary(secondaryURL, token, domain string) error {
	return callSecondary("POST", secondaryURL, "/bind/refresh/"+domain, token, nil)
}

// --- Orchestration helpers (primary mode) ---

// SetSlavesBindConfig generates the secondary bind config from the DB and pushes it to all secondary instances.
func SetSlavesBindConfig() {
	defer func() {
		if r := recover(); r != nil {
			captureRecoveredPanic(r)
			log.Errorf("%v", r)
		}
	}()

	var zones []Zone
	db := GetDatabaseConnection()
	if err := db.Find(&zones).Error; err != nil {
		log.Errorf("SetSlavesBindConfig: db: %v", err)
		return
	}

	var buf strings.Builder
	for _, zone := range zones {
		buf.WriteString(zone.RenderSecondary())
		buf.WriteString("\n")
	}
	bindConfig := buf.String()

	for _, secondary := range config.SecondaryInstanceList() {
		go func(url string) {
			defer func() {
				if r := recover(); r != nil {
					captureRecoveredPanic(r)
					log.Errorf("%v", r)
				}
			}()
			if err := SyncBindConfigToSecondary(url, config.APIToken, bindConfig); err != nil {
				log.Errorf("SyncBindConfigToSecondary(%s): %v", url, err)
			}
		}(secondary)
	}
}

// SetMasterBindConfig generates the primary bind config from the DB, writes it locally, then reloads bind.
func SetMasterBindConfig() {
	defer func() {
		if r := recover(); r != nil {
			captureRecoveredPanic(r)
			log.Errorf("%v", r)
		}
	}()

	var zones []Zone
	db := GetDatabaseConnection()
	if err := db.Find(&zones).Error; err != nil {
		log.Errorf("SetMasterBindConfig: db: %v", err)
		return
	}

	var buf strings.Builder
	for _, zone := range zones {
		buf.WriteString(zone.RenderPrimary())
		buf.WriteString("\n")
	}

	if err := WriteBindConfig(PrimaryBindConfigPath, buf.String()); err != nil {
		log.Errorf("SetMasterBindConfig: write config: %v", err)
		return
	}
	if err := ReloadBind(); err != nil {
		log.Errorf("SetMasterBindConfig: reload bind: %v", err)
	}
}

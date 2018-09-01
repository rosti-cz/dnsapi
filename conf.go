package main

import (
	"github.com/pkg/errors"
	"strings"
)

const (
	// Where zones are saved in bind's config directory
	PrimaryZonePath = "/var/cache/bind"
	// Where bind's configuration is saved in bind's directory (master)
	PrimaryBindConfigPath = "/etc/bind/named.conf.rosti"
	// Where bind's configuration is saved in bind's directory (slave)
	SecondaryBindConfigPath = "/etc/bind/named.conf.rosti"

	RECORD_NOT_FOUND_MESSAGE = "record not found"
)

// Configuration struct. All input form the maintainer is available through this struct.
type Config struct {
	PrimaryNameServerIP string `split_words:"true"` // If not set, automatically resolved from PrimaryNameServer
	// TODO: in the following field the split_words doesn't work as expected
	SecondaryNameServerIPs []string ``                                         // If not set, automatically resolved from NameServers
	PrimaryNameServer      string   `split_words:"true"`                       // Bind's master server
	NameServers            []string `split_words:"true"`                       // Bind's slave servers
	AbuseEmail             string   `split_words:"true"`                       // Abuse email
	TimeToRefresh          int      `default:"300" split_words:"true"`         // Time to refresh the records on slaves
	TimeToRetry            int      `default:"180" split_words:"true"`         // Time to waif for another try if connection fails
	TimeToExpire           int      `default:"604800" split_words:"true"`      // Time to expire when the domain is not available on master
	MinimalTTL             int      `default:"30" split_words:"true"`          // Minimal TTL
	TTL                    int      `default:"3600"`                           // Default TTL
	DatabasePath           string   `default:"gorm.sqlite" split_words:"true"` // Path to the database
	SSHKey                 string   `split_words:"yes"`                        // SSH key used for set Bind's config files (path to file)
	SSHUser                string   `default:"root" split_words:"yes"`         // SSH user used for saving config files
	APIToken               string   `default:"" split_words:"yes"`             // Token to access the API
	Port                   uint16   `default:"1323"`                           // Port where the API listens
}

// Validates data inside the config struct
func (c *Config) Validate() error {
	if c.PrimaryNameServer == "" {
		return errors.New("DNSAPI_PRIMARY_NAME_SERVER has to be defined")
	}
	if len(c.NameServers) < 2 {
		return errors.New("DNSAPI_NAME_SERVERS has to be defined and contains at least two servers")
	}
	if c.AbuseEmail == "" || !strings.Contains(c.AbuseEmail, "@") || !strings.Contains(c.AbuseEmail, ".") {
		return errors.New("DNSAPI_ABUSE_EMAIL has to be defined and contains a valid email address")
	}

	return nil
}

// Reformat the email so it can be used in zone files
func (c *Config) RenderEmail() string {
	return strings.Replace(c.AbuseEmail, "@", ".", -1)
}

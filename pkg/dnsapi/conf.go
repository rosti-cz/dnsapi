package dnsapi

import (
	"strings"

	"github.com/pkg/errors"
)

const (
	RuntimeModePrimary   = "primary"
	RuntimeModeSecondary = "secondary"
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
	Mode                   string   `default:"primary" split_words:"true"`                // Runtime mode: primary or secondary
	PrimaryNameServerIP    string   `split_words:"true"`                                  // If not set, automatically resolved from PrimaryNameServer
	SecondaryNameServerIPs []string `envconfig:"SECONDARY_NAME_SERVER_IPS"`               // If not set, automatically resolved from NameServers
	SecondaryInstances     []string `split_words:"true"`                                  // Secondary dnsapi instances for sync fanout
	PrimaryNameServer      string   `split_words:"true"`                                  // Bind's master server
	NameServers            []string `split_words:"true"`                                  // Bind's slave servers
	AbuseEmail             string   `split_words:"true"`                                  // Abuse email
	TimeToRefresh          int      `default:"300" split_words:"true"`                    // Time to refresh the records on slaves
	TimeToRetry            int      `default:"180" split_words:"true"`                    // Time to waif for another try if connection fails
	TimeToExpire           int      `default:"604800" split_words:"true"`                 // Time to expire when the domain is not available on master
	MinimalTTL             int      `default:"30" split_words:"true"`                     // Minimal TTL
	TTL                    int      `default:"3600"`                                      // Default TTL
	DatabasePath           string   `default:"gorm.sqlite" split_words:"true"`            // Path to the database
	BindReloadCommand      string   `default:"systemctl reload bind9" split_words:"true"` // Command used to reload bind9
	BindRefreshCommand     string   `default:"rndc refresh" split_words:"true"`           // Command used to refresh a zone (domain appended as last argument)
	APIToken               string   `default:"" envconfig:"API_TOKEN"`                    // Token to access the API
	PublicDomain           string   `split_words:"true"`                                  // Public domain for this instance
	HTTPS                  bool     `split_words:"true"`                                  // Serve HTTPS and obtain Let's Encrypt certs
	ACMEEmail              string   `split_words:"true"`                                  // Email for Let's Encrypt
	SentryDSN              string   `split_words:"true"`                                  // Sentry DSN
	SentryENV              string   `default:"dev" split_words:"true"`                    // Sentry environment
	Port                   uint16   `default:"1323"`                                      // Port where the API listens
}

// Validates data inside the config struct
func (c *Config) Validate() error {
	switch c.Mode {
	case "", RuntimeModePrimary, RuntimeModeSecondary:
	default:
		return errors.New("DNSAPI_MODE has to be primary or secondary")
	}

	if c.IsPrimaryMode() {
		if c.PrimaryNameServer == "" {
			return errors.New("DNSAPI_PRIMARY_NAME_SERVER has to be defined")
		}
		if len(c.NameServers) < 2 {
			return errors.New("DNSAPI_NAME_SERVERS has to be defined and contains at least two servers")
		}
		if c.AbuseEmail == "" || !strings.Contains(c.AbuseEmail, "@") || !strings.Contains(c.AbuseEmail, ".") {
			return errors.New("DNSAPI_ABUSE_EMAIL has to be defined and contains a valid email address")
		}
	}

	if c.HTTPS && c.PublicDomain == "" {
		return errors.New("DNSAPI_PUBLIC_DOMAIN has to be defined when HTTPS is enabled")
	}
	if c.HTTPS && c.ACMEEmail == "" {
		return errors.New("DNSAPI_ACME_EMAIL has to be defined when HTTPS is enabled")
	}

	return nil
}

func (c *Config) IsPrimaryMode() bool {
	return c.Mode == "" || c.Mode == RuntimeModePrimary
}

func (c *Config) IsSecondaryMode() bool {
	return c.Mode == RuntimeModeSecondary
}

func (c *Config) SecondaryInstanceList() []string {
	instances := make([]string, 0, len(c.SecondaryInstances))
	seen := make(map[string]struct{}, len(c.SecondaryInstances))

	for _, instance := range c.SecondaryInstances {
		instance = strings.TrimSpace(instance)
		if instance == "" {
			continue
		}
		if _, ok := seen[instance]; ok {
			continue
		}
		seen[instance] = struct{}{}
		instances = append(instances, instance)
	}

	return instances
}

// Reformat the email so it can be used in zone files
func (c *Config) RenderEmail() string {
	return strings.Replace(c.AbuseEmail, "@", ".", -1)
}

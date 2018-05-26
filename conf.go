package main

import (
	"github.com/pkg/errors"
	"strings"
)

type Config struct {
	PrimaryNameServerIP string   `split_words:"true"`
	SecondaryNameServerIPs []string   `split_words:"true"`
	PrimaryNameServer string   `split_words:"true"`
	NameServers       []string `split_words:"true"`
	AbuseEmail        string   `split_words:"true"`
	TimeToRefresh     int      `default:"300" split_words:"true"`
	TimeToRetry       int      `default:"180" split_words:"true"`
	TimeToExpire      int      `default:"604800" split_words:"true"`
	MinimalTTL        int      `default:"30" split_words:"true"`
	TTL               int      `default:"3600"`
	DatabasePath      string   `default:"gorm.sqlite"`
}

func (c *Config) Validate() error {
	if c.PrimaryNameServer == "" {
		return errors.New("DNSAPI_PRIMARY_NAME_SERVER has to be defined")
	}
	if len(c.NameServers) < 2 {
		return errors.New("DNSAPI_NAME_SERVER has to be defined and contains at least two servers")
	}
	if c.AbuseEmail == "" || !strings.Contains(c.AbuseEmail, "@") || !strings.Contains(c.AbuseEmail, ".") {
		return errors.New("DNSAPI_ABUSE_EMAIL has to be defined and contains a valid email address")
	}

	if c.PrimaryNameServerIP == "" {
		return errors.New("DNSAPI_PRIMARY_NAME_SERVER_IP has to be defined")
	}
	if len(c.SecondaryNameServerIPs) == 0 {
		return errors.New("DNSAPI_SECONDARY_NAME_SERVER_IPS has to be defined and contains at least one server")
	}

	return nil
}

func (c *Config) RenderEmail() string {
	return strings.Replace(c.AbuseEmail, "@", ".", -1)
}

package main

import (
	"github.com/kelseyhightower/envconfig"
	"log"
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

	var zone Zone
	zone.AddRecord("@", 300, "A", 0, "1.2.3.4")
	zone.AddRecord("@", 300, "AAAA", 0, "2001::2")
	zone.AddRecord("www", 300, "CNAME", 0, "@")
	zone.AddRecord("@", 300, "MX", 10, "mail.rosti.cz.")

	errs := zone.Validate()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
		}
	} else {
		log.Println(zone.Render())
	}
}

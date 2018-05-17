package main

import (
	"strconv"
	"github.com/pkg/errors"
	"regexp"
	"net"
	"strings"
)

// Record struct

type Record struct {
	Name  string
	TTL   int
	Type  string // A, AAAA, CNAME, TXT, SRV
	Prio  int
	Value string
}

// Validates the record
func (r *Record) Validate() error {
	if r.TTL < 60 && r.TTL > 2592000 {
		return errors.New(r.Name + ": TTL has to be number between 60 and 2592000")
	}

	if r.Type == "A" {
		parsed := net.ParseIP(r.Value)

		if parsed == nil || !strings.Contains(r.Value, ".") {
			return errors.New(r.Name + ": IP address of A record is not valid")
		}
	} else if r.Type == "AAAA" {
		parsed := net.ParseIP(r.Value)

		if parsed == nil || !strings.Contains(r.Value, ":") {
			return errors.New(r.Name + ": IP address of AAAA record is not valid")
		}
	} else if r.Type == "CNAME" {
		matched, err := regexp.MatchString(`[a-z\.0-9]{3,64}`, r.Value)
		if err != nil {
			panic(err)
		}
		if !matched {
			return errors.New(r.Name + ": CNAME has not a valid value")
		}
	} else if r.Type == "TXT" {
		if strings.Contains(r.Value, "\"") || strings.Contains(r.Value, "'") || strings.Contains(r.Value, "`") {
			return errors.New(r.Name + ": characters \"' or ` are not allowed in TXT records")
		}
		// TODO: just check if there are not unsupported characters (" or `)
	} else if r.Type == "SRV" {
	} else if r.Type == "MX" {
		if r.Prio <= 0 && r.Prio <= 100 {
			return errors.New(r.Name + ": Prio has to be bigger than 0 and smaller than 100")
		}
		//TODO: Has to be domain and valid A/AAAA record (even in different location)
	} else {
		return errors.New("Unknown record type")
	}

	return nil
}

// Renders one record
func (r *Record) Render() string {
	if r.Type == "TXT" {
		// TODO: split the value
	}

	// TODO: prio only in case of MX record
	return r.Name + "    " +
		strconv.Itoa(r.TTL) + "s    " +
		r.Type + "    " +
		strconv.Itoa(r.Prio) + "    " +
		r.Value
}

// Zone struct

type Zone struct {
	Serial  string
	Records []Record
	Tags    []string
}

// Validates records in the zone
func (z *Zone) Validate() []error {
	var errorsMsgs []error
	var usedNames []string

	for _, record := range z.Records {
		err := record.Validate()
		if err != nil {
			errorsMsgs = append(errorsMsgs, err)
		}

		if record.Type == "A" || record.Type == "AAAA" || record.Type == "CNAME" {
			usedNames = append(usedNames, record.Name)
		}
	}

	// Additional checks

	// CNAME record can't have same name as another AAAA record, A record or CNAME record
	for _, record := range z.Records {
		if record.Type == "CNAME" {
			count := 0
			for _, usedName := range usedNames {
				if usedName == record.Name {
					count += 1
				}
			}
			if count > 1 {
				errorsMsgs = append(errorsMsgs, errors.New(record.Name + " is already used in another A/AAAA/CNAME record"))
			}
		}
	}


	return errorsMsgs
}

// Renders whole zone
func (z *Zone) Render() string {
	var zone string

	/*
		@     IN     SOA    <primary-name-server>	<hostmaster-email> (
		<serial-number>
		<time-to-refresh>
		<time-to-retry>
		<time-to-expire>
		<minimum-TTL> )
	*/

	zone = `$TTL 86400s
@       IN      SOA     ns1.rosti.cz. abuse.rosti.cz.  (
		` + z.Serial + `
		300
		180
		604800
		30
)

	IN	NS	ns1.rosti.cz.
	IN	NS	ns2.rosti.cz.

`

	for _, record := range z.Records {
		zone += record.Render()
		zone += "\n"
	}

	return zone
}

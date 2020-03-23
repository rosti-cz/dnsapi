package main

import (
	"path"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
)

// Create a new zone
func NewZone(domain string, tags []string, abuseEmail string) (*Zone, []error) {
	zone := Zone{
		Domain:     strings.ToLower(domain),
		Tags:       strings.Join(tags, ","),
		AbuseEmail: abuseEmail,
		Delete:     false,
	}

	errs := zone.Validate()
	if len(errs) > 0 {
		return &zone, errs
	}

	db := GetDatabaseConnection()
	db.NewRecord(&zone)
	err := db.Create(&zone).Error
	if err != nil {
		return &zone, []error{err}
	}

	return &zone, nil
}

// Update existing zone
func UpdateZone(zoneId uint, tags []string, abuseEmail string) (*Zone, []error) {
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", zoneId).Find(&zone).Error
	if err != nil {
		return nil, []error{err}
	}

	zone.Tags = strings.Join(tags, ",")
	zone.AbuseEmail = abuseEmail

	errs := zone.Validate()
	if len(errs) > 0 {
		return nil, errs
	}

	err = db.Model(&zone).Update("tags", zone.Tags).
		Update("abuse_email", zone.AbuseEmail).
		Update("serial", zone.Serial).Error
	if err != nil {
		return nil, []error{err}
	}

	err = db.Where("id = ?", zoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		return nil, []error{err}
	}

	return &zone, nil
}

// Delete existing zone
func DeleteZone(zoneId uint) error {
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", zoneId).Find(&zone).Error
	if err != nil {
		return err
	}

	tx := db.Begin()

	err = tx.Where("zone_id = ?", zoneId).Delete(&Record{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Where("id = ?", zoneId).Delete(&zone).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}

	// Delete the zone file
	_, err = SendCommandViaSSH(config.PrimaryNameServerIP, "rm -f "+path.Join(PrimaryZonePath, zone.Domain+".zone"))
	if err != nil {
		panic(err)
	}

	go SetSlavesBindConfig()
	go SetMasterBindConfig()

	return nil
}

// Create a new record
func NewRecord(zoneId uint, name string, ttl int, recordType string, prio int, value string) (*Record, []error) {
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", zoneId).Find(&zone).Error
	if err != nil {
		return nil, []error{err}
	}

	record, errs := zone.AddRecord(name, ttl, recordType, prio, value)
	if len(errs) > 0 {
		return nil, errs
	}

	err = db.Create(record).Error
	if err != nil {
		return record, []error{err}
	}

	return record, nil
}

// UpdateRecord updates existing record
func UpdateRecord(recordId uint, name string, ttl int, prio int, value string) (*Record, []error) {
	var record Record = Record{}
	var zone Zone = Zone{}

	db := GetDatabaseConnection()
	err := db.Where("id = ?", recordId).Find(&record).Error
	if err != nil {
		return nil, []error{err}
	}

	err = db.Where("id = ?", record.ZoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		return nil, []error{err}
	}

	for _, recordTmp := range zone.Records {
		if recordTmp.ID == recordId {
			record = recordTmp
		}
	}
	if record.ID == 0 {
		panic(errors.New("record not found"))
	}

	record.Name = name
	record.TTL = ttl
	record.Prio = prio
	record.Value = value

	errs := zone.Validate()
	if len(errs) > 0 {
		return nil, errs
	}

	tx := db.Begin()
	err = tx.Model(&zone).Update("serial", zone.Serial).Error
	if err != nil {
		tx.Rollback()
		return nil, []error{err}
	}
	err = tx.Model(&record).Update("name", name).
		Update("ttl", ttl).
		Update("prio", prio).
		Update("value", value).Error
	if err != nil {
		tx.Rollback()
		return nil, []error{err}
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, []error{err}
	}

	err = db.Where("id = ?", recordId).Find(&record).Error
	if err != nil {
		return nil, []error{err}
	}

	return &record, nil
}

// Delete existing record
func DeleteRecord(recordId uint) error {
	db := GetDatabaseConnection()

	err := db.Where("id = ?", recordId).Delete(&Record{}).Error
	if err != nil {
		return err
	}

	return nil
}

// Write new zone into DNS servers
// TODO: here is a lot of SSH stuff we can do in parallel
func Commit(zoneId uint) error {
	var zones []Zone // all zones
	var zone Zone    // updating zone

	// Get the committing zone and all zones from db
	db := GetDatabaseConnection()
	err := db.Model(&zone).Where("id = ?", zoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("Zone not found")
		}
		return err
	}

	err = db.Find(&zones).Error
	if err != nil {
		return err
	}

	// Set new serial
	zone.SetNewSerial()
	err = db.Model(&zone).Update("serial", zone.Serial).Error
	if err != nil {
		return err
	}

	// Generate all config files for bind
	var allZonesPrimaryConfig string
	var allZonesSecondaryConfig string
	for _, zone := range zones {
		allZonesPrimaryConfig += zone.RenderPrimary()
		allZonesPrimaryConfig += "\n"

		allZonesSecondaryConfig += zone.RenderSecondary()
		allZonesSecondaryConfig += "\n"
	}

	// Save slaves' main config
	go SetSlavesBindConfig()

	go func(zone *Zone, IP string, bindConfig string) {
		// This is called as goroutine so we need to recover from panicing
		defer func() {
			// TODO: implement sentry here
			if r := recover(); r != nil {
				log.Errorf(r.(error).Error())
			}
		}()
		// Save zone file
		err = SendFileViaSSH(IP, path.Join(PrimaryZonePath, zone.Domain+".zone"), zone.Render())
		if err != nil {
			panic(err)
		}

		SetMasterBindConfig()
	}(&zone, config.PrimaryNameServer, allZonesPrimaryConfig)

	// Force zone refresh a few moments after everything is done
	go func(config *Config, zone *Zone) {
		// This is called as goroutine so we need to recover from panicing
		defer func() {
			// TODO: implement sentry here
			if r := recover(); r != nil {
				log.Errorf(r.(error).Error())
			}
		}()
		// Wait for 10 second to settle things up
		time.Sleep(10 * time.Second)

		// When reload is done, force to refresh
		for _, server := range config.SecondaryNameServerIPs {
			_, err = SendCommandViaSSH(server, "rndc refresh "+zone.Domain)
			if err != nil {
				panic(err)
			}
		}
	}(&config, &zone)

	return nil
}

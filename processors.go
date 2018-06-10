package main

import (
	"strings"
	"path"
	"time"
)

// Create a new zone
func NewZone(domain string, tags []string, abuseEmail string) (*Zone, []error) {
	zone := Zone{
		Domain: domain,
		Tags: strings.Join(tags, ","),
		AbuseEmail: abuseEmail,
	}
	zone.SetNewSerial()

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
func UpdateZone(zoneId uint, tags[]string, abuseEmail string) (*Zone, []error) {
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", zoneId).Find(&zone).Error
	if err != nil {
		return nil, []error{err}
	}

	zone.Tags = strings.Join(tags, ",")
	zone.AbuseEmail = abuseEmail
	zone.SetNewSerial()

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
	db := GetDatabaseConnection()
	tx := db.Begin()

	err := tx.Where("zone_id = ?", zoneId).Delete(&Record{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Where("id = ?", zoneId).Delete(&Zone{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}


// Create a new record
func NewRecord(zoneId uint, name string, ttl int, recordType string, prio int, value string) (*Record, []error) {
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", zoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		return nil, []error{err}
	}

	record, errs := zone.AddRecord(name, ttl, recordType, prio, value)
	if len(errs) > 0 {
		return nil, errs
	}

	zone.SetNewSerial()
	tx := db.Begin()
	err = tx.Model(&zone).Update("serial", zone.Serial).Error
	if err != nil {
		tx.Rollback()
		return nil, []error{err}
	}

	err = tx.Create(record).Error
	if err != nil {
		tx.Rollback()
		return record, []error{err}
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, []error{err}
	}

	return record, nil
}

// Update existing record
func UpdateRecord(recordId uint, name string, ttl int, prio int, value string) (*Record, []error) {
	var record Record
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", recordId).Find(&record).Error
	if err != nil {
		return nil, []error{err}
	}

	err = db.Where("id = ?", record.ZoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		return nil, []error{err}
	}

	zone.SetNewSerial()
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
	var zone Zone // updating zone

	db := GetDatabaseConnection()
	err := db.Model(&zone).Where("id = ?", zoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		return err
	}

	err = db.Find(&zones).Error
	if err != nil {
		return err
	}

	var allZonesPrimaryConfig string
	var allZonesSecondaryConfig string
	for _, zone := range zones {
		allZonesPrimaryConfig += zone.RenderPrimary()
		allZonesPrimaryConfig += "\n"

		allZonesSecondaryConfig += zone.RenderSecondary()
		allZonesSecondaryConfig += "\n"
	}

	// Save slaves' main config
	for _, server := range config.SecondaryNameServerIPs {
		err = SendFileViaSSH(server, SecondaryBindConfigPath, allZonesSecondaryConfig)
		if err != nil {
			return err
		}
		_, err = SendCommandViaSSH(server, "systemctl reload bind9")
		if err != nil {
			return err
		}
	}

	// Save zone file
	err = SendFileViaSSH(config.PrimaryNameServer, path.Join(PrimaryZonePath, zone.Domain + ".zone"), zone.Render())
	if err != nil {
		return err
	}
	// Save master's main config
	err = SendFileViaSSH(config.PrimaryNameServer, PrimaryBindConfigPath, allZonesPrimaryConfig)
	if err != nil {
		return err
	}
	_, err = SendCommandViaSSH(config.PrimaryNameServer, "systemctl reload bind9")
	if err != nil {
		return err
	}

	// Force zone refresh a few moments after everything is done
	go func (config *Config, zone *Zone) {
		// Wait for 10 second to settle things up
		time.Sleep(10*time.Second)

		// When reload is done, force to refresh
		for _, server := range config.SecondaryNameServerIPs {
			err = SendFileViaSSH(server, SecondaryBindConfigPath, allZonesSecondaryConfig)
			if err != nil {
				panic(err)
			}
			_, err = SendCommandViaSSH(server, "rndc refresh " + zone.Domain)
			if err != nil {
				panic(err)
			}
		}
	}(&config, &zone)


	return nil
}

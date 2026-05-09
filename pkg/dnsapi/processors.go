package dnsapi

import (
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
)

// Create a new zone
func NewZone(domain string, tags []string, abuseEmail string, owner string) (*Zone, []error) {
	zone := Zone{
		Domain:     strings.ToLower(domain),
		Tags:       strings.Join(tags, ","),
		AbuseEmail: abuseEmail,
		Owner:      owner,
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
func UpdateZone(zoneId uint, tags []string, abuseEmail string, owner string) (*Zone, []error) {
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", zoneId).Find(&zone).Error
	if err != nil {
		return nil, []error{err}
	}

	zone.Tags = strings.Join(tags, ",")
	zone.AbuseEmail = abuseEmail
	zone.Owner = owner

	errs := zone.Validate()
	if len(errs) > 0 {
		return nil, errs
	}

	err = db.Model(&zone).Update("tags", zone.Tags).
		Update("abuse_email", zone.AbuseEmail).
		Update("owner", zone.Owner).
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

	// Delete the zone file locally
	if err := DeleteZoneFile(zone.Domain); err != nil {
		log.Errorf("DeleteZone: delete zone file: %v", err)
	}

	// Delete the zone file on each secondary instance
	for _, secondary := range config.SecondaryInstanceList() {
		go func(url string) {
			if err := DeleteZoneOnSecondary(url, config.APIToken, zone.Domain); err != nil {
				log.Errorf("DeleteZoneOnSecondary(%s, %s): %v", url, zone.Domain, err)
			}
		}(secondary)
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
func UpdateRecord(zoneId uint, recordId uint, name string, ttl int, prio int, value string) (*Record, []error) {
	var record Record = Record{}
	var zone Zone = Zone{}

	db := GetDatabaseConnection()
	err := db.Where("id = ? AND zone_id = ?", recordId, zoneId).Find(&record).Error
	if err != nil {
		return nil, []error{err}
	}

	err = db.Where("id = ?", zoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		return nil, []error{err}
	}

	for _, recordTmp := range zone.Records {
		if recordTmp.ID == recordId {
			record = recordTmp
		}
	}
	if record.ID == 0 {
		return nil, []error{gorm.ErrRecordNotFound}
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
func DeleteRecord(zoneId uint, recordId uint) error {
	db := GetDatabaseConnection()

	result := db.Where("id = ? AND zone_id = ?", recordId, zoneId).Delete(&Record{})
	err := result.Error
	if err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// Write new zone into DNS servers
func Commit(zoneId uint) error {
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Model(&zone).Where("id = ?", zoneId).Preload("Records").Find(&zone).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("Zone not found")
		}
		return err
	}

	// Set new serial and mark as committed
	zone.SetNewSerial()
	err = db.Model(&zone).Update("serial", zone.Serial).Update("committed_serial", zone.Serial).Error
	if err != nil {
		return err
	}

	// Push secondary bind config to all secondary instances
	go SetSlavesBindConfig()

	// Write zone file locally and sync to secondaries, then rebuild primary bind config
	go func(zone *Zone) {
		defer func() {
			if r := recover(); r != nil {
				captureRecoveredPanic(r)
				log.Errorf("%v", r)
			}
		}()

		renderedZone := zone.Render()

		// Write zone file to local filesystem (primary bind reads this)
		if err := WriteZoneFile(zone.Domain, renderedZone); err != nil {
			log.Errorf("Commit: write zone file: %v", err)
			return
		}

		// Push zone file to each secondary instance
		for _, secondary := range config.SecondaryInstanceList() {
			go func(url string) {
				if err := SyncZoneToSecondary(url, config.APIToken, zone.Domain, renderedZone); err != nil {
					log.Errorf("SyncZoneToSecondary(%s, %s): %v", url, zone.Domain, err)
				}
			}(secondary)
		}

		SetMasterBindConfig()
	}(&zone)

	// Force zone refresh on primary and all secondaries after settling
	go func(cfg *Config, zone *Zone) {
		defer func() {
			if r := recover(); r != nil {
				captureRecoveredPanic(r)
				log.Errorf("%v", r)
			}
		}()
		time.Sleep(10 * time.Second)

		if err := RefreshZone(zone.Domain); err != nil {
			log.Errorf("RefreshZone(%s): %v", zone.Domain, err)
		}

		for _, secondary := range cfg.SecondaryInstanceList() {
			if err := RefreshZoneOnSecondary(secondary, cfg.APIToken, zone.Domain); err != nil {
				log.Errorf("RefreshZoneOnSecondary(%s, %s): %v", secondary, zone.Domain, err)
			}
		}
	}(&config, &zone)

	return nil
}

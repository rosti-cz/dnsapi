package main

import (
	"strings"
	"log"
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

	err = db.Where("id = ?", zoneId).Find(&zone).Error
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

// Update existing record
func UpdateRecord(recordId uint, name string, ttl int, prio int, value string) (*Record, []error) {
	var record Record
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", recordId).Find(&record).Error
	if err != nil {
		return nil, []error{err}
	}

	err = db.Where("id = ?", record.ZoneId).Find(&zone).Error
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
	tx.Commit()

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
func Commit(zoneId uint) error {
	var zones []Zone // all zones
	var zone Zone // updating zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", zoneId).Find(&zone).Error
	if err != nil {
		return err
	}

	err = db.Find(&zones).Error
	if err != nil {
		return err
	}

	var zoneConfig = zone.Render()
	var allZonesConfig string
	for _, zone := range zones {
		allZonesConfig += zone.Render()
	}

	// TODO: Save zone
	log.Println(zoneConfig)
	// TODO: Save named config
	log.Println(allZonesConfig)

	return nil
}

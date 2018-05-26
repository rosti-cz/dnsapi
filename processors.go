package main

import (
	"strings"
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
func UpdateZone(zoneId uint, tags[]string, abuseEmail string) []error {
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", zoneId).Find(&zone).Error
	if err != nil {
		return []error{err}
	}

	zone.Tags = strings.Join(tags, ",")
	zone.AbuseEmail = abuseEmail
	zone.SetNewSerial()

	errs := zone.Validate()
	if len(errs) > 0 {
		return errs
	}

	err = db.Update(&zone).Error
	if err != nil {
		return []error{err}
	}

	return nil
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
func UpdateRecord(recordId uint, name string, ttl int, prio int, value string) []error {
	var record Record
	var zone Zone

	db := GetDatabaseConnection()
	err := db.Where("id = ?", recordId).Find(&record).Error
	if err != nil {
		return []error{err}
	}

	err = db.Where("id = ?", record.ZoneId).Find(&zone).Error
	if err != nil {
		return []error{err}
	}

	zone.SetNewSerial()
	record.Name = name
	record.TTL = ttl
	record.Prio = prio
	record.Value = value

	errs := zone.Validate()
	if len(errs) > 0 {
		return errs
	}

	tx := db.Begin()
	err = tx.Update(&zone).Error
	if err != nil {
		tx.Rollback()
		return []error{err}
	}
	err = tx.Update(&record).Error
	if err != nil {
		tx.Rollback()
		return []error{err}
	}
	tx.Commit()

	return nil
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
func Commit(zoneId uint) {

}
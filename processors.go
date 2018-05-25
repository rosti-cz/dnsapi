package main

// Create a new zone
func NewZone(domain string, tags []string, abuseEmail string) {}

// Update existing zone
func UpdateZone(zoneId uint, tags[]string, abuseEmail string) {}

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
func NewRecord(zoneId uint, name string, ttl int, recordType string, prio int, value string) {}

// Update existing record
func UpdateRecord(recordId uint, name string, ttl int, prio int, value string) {}

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
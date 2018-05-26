package main

import (
	"testing"
	"os"
	"fmt"
)

const TEST_DOMAIN = "ohphiuhi.txt"
const TEST_ABUSE_EMAIL = "t@ohphiuhi.txt"


func TestMain(m *testing.M) {
	config.DatabasePath = "/tmp/dnsapi_test_database.sqlite"

	db := GetDatabaseConnection()
	defer db.Close()

	os.Exit(m.Run())

	err := os.Remove(config.DatabasePath)
	if err != nil {
		fmt.Println("Can't remove test database")
	}
}


func TestNewZone(t *testing.T) {
	_, errs := NewZone(TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
	if len(errs) > 0 {
		t.Error(errs)
	}

}

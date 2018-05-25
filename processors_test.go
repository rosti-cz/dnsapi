package main

import (
	"testing"
	"os"
)

const TEST_DOMAIN = "ohphiuhi.txt"
const TEST_ABUSE_EMAIL = "t@ohphiuhi.txt"


func TestMain(m *testing.M) {
	db := GetDatabaseConnection()
	defer db.Close()

	os.Exit(m.Run())
}


func TestNewZone(t *testing.T) {
	errs := NewZone(TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
	if len(errs) > 0 {
		t.Error(errs)
	}
}

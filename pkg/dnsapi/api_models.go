package dnsapi

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// CreateZoneRequest is the body accepted by POST /zones/.
type CreateZoneRequest struct {
	Domain     string `json:"domain" example:"example.com"`
	Tags       string `json:"tags" example:"tag1,tag2"`
	AbuseEmail string `json:"abuse_email" example:"admin@example.com"`
	Owner      string `json:"owner" example:"alice"`
}

// UpdateZoneRequest is the body accepted by PUT /zones/:id.
// Only tags, abuse_email and owner can be changed; domain and records are managed separately.
type UpdateZoneRequest struct {
	Tags       string `json:"tags" example:"tag1,tag2"`
	AbuseEmail string `json:"abuse_email" example:"admin@example.com"`
	Owner      string `json:"owner" example:"alice"`
}

// CreateRecordRequest is the body accepted by POST /zones/:id/records/.
type CreateRecordRequest struct {
	Name  string `json:"name" example:"www"`
	TTL   int    `json:"ttl" example:"3600"`
	Type  string `json:"type" example:"A"`
	Prio  int    `json:"prio" example:"0"`
	Value string `json:"value" example:"1.2.3.4"`
}

// UpdateRecordRequest is the body accepted by PUT /zones/:id/records/:record_id.
// Type cannot be changed; delete and recreate the record instead.
type UpdateRecordRequest struct {
	Name  string `json:"name" example:"www"`
	TTL   int    `json:"ttl" example:"3600"`
	Prio  int    `json:"prio" example:"0"`
	Value string `json:"value" example:"1.2.3.4"`
}

func captureRecoveredPanic(value interface{}) {
	if value == nil {
		return
	}

	if err, ok := value.(error); ok {
		sentry.CaptureException(err)
		return
	}

	sentry.CaptureMessage(fmt.Sprint(value))
}

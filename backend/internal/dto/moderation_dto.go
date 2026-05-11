package dto

import (
	"anirank/api/internal/domain"
)

// ReportWithSnapshot is a helper DTO to ensure Snapshot is properly included in responses
// and follows the security rule of not exposing internal numeric IDs if possible.
// However, since these are Admin endpoints, we focus on the Snapshot field availability.

type SongReportResponse struct {
	domain.SongReport
	Snapshot *string `json:"snapshot,omitempty"`
}

type CommentReportResponse struct {
	domain.CommentReport
	Snapshot *string `json:"snapshot,omitempty"`
}

type UserReportResponse struct {
	domain.UserReport
	Snapshot *string `json:"snapshot,omitempty"`
}

package dto

import (
	"encoding/json"
	"time"
)

type NotificationDTO struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	SubjectID   *string         `json:"subject_id,omitempty"` // UUID-ready
	SubjectType *string         `json:"subject_type,omitempty"`
	Data        json.RawMessage `json:"data"`
	ReadAt      *time.Time      `json:"read_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

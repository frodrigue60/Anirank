package domain

import (
	"context"
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID            uint64          `db:"id" json:"id"`
	UserID        uint64          `db:"user_id" json:"user_id"`
	Event         string          `db:"event" json:"event"`
	AuditableID   uint64          `db:"auditable_id" json:"auditable_id"`
	AuditableType string          `db:"auditable_type" json:"auditable_type"`
	OldValues     *json.RawMessage `db:"old_values" json:"old_values"`
	NewValues     *json.RawMessage `db:"new_values" json:"new_values"`
	URL           *string         `db:"url" json:"url"`
	IPAddress     *string         `db:"ip_address" json:"ip_address"`
	UserAgent     *string         `db:"user_agent" json:"user_agent"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
	
	// Join fields
	UserName      string          `db:"user_name" json:"user_name,omitempty"`
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetList(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]AuditLog, int, error)
	GetByID(ctx context.Context, id uint64) (*AuditLog, error)
}

type AuditLogUsecase interface {
	LogActions(ctx context.Context, userID uint64, event string, auditableID uint64, auditableType string, oldValues, newValues interface{}, url, ip, ua *string) error
	GetAuditLogs(ctx context.Context, page, limit int, filters map[string]interface{}) ([]AuditLog, int, error)
	GetAuditLog(ctx context.Context, id uint64) (*AuditLog, error)
}

type AuditMetadata struct {
	ActorID   uint64
	Role      string // added for status guard logic
	URL       string
	IPAddress string
	UserAgent string
}

package audit

import (
	"context"
	"encoding/json"

	"anirank/api/internal/domain"
)

type auditLogUsecase struct {
	repo domain.AuditLogRepository
}

func NewAuditLogUsecase(repo domain.AuditLogRepository) domain.AuditLogUsecase {
	return &auditLogUsecase{repo: repo}
}

func (u *auditLogUsecase) LogActions(ctx context.Context, userID uint64, event string, auditableID uint64, auditableType string, oldValues, newValues interface{}, url, ip, ua *string) error {
	var oldRaw, newRaw *json.RawMessage

	if oldValues != nil {
		oldJSON, err := json.Marshal(oldValues)
		if err != nil {
			return err
		}
		raw := json.RawMessage(oldJSON)
		oldRaw = &raw
	}

	if newValues != nil {
		newJSON, err := json.Marshal(newValues)
		if err != nil {
			return err
		}
		raw := json.RawMessage(newJSON)
		newRaw = &raw
	}

	log := &domain.AuditLog{
		UserID:        userID,
		Event:         event,
		AuditableID:   auditableID,
		AuditableType: auditableType,
		OldValues:     oldRaw,
		NewValues:     newRaw,
		URL:           url,
		IPAddress:     ip,
		UserAgent:     ua,
	}

	return u.repo.Create(ctx, log)
}

func (u *auditLogUsecase) GetAuditLogs(ctx context.Context, page, limit int, filters map[string]interface{}) ([]domain.AuditLog, int, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	return u.repo.GetList(ctx, limit, offset, filters)
}

func (u *auditLogUsecase) GetAuditLog(ctx context.Context, id uint64) (*domain.AuditLog, error) {
	return u.repo.GetByID(ctx, id)
}

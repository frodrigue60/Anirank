package postgres

import (
	"context"
	"fmt"
	"strings"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type auditLogRepository struct {
	db *sqlx.DB
}

func NewAuditLogRepository(db *sqlx.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, event, auditable_id, auditable_type, old_values, new_values, url, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.UserID,
		log.Event,
		log.AuditableID,
		log.AuditableType,
		log.OldValues,
		log.NewValues,
		log.URL,
		log.IPAddress,
		log.UserAgent,
	)
	return err
}

func (r *auditLogRepository) GetList(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]domain.AuditLog, int, error) {
	var logs []domain.AuditLog
	var total int

	query := `
		SELECT al.*, u.name as user_name
		FROM audit_logs al
		JOIN users u ON al.user_id = u.id
	`
	countQuery := `SELECT COUNT(*) FROM audit_logs al JOIN users u ON al.user_id = u.id`

	var whereClauses []string
	var args []interface{}
	i := 1

	if userID, ok := filters["user_id"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("al.user_id = $%d", i))
		args = append(args, userID)
		i++
	}
	if event, ok := filters["event"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("al.event = $%d", i))
		args = append(args, event)
		i++
	}
	if auditableType, ok := filters["auditable_type"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("al.auditable_type = $%d", i))
		args = append(args, auditableType)
		i++
	}
	if auditableID, ok := filters["auditable_id"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("al.auditable_id = $%d", i))
		args = append(args, auditableID)
		i++
	}

	if len(whereClauses) > 0 {
		where := " WHERE " + strings.Join(whereClauses, " AND ")
		query += where
		countQuery += where
	}

	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(" ORDER BY al.created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	if err := r.db.SelectContext(ctx, &logs, query, args...); err != nil {
		return nil, 0, err
	}

	if logs == nil {
		logs = []domain.AuditLog{}
	}

	return logs, total, nil
}

func (r *auditLogRepository) GetByID(ctx context.Context, id uint64) (*domain.AuditLog, error) {
	var log domain.AuditLog
	query := `
		SELECT al.*, u.name as user_name
		FROM audit_logs al
		JOIN users u ON al.user_id = u.id
		WHERE al.id = $1
	`
	if err := r.db.GetContext(ctx, &log, query, id); err != nil {
		return nil, err
	}
	return &log, nil
}

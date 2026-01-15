package audit

import (
	"context"

	"pin/internal/domain"
)

// Repository defines persistence operations for audit logs.
type Repository interface {
	WriteAuditLog(ctx context.Context, actorID int, action, target string, metadata map[string]string) error
	ListAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error)
	CountAuditLogs(ctx context.Context) (int, error)
	ListAllAuditLogs(ctx context.Context) ([]domain.AuditLog, error)
}

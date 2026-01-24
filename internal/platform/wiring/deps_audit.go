package wiring

import (
	"context"

	"pin/internal/domain"
)

// Audit.
func (d Deps) ListAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	return d.repos.Audit.ListAuditLogs(ctx, limit, offset)
}

// CountAuditLogs returns audit logs.
func (d Deps) CountAuditLogs(ctx context.Context) (int, error) {
	return d.repos.Audit.CountAuditLogs(ctx)
}

// ListAllAuditLogs returns the all audit logs list by delegating to configured services.
func (d Deps) ListAllAuditLogs(ctx context.Context) ([]domain.AuditLog, error) {
	return d.repos.Audit.ListAllAuditLogs(ctx)
}

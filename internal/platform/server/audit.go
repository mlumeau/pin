package server

import (
	"context"
	"errors"
)

// auditStatus records status as an audit event.
func auditStatus(err error) string {
	if err == nil {
		return "success"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	return "failure"
}

// mergeAuditMeta merges audit meta values into the target.
func mergeAuditMeta(meta map[string]string, extra map[string]string) map[string]string {
	if meta == nil {
		meta = map[string]string{}
	}
	for key, value := range extra {
		meta[key] = value
	}
	return meta
}

// auditAttempt records attempt as an audit event.
func (s *Server) auditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string) {
	meta = mergeAuditMeta(meta, map[string]string{"status": "attempt"})
	_ = s.repos.Audit.WriteAuditLog(ctx, actorID, action, target, meta)
}

// auditOutcome records outcome as an audit event.
func (s *Server) auditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string) {
	status := auditStatus(err)
	meta = mergeAuditMeta(meta, map[string]string{"status": status})
	if err != nil {
		meta["error"] = err.Error()
	}
	_ = s.repos.Audit.WriteAuditLog(ctx, actorID, action, target, meta)
}

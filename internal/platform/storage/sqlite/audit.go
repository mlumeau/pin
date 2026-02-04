package sqlitestore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"pin/internal/domain"
)

// WriteAuditLog writes audit log to the response/output.
func WriteAuditLog(ctx context.Context, db *sql.DB, actorID int, action, target string, metadata map[string]string) error {
	metaJSON := ""
	if metadata != nil {
		if raw, err := json.Marshal(metadata); err == nil {
			metaJSON = string(raw)
		}
	}

	actorName := ""
	if actorID > 0 {
		row := db.QueryRowContext(ctx, "SELECT COALESCE(handle,'') FROM identity WHERE user_id = ? LIMIT 1", actorID)
		if err := row.Scan(&actorName); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	_, err := db.ExecContext(
		ctx,
		"INSERT INTO audit_log (actor_id, actor_name, action, target, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		actorID,
		actorName,
		action,
		target,
		metaJSON,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// ListAuditLogs returns a page of audit logs using limit/offset in the SQLite store.
func ListAuditLogs(ctx context.Context, db *sql.DB, limit, offset int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := db.QueryContext(ctx, "SELECT audit_log.id, COALESCE(audit_log.actor_id, 0), COALESCE(audit_log.actor_name,''), audit_log.action, COALESCE(audit_log.target,''), COALESCE(audit_log.metadata,''), audit_log.created_at FROM audit_log WHERE COALESCE(audit_log.metadata,'') NOT LIKE '%\"status\":\"attempt\"%' ORDER BY audit_log.id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var logEntry domain.AuditLog
		var actorID int64
		var created string
		if err := rows.Scan(&logEntry.ID, &actorID, &logEntry.ActorName, &logEntry.Action, &logEntry.Target, &logEntry.Metadata, &created); err != nil {
			return nil, err
		}
		if actorID > 0 {
			logEntry.ActorID = sql.NullInt64{Int64: actorID, Valid: true}
		}
		logEntry.CreatedAt, _ = time.Parse(time.RFC3339, created)
		logs = append(logs, logEntry)
	}
	return logs, nil
}

// CountAuditLogs returns audit logs.
func CountAuditLogs(ctx context.Context, db *sql.DB) (int, error) {
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_log WHERE COALESCE(audit_log.metadata,'') NOT LIKE '%\"status\":\"attempt\"%'")
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// ListAllAuditLogs returns the all audit logs list in the SQLite store.
func ListAllAuditLogs(ctx context.Context, db *sql.DB) ([]domain.AuditLog, error) {
	rows, err := db.QueryContext(ctx, "SELECT audit_log.id, COALESCE(audit_log.actor_id, 0), COALESCE(audit_log.actor_name,''), audit_log.action, COALESCE(audit_log.target,''), COALESCE(audit_log.metadata,''), audit_log.created_at FROM audit_log WHERE COALESCE(audit_log.metadata,'') NOT LIKE '%\"status\":\"attempt\"%' ORDER BY audit_log.id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var logEntry domain.AuditLog
		var actorID int64
		var created string
		if err := rows.Scan(&logEntry.ID, &actorID, &logEntry.ActorName, &logEntry.Action, &logEntry.Target, &logEntry.Metadata, &created); err != nil {
			return nil, err
		}
		if actorID > 0 {
			logEntry.ActorID = sql.NullInt64{Int64: actorID, Valid: true}
		}
		logEntry.CreatedAt, _ = time.Parse(time.RFC3339, created)
		logs = append(logs, logEntry)
	}
	return logs, nil
}

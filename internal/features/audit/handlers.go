package audit

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pin/internal/domain"
)

type Dependencies interface {
	CurrentUser(r *http.Request) (domain.User, error)
	ListAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error)
	ListAllAuditLogs(ctx context.Context) ([]domain.AuditLog, error)
}

type Handler struct {
	deps Dependencies
}

// NewHandler constructs a new handler.
func NewHandler(deps Dependencies) Handler {
	return Handler{deps: deps}
}

// Download exports audit logs as CSV, JSON, or text, scoped by the request query.
func (h Handler) Download(w http.ResponseWriter, r *http.Request) {
	current, err := h.deps.CurrentUser(r)
	if err != nil || !isAdmin(current) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// Normalize query params with safe defaults.
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if format == "" {
		format = "csv"
	}
	if scope == "" {
		scope = "page"
	}

	var logs []domain.AuditLog
	// Load either the entire log or a paged slice based on scope.
	if scope == "all" {
		logs, err = h.deps.ListAllAuditLogs(r.Context())
	} else {
		page := 1
		if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
			page = p
		}
		const auditPageSize = 20
		offset := (page - 1) * auditPageSize
		logs, err = h.deps.ListAuditLogs(r.Context(), auditPageSize, offset)
	}
	if err != nil {
		http.Error(w, "Failed to load audit log", http.StatusInternalServerError)
		return
	}

	filename := "audit-log"
	if scope == "all" {
		filename += "-all"
	} else {
		filename += "-page"
	}
	filename += "." + format

	// Render the export in the requested format.
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		_ = json.NewEncoder(w).Encode(logs)
	case "txt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		for _, logEntry := range logs {
			actor := logEntry.ActorName
			if actor == "" {
				actor = "system"
			}
			target := logEntry.Target
			if target == "" {
				target = "n/a"
			}
			fmt.Fprintf(
				w,
				"%s | %s | by %s | object %s | log #%d\n",
				logEntry.CreatedAt.Format("2006-01-02 15:04:05"),
				logEntry.Action,
				actor,
				target,
				logEntry.ID,
			)
		}
	default:
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		writer := csv.NewWriter(w)
		_ = writer.Write([]string{"id", "timestamp", "action", "actor", "target", "metadata"})
		for _, logEntry := range logs {
			actor := logEntry.ActorName
			if actor == "" {
				actor = "system"
			}
			_ = writer.Write([]string{
				strconv.Itoa(logEntry.ID),
				logEntry.CreatedAt.Format(time.RFC3339),
				logEntry.Action,
				actor,
				logEntry.Target,
				logEntry.Metadata,
			})
		}
		writer.Flush()
	}
}

// isAdmin reports whether admin is true.
func isAdmin(user domain.User) bool {
	return strings.EqualFold(user.Role, "admin") || strings.EqualFold(user.Role, "owner")
}

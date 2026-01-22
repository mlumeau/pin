package domains

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"pin/internal/domain"
	"pin/internal/features/identity"
)

type Dependencies interface {
	Store
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	ValidateCSRF(session *sessions.Session, token string) bool
	CurrentUser(r *http.Request) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

type Handler struct {
	deps Dependencies
	svc  Service
}

func NewHandler(deps Dependencies) Handler {
	return Handler{deps: deps, svc: NewService(deps)}
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	domainList := parseDomains(r.FormValue("domains"))
	if len(domainList) == 0 {
		http.Error(w, "No domains provided", http.StatusBadRequest)
		return
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "domain.create", strings.Join(domainList, ","), nil)
	rows, verified, err := h.svc.CreateDomains(r.Context(), current.ID, domainList, func() string {
		return RandomTokenURL(12)
	})
	if err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "domain.create", strings.Join(domainList, ","), err, nil)
		http.Error(w, "Failed to save domains", http.StatusInternalServerError)
		return
	}
	domainsJSON := identity.EncodeStringSlice(verified)
	meta := map[string]string{"verified_domains": domainsJSON}
	current.VerifiedDomainsJSON = domainsJSON
	if err := h.deps.UpdateUser(r.Context(), current); err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "domain.create", strings.Join(domainList, ","), err, meta)
		http.Error(w, "Failed to save profile", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "domain.create", strings.Join(domainList, ","), nil, meta)
	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"ok":       true,
			"domains":  identity.DomainsToText(rows),
			"verified": verified,
		})
		return
	}
	http.Redirect(w, r, "/settings/profile?toast=Domain%20records%20seeded#section-verified-domains", http.StatusFound)
}

func (h Handler) Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	domain := normalizeDomain(r.FormValue("domain"))
	if domain == "" {
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "domain.verify", domain, nil)
	rows, verified, err := h.svc.VerifyDomain(r.Context(), current.ID, domain)
	if err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "domain.verify", domain, err, nil)
		if err.Error() == "domain not found" {
			http.Error(w, "Domain not found", http.StatusBadRequest)
			return
		}
		if err.Error() == "token not found" {
			http.Error(w, "Token not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Verification failed", http.StatusBadRequest)
		return
	}
	var updateErr error
	if domainsJSON, err := json.Marshal(verified); err == nil {
		current.VerifiedDomainsJSON = string(domainsJSON)
		updateErr = h.deps.UpdateUser(r.Context(), current)
	} else {
		updateErr = err
	}
	if updateErr != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "domain.verify", domain, updateErr, nil)
		http.Error(w, "Failed to save domains", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "domain.verify", domain, nil, nil)
	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"ok":       true,
			"domains":  identity.DomainsToText(rows),
			"verified": verified,
		})
		return
	}
	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	domain := normalizeDomain(r.FormValue("domain"))
	if domain == "" {
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "domain.delete", domain, nil)
	rows, remaining, err := h.svc.DeleteDomain(r.Context(), current.ID, domain)
	if err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "domain.delete", domain, err, nil)
		http.Error(w, "Failed to delete domain", http.StatusInternalServerError)
		return
	}
	current.VerifiedDomainsJSON = identity.EncodeStringSlice(remaining)
	if err := h.deps.UpdateUser(r.Context(), current); err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "domain.delete", domain, err, nil)
		http.Error(w, "Failed to save profile", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "domain.delete", domain, nil, map[string]string{"verified": strings.Join(remaining, ",")})
	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"ok":       true,
			"domains":  identity.DomainsToText(rows),
			"verified": remaining,
		})
		return
	}
	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(data)
}

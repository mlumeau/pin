package passkeys

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
	"pin/internal/config"
	"pin/internal/domain"
	"pin/internal/platform/core"
)

const (
	passkeyRegisterSessionKey = "webauthn_register"
	passkeyLoginSessionKey    = "webauthn_login"
	passkeyRegisterNameKey    = "webauthn_register_name"
	passkeyLoginUserKey       = "webauthn_login_user"
	passkeyLoginNextKey       = "webauthn_login_next"
)

type Dependencies interface {
	Config() config.Config
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	ValidateCSRF(session *sessions.Session, token string) bool
	CurrentUser(r *http.Request) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
	LoadPasskeyCredentials(ctx context.Context, userID int) ([]webauthn.Credential, error)
	InsertPasskey(ctx context.Context, userID int, name string, credential webauthn.Credential) error
	UpdatePasskeyCredential(ctx context.Context, userID int, credentialID string, credential webauthn.Credential) error
	DeletePasskey(ctx context.Context, userID, id int) error
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

type Handler struct {
	deps Dependencies
}

func NewHandler(deps Dependencies) Handler {
	return Handler{deps: deps}
}

type passkeyUser struct {
	user        domain.User
	credentials []webauthn.Credential
}

func (u passkeyUser) WebAuthnID() []byte {
	return []byte(strconv.Itoa(u.user.ID))
}

func (u passkeyUser) WebAuthnName() string {
	return u.user.Username
}

func (u passkeyUser) WebAuthnDisplayName() string {
	return core.FirstNonEmpty(u.user.DisplayName, u.user.Username)
}

func (u passkeyUser) WebAuthnIcon() string {
	return ""
}

func (u passkeyUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (h Handler) RegisterOptions(w http.ResponseWriter, r *http.Request) {
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
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = "Passkey"
	}

	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	creds, err := h.deps.LoadPasskeyCredentials(r.Context(), current.ID)
	if err != nil {
		http.Error(w, "Failed to load passkeys", http.StatusInternalServerError)
		return
	}
	wa, err := h.webauthnForRequest(r)
	if err != nil {
		http.Error(w, "Passkey unavailable", http.StatusInternalServerError)
		return
	}
	user := passkeyUser{user: current, credentials: creds}
	options, sessionData, err := wa.BeginRegistration(
		user,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementPreferred,
			UserVerification: protocol.VerificationPreferred,
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
	)
	if err != nil {
		http.Error(w, "Failed to start passkey registration", http.StatusBadRequest)
		return
	}
	if err := storeWebauthnSession(session, passkeyRegisterSessionKey, sessionData); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	session.Values[passkeyRegisterNameKey] = name
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	core.WriteJSON(w, options)
}

func (h Handler) RegisterFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	sessionData, err := loadWebauthnSession(session, passkeyRegisterSessionKey)
	if err != nil {
		http.Error(w, "Passkey session expired", http.StatusBadRequest)
		return
	}
	creds, err := h.deps.LoadPasskeyCredentials(r.Context(), current.ID)
	if err != nil {
		http.Error(w, "Failed to load passkeys", http.StatusInternalServerError)
		return
	}
	wa, err := h.webauthnForRequest(r)
	if err != nil {
		http.Error(w, "Passkey unavailable", http.StatusInternalServerError)
		return
	}
	user := passkeyUser{user: current, credentials: creds}
	credential, err := wa.FinishRegistration(user, *sessionData, r)
	if err != nil {
		http.Error(w, "Passkey registration failed", http.StatusBadRequest)
		return
	}
	name, _ := session.Values[passkeyRegisterNameKey].(string)
	if strings.TrimSpace(name) == "" {
		name = "Passkey"
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "passkey.register", name, nil)
	if err := h.deps.InsertPasskey(r.Context(), current.ID, name, *credential); err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "passkey.register", name, err, nil)
		http.Error(w, "Failed to save passkey", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "passkey.register", name, nil, nil)
	delete(session.Values, passkeyRegisterSessionKey)
	delete(session.Values, passkeyRegisterNameKey)
	_ = session.Save(r, w)
	core.WriteJSON(w, map[string]interface{}{"ok": true})
}

func (h Handler) LoginOptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	user, err := h.deps.GetUserByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "Unknown user", http.StatusBadRequest)
		return
	}
	creds, err := h.deps.LoadPasskeyCredentials(r.Context(), user.ID)
	if err != nil || len(creds) == 0 {
		http.Error(w, "No passkeys enrolled", http.StatusBadRequest)
		return
	}
	wa, err := h.webauthnForRequest(r)
	if err != nil {
		http.Error(w, "Passkey unavailable", http.StatusInternalServerError)
		return
	}
	pkUser := passkeyUser{user: user, credentials: creds}
	options, sessionData, err := wa.BeginLogin(pkUser)
	if err != nil {
		http.Error(w, "Failed to start passkey login", http.StatusBadRequest)
		return
	}

	session, _ := h.deps.GetSession(r, "pin_session")
	if err := storeWebauthnSession(session, passkeyLoginSessionKey, sessionData); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	session.Values[passkeyLoginUserKey] = user.Username
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/settings"
	}
	if !core.IsSafeRedirect(r, next) {
		next = "/settings"
	}
	session.Values[passkeyLoginNextKey] = next
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	core.WriteJSON(w, options)
}

func (h Handler) LoginFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	sessionData, err := loadWebauthnSession(session, passkeyLoginSessionKey)
	if err != nil {
		http.Error(w, "Passkey session expired", http.StatusBadRequest)
		return
	}
	username, _ := session.Values[passkeyLoginUserKey].(string)
	if strings.TrimSpace(username) == "" {
		http.Error(w, "Passkey session expired", http.StatusBadRequest)
		return
	}
	user, err := h.deps.GetUserByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "Unknown user", http.StatusBadRequest)
		return
	}
	creds, err := h.deps.LoadPasskeyCredentials(r.Context(), user.ID)
	if err != nil || len(creds) == 0 {
		http.Error(w, "No passkeys enrolled", http.StatusBadRequest)
		return
	}
	wa, err := h.webauthnForRequest(r)
	if err != nil {
		http.Error(w, "Passkey unavailable", http.StatusInternalServerError)
		return
	}
	pkUser := passkeyUser{user: user, credentials: creds}
	credential, err := wa.FinishLogin(pkUser, *sessionData, r)
	if err != nil {
		http.Error(w, "Passkey login failed", http.StatusBadRequest)
		return
	}
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	h.deps.AuditAttempt(r.Context(), user.ID, "passkey.update", credentialID, nil)
	if err := h.deps.UpdatePasskeyCredential(r.Context(), user.ID, credentialID, *credential); err != nil {
		h.deps.AuditOutcome(r.Context(), user.ID, "passkey.update", credentialID, err, nil)
		http.Error(w, "Failed to update passkey", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), user.ID, "passkey.update", credentialID, nil, nil)

	session.Values["user_id"] = user.ID
	delete(session.Values, passkeyLoginSessionKey)
	delete(session.Values, passkeyLoginUserKey)
	next, _ := session.Values[passkeyLoginNextKey].(string)
	delete(session.Values, passkeyLoginNextKey)
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	if next == "" {
		next = "/settings"
	}
	core.WriteJSON(w, map[string]interface{}{"ok": true, "redirect": next})
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
	id, err := strconv.Atoi(strings.TrimSpace(r.FormValue("id")))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid passkey", http.StatusBadRequest)
		return
	}
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "passkey.delete", strconv.Itoa(id), nil)
	if err := h.deps.DeletePasskey(r.Context(), current.ID, id); err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "passkey.delete", strconv.Itoa(id), err, nil)
		http.Error(w, "Failed to delete passkey", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "passkey.delete", strconv.Itoa(id), nil, nil)
	core.WriteJSON(w, map[string]interface{}{"ok": true})
}

func (h Handler) webauthnForRequest(r *http.Request) (*webauthn.WebAuthn, error) {
	origin := strings.TrimRight(h.deps.Config().BaseURL, "/")
	if origin == "" {
		origin = core.BaseURL(r)
	}
	rpID := rpIDFromOrigin(origin)
	if rpID == "" {
		rpID = stripPort(r.Host)
	}
	return webauthn.New(&webauthn.Config{
		RPDisplayName: "Pin",
		RPID:          rpID,
		RPOrigins:     []string{origin},
	})
}

func rpIDFromOrigin(origin string) string {
	parsed, err := url.Parse(origin)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func stripPort(host string) string {
	if host == "" {
		return ""
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	}
	return host
}

func storeWebauthnSession(session *sessions.Session, key string, data *webauthn.SessionData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	session.Values[key] = string(b)
	return nil
}

func loadWebauthnSession(session *sessions.Session, key string) (*webauthn.SessionData, error) {
	raw, _ := session.Values[key].(string)
	if raw == "" {
		return nil, errors.New("missing session")
	}
	var data webauthn.SessionData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

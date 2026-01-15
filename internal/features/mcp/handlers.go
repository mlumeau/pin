package mcp

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"pin/internal/domain"
	"pin/internal/features/identity"
	"pin/internal/features/identity/export"
	"pin/internal/features/profilepicture"
)

const protocolVersion = "2024-11-05"

type Config struct {
	Enabled  bool
	Token    string
	ReadOnly bool
}

type Dependencies interface {
	profilepicture.Store
	ListUsers(ctx context.Context) ([]domain.User, error)
	FindUserByIdentifier(ctx context.Context, identifier string) (domain.User, error)
	BaseURL(r *http.Request) string
	GetOwnerUser(ctx context.Context) (domain.User, error)
}

type Handler struct {
	cfg  Config
	deps Dependencies
}

func NewHandler(cfg Config, deps Dependencies) Handler {
	return Handler{cfg: cfg, deps: deps}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type readParams struct {
	URI string `json:"uri"`
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.cfg.Enabled {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if !h.authorize(r) {
		h.writeError(w, nil, -32001, "Unauthorized")
		return
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, nil, -32700, "Invalid JSON")
		return
	}
	if req.JSONRPC != "2.0" {
		h.writeError(w, req.ID, -32600, "Invalid JSON-RPC version")
		return
	}
	if h.cfg.ReadOnly && !isReadMethod(req.Method) {
		h.writeError(w, req.ID, -32601, "Method not allowed")
		return
	}
	switch req.Method {
	case "initialize":
		h.writeResult(w, req.ID, map[string]interface{}{
			"protocolVersion": protocolVersion,
			"serverInfo": map[string]string{
				"name":    "pin",
				"version": "1.0",
			},
			"capabilities": map[string]interface{}{
				"resources": map[string]interface{}{},
			},
		})
	case "resources/list":
		resources, err := h.listIdentityResources(r)
		if err != nil {
			h.writeError(w, req.ID, -32000, "Failed to list resources")
			return
		}
		h.writeResult(w, req.ID, map[string]interface{}{
			"resources": resources,
		})
	case "resources/read":
		var params readParams
		if err := json.Unmarshal(req.Params, &params); err != nil || strings.TrimSpace(params.URI) == "" {
			h.writeError(w, req.ID, -32602, "Invalid params")
			return
		}
		contents, err := h.readIdentityResource(r, params.URI)
		if err != nil {
			h.writeError(w, req.ID, -32004, err.Error())
			return
		}
		h.writeResult(w, req.ID, map[string]interface{}{
			"contents": contents,
		})
	default:
		h.writeError(w, req.ID, -32601, "Method not found")
	}
}

func (h Handler) listIdentityResources(r *http.Request) ([]resource, error) {
	users, err := h.deps.ListUsers(r.Context())
	if err != nil {
		return nil, err
	}
	var resources []resource
	for _, user := range users {
		if user.Username == "" {
			continue
		}
		base := "identity://" + user.Username
		resources = append(resources, resource{
			URI:         base,
			Name:        user.Username,
			Description: "Identity export for " + user.Username,
			MimeType:    "application/json",
		})
		resources = append(resources, resource{
			URI:         base + "/profile-picture",
			Name:        user.Username + " profile picture",
			Description: "Active profile picture for " + user.Username,
			MimeType:    "application/json",
		})
	}
	return resources, nil
}

func (h Handler) readIdentityResource(r *http.Request, uri string) ([]map[string]interface{}, error) {
	target, err := parseIdentityURI(uri)
	if err != nil {
		return nil, err
	}
	user, err := h.deps.FindUserByIdentifier(r.Context(), target.Ident)
	if err != nil || !identity.MatchesIdentity(user, target.Ident) {
		return nil, errors.New("Identity not found")
	}
	if target.ProfilePicture {
		payload := map[string]string{
			"url": h.profilePictureURL(r, user),
			"alt": profilepicture.NewService(h.deps).ActiveAlt(r.Context(), user),
		}
		raw, _ := json.Marshal(payload)
		return []map[string]interface{}{
			{
				"uri":      uri,
				"mimeType": "application/json",
				"text":     string(raw),
			},
		}, nil
	}
	publicUser, customFields := identity.VisibleIdentity(user, false)
	handler := export.NewHandler(source{deps: h.deps})
	identityExport, err := handler.Build(r.Context(), r, publicUser, customFields, h.deps.BaseURL(r)+"/"+url.PathEscape(user.Username))
	if err != nil {
		return nil, errors.New("Failed to load identity")
	}
	wrapped := map[string]interface{}{"identity": identityExport}
	raw, _ := json.Marshal(wrapped)
	return []map[string]interface{}{
		{
			"uri":      uri,
			"mimeType": "application/json",
			"text":     string(raw),
		},
	}, nil
}

type identityResourceTarget struct {
	Ident          string
	ProfilePicture bool
}

func parseIdentityURI(uri string) (identityResourceTarget, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return identityResourceTarget{}, errors.New("Invalid resource URI")
	}
	if parsed.Scheme != "identity" {
		return identityResourceTarget{}, errors.New("Unsupported resource URI")
	}
	ident := strings.Trim(parsed.Host+parsed.Path, "/")
	if ident == "" {
		return identityResourceTarget{}, errors.New("Invalid resource URI")
	}
	if strings.HasSuffix(strings.ToLower(ident), "/profile-picture") {
		name := strings.TrimSuffix(ident, "/profile-picture")
		return identityResourceTarget{Ident: name, ProfilePicture: true}, nil
	}
	return identityResourceTarget{Ident: ident}, nil
}

func (h Handler) profilePictureURL(r *http.Request, user domain.User) string {
	return h.deps.BaseURL(r) + "/profile-picture/" + url.PathEscape(user.Username)
}

func (h Handler) writeResult(w http.ResponseWriter, id interface{}, result interface{}) {
	resp := response{JSONRPC: "2.0", ID: id, Result: result}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h Handler) writeError(w http.ResponseWriter, id interface{}, code int, message string) {
	resp := response{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message}}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h Handler) authorize(r *http.Request) bool {
	if strings.TrimSpace(h.cfg.Token) == "" {
		return true
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		token := strings.TrimSpace(auth[7:])
		return subtleCompare(token, h.cfg.Token)
	}
	if token := strings.TrimSpace(r.Header.Get("X-MCP-Token")); token != "" {
		return subtleCompare(token, h.cfg.Token)
	}
	return false
}

func subtleCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func isReadMethod(method string) bool {
	switch method {
	case "initialize", "resources/list", "resources/read":
		return true
	default:
		return false
	}
}

type source struct {
	deps Dependencies
}

func (s source) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return s.deps.GetOwnerUser(ctx)
}

func (s source) VisibleIdentity(user domain.User, isPrivate bool) (domain.User, map[string]string) {
	return identity.VisibleIdentity(user, isPrivate)
}

func (s source) ActiveProfilePictureAlt(ctx context.Context, user domain.User) string {
	return profilepicture.NewService(s.deps).ActiveAlt(ctx, user)
}

func (s source) BaseURL(r *http.Request) string {
	return s.deps.BaseURL(r)
}

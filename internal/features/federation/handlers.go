package federation

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pin/internal/config"
	"pin/internal/domain"
	"pin/internal/features/identity"
	"pin/internal/platform/core"
)

type Dependencies interface {
	GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error)
	GetOwnerIdentity(ctx context.Context) (domain.Identity, error)
}

// Handler hosts federation and well-known endpoints.
type Handler struct {
	cfg  config.Config
	deps Dependencies
}

// NewHandler creates a federation handler with required dependencies.
func NewHandler(cfg config.Config, deps Dependencies) Handler {
	return Handler{cfg: cfg, deps: deps}
}

// Actor handles the HTTP request.
func (h Handler) Actor(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	if strings.HasSuffix(path, "/inbox") {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	handle := strings.Trim(path, "/")
	user, err := h.deps.GetIdentityByHandle(r.Context(), handle)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !identity.MatchesIdentity(user, handle) {
		http.NotFound(w, r)
		return
	}

	publicUser, _ := identity.VisibleIdentity(user, false)
	baseURL := core.BaseURL(r)
	actorURL := baseURL + "/users/" + publicUser.Handle
	inboxURL := baseURL + "/users/" + publicUser.Handle + "/inbox"
	profilePictureURL := baseURL + "/" + publicUser.Handle + "/profile-picture"

	response := map[string]interface{}{
		"@context": []string{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
		},
		"id":                actorURL,
		"type":              "Person",
		"preferredUsername": publicUser.Handle,
		"name":              identity.FirstNonEmpty(publicUser.DisplayName, publicUser.Handle),
		"summary":           publicUser.Bio,
		"inbox":             inboxURL,
		"url":               baseURL,
		"icon": map[string]string{
			"type":      "Image",
			"mediaType": "image/png",
			"url":       profilePictureURL,
		},
	}
	if key := strings.TrimSpace(identity.DecodeStringMap(publicUser.PublicKeysJSON)["activitypub"]); key != "" {
		response["publicKey"] = map[string]string{
			"id":           actorURL + "#main-key",
			"owner":        actorURL,
			"publicKeyPem": key,
		}
	}
	var socialProfiles []domain.SocialProfile
	if publicUser.SocialProfilesJSON != "" {
		_ = json.Unmarshal([]byte(publicUser.SocialProfilesJSON), &socialProfiles)
	}
	if attachment := identity.BuildAttachments(publicUser, identity.DecodeStringMap(publicUser.WalletsJSON), identity.DecodeStringMap(publicUser.PublicKeysJSON), identity.DecodeStringSlice(publicUser.VerifiedDomainsJSON), socialProfiles); len(attachment) > 0 {
		response["attachment"] = attachment
	}

	w.Header().Set("Content-Type", "application/activity+json")
	_ = json.NewEncoder(w).Encode(response)
}

// WellKnownPinVerify serves .well-known/pin-verify from the static directory.
func (h Handler) WellKnownPinVerify(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.cfg.StaticDir, ".well-known", "pin-verify")
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.ServeFile(w, r, path)
}

// Webfinger handles the HTTP request.
func (h Handler) Webfinger(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	if !strings.HasPrefix(resource, "acct:") {
		http.Error(w, "resource must start with acct:", http.StatusBadRequest)
		return
	}
	acct := strings.TrimPrefix(resource, "acct:")
	parts := strings.SplitN(acct, "@", 2)
	if len(parts) != 2 {
		http.Error(w, "invalid resource", http.StatusBadRequest)
		return
	}
	handle := parts[0]

	user, err := h.deps.GetIdentityByHandle(r.Context(), handle)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !identity.MatchesIdentity(user, handle) {
		http.NotFound(w, r)
		return
	}

	baseURL := core.BaseURL(r)
	actorURL := baseURL + "/users/" + user.Handle
	profileURL := baseURL

	response := map[string]interface{}{
		"subject": resource,
		"aliases": []string{actorURL, profileURL},
		"links": []map[string]string{
			{
				"rel":  "self",
				"type": "application/activity+json",
				"href": actorURL,
			},
			{
				"rel":  "http://webfinger.net/rel/profile-page",
				"type": "text/html",
				"href": profileURL,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// AtprotoDID returns the configured atproto DID if present.
func (h Handler) AtprotoDID(w http.ResponseWriter, r *http.Request) {
	user, err := h.deps.GetOwnerIdentity(r.Context())
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}
	did := strings.TrimSpace(user.ATProtoDID)
	if did == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(did))
}

// PincCapability handles HTTP requests for capability.
func (h Handler) PincCapability(w http.ResponseWriter, r *http.Request) {
	base := core.BaseURL(r)
	payload := map[string]interface{}{
		"pinc_version":   identity.PincVersion,
		"base_url":       base,
		"export_formats": []string{"json", "xml", "txt", "vcf"},
		"views":          []string{"public", "private"},
		"media_formats":  []string{"webp", "png", "jpeg"},
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(payload)
}

// PincIdentitySchema returns the JSON schema for the PINC canonical identity.
func (h Handler) PincIdentitySchema(w http.ResponseWriter, r *http.Request) {
	base := core.BaseURL(r)
	schema := map[string]interface{}{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$id":     base + "/.well-known/pinc/identity",
		"title":   "PINC Canonical Identity",
		"type":    "object",
		"properties": map[string]interface{}{
			"meta": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version":  map[string]interface{}{"type": "string"},
					"base_url": map[string]interface{}{"type": "string"},
					"view": map[string]interface{}{
						"type": "string",
						"enum": []string{"public", "private"},
					},
					"subject": map[string]interface{}{"type": "string"},
					"rev":     map[string]interface{}{"type": "string"},
					"self":    map[string]interface{}{"type": "string"},
				},
				"required": []string{"version", "base_url", "view", "subject", "rev"},
			},
			"identity": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"handle":       map[string]interface{}{"type": "string"},
					"display_name": map[string]interface{}{"type": "string"},
					"url":          map[string]interface{}{"type": "string"},
					"updated_at":   map[string]interface{}{"type": "string"},
				},
				"required": []string{"handle", "display_name", "url", "updated_at"},
			},
		},
		"required": []string{"meta", "identity"},
	}
	w.Header().Set("Content-Type", "application/schema+json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(schema)
}

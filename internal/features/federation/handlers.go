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
	FindUserByIdentifier(ctx context.Context, identifier string) (domain.User, error)
	GetOwnerUser(ctx context.Context) (domain.User, error)
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

// Actor serves ActivityPub actor JSON and inbox endpoint.
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

	username := strings.Trim(path, "/")
	user, err := h.deps.FindUserByIdentifier(r.Context(), username)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !identity.MatchesIdentity(user, username) {
		http.NotFound(w, r)
		return
	}

	publicUser, _ := identity.VisibleIdentity(user, false)
	baseURL := core.BaseURL(r)
	actorURL := baseURL + "/users/" + publicUser.Username
	inboxURL := baseURL + "/users/" + publicUser.Username + "/inbox"
	profilePictureURL := baseURL + "/profile-picture/" + publicUser.Username

	response := map[string]interface{}{
		"@context": []string{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
		},
		"id":                actorURL,
		"type":              "Person",
		"preferredUsername": publicUser.Username,
		"name":              identity.FirstNonEmpty(publicUser.DisplayName, publicUser.Username),
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

// Webfinger serves the WebFinger response for ActivityPub discovery.
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
	username := parts[0]

	user, err := h.deps.FindUserByIdentifier(r.Context(), username)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !identity.MatchesIdentity(user, username) {
		http.NotFound(w, r)
		return
	}

	baseURL := core.BaseURL(r)
	actorURL := baseURL + "/users/" + user.Username
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
	user, err := h.deps.GetOwnerUser(r.Context())
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

// IdentitySchema returns the JSON schema for the identity export.
func (h Handler) IdentitySchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/schema+json; charset=utf-8")
	identity.WriteIdentityCacheHeaders(w)
	base := core.BaseURL(r)
	schema := map[string]interface{}{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$id":     base + "/identity.schema.json",
		"title":   "PIN Identity",
		"type":    "object",
		"properties": map[string]interface{}{
			"identity": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"username":     map[string]interface{}{"type": "string"},
					"display_name": map[string]interface{}{"type": "string"},
					"email":        map[string]interface{}{"type": "string"},
					"bio":          map[string]interface{}{"type": "string"},
					"organization": map[string]interface{}{"type": "string"},
					"job_title":    map[string]interface{}{"type": "string"},
					"birthdate":    map[string]interface{}{"type": "string"},
					"languages":    map[string]interface{}{"type": "string"},
					"phone":        map[string]interface{}{"type": "string"},
					"address":      map[string]interface{}{"type": "string"},
					"location":     map[string]interface{}{"type": "string"},
					"website":      map[string]interface{}{"type": "string"},
					"pronouns":     map[string]interface{}{"type": "string"},
					"timezone":     map[string]interface{}{"type": "string"},
					"custom_fields": map[string]interface{}{
						"type":                 "object",
						"additionalProperties": map[string]interface{}{"type": "string"},
					},
					"profile_url":       map[string]interface{}{"type": "string"},
					"profile_image":     map[string]interface{}{"type": "string"},
					"profile_image_alt": map[string]interface{}{"type": "string"},
					"aliases": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"type": "string"},
					},
					"links": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"label": map[string]interface{}{"type": "string"},
								"url":   map[string]interface{}{"type": "string"},
							},
						},
					},
					"social": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"label":    map[string]interface{}{"type": "string"},
								"url":      map[string]interface{}{"type": "string"},
								"provider": map[string]interface{}{"type": "string"},
								"verified": map[string]interface{}{"type": "boolean"},
							},
						},
					},
					"wallets": map[string]interface{}{
						"type":                 "object",
						"additionalProperties": map[string]interface{}{"type": "string"},
					},
					"public_keys": map[string]interface{}{
						"type":                 "object",
						"additionalProperties": map[string]interface{}{"type": "string"},
					},
					"verified_domains": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"type": "string"},
					},
					"atproto_handle": map[string]interface{}{"type": "string"},
					"atproto_did":    map[string]interface{}{"type": "string"},
					"updated_at":     map[string]interface{}{"type": "string"},
				},
				"required": []string{"username", "display_name", "profile_url", "profile_image"},
			},
		},
		"required": []string{"identity"},
	}
	_ = json.NewEncoder(w).Encode(schema)
}

package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"pin/internal/domain"
	"pin/internal/features/identity"
	"pin/internal/platform/core"
)

type Config struct {
	BaseURL            string
	GitHubClientID     string
	GitHubClientSecret string
	RedditClientID     string
	RedditClientSecret string
	RedditUserAgent    string
	BlueskyPDS         string
}

type Dependencies interface {
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	ValidateCSRF(session *sessions.Session, token string) bool
	CurrentUser(r *http.Request) (domain.User, error)
	CurrentIdentity(r *http.Request) (domain.Identity, error)
	UpdateIdentity(ctx context.Context, identity domain.Identity) error
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

type Handler struct {
	cfg  Config
	deps Dependencies
}

// NewHandler constructs a new handler.
func NewHandler(cfg Config, deps Dependencies) Handler {
	return Handler{cfg: cfg, deps: deps}
}

// GitHubStart redirects to GitHub OAuth with a stored state token.
func (h Handler) GitHubStart(w http.ResponseWriter, r *http.Request) {
	if h.cfg.GitHubClientID == "" || h.cfg.GitHubClientSecret == "" || h.cfg.BaseURL == "" {
		http.Error(w, "GitHub OAuth not configured", http.StatusBadRequest)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	// Bind the authorization request to a session-scoped state token.
	state := core.RandomToken(16)
	session.Values["oauth_github_state"] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	authURL := "https://github.com/login/oauth/authorize"
	params := url.Values{}
	params.Set("client_id", h.cfg.GitHubClientID)
	params.Set("redirect_uri", h.cfg.BaseURL+"/oauth/github/callback")
	params.Set("scope", "read:user")
	params.Set("state", state)
	http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound)
}

// GitHubCallback exchanges the code for a token and persists the verified profile.
func (h Handler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	session, _ := h.deps.GetSession(r, "pin_session")
	// Validate state to protect against CSRF in the OAuth callback.
	state, _ := session.Values["oauth_github_state"].(string)
	if state == "" || r.URL.Query().Get("state") != state {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	token, err := exchangeGitHubToken(h.cfg, code)
	if err != nil {
		http.Error(w, "GitHub auth failed", http.StatusBadRequest)
		return
	}
	_, profileURL, err := fetchGitHubProfile(token)
	if err != nil {
		http.Error(w, "GitHub profile fetch failed", http.StatusBadRequest)
		return
	}

	if err := h.addOrUpdateSocialProfile(r, domain.SocialProfile{
		Label:    "GitHub",
		URL:      profileURL,
		Provider: "github",
		Verified: true,
	}); err != nil {
		http.Error(w, "Failed to save profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

// RedditStart redirects to Reddit OAuth with a stored state token.
func (h Handler) RedditStart(w http.ResponseWriter, r *http.Request) {
	if h.cfg.RedditClientID == "" || h.cfg.RedditClientSecret == "" || h.cfg.BaseURL == "" {
		http.Error(w, "Reddit OAuth not configured", http.StatusBadRequest)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	// Bind the authorization request to a session-scoped state token.
	state := core.RandomToken(16)
	session.Values["oauth_reddit_state"] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	authURL := "https://www.reddit.com/api/v1/authorize"
	params := url.Values{}
	params.Set("client_id", h.cfg.RedditClientID)
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("redirect_uri", h.cfg.BaseURL+"/oauth/reddit/callback")
	params.Set("duration", "permanent")
	params.Set("scope", "identity")
	http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound)
}

// RedditCallback exchanges the code for a token and persists the verified profile.
func (h Handler) RedditCallback(w http.ResponseWriter, r *http.Request) {
	session, _ := h.deps.GetSession(r, "pin_session")
	// Validate state to protect against CSRF in the OAuth callback.
	state, _ := session.Values["oauth_reddit_state"].(string)
	if state == "" || r.URL.Query().Get("state") != state {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	token, err := exchangeRedditToken(h.cfg, code)
	if err != nil {
		http.Error(w, "Reddit auth failed", http.StatusBadRequest)
		return
	}
	_, profileURL, err := fetchRedditProfile(token, h.cfg.RedditUserAgent)
	if err != nil {
		http.Error(w, "Reddit profile fetch failed", http.StatusBadRequest)
		return
	}

	if err := h.addOrUpdateSocialProfile(r, domain.SocialProfile{
		Label:    "Reddit",
		URL:      profileURL,
		Provider: "reddit",
		Verified: true,
	}); err != nil {
		http.Error(w, "Failed to save profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

// BlueskyConnect verifies a Bluesky handle using an app password.
func (h Handler) BlueskyConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.cfg.BlueskyPDS == "" {
		http.Error(w, "Bluesky not configured", http.StatusBadRequest)
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
	handle := strings.TrimSpace(r.FormValue("bsky_handle"))
	appPassword := strings.TrimSpace(r.FormValue("bsky_app_password"))
	if handle == "" || appPassword == "" {
		http.Error(w, "Missing handle or app password", http.StatusBadRequest)
		return
	}
	did, err := verifyBlueskyHandle(h.cfg.BlueskyPDS, handle, appPassword)
	if err != nil {
		http.Error(w, "Bluesky verification failed", http.StatusBadRequest)
		return
	}
	profileURL := "https://bsky.app/profile/" + handle
	if err := h.addOrUpdateSocialProfile(r, domain.SocialProfile{
		Label:    "Bluesky",
		URL:      profileURL,
		Provider: "bluesky",
		Verified: true,
	}); err != nil {
		http.Error(w, "Failed to save profile", http.StatusInternalServerError)
		return
	}
	if err := h.updateATProtoProfile(r, handle, did); err != nil {
		http.Error(w, "Failed to save profile", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

// exchangeGitHubToken exchanges an OAuth code for a GitHub access token.
func exchangeGitHubToken(cfg Config, code string) (string, error) {
	payload := url.Values{}
	payload.Set("client_id", cfg.GitHubClientID)
	payload.Set("client_secret", cfg.GitHubClientSecret)
	payload.Set("code", code)
	payload.Set("redirect_uri", cfg.BaseURL+"/oauth/github/callback")

	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(payload.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.AccessToken == "" {
		return "", errors.New("missing access token")
	}
	return data.AccessToken, nil
}

// fetchGitHubProfile returns the GitHub login and profile URL for a token.
func fetchGitHubProfile(token string) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var data struct {
		Login string `json:"login"`
		HTML  string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", err
	}
	if data.Login == "" || data.HTML == "" {
		return "", "", errors.New("missing profile data")
	}
	return data.Login, data.HTML, nil
}

// exchangeRedditToken exchanges an OAuth code for a Reddit access token.
func exchangeRedditToken(cfg Config, code string) (string, error) {
	payload := url.Values{}
	payload.Set("grant_type", "authorization_code")
	payload.Set("code", code)
	payload.Set("redirect_uri", cfg.BaseURL+"/oauth/reddit/callback")

	req, err := http.NewRequest(http.MethodPost, "https://www.reddit.com/api/v1/access_token", strings.NewReader(payload.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(cfg.RedditClientID, cfg.RedditClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", cfg.RedditUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.AccessToken == "" {
		return "", errors.New("missing access token")
	}
	return data.AccessToken, nil
}

// fetchRedditProfile returns the Reddit username and profile URL for a token.
func fetchRedditProfile(token, userAgent string) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://oauth.reddit.com/api/v1/me", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var data struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", err
	}
	if data.Name == "" {
		return "", "", errors.New("missing profile data")
	}
	return data.Name, "https://www.reddit.com/user/" + data.Name, nil
}

// verifyBlueskyHandle verifies a Bluesky handle and returns its DID.
func verifyBlueskyHandle(pdsURL, handle, appPassword string) (string, error) {
	body := map[string]string{
		"identifier": handle,
		"password":   appPassword,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, pdsURL+"/xrpc/com.atproto.server.createSession", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Handle string `json:"handle"`
		DID    string `json:"did"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if !strings.EqualFold(data.Handle, handle) {
		return "", errors.New("handle mismatch")
	}
	return strings.TrimSpace(data.DID), nil
}

// updateATProtoProfile persists ATProto handle and DID on the identity.
func (h Handler) updateATProtoProfile(r *http.Request, handle, did string) error {
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		return err
	}
	identityRecord, err := h.deps.CurrentIdentity(r)
	if err != nil {
		return err
	}
	handle = strings.TrimSpace(handle)
	did = strings.TrimSpace(did)
	if handle != "" {
		identityRecord.ATProtoHandle = handle
	}
	if did != "" {
		identityRecord.ATProtoDID = did
	}
	target := identityRecord.ATProtoHandle
	if target == "" {
		target = identityRecord.Handle
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "atproto.update", target, nil)
	err = h.deps.UpdateIdentity(r.Context(), identityRecord)
	h.deps.AuditOutcome(r.Context(), current.ID, "atproto.update", target, err, nil)
	return err
}

// addOrUpdateSocialProfile upserts a social profile entry by provider or URL.
func (h Handler) addOrUpdateSocialProfile(r *http.Request, profile domain.SocialProfile) error {
	user, err := h.deps.CurrentUser(r)
	if err != nil {
		return err
	}
	identityRecord, err := h.deps.CurrentIdentity(r)
	if err != nil {
		return err
	}
	profiles := identity.DecodeSocialProfiles(identityRecord.SocialProfilesJSON)

	found := false
	for i, existing := range profiles {
		if strings.EqualFold(existing.Provider, profile.Provider) && existing.Provider != "" {
			profiles[i] = profile
			found = true
			break
		}
		if strings.EqualFold(existing.URL, profile.URL) && existing.URL != "" {
			profiles[i] = profile
			found = true
			break
		}
	}
	if !found {
		profiles = append(profiles, profile)
	}

	payload, err := json.Marshal(profiles)
	if err != nil {
		return err
	}
	identityRecord.SocialProfilesJSON = string(payload)
	target := profile.Provider
	if strings.TrimSpace(target) == "" {
		target = profile.URL
	}
	h.deps.AuditAttempt(r.Context(), user.ID, "social.update", target, nil)
	err = h.deps.UpdateIdentity(r.Context(), identityRecord)
	h.deps.AuditOutcome(r.Context(), user.ID, "social.update", target, err, nil)
	return err
}

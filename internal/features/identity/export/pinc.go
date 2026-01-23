package export

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"pin/internal/domain"
	"pin/internal/features/identity"
)

type pincMeta struct {
	Version string `json:"version"`
	BaseURL string `json:"base_url"`
	View    string `json:"view"`
	Subject string `json:"subject"`
	Rev     string `json:"rev"`
	Self    string `json:"self,omitempty"`
}

type pincIdentity struct {
	Handle          string                 `json:"handle"`
	DisplayName     string                 `json:"display_name"`
	URL             string                 `json:"url"`
	UpdatedAt       string                 `json:"updated_at"`
	Email           string                 `json:"email,omitempty"`
	Bio             string                 `json:"bio,omitempty"`
	Organization    string                 `json:"organization,omitempty"`
	JobTitle        string                 `json:"job_title,omitempty"`
	Birthdate       string                 `json:"birthdate,omitempty"`
	Languages       string                 `json:"languages,omitempty"`
	Phone           string                 `json:"phone,omitempty"`
	Address         string                 `json:"address,omitempty"`
	Location        string                 `json:"location,omitempty"`
	Website         string                 `json:"website,omitempty"`
	Pronouns        string                 `json:"pronouns,omitempty"`
	Timezone        string                 `json:"timezone,omitempty"`
	CustomFields    map[string]string      `json:"custom_fields,omitempty"`
	ProfileImage    string                 `json:"profile_image,omitempty"`
	ImageAltText    string                 `json:"profile_image_alt,omitempty"`
	Links           []domain.Link          `json:"links,omitempty"`
	Social          []domain.SocialProfile `json:"social,omitempty"`
	Wallets         map[string]string      `json:"wallets,omitempty"`
	PublicKeys      map[string]string      `json:"public_keys,omitempty"`
	VerifiedDomains []string               `json:"verified_domains,omitempty"`
	ATProtoHandle   string                 `json:"atproto_handle,omitempty"`
	ATProtoDID      string                 `json:"atproto_did,omitempty"`
}

type pincEnvelope struct {
	Meta     pincMeta     `json:"meta"`
	Identity pincIdentity `json:"identity"`
}

type pincPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type pincIdentityRev struct {
	Handle          string                 `json:"handle"`
	DisplayName     string                 `json:"display_name"`
	URL             string                 `json:"url"`
	UpdatedAt       string                 `json:"updated_at"`
	Email           string                 `json:"email,omitempty"`
	Bio             string                 `json:"bio,omitempty"`
	Organization    string                 `json:"organization,omitempty"`
	JobTitle        string                 `json:"job_title,omitempty"`
	Birthdate       string                 `json:"birthdate,omitempty"`
	Languages       string                 `json:"languages,omitempty"`
	Phone           string                 `json:"phone,omitempty"`
	Address         string                 `json:"address,omitempty"`
	Location        string                 `json:"location,omitempty"`
	Website         string                 `json:"website,omitempty"`
	Pronouns        string                 `json:"pronouns,omitempty"`
	Timezone        string                 `json:"timezone,omitempty"`
	CustomFields    []pincPair             `json:"custom_fields,omitempty"`
	ProfileImage    string                 `json:"profile_image,omitempty"`
	ImageAltText    string                 `json:"profile_image_alt,omitempty"`
	Links           []domain.Link          `json:"links,omitempty"`
	Social          []domain.SocialProfile `json:"social,omitempty"`
	Wallets         []pincPair             `json:"wallets,omitempty"`
	PublicKeys      []pincPair             `json:"public_keys,omitempty"`
	VerifiedDomains []string               `json:"verified_domains,omitempty"`
	ATProtoHandle   string                 `json:"atproto_handle,omitempty"`
	ATProtoDID      string                 `json:"atproto_did,omitempty"`
}

func (h Handler) BuildPINC(ctx context.Context, r *http.Request, user domain.Identity, customFields map[string]string, view string, selfURL string) (pincEnvelope, error) {
	baseURL := h.source.BaseURL(r)
	handle := strings.TrimSpace(user.Handle)
	profileURL := baseURL + "/" + url.PathEscape(handle)
	profileImageURL := baseURL + "/" + url.PathEscape(handle) + "/profile-picture"
	if strings.EqualFold(view, "private") && strings.TrimSpace(selfURL) != "" {
		if privateImage := profileImageFromSelf(selfURL); privateImage != "" {
			profileImageURL = privateImage
		}
	}
	updatedAt := user.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	customFields = identity.StripEmptyMap(customFields)
	if len(customFields) == 0 {
		customFields = nil
	}

	links := identity.DecodeLinks(user.LinksJSON)
	if len(links) == 0 {
		links = nil
	}
	socialProfiles := identity.DecodeSocialProfiles(user.SocialProfilesJSON)
	if len(socialProfiles) == 0 {
		socialProfiles = nil
	}
	wallets := identity.DecodeStringMap(user.WalletsJSON)
	if len(wallets) == 0 {
		wallets = nil
	}
	publicKeys := identity.DecodeStringMap(user.PublicKeysJSON)
	if len(publicKeys) == 0 {
		publicKeys = nil
	}
	verifiedDomains := identity.DecodeStringSlice(user.VerifiedDomainsJSON)
	if len(verifiedDomains) == 0 {
		verifiedDomains = nil
	}

	identityPayload := pincIdentity{
		Handle:          handle,
		DisplayName:     identity.FirstNonEmpty(user.DisplayName, handle),
		URL:             profileURL,
		UpdatedAt:       updatedAt.Format(time.RFC3339),
		Email:           strings.TrimSpace(user.Email),
		Bio:             strings.TrimSpace(user.Bio),
		Organization:    strings.TrimSpace(user.Organization),
		JobTitle:        strings.TrimSpace(user.JobTitle),
		Birthdate:       strings.TrimSpace(user.Birthdate),
		Languages:       strings.TrimSpace(user.Languages),
		Phone:           strings.TrimSpace(user.Phone),
		Address:         strings.TrimSpace(user.Address),
		Location:        strings.TrimSpace(user.Location),
		Website:         strings.TrimSpace(user.Website),
		Pronouns:        strings.TrimSpace(user.Pronouns),
		Timezone:        strings.TrimSpace(user.Timezone),
		CustomFields:    customFields,
		ProfileImage:    profileImageURL,
		ImageAltText:    h.source.ActiveProfilePictureAlt(ctx, user),
		Links:           links,
		Social:          socialProfiles,
		Wallets:         wallets,
		PublicKeys:      publicKeys,
		VerifiedDomains: verifiedDomains,
		ATProtoHandle:   strings.TrimSpace(user.ATProtoHandle),
		ATProtoDID:      strings.TrimSpace(user.ATProtoDID),
	}

	meta := pincMeta{
		Version: identity.PincVersion,
		BaseURL: baseURL,
		View:    view,
		Subject: identity.SubjectForIdentity(user),
		Rev:     computePINCRev(identityPayload),
	}
	if strings.TrimSpace(selfURL) != "" {
		meta.Self = selfURL
	}

	return pincEnvelope{
		Meta:     meta,
		Identity: identityPayload,
	}, nil
}

func profileImageFromSelf(selfURL string) string {
	parsed, err := url.Parse(selfURL)
	if err != nil {
		return ""
	}
	path := strings.TrimSuffix(parsed.Path, ".json")
	if path == "" || path == "/" {
		return ""
	}
	parsed.Path = path + "/profile-picture"
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func (h Handler) ServePINCJSON(w http.ResponseWriter, r *http.Request, user domain.Identity, customFields map[string]string, view string, selfURL string) error {
	payload, err := h.BuildPINC(r.Context(), r, user, customFields, view, selfURL)
	if err != nil {
		return err
	}
	if view == "private" {
		identity.WritePrivateIdentityCacheHeaders(w)
	} else {
		identity.WriteIdentityCacheHeaders(w)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(payload)
}

func computePINCRev(identityPayload pincIdentity) string {
	rev := pincIdentityRev{
		Handle:          identityPayload.Handle,
		DisplayName:     identityPayload.DisplayName,
		URL:             identityPayload.URL,
		UpdatedAt:       identityPayload.UpdatedAt,
		Email:           identityPayload.Email,
		Bio:             identityPayload.Bio,
		Organization:    identityPayload.Organization,
		JobTitle:        identityPayload.JobTitle,
		Birthdate:       identityPayload.Birthdate,
		Languages:       identityPayload.Languages,
		Phone:           identityPayload.Phone,
		Address:         identityPayload.Address,
		Location:        identityPayload.Location,
		Website:         identityPayload.Website,
		Pronouns:        identityPayload.Pronouns,
		Timezone:        identityPayload.Timezone,
		CustomFields:    sortedPairs(identityPayload.CustomFields),
		ProfileImage:    identityPayload.ProfileImage,
		ImageAltText:    identityPayload.ImageAltText,
		Links:           identityPayload.Links,
		Social:          identityPayload.Social,
		Wallets:         sortedPairs(identityPayload.Wallets),
		PublicKeys:      sortedPairs(identityPayload.PublicKeys),
		VerifiedDomains: identityPayload.VerifiedDomains,
		ATProtoHandle:   identityPayload.ATProtoHandle,
		ATProtoDID:      identityPayload.ATProtoDID,
	}
	raw, _ := json.Marshal(rev)
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func sortedPairs(values map[string]string) []pincPair {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		if strings.TrimSpace(key) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]pincPair, 0, len(keys))
	for _, key := range keys {
		value := strings.TrimSpace(values[key])
		if value == "" {
			continue
		}
		out = append(out, pincPair{Key: key, Value: value})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

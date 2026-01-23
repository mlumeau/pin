package export

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pin/internal/domain"
	"pin/internal/features/identity"
)

type identityExport struct {
	XMLName       xml.Name               `xml:"identity" json:"-"`
	Handle        string                 `xml:"handle" json:"handle"`
	DisplayName   string                 `xml:"display_name" json:"display_name"`
	Email         string                 `xml:"email,omitempty" json:"email,omitempty"`
	Bio           string                 `xml:"bio,omitempty" json:"bio,omitempty"`
	Organization  string                 `xml:"organization,omitempty" json:"organization,omitempty"`
	JobTitle      string                 `xml:"job_title,omitempty" json:"job_title,omitempty"`
	Birthdate     string                 `xml:"birthdate,omitempty" json:"birthdate,omitempty"`
	Languages     string                 `xml:"languages,omitempty" json:"languages,omitempty"`
	Phone         string                 `xml:"phone,omitempty" json:"phone,omitempty"`
	Address       string                 `xml:"address,omitempty" json:"address,omitempty"`
	Location      string                 `xml:"location,omitempty" json:"location,omitempty"`
	Website       string                 `xml:"website,omitempty" json:"website,omitempty"`
	Pronouns      string                 `xml:"pronouns,omitempty" json:"pronouns,omitempty"`
	Timezone      string                 `xml:"timezone,omitempty" json:"timezone,omitempty"`
	CustomFields  map[string]string      `xml:"-" json:"custom_fields,omitempty"`
	CustomList    []identityField        `xml:"custom_fields>field,omitempty" json:"-"`
	ProfileURL    string                 `xml:"profile_url" json:"profile_url"`
	ProfileImage  string                 `xml:"profile_image" json:"profile_image"`
	ImageAltText  string                 `xml:"profile_image_alt,omitempty" json:"profile_image_alt,omitempty"`
	Links         []domain.Link          `xml:"links>link,omitempty" json:"links,omitempty"`
	Social        []domain.SocialProfile `xml:"social>profile,omitempty" json:"social,omitempty"`
	Wallets       map[string]string      `xml:"wallets,omitempty" json:"wallets,omitempty"`
	PublicKeys    map[string]string      `xml:"public_keys,omitempty" json:"public_keys,omitempty"`
	Domains       []string               `xml:"verified_domains>domain,omitempty" json:"verified_domains,omitempty"`
	ATProtoHandle string                 `xml:"atproto_handle,omitempty" json:"atproto_handle,omitempty"`
	ATProtoDID    string                 `xml:"atproto_did,omitempty" json:"atproto_did,omitempty"`
	UpdatedAt     string                 `xml:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type identityField struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

// Source provides identity data and helpers.
type Source interface {
	GetOwnerIdentity(ctx context.Context) (domain.Identity, error)
	VisibleIdentity(user domain.Identity, isPrivate bool) (domain.Identity, map[string]string)
	ActiveProfilePictureAlt(ctx context.Context, user domain.Identity) string
	BaseURL(r *http.Request) string
}

type Handler struct {
	source Source
}

func NewHandler(source Source) Handler {
	return Handler{source: source}
}

// Build constructs the identity export payload for a user.
func (h Handler) Build(ctx context.Context, r *http.Request, user domain.Identity, customFields map[string]string, profileURL string) (identityExport, error) {
	links := identity.DecodeLinks(user.LinksJSON)
	socialProfiles := identity.DecodeSocialProfiles(user.SocialProfilesJSON)
	wallets := identity.DecodeStringMap(user.WalletsJSON)
	publicKeys := identity.DecodeStringMap(user.PublicKeysJSON)
	verifiedDomains := identity.DecodeStringSlice(user.VerifiedDomainsJSON)
	customFields = identity.StripEmptyMap(customFields)
	customList := []identityField{}
	for key, value := range customFields {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		customList = append(customList, identityField{Key: key, Value: value})
	}
	if len(customFields) == 0 {
		customFields = nil
		customList = nil
	}
	updatedAt := user.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	return identityExport{
		Handle:        user.Handle,
		DisplayName:   identity.FirstNonEmpty(user.DisplayName, user.Handle),
		Email:         user.Email,
		Bio:           user.Bio,
		Organization:  user.Organization,
		JobTitle:      user.JobTitle,
		Birthdate:     user.Birthdate,
		Languages:     user.Languages,
		Phone:         user.Phone,
		Address:       user.Address,
		Location:      user.Location,
		Website:       user.Website,
		Pronouns:      user.Pronouns,
		Timezone:      user.Timezone,
		CustomFields:  customFields,
		CustomList:    customList,
		ProfileURL:    profileURL,
		ProfileImage:  profileURL + "/profile-picture",
		ImageAltText:  h.source.ActiveProfilePictureAlt(ctx, user),
		Links:         links,
		Social:        socialProfiles,
		Wallets:       wallets,
		PublicKeys:    publicKeys,
		Domains:       verifiedDomains,
		ATProtoHandle: user.ATProtoHandle,
		ATProtoDID:    user.ATProtoDID,
		UpdatedAt:     updatedAt.Format(time.RFC3339),
	}, nil
}

func (h Handler) IdentityJSON(w http.ResponseWriter, r *http.Request) {
	user, err := h.source.GetOwnerIdentity(r.Context())
	if err != nil {
		http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		return
	}
	publicUser, customFields := h.source.VisibleIdentity(user, false)
	selfURL := h.source.BaseURL(r) + r.URL.Path
	if r.URL.RawQuery != "" {
		selfURL += "?" + r.URL.RawQuery
	}
	if err := h.ServePINCJSON(w, r, publicUser, customFields, "public", selfURL); err != nil {
		http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		return
	}
}

func (h Handler) IdentityXML(w http.ResponseWriter, r *http.Request) {
	user, err := h.source.GetOwnerIdentity(r.Context())
	if err != nil {
		http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		return
	}
	publicUser, customFields := h.source.VisibleIdentity(user, false)
	identityExport, err := h.Build(r.Context(), r, publicUser, customFields, h.source.BaseURL(r)+"/"+url.PathEscape(user.Handle))
	if err != nil {
		http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		return
	}
	identity.WriteIdentityCacheHeaders(w)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return
	}
	_ = enc.Encode(identityExport)
}

func (h Handler) IdentityTXT(w http.ResponseWriter, r *http.Request) {
	user, err := h.source.GetOwnerIdentity(r.Context())
	if err != nil {
		http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		return
	}
	publicUser, customFields := h.source.VisibleIdentity(user, false)
	identityExport, err := h.Build(r.Context(), r, publicUser, customFields, h.source.BaseURL(r)+"/"+url.PathEscape(user.Handle))
	if err != nil {
		http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		return
	}
	identity.WriteIdentityCacheHeaders(w)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	lines := []string{
		"handle: " + identityExport.Handle,
		"display_name: " + identityExport.DisplayName,
	}
	if identityExport.Email != "" {
		lines = append(lines, "email: "+identityExport.Email)
	}
	if identityExport.Bio != "" {
		lines = append(lines, "bio: "+identityExport.Bio)
	}
	if identityExport.Organization != "" {
		lines = append(lines, "organization: "+identityExport.Organization)
	}
	if identityExport.JobTitle != "" {
		lines = append(lines, "job_title: "+identityExport.JobTitle)
	}
	if identityExport.Birthdate != "" {
		lines = append(lines, "birthdate: "+identityExport.Birthdate)
	}
	if identityExport.Languages != "" {
		lines = append(lines, "languages: "+identityExport.Languages)
	}
	if identityExport.Phone != "" {
		lines = append(lines, "phone: "+identityExport.Phone)
	}
	if len(identityExport.CustomFields) > 0 {
		for key, value := range identityExport.CustomFields {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
				continue
			}
			lines = append(lines, "custom."+strings.ToLower(key)+": "+value)
		}
	}
	if identityExport.Location != "" {
		lines = append(lines, "location: "+identityExport.Location)
	}
	if identityExport.Website != "" {
		lines = append(lines, "website: "+identityExport.Website)
	}
	if identityExport.Pronouns != "" {
		lines = append(lines, "pronouns: "+identityExport.Pronouns)
	}
	if identityExport.Timezone != "" {
		lines = append(lines, "timezone: "+identityExport.Timezone)
	}
	if len(identityExport.Wallets) > 0 {
		for key, value := range identityExport.Wallets {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
				continue
			}
			lines = append(lines, "wallet."+strings.ToLower(key)+": "+value)
		}
	}
	if len(identityExport.PublicKeys) > 0 {
		for key, value := range identityExport.PublicKeys {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
				continue
			}
			lines = append(lines, "key."+strings.ToLower(key)+": "+value)
		}
	}
	if len(identityExport.Domains) > 0 {
		for _, domain := range identityExport.Domains {
			if strings.TrimSpace(domain) == "" {
				continue
			}
			lines = append(lines, "verified_domain: "+domain)
		}
	}
	if identityExport.ATProtoHandle != "" {
		lines = append(lines, "atproto_handle: "+identityExport.ATProtoHandle)
	}
	if identityExport.ATProtoDID != "" {
		lines = append(lines, "atproto_did: "+identityExport.ATProtoDID)
	}
	if identityExport.ProfileURL != "" {
		lines = append(lines, "profile_url: "+identityExport.ProfileURL)
	}
	if identityExport.ProfileImage != "" {
		lines = append(lines, "profile_image: "+identityExport.ProfileImage)
	}
	if identityExport.ImageAltText != "" {
		lines = append(lines, "profile_image_alt: "+identityExport.ImageAltText)
	}
	if identityExport.UpdatedAt != "" {
		lines = append(lines, "updated_at: "+identityExport.UpdatedAt)
	}
	_, _ = w.Write([]byte(strings.Join(lines, "\n")))
}

func (h Handler) IdentityVCF(w http.ResponseWriter, r *http.Request) {
	user, err := h.source.GetOwnerIdentity(r.Context())
	if err != nil {
		http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		return
	}
	publicUser, customFields := h.source.VisibleIdentity(user, false)
	identityExport, err := h.Build(r.Context(), r, publicUser, customFields, h.source.BaseURL(r)+"/"+url.PathEscape(user.Handle))
	if err != nil {
		http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		return
	}
	identity.WriteIdentityCacheHeaders(w)
	w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
	lines := []string{
		"BEGIN:VCARD",
		"VERSION:4.0",
		"N:;" + identity.EscapeVCard(identityExport.DisplayName) + ";;;",
		"FN:" + identity.EscapeVCard(identityExport.DisplayName),
	}
	if identityExport.Email != "" {
		lines = append(lines, "EMAIL;TYPE=work:"+identity.EscapeVCard(identityExport.Email))
	}
	if identityExport.Phone != "" {
		lines = append(lines, "TEL;TYPE=cell:"+identity.EscapeVCard(identityExport.Phone))
	}
	if identityExport.Organization != "" {
		lines = append(lines, "ORG:"+identity.EscapeVCard(identityExport.Organization))
	}
	if identityExport.JobTitle != "" {
		lines = append(lines, "TITLE:"+identity.EscapeVCard(identityExport.JobTitle))
	}
	if identityExport.Website != "" {
		lines = append(lines, "URL:"+identity.EscapeVCard(identityExport.Website))
	}
	if identityExport.ProfileURL != "" && identityExport.Website == "" {
		lines = append(lines, "URL:"+identity.EscapeVCard(identityExport.ProfileURL))
	}
	if identityExport.Location != "" {
		lines = append(lines, "ADR;TYPE=home:;;"+identity.EscapeVCard(identityExport.Location)+";;;;")
	}
	if identityExport.Address != "" && identityExport.Location == "" {
		lines = append(lines, "ADR;TYPE=home:;;"+identity.EscapeVCard(identityExport.Address)+";;;;")
	}
	if identityExport.ProfileImage != "" {
		lines = append(lines, "PHOTO;MEDIATYPE=image/webp:"+identity.EscapeVCard(identityExport.ProfileImage))
	}
	if identityExport.Bio != "" {
		lines = append(lines, "NOTE:"+identity.EscapeVCard(identityExport.Bio))
	}
	if identityExport.Languages != "" {
		lines = append(lines, "LANG:"+identity.EscapeVCard(identityExport.Languages))
	}
	if len(identityExport.CustomFields) > 0 {
		for key, value := range identityExport.CustomFields {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
				continue
			}
			lines = append(lines, "X-"+identity.SanitizeVCardKey(key)+":"+identity.EscapeVCard(value))
		}
	}
	lines = append(lines, "END:VCARD")
	_, _ = w.Write([]byte(strings.Join(lines, "\r\n")))
}

// ServeIdentity renders an identity export for a given user and extension.
func (h Handler) ServeIdentity(w http.ResponseWriter, r *http.Request, user domain.Identity, customFields map[string]string, profileURL string, ext string, isPrivate bool) error {
	identityExport, err := h.Build(r.Context(), r, user, customFields, profileURL)
	if err != nil {
		return err
	}
	if isPrivate {
		identity.WritePrivateIdentityCacheHeaders(w)
	} else {
		identity.WriteIdentityCacheHeaders(w)
	}
	switch strings.ToLower(ext) {
	case "xml":
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		if _, err := w.Write([]byte(xml.Header)); err != nil {
			return err
		}
		return enc.Encode(identityExport)
	case "txt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		lines := []string{
			"handle: " + identityExport.Handle,
			"display_name: " + identityExport.DisplayName,
		}
		if identityExport.Email != "" {
			lines = append(lines, "email: "+identityExport.Email)
		}
		if identityExport.Bio != "" {
			lines = append(lines, "bio: "+identityExport.Bio)
		}
		if identityExport.Organization != "" {
			lines = append(lines, "organization: "+identityExport.Organization)
		}
		if identityExport.JobTitle != "" {
			lines = append(lines, "job_title: "+identityExport.JobTitle)
		}
		if identityExport.Birthdate != "" {
			lines = append(lines, "birthdate: "+identityExport.Birthdate)
		}
		if identityExport.Languages != "" {
			lines = append(lines, "languages: "+identityExport.Languages)
		}
		if identityExport.Phone != "" {
			lines = append(lines, "phone: "+identityExport.Phone)
		}
		if len(identityExport.CustomFields) > 0 {
			for key, value := range identityExport.CustomFields {
				if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
					continue
				}
				lines = append(lines, "custom."+strings.ToLower(key)+": "+value)
			}
		}
		if identityExport.Location != "" {
			lines = append(lines, "location: "+identityExport.Location)
		}
		if identityExport.Website != "" {
			lines = append(lines, "website: "+identityExport.Website)
		}
		if identityExport.Pronouns != "" {
			lines = append(lines, "pronouns: "+identityExport.Pronouns)
		}
		if identityExport.Timezone != "" {
			lines = append(lines, "timezone: "+identityExport.Timezone)
		}
		if len(identityExport.Wallets) > 0 {
			for key, value := range identityExport.Wallets {
				if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
					continue
				}
				lines = append(lines, "wallet."+strings.ToLower(key)+": "+value)
			}
		}
		if len(identityExport.PublicKeys) > 0 {
			for key, value := range identityExport.PublicKeys {
				if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
					continue
				}
				lines = append(lines, "key."+strings.ToLower(key)+": "+value)
			}
		}
		if len(identityExport.Domains) > 0 {
			for _, domain := range identityExport.Domains {
				if strings.TrimSpace(domain) == "" {
					continue
				}
				lines = append(lines, "verified_domain: "+domain)
			}
		}
		if identityExport.ATProtoHandle != "" {
			lines = append(lines, "atproto_handle: "+identityExport.ATProtoHandle)
		}
		if identityExport.ATProtoDID != "" {
			lines = append(lines, "atproto_did: "+identityExport.ATProtoDID)
		}
		if identityExport.ProfileURL != "" {
			lines = append(lines, "profile_url: "+identityExport.ProfileURL)
		}
		if identityExport.ProfileImage != "" {
			lines = append(lines, "profile_image: "+identityExport.ProfileImage)
		}
		if identityExport.ImageAltText != "" {
			lines = append(lines, "profile_image_alt: "+identityExport.ImageAltText)
		}
		if identityExport.UpdatedAt != "" {
			lines = append(lines, "updated_at: "+identityExport.UpdatedAt)
		}
		_, _ = w.Write([]byte(strings.Join(lines, "\n")))
		return nil
	case "vcf":
		w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
		lines := []string{
			"BEGIN:VCARD",
			"VERSION:4.0",
			"N:;" + identity.EscapeVCard(identityExport.DisplayName) + ";;;",
			"FN:" + identity.EscapeVCard(identityExport.DisplayName),
		}
		if identityExport.Email != "" {
			lines = append(lines, "EMAIL;TYPE=work:"+identity.EscapeVCard(identityExport.Email))
		}
		if identityExport.Phone != "" {
			lines = append(lines, "TEL;TYPE=cell:"+identity.EscapeVCard(identityExport.Phone))
		}
		if identityExport.Organization != "" {
			lines = append(lines, "ORG:"+identity.EscapeVCard(identityExport.Organization))
		}
		if identityExport.JobTitle != "" {
			lines = append(lines, "TITLE:"+identity.EscapeVCard(identityExport.JobTitle))
		}
		if identityExport.Website != "" {
			lines = append(lines, "URL:"+identity.EscapeVCard(identityExport.Website))
		}
		if identityExport.ProfileURL != "" && identityExport.Website == "" {
			lines = append(lines, "URL:"+identity.EscapeVCard(identityExport.ProfileURL))
		}
		if identityExport.Location != "" {
			lines = append(lines, "ADR;TYPE=home:;;"+identity.EscapeVCard(identityExport.Location)+";;;;")
		}
		if identityExport.Address != "" && identityExport.Location == "" {
			lines = append(lines, "ADR;TYPE=home:;;"+identity.EscapeVCard(identityExport.Address)+";;;;")
		}
		if identityExport.ProfileImage != "" {
			lines = append(lines, "PHOTO;MEDIATYPE=image/webp:"+identity.EscapeVCard(identityExport.ProfileImage))
		}
		if identityExport.Bio != "" {
			lines = append(lines, "NOTE:"+identity.EscapeVCard(identityExport.Bio))
		}
		if identityExport.Languages != "" {
			lines = append(lines, "LANG:"+identity.EscapeVCard(identityExport.Languages))
		}
		if len(identityExport.CustomFields) > 0 {
			for key, value := range identityExport.CustomFields {
				if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
					continue
				}
				lines = append(lines, "X-"+identity.SanitizeVCardKey(key)+":"+identity.EscapeVCard(value))
			}
		}
		lines = append(lines, "END:VCARD")
		_, _ = w.Write([]byte(strings.Join(lines, "\r\n")))
		return nil
	default:
		http.NotFound(w, r)
		return nil
	}
}

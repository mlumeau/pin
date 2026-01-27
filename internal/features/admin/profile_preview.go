package admin

import (
	"net/url"
	"strings"
	"time"

	"pin/internal/domain"
	"pin/internal/features/identity"
	featuresettings "pin/internal/features/settings"
	"pin/internal/platform/core"
)

type ProfilePreviewData struct {
	Theme featuresettings.ThemeSettings
	Data  ProfilePreviewModes
}

type ProfilePreviewModes struct {
	Public  ProfilePreviewSnapshot
	Private ProfilePreviewSnapshot
}

type ProfilePreviewSnapshot struct {
	User              domain.Identity
	CustomFields      map[string]string
	Links             []domain.Link
	SocialProfiles    []domain.SocialProfile
	Wallets           []domain.Wallet
	PublicKeys        []domain.PublicKey
	VerifiedDomains   []domain.VerifiedDomain
	ProfilePictureURL string
	ProfilePictureAlt string
	ProfileURL        string
	ExportBase        string
	UpdatedAt         time.Time
}

func buildProfilePreviewData(identityRecord domain.Identity, theme featuresettings.ThemeSettings) ProfilePreviewData {
	updatedAt := identityRecord.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	publicUser, publicCustom := identity.VisibleIdentity(identityRecord, false)
	publicPath := "/" + url.PathEscape(publicUser.Handle)

	privateUser, privateCustom := identity.VisibleIdentity(identityRecord, true)
	privatePath := publicPath
	if strings.TrimSpace(identityRecord.PrivateToken) != "" {
		privateHash := core.ShortHash(strings.ToLower(strings.TrimSpace(identityRecord.Handle)), 7)
		privatePath = "/p/" + url.PathEscape(privateHash) + "/" + url.PathEscape(identityRecord.PrivateToken)
	}

	publicPreview := buildPreviewSnapshot(publicUser, publicCustom, publicPath, updatedAt)
	privatePreview := buildPreviewSnapshot(privateUser, privateCustom, privatePath, updatedAt)

	return ProfilePreviewData{
		Theme: theme,
		Data: ProfilePreviewModes{
			Public:  publicPreview,
			Private: privatePreview,
		},
	}
}

func buildPreviewSnapshot(user domain.Identity, customFields map[string]string, profilePath string, updatedAt time.Time) ProfilePreviewSnapshot {
	return ProfilePreviewSnapshot{
		User:              user,
		CustomFields:      customFields,
		Links:             identity.DecodeLinks(user.LinksJSON),
		SocialProfiles:    identity.DecodeSocialProfiles(user.SocialProfilesJSON),
		Wallets:           identity.WalletsMapToStructs(identity.DecodeStringMap(user.WalletsJSON)),
		PublicKeys:        identity.PublicKeysMapToStructs(identity.DecodeStringMap(user.PublicKeysJSON)),
		VerifiedDomains:   identity.VerifiedDomainsSliceToStructs(identity.DecodeStringSlice(user.VerifiedDomainsJSON)),
		ProfilePictureURL: profilePath + "/profile-picture",
		ProfilePictureAlt: "",
		ProfileURL:        profilePath,
		ExportBase:        profilePath,
		UpdatedAt:         updatedAt,
	}
}

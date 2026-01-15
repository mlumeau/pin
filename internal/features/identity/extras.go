package identity

import (
	"strings"

	"pin/internal/domain"
)

// DomainsToText renders domain verifications as a comma-separated list.
func DomainsToText(rows []domain.DomainVerification) string {
	if len(rows) == 0 {
		return ""
	}
	var parts []string
	for _, row := range rows {
		if strings.TrimSpace(row.Domain) != "" {
			parts = append(parts, row.Domain)
		}
	}
	return strings.Join(parts, ", ")
}

// VerifiedDomains returns verified domain strings.
func VerifiedDomains(rows []domain.DomainVerification) []string {
	var out []string
	for _, row := range rows {
		if row.VerifiedAt.Valid && strings.TrimSpace(row.Domain) != "" {
			out = append(out, row.Domain)
		}
	}
	return out
}

// IsATProtoHandleVerified checks if the handle suffix matches any verified domain.
func IsATProtoHandleVerified(handle string, verified []string) bool {
	handle = strings.ToLower(strings.TrimSpace(handle))
	if handle == "" {
		return false
	}
	for _, domain := range verified {
		d := strings.ToLower(strings.TrimSpace(domain))
		if d == "" {
			continue
		}
		if strings.HasSuffix(handle, d) {
			return true
		}
	}
	return false
}

// MergeSocialProfiles preserves existing verification flags when updating socials.
func MergeSocialProfiles(updated []domain.SocialProfile, existing []domain.SocialProfile) []domain.SocialProfile {
	if len(updated) == 0 {
		return existing
	}
	existingByURL := map[string]domain.SocialProfile{}
	for _, profile := range existing {
		key := strings.ToLower(strings.TrimSpace(profile.URL))
		if key != "" {
			existingByURL[key] = profile
		}
	}
	var out []domain.SocialProfile
	for _, profile := range updated {
		key := strings.ToLower(strings.TrimSpace(profile.URL))
		if key == "" {
			continue
		}
		if prev, ok := existingByURL[key]; ok {
			profile.Verified = prev.Verified
			if profile.Provider == "" {
				profile.Provider = prev.Provider
			}
		}
		out = append(out, profile)
	}
	return out
}

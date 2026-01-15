package identity

import (
	"encoding/json"
	"strings"

	"pin/internal/domain"
)

// VisibleIdentity filters fields based on visibility flags and private mode.
func VisibleIdentity(user domain.User, isPrivate bool) (domain.User, map[string]string) {
	customFields := StripEmptyMap(DecodeStringMap(user.CustomFieldsJSON))
	if isPrivate {
		if len(customFields) == 0 {
			customFields = nil
		}
		return user, customFields
	}
	fieldVisibility := DecodeStringMap(user.VisibilityJSON)
	defaultPrivate := map[string]bool{
		"email":     true,
		"phone":     true,
		"address":   true,
		"birthdate": true,
	}
	customVisibility := DecodeStringMap(user.VisibilityJSON)
	applyVisibilityToStringFields(map[string]*string{
		"email":          &user.Email,
		"organization":   &user.Organization,
		"job_title":      &user.JobTitle,
		"birthdate":      &user.Birthdate,
		"languages":      &user.Languages,
		"phone":          &user.Phone,
		"address":        &user.Address,
		"location":       &user.Location,
		"website":        &user.Website,
		"pronouns":       &user.Pronouns,
		"timezone":       &user.Timezone,
		"atproto_handle": &user.ATProtoHandle,
		"atproto_did":    &user.ATProtoDID,
	}, fieldVisibility, defaultPrivate)

	if user.LinksJSON != "" {
		var links []domain.Link
		if err := json.Unmarshal([]byte(user.LinksJSON), &links); err == nil {
			var out []domain.Link
			for _, link := range links {
				if NormalizeVisibility(link.Visibility) == "private" {
					continue
				}
				out = append(out, link)
			}
			if data, err := json.Marshal(out); err == nil {
				user.LinksJSON = string(data)
			}
		}
	}

	if user.SocialProfilesJSON != "" {
		var socialProfiles []domain.SocialProfile
		if err := json.Unmarshal([]byte(user.SocialProfilesJSON), &socialProfiles); err == nil {
			var out []domain.SocialProfile
			for _, profile := range socialProfiles {
				if NormalizeVisibility(profile.Visibility) == "private" {
					continue
				}
				out = append(out, profile)
			}
			if data, err := json.Marshal(out); err == nil {
				user.SocialProfilesJSON = string(data)
			}
		}
	}

	if user.WalletsJSON != "" {
		if m := DecodeStringMap(user.WalletsJSON); len(m) > 0 {
			for k := range m {
				if NormalizeVisibility(customVisibility["wallet."+k]) == "private" {
					delete(m, k)
				}
			}
			user.WalletsJSON = EncodeStringMap(m)
		}
	}
	if user.PublicKeysJSON != "" {
		if m := DecodeStringMap(user.PublicKeysJSON); len(m) > 0 {
			for k := range m {
				if NormalizeVisibility(customVisibility["key."+k]) == "private" {
					delete(m, k)
				}
			}
			user.PublicKeysJSON = EncodeStringMap(m)
		}
	}

	if user.CustomFieldsJSON != "" {
		if m := DecodeStringMap(user.CustomFieldsJSON); len(m) > 0 {
			for k := range m {
				if NormalizeVisibility(customVisibility["custom."+k]) == "private" {
					delete(m, k)
				}
			}
			user.CustomFieldsJSON = EncodeStringMap(m)
		}
	}

	if user.Email != "" && NormalizeVisibility(fieldVisibility["email"]) == "private" {
		user.Email = ""
	}
	if user.Organization != "" && NormalizeVisibility(fieldVisibility["organization"]) == "private" {
		user.Organization = ""
	}
	if user.JobTitle != "" && NormalizeVisibility(fieldVisibility["job_title"]) == "private" {
		user.JobTitle = ""
	}
	if user.Birthdate != "" && NormalizeVisibility(fieldVisibility["birthdate"]) == "private" {
		user.Birthdate = ""
	}
	if user.Languages != "" && NormalizeVisibility(fieldVisibility["languages"]) == "private" {
		user.Languages = ""
	}
	if user.Phone != "" && NormalizeVisibility(fieldVisibility["phone"]) == "private" {
		user.Phone = ""
	}
	if user.Address != "" && NormalizeVisibility(fieldVisibility["address"]) == "private" {
		user.Address = ""
	}
	if user.Location != "" && NormalizeVisibility(fieldVisibility["location"]) == "private" {
		user.Location = ""
	}
	if user.Website != "" && NormalizeVisibility(fieldVisibility["website"]) == "private" {
		user.Website = ""
	}
	if user.Pronouns != "" && NormalizeVisibility(fieldVisibility["pronouns"]) == "private" {
		user.Pronouns = ""
	}
	if user.Timezone != "" && NormalizeVisibility(fieldVisibility["timezone"]) == "private" {
		user.Timezone = ""
	}
	if user.ATProtoHandle != "" && NormalizeVisibility(fieldVisibility["atproto_handle"]) == "private" {
		user.ATProtoHandle = ""
	}
	if user.ATProtoDID != "" && NormalizeVisibility(fieldVisibility["atproto_did"]) == "private" {
		user.ATProtoDID = ""
	}
	return user, customFields
}

func DecodeStringMap(jsonStr string) map[string]string {
	out := map[string]string{}
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

func EncodeStringMap(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

func DecodeStringSlice(jsonStr string) []string {
	var out []string
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

func EncodeStringSlice(values []string) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

func StripEmptyMap(values map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out[k] = v
	}
	return out
}

func BuildAttachments(user domain.User, wallets, keys map[string]string, domains []string, social []domain.SocialProfile) []map[string]string {
	var attachments []map[string]string
	add := func(name, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		attachments = append(attachments, map[string]string{
			"type":  "PropertyValue",
			"name":  name,
			"value": value,
		})
	}
	add("Location", user.Location)
	add("Website", user.Website)
	add("Pronouns", user.Pronouns)
	add("Timezone", user.Timezone)
	for _, domain := range domains {
		add("Verified Domain", domain)
	}
	for _, key := range []string{"btc", "eth", "sol", "xrp"} {
		add(strings.ToUpper(key), wallets[key])
	}
	add("PGP", keys["pgp"])
	add("SSH", keys["ssh"])
	add("age", keys["age"])
	for _, profile := range social {
		if strings.TrimSpace(profile.URL) == "" {
			continue
		}
		label := strings.TrimSpace(profile.Label)
		if label == "" {
			label = "Social"
		}
		if profile.Verified {
			label = label + " (verified)"
		}
		add(label, profile.URL)
	}
	return attachments
}

func ParseSocialForm(labels, urls, visibilities []string) []domain.SocialProfile {
	var out []domain.SocialProfile
	maxLen := len(labels)
	if len(urls) > maxLen {
		maxLen = len(urls)
	}
	if len(visibilities) > maxLen {
		maxLen = len(visibilities)
	}
	for i := 0; i < maxLen; i++ {
		label := ""
		if i < len(labels) {
			label = strings.TrimSpace(labels[i])
		}
		urlStr := ""
		if i < len(urls) {
			urlStr = strings.TrimSpace(urls[i])
		}
		if label == "" && urlStr == "" {
			continue
		}
		if label == "" || urlStr == "" {
			continue
		}
		visibility := ""
		if i < len(visibilities) {
			visibility = visibilities[i]
		}
		out = append(out, domain.SocialProfile{
			Label:      label,
			URL:        urlStr,
			Visibility: NormalizeVisibility(visibility),
		})
	}
	return out
}

func NormalizeVisibility(input string) string {
	v := strings.ToLower(strings.TrimSpace(input))
	if v == "private" || v == "public" {
		return v
	}
	return "public"
}

func applyVisibilityToStringFields(fields map[string]*string, visibility map[string]string, defaults map[string]bool) {
	for key, ptr := range fields {
		if ptr == nil {
			continue
		}
		value := strings.TrimSpace(*ptr)
		if value == "" {
			continue
		}
		vis := NormalizeVisibility(visibility[key])
		if vis == "" && defaults[key] {
			vis = "private"
		}
		if vis == "private" {
			*ptr = ""
		}
	}
}

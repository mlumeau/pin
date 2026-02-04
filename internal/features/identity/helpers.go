package identity

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"pin/internal/domain"
)

// VisibleIdentity filters fields based on visibility flags and private mode.
func VisibleIdentity(user domain.Identity, isPrivate bool) (domain.Identity, map[string]string) {
	customFields := StripEmptyMap(DecodeStringMap(user.CustomFieldsJSON))
	if isPrivate {
		if len(customFields) == 0 {
			customFields = nil
		}
		return user, customFields
	}
	fieldVisibility := DecodeVisibilityMap(user.VisibilityJSON)
	defaultPrivate := map[string]bool{
		"email":     true,
		"phone":     true,
		"address":   true,
		"birthdate": true,
	}
	customVisibility := fieldVisibility
	applyVisibilityToStringFields(map[string]*string{
		"display_name":   &user.DisplayName,
		"bio":            &user.Bio,
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
		links := DecodeLinks(user.LinksJSON)
		if len(links) > 0 {
			var out []domain.Link
			for i, link := range links {
				if isVisibilityPrivate(customVisibility, LinkVisibilityKey(i), "") {
					continue
				}
				out = append(out, link)
			}
			user.LinksJSON = EncodeLinks(out)
		}
	}

	if user.SocialProfilesJSON != "" {
		socialProfiles := DecodeSocialProfiles(user.SocialProfilesJSON)
		if len(socialProfiles) > 0 {
			var out []domain.SocialProfile
			for i, profile := range socialProfiles {
				if isVisibilityPrivate(customVisibility, SocialVisibilityKey(i), "") {
					continue
				}
				out = append(out, profile)
			}
			user.SocialProfilesJSON = EncodeSocialProfiles(out)
		}
	}

	if user.WalletsJSON != "" {
		if m := DecodeStringMap(user.WalletsJSON); len(m) > 0 {
			for k := range m {
				if isVisibilityPrivate(customVisibility, "wallet."+strings.ToLower(k), "") {
					delete(m, k)
				}
			}
			user.WalletsJSON = EncodeStringMap(m)
		}
	}
	if user.PublicKeysJSON != "" {
		if m := DecodeStringMap(user.PublicKeysJSON); len(m) > 0 {
			for k := range m {
				if isVisibilityPrivate(customVisibility, "key."+strings.ToLower(k), "") {
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

	if user.DisplayName != "" && NormalizeVisibility(fieldVisibility["display_name"]) == "private" {
		user.DisplayName = ""
	}
	if user.Bio != "" && NormalizeVisibility(fieldVisibility["bio"]) == "private" {
		user.Bio = ""
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

// DecodeStringMap decodes a JSON string map into a Go map.
func DecodeStringMap(jsonStr string) map[string]string {
	out := map[string]string{}
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

// EncodeStringMap encodes a string map as JSON.
func EncodeStringMap(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

// DecodeVisibilityMap decodes JSON visibility entries into a key/value map.
func DecodeVisibilityMap(jsonStr string) map[string]string {
	out := map[string]string{}
	if jsonStr == "" {
		return out
	}
	var entries []domain.Visibility
	if err := json.Unmarshal([]byte(jsonStr), &entries); err != nil {
		return out
	}
	for _, entry := range entries {
		key := strings.TrimSpace(entry.Key)
		if key == "" {
			continue
		}
		out[key] = NormalizeVisibility(entry.Visibility)
	}
	return out
}

// EncodeVisibilityMap encodes a visibility map as a stable JSON list.
func EncodeVisibilityMap(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	entries := make([]domain.Visibility, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, domain.Visibility{
			Key:        key,
			Visibility: NormalizeVisibility(values[key]),
		})
	}
	if data, err := json.Marshal(entries); err == nil {
		return string(data)
	}
	return ""
}

// DecodeStringSlice decodes a JSON string slice.
func DecodeStringSlice(jsonStr string) []string {
	var out []string
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

// EncodeStringSlice encodes a string slice as JSON.
func EncodeStringSlice(values []string) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

// DecodeCustomFields decodes JSON custom fields into structs.
func DecodeCustomFields(jsonStr string) []domain.CustomField {
	var out []domain.CustomField
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

// EncodeCustomFields encodes custom field structs as JSON.
func EncodeCustomFields(values []domain.CustomField) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

// DecodeWallets decodes JSON wallet entries into structs.
func DecodeWallets(jsonStr string) []domain.Wallet {
	var out []domain.Wallet
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

// EncodeWallets encodes wallet structs as JSON.
func EncodeWallets(values []domain.Wallet) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

// DecodePublicKeys decodes JSON public keys into structs.
func DecodePublicKeys(jsonStr string) []domain.PublicKey {
	var out []domain.PublicKey
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

// EncodePublicKeys encodes public key structs as JSON.
func EncodePublicKeys(values []domain.PublicKey) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

// DecodeVerifiedDomains decodes JSON verified domains into structs.
func DecodeVerifiedDomains(jsonStr string) []domain.VerifiedDomain {
	var out []domain.VerifiedDomain
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

// EncodeVerifiedDomains encodes verified domain structs as JSON.
func EncodeVerifiedDomains(values []domain.VerifiedDomain) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

// DecodeLinks decodes links from a string representation.
func DecodeLinks(jsonStr string) []domain.Link {
	var out []domain.Link
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

// EncodeLinks encodes links into a string representation.
func EncodeLinks(values []domain.Link) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

// DecodeSocialProfiles decodes social profiles from a string representation.
func DecodeSocialProfiles(jsonStr string) []domain.SocialProfile {
	var out []domain.SocialProfile
	if jsonStr == "" {
		return out
	}
	_ = json.Unmarshal([]byte(jsonStr), &out)
	return out
}

// EncodeSocialProfiles encodes social profiles into a string representation.
func EncodeSocialProfiles(values []domain.SocialProfile) string {
	if len(values) == 0 {
		return ""
	}
	if data, err := json.Marshal(values); err == nil {
		return string(data)
	}
	return ""
}

// WalletsMapToStructs converts a string map into wallet structs, skipping empty entries.
func WalletsMapToStructs(wallets map[string]string) []domain.Wallet {
	if len(wallets) == 0 {
		return nil
	}
	var out []domain.Wallet
	for label, address := range wallets {
		if strings.TrimSpace(label) == "" || strings.TrimSpace(address) == "" {
			continue
		}
		out = append(out, domain.Wallet{
			Label:   label,
			Address: address,
		})
	}
	return out
}

// PublicKeysMapToStructs converts a key map into public key structs in preferred order.
func PublicKeysMapToStructs(keys map[string]string) []domain.PublicKey {
	if len(keys) == 0 {
		return nil
	}
	var out []domain.PublicKey
	algorithms := []string{"pgp", "ssh", "age", "activitypub"}
	for _, algo := range algorithms {
		if key := strings.TrimSpace(keys[algo]); key != "" {
			out = append(out, domain.PublicKey{
				Algorithm: algo,
				Key:       key,
			})
		}
	}
	return out
}

// VerifiedDomainsSliceToStructs converts a string slice into verified domain structs.
func VerifiedDomainsSliceToStructs(domains []string) []domain.VerifiedDomain {
	if len(domains) == 0 {
		return nil
	}
	var out []domain.VerifiedDomain
	for _, d := range domains {
		if strings.TrimSpace(d) == "" {
			continue
		}
		out = append(out, domain.VerifiedDomain{
			Domain: d,
		})
	}
	return out
}

// StripEmptyMap filters out empty values from a string map.
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

// BuildAttachments builds ActivityPub attachment payloads for public identity fields.
func BuildAttachments(user domain.Identity, wallets, keys map[string]string, domains []string, social []domain.SocialProfile) []map[string]string {
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

// ParseSocialForm parses social profiles and visibility from form input arrays.
func ParseSocialForm(labels, urls, visibilities []string) ([]domain.SocialProfile, map[string]string) {
	var out []domain.SocialProfile
	visibility := map[string]string{}
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
		vis := ""
		if i < len(visibilities) {
			vis = visibilities[i]
		}
		index := len(out)
		out = append(out, domain.SocialProfile{
			Label: label,
			URL:   urlStr,
		})
		visibility[SocialVisibilityKey(index)] = NormalizeVisibility(vis)
	}
	return out, visibility
}

// NormalizeVisibility normalizes visibility values to "public" or "private".
func NormalizeVisibility(input string) string {
	v := strings.ToLower(strings.TrimSpace(input))
	if v == "private" || v == "public" {
		return v
	}
	return "public"
}

// applyVisibilityToStringFields applies visibility to string fields to the target.
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

// isVisibilityPrivate reports whether visibility private is true.
func isVisibilityPrivate(values map[string]string, key string, fallback string) bool {
	if key != "" {
		if vis, ok := values[key]; ok {
			return NormalizeVisibility(vis) == "private"
		}
	}
	return fallback == "private"
}

// LinkVisibilityKey returns the visibility map key for a link index.
func LinkVisibilityKey(index int) string {
	return "link:" + strconv.Itoa(index)
}

// SocialVisibilityKey returns the visibility map key for a social profile index.
func SocialVisibilityKey(index int) string {
	return "social:" + strconv.Itoa(index)
}

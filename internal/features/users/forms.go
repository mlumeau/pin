package users

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"pin/internal/domain"
	"pin/internal/features/identity"
	"pin/internal/platform/core"
)

type WalletEntry struct {
	Label      string
	Address    string
	Visibility string
}

type LinkEntry struct {
	Label      string
	URL        string
	Visibility string
}

type SocialEntry struct {
	Label      string
	URL        string
	Provider   string
	Verified   bool
	Visibility string
}

// BuildWalletEntries builds wallet entries from the supplied inputs.
func BuildWalletEntries(wallets map[string]string, visibility map[string]string) []WalletEntry {
	if wallets == nil {
		return []WalletEntry{}
	}
	keys := make([]string, 0, len(wallets))
	for key := range wallets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]WalletEntry, 0, len(keys))
	for _, key := range keys {
		val := strings.TrimSpace(wallets[key])
		if val == "" {
			continue
		}
		visKey := "wallet_" + strings.ToUpper(strings.TrimSpace(key))
		out = append(out, WalletEntry{
			Label:      strings.ToUpper(strings.TrimSpace(key)),
			Address:    val,
			Visibility: visibilityValue(visibility, visKey, nil),
		})
	}
	return out
}

// BuildLinkEntries builds link entries from the supplied inputs.
func BuildLinkEntries(links []domain.Link, visibility map[string]string) []LinkEntry {
	if len(links) == 0 {
		return []LinkEntry{}
	}
	out := make([]LinkEntry, 0, len(links))
	for i, link := range links {
		if strings.TrimSpace(link.Label) == "" || strings.TrimSpace(link.URL) == "" {
			continue
		}
		visKey := identity.LinkVisibilityKey(i)
		out = append(out, LinkEntry{
			Label:      link.Label,
			URL:        link.URL,
			Visibility: visibilityValue(visibility, visKey, nil),
		})
	}
	return out
}

// BuildSocialEntries builds social entries from the supplied inputs.
func BuildSocialEntries(social []domain.SocialProfile, visibility map[string]string) []SocialEntry {
	if len(social) == 0 {
		return []SocialEntry{}
	}
	out := make([]SocialEntry, 0, len(social))
	for i, profile := range social {
		if strings.TrimSpace(profile.URL) == "" {
			continue
		}
		if strings.TrimSpace(profile.Label) == "" && strings.TrimSpace(profile.Provider) == "" {
			continue
		}
		visKey := identity.SocialVisibilityKey(i)
		out = append(out, SocialEntry{
			Label:      profile.Label,
			URL:        profile.URL,
			Provider:   profile.Provider,
			Verified:   profile.Verified,
			Visibility: visibilityValue(visibility, visKey, nil),
		})
	}
	return out
}

// ParseLinksForm parses links form from the provided input.
func ParseLinksForm(labels, urls, visibilities []string) ([]domain.Link, map[string]string) {
	var links []domain.Link
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
		index := len(links)
		links = append(links, domain.Link{Label: label, URL: urlStr})
		visibility[identity.LinkVisibilityKey(index)] = NormalizeVisibility(vis)
	}
	return links, visibility
}

// ParseCustomFieldsForm parses custom fields form from the provided input.
func ParseCustomFieldsForm(keys, values []string) map[string]string {
	out := map[string]string{}
	maxLen := len(keys)
	if len(values) > maxLen {
		maxLen = len(values)
	}
	for i := 0; i < maxLen; i++ {
		key := ""
		if i < len(keys) {
			key = strings.TrimSpace(keys[i])
		}
		value := ""
		if i < len(values) {
			value = strings.TrimSpace(values[i])
		}
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	return out
}

// ParseVisibilityForm parses visibility form from the provided input.
func ParseVisibilityForm(values map[string][]string, fields []string) map[string]string {
	out := map[string]string{}
	for _, field := range fields {
		key := "visibility_" + field
		val := ""
		if v, ok := values[key]; ok && len(v) > 0 {
			val = strings.TrimSpace(v[0])
		}
		out[field] = NormalizeVisibility(val)
	}
	return out
}

// NormalizeVisibility normalizes visibility values to "public" or "private".
func NormalizeVisibility(value string) string {
	value = strings.TrimSpace(value)
	if value != "private" && value != "public" {
		return "public"
	}
	return value
}

// BuildVisibilityMap builds visibility map from the supplied inputs.
func BuildVisibilityMap(fieldVisibility map[string]string, customVisibility map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range fieldVisibility {
		if strings.TrimSpace(key) == "" {
			continue
		}
		out[key] = NormalizeVisibility(value)
	}
	for key, value := range customVisibility {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out["custom:"+key] = NormalizeVisibility(value)
	}
	return out
}

// VisibilityCustomMap returns visibility entries scoped to custom fields.
func VisibilityCustomMap(visibility map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range visibility {
		if !strings.HasPrefix(key, "custom:") {
			continue
		}
		base := strings.TrimPrefix(key, "custom:")
		if strings.TrimSpace(base) == "" {
			continue
		}
		out[base] = NormalizeVisibility(value)
	}
	return out
}

// DomainVisibilityMap returns visibility entries scoped to verified domains.
func DomainVisibilityMap(visibility map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range visibility {
		if !strings.HasPrefix(key, "verified_domain:") {
			continue
		}
		base := strings.TrimPrefix(key, "verified_domain:")
		base = core.NormalizeDomain(base)
		if base == "" {
			continue
		}
		out[base] = NormalizeVisibility(value)
	}
	return out
}

// ParseCustomVisibilityForm parses custom visibility form from the provided input.
func ParseCustomVisibilityForm(keys, values, visibilities []string) map[string]string {
	out := map[string]string{}
	maxLen := len(keys)
	if len(values) > maxLen {
		maxLen = len(values)
	}
	if len(visibilities) > maxLen {
		maxLen = len(visibilities)
	}
	for i := 0; i < maxLen; i++ {
		key := ""
		if i < len(keys) {
			key = strings.TrimSpace(keys[i])
		}
		value := ""
		if i < len(values) {
			value = strings.TrimSpace(values[i])
		}
		if key == "" || value == "" {
			continue
		}
		vis := "public"
		if i < len(visibilities) {
			vis = strings.TrimSpace(visibilities[i])
		}
		if vis != "private" && vis != "public" {
			vis = "public"
		}
		out[key] = vis
	}
	return out
}

// FilterCustomVisibility filters custom visibility using the provided criteria.
func FilterCustomVisibility(fields, visibility map[string]string) map[string]string {
	out := map[string]string{}
	for key := range fields {
		vis := "public"
		if visibility != nil {
			if v, ok := visibility[key]; ok {
				vis = v
			}
		}
		if vis != "private" && vis != "public" {
			vis = "public"
		}
		out[key] = vis
	}
	return out
}

// ParseWalletForm parses wallet form from the provided input.
func ParseWalletForm(labels, addresses, visibilities []string) (map[string]string, map[string]string, error) {
	out := map[string]string{}
	visibility := map[string]string{}
	maxLen := len(labels)
	if len(addresses) > maxLen {
		maxLen = len(addresses)
	}
	if len(visibilities) > maxLen {
		maxLen = len(visibilities)
	}
	seen := map[string]bool{}
	for i := 0; i < maxLen; i++ {
		label := ""
		if i < len(labels) {
			label = strings.TrimSpace(labels[i])
		}
		address := ""
		if i < len(addresses) {
			address = strings.TrimSpace(addresses[i])
		}
		if label == "" && address == "" {
			continue
		}
		if label == "" || address == "" {
			continue
		}
		label = strings.ToUpper(label)
		if seen[label] {
			return nil, nil, errors.New("Wallet labels must be unique")
		}
		seen[label] = true
		out[label] = address
		vis := "public"
		if i < len(visibilities) {
			vis = NormalizeVisibility(visibilities[i])
		}
		visibility["wallet_"+label] = vis
	}
	return out, visibility, nil
}

// ParseVerifiedDomainsText parses verified domains text from the provided input.
func ParseVerifiedDomainsText(input string) []string {
	var domains []string
	for _, line := range strings.Split(strings.TrimSpace(input), "\n") {
		line = core.NormalizeDomain(line)
		if line == "" {
			continue
		}
		domains = append(domains, line)
	}
	return domains
}

// ParseVerifiedDomainVisibilityForm parses verified domain visibility form from the provided input.
func ParseVerifiedDomainVisibilityForm(domains, visibilities []string) map[string]string {
	out := map[string]string{}
	maxLen := len(domains)
	if len(visibilities) > maxLen {
		maxLen = len(visibilities)
	}
	for i := 0; i < maxLen; i++ {
		domain := ""
		if i < len(domains) {
			domain = core.NormalizeDomain(domains[i])
		}
		if domain == "" {
			continue
		}
		visibility := ""
		if i < len(visibilities) {
			visibility = visibilities[i]
		}
		out[domain] = NormalizeVisibility(visibility)
	}
	return out
}

// VerifiedDomainsToText renders verified domains as newline-delimited text.
func VerifiedDomainsToText(jsonStr string) string {
	var domains []string
	if jsonStr == "" {
		return ""
	}
	if err := json.Unmarshal([]byte(jsonStr), &domains); err != nil {
		return ""
	}
	return strings.Join(domains, "\n")
}

// visibilityValue returns a normalized visibility value with defaults applied.
func visibilityValue(visibility map[string]string, field string, defaultPrivate map[string]bool) string {
	val, ok := visibility[field]
	if !ok && defaultPrivate != nil && defaultPrivate[field] {
		val = "private"
	}
	return NormalizeVisibility(val)
}

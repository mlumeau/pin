package users

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"pin/internal/domain"
	"pin/internal/platform/core"
)

type WalletEntry struct {
	Label      string
	Address    string
	Visibility string
}

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

func ParseLinksForm(labels, urls, visibilities []string) []domain.Link {
	var links []domain.Link
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
		links = append(links, domain.Link{Label: label, URL: urlStr, Visibility: NormalizeVisibility(visibility)})
	}
	return links
}

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

func NormalizeVisibility(value string) string {
	value = strings.TrimSpace(value)
	if value != "private" && value != "public" {
		return "public"
	}
	return value
}

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

func ParseAliasesText(input string) []string {
	var aliases []string
	for _, line := range strings.Split(strings.TrimSpace(input), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		aliases = append(aliases, line)
	}
	return aliases
}

func AliasesToText(jsonStr string) string {
	var aliases []string
	if jsonStr == "" {
		return ""
	}
	if err := json.Unmarshal([]byte(jsonStr), &aliases); err != nil {
		return ""
	}
	return strings.Join(aliases, "\n")
}

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

func visibilityValue(visibility map[string]string, field string, defaultPrivate map[string]bool) string {
	val, ok := visibility[field]
	if !ok && defaultPrivate != nil && defaultPrivate[field] {
		val = "private"
	}
	return NormalizeVisibility(val)
}

package identity

import (
	"html/template"
	"strings"
)

// TemplateFuncs exposes identity-specific template helpers.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"isKnownWallet": func(label string) bool {
			label = strings.ToUpper(strings.TrimSpace(label))
			if label == "" {
				return false
			}
			switch label {
			case "BTC", "ETH", "USDT", "USDC", "BNB", "SOL", "XRP", "ADA", "DOGE", "TRX", "MATIC", "DOT":
				return true
			default:
				return false
			}
		},
		"walletSelectValue": func(label string) string {
			label = strings.ToUpper(strings.TrimSpace(label))
			if label == "" {
				return "OTHER"
			}
			switch label {
			case "BTC", "ETH", "USDT", "USDC", "BNB", "SOL", "XRP", "ADA", "DOGE", "TRX", "MATIC", "DOT":
				return label
			default:
				return "OTHER"
			}
		},
	}
}

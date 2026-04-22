package service

import "strings"

// CORSAllowedOriginsSettingKey stores a comma-separated list of additional
// origins (beyond cfg.SiteURL) that the CORS middleware will accept for
// credentialed cross-origin requests. Operators edit this from the admin
// settings UI whenever they run the admin console on a different host than
// the API server (e.g. a staging subdomain, a custom desktop client).
const CORSAllowedOriginsSettingKey = "cors_allowed_origins"

// ParseCORSAllowedOrigins splits the persisted comma/whitespace/newline
// separated value into discrete origin strings, drops the empty ones and
// preserves the original casing for later scheme/host normalization by the
// middleware.
func ParseCORSAllowedOrigins(raw string) []string {
	if raw == "" {
		return nil
	}
	// Admin-entered values commonly use commas, spaces, or newlines. Accept
	// all three so operators don't have to pick a specific separator.
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '\n', '\r', '\t', ' ':
			return true
		}
		return false
	})
	var result []string
	for _, field := range fields {
		field = strings.TrimRight(strings.TrimSpace(field), "/")
		if field != "" {
			result = append(result, field)
		}
	}
	return result
}

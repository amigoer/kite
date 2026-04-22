package service

import "strings"

// JWTSecretSettingKey is the settings key that stores the JWT signing secret.
// It MUST NEVER be exposed via the admin GET /settings endpoint, and admins
// MUST NOT be allowed to overwrite it through PUT /settings — rotation is a
// deliberate operator action, not a UI toggle.
const JWTSecretSettingKey = "jwt_secret"

// oauthClientSecretSuffix matches any per-provider OAuth client secret key
// produced by OAuthConfigService.clientSecretKey, i.e. "oauth_<provider>_client_secret".
const (
	oauthSettingsPrefix       = "oauth_"
	oauthClientSecretSuffix   = "_client_secret"
	secretConfiguredKeySuffix = "_configured"
)

// IsSecretSettingKey reports whether a settings key stores sensitive material
// (JWT secret, SMTP password, OAuth client secrets) that must never leave the
// server. Callers use this to strip values from GET responses and to surface
// a boolean "<key>_configured" flag to the UI instead.
func IsSecretSettingKey(key string) bool {
	switch key {
	case JWTSecretSettingKey, SMTPPasswordSettingKey:
		return true
	}
	return isOAuthClientSecretKey(key)
}

// IsReadOnlySecretSettingKey reports whether a secret key is locked against
// writes through the generic PUT /settings endpoint. Rotating the JWT secret
// via the settings UI would sign new tokens with an attacker-chosen value and
// immediately invalidate every active session.
func IsReadOnlySecretSettingKey(key string) bool {
	return key == JWTSecretSettingKey
}

// SecretConfiguredKey returns the sentinel key used in GET responses to tell
// the UI whether a secret has been configured without exposing its value.
// Example: smtp_password -> smtp_password_configured.
func SecretConfiguredKey(key string) string {
	return key + secretConfiguredKeySuffix
}

// IsSecretConfiguredKey reports whether a key was emitted by SecretConfiguredKey
// and therefore must be stripped from PUT /settings payloads.
func IsSecretConfiguredKey(key string) bool {
	return strings.HasSuffix(key, secretConfiguredKeySuffix)
}

func isOAuthClientSecretKey(key string) bool {
	// Key must parse as "oauth_<provider>_client_secret" with a non-empty
	// provider segment AND both underscores intact. strings.CutPrefix +
	// strings.CutSuffix enforce that both borders were actually consumed, so
	// "oauth_client_secret" (missing provider slot) is correctly rejected.
	middle, ok := strings.CutPrefix(key, oauthSettingsPrefix)
	if !ok {
		return false
	}
	provider, ok := strings.CutSuffix(middle, oauthClientSecretSuffix)
	if !ok {
		return false
	}
	return provider != ""
}

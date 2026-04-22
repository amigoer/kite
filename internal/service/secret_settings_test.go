package service

import "testing"

func TestIsSecretSettingKey(t *testing.T) {
	cases := []struct {
		key  string
		want bool
	}{
		{JWTSecretSettingKey, true},
		{SMTPPasswordSettingKey, true},
		{"oauth_github_client_secret", true},
		{"oauth_google_client_secret", true},
		{"oauth_wechat_client_secret", true},
		{"oauth_github_client_id", false},
		{"oauth_github_enabled", false},
		{"oauth__client_secret", false},
		{"oauth_client_secret", false},
		{"_client_secret", false},
		{SMTPHostSettingKey, false},
		{SMTPUsernameSettingKey, false},
		{SiteNameSettingKey, false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IsSecretSettingKey(tc.key); got != tc.want {
			t.Fatalf("IsSecretSettingKey(%q) = %v, want %v", tc.key, got, tc.want)
		}
	}
}

func TestIsReadOnlySecretSettingKey(t *testing.T) {
	if !IsReadOnlySecretSettingKey(JWTSecretSettingKey) {
		t.Fatal("jwt_secret must be read-only via /settings")
	}
	if IsReadOnlySecretSettingKey(SMTPPasswordSettingKey) {
		t.Fatal("smtp_password must remain writable so operators can configure it")
	}
	if IsReadOnlySecretSettingKey("oauth_github_client_secret") {
		t.Fatal("oauth secrets are written via /settings today; only jwt_secret is locked")
	}
}

func TestSecretConfiguredKey(t *testing.T) {
	if got := SecretConfiguredKey(SMTPPasswordSettingKey); got != SMTPPasswordConfiguredSettingKey {
		t.Fatalf("SecretConfiguredKey(smtp_password) = %q, want %q", got, SMTPPasswordConfiguredSettingKey)
	}
	if got := SecretConfiguredKey(JWTSecretSettingKey); got != "jwt_secret_configured" {
		t.Fatalf("SecretConfiguredKey(jwt_secret) = %q, want jwt_secret_configured", got)
	}
}

func TestIsSecretConfiguredKey(t *testing.T) {
	if !IsSecretConfiguredKey(SMTPPasswordConfiguredSettingKey) {
		t.Fatal("smtp_password_configured must be recognized")
	}
	if !IsSecretConfiguredKey("jwt_secret_configured") {
		t.Fatal("jwt_secret_configured must be recognized")
	}
	if IsSecretConfiguredKey(SMTPPasswordSettingKey) {
		t.Fatal("smtp_password is the secret itself, not the sentinel")
	}
	if IsSecretConfiguredKey(SiteNameSettingKey) {
		t.Fatal("site_name must not be treated as a configured-flag")
	}
}

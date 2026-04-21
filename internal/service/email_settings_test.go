package service

import "testing"

func TestResolveSMTPConfig_DefaultPortAndTLS(t *testing.T) {
	cfg, err := ResolveSMTPConfig(map[string]string{
		SMTPHostSettingKey: " smtp.example.com ",
		SMTPFromSettingKey: "no-reply@example.com",
		SMTPTLSSettingKey:  "true",
	})
	if err != nil {
		t.Fatalf("ResolveSMTPConfig: %v", err)
	}
	if cfg.Host != "smtp.example.com" {
		t.Fatalf("unexpected host: %q", cfg.Host)
	}
	if cfg.Port != 587 {
		t.Fatalf("unexpected port: %d", cfg.Port)
	}
	if !cfg.UseTLS {
		t.Fatal("expected TLS to be enabled")
	}
}

func TestResolveSMTPConfig_RejectsMissingHost(t *testing.T) {
	_, err := ResolveSMTPConfig(map[string]string{
		SMTPFromSettingKey: "no-reply@example.com",
	})
	if err == nil {
		t.Fatal("expected missing host error")
	}
}

func TestResolveSMTPConfig_RejectsPasswordWithoutUsername(t *testing.T) {
	_, err := ResolveSMTPConfig(map[string]string{
		SMTPHostSettingKey:     "smtp.example.com",
		SMTPFromSettingKey:     "no-reply@example.com",
		SMTPPasswordSettingKey: "secret",
	})
	if err == nil {
		t.Fatal("expected username validation error")
	}
}

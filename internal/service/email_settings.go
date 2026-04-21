package service

import (
	"fmt"
	"net/mail"
	"strconv"
	"strings"
)

const (
	SMTPHostSettingKey               = "smtp_host"
	SMTPPortSettingKey               = "smtp_port"
	SMTPTLSSettingKey                = "smtp_tls"
	SMTPFromSettingKey               = "smtp_from"
	SMTPUsernameSettingKey           = "smtp_username"
	SMTPPasswordSettingKey           = "smtp_password"
	SMTPPasswordConfiguredSettingKey = "smtp_password_configured"
	defaultSMTPPort                  = "587"
)

type SMTPConfig struct {
	Host     string
	Port     int
	UseTLS   bool
	From     string
	Username string
	Password string
}

func DefaultSMTPPort() string {
	return defaultSMTPPort
}

func NormalizeSMTPHost(raw string) string {
	return strings.TrimSpace(raw)
}

func NormalizeSMTPPort(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	port, err := strconv.Atoi(trimmed)
	if err != nil {
		return "", fmt.Errorf("must be a valid port number")
	}
	if port < 1 || port > 65535 {
		return "", fmt.Errorf("must be between 1 and 65535")
	}

	return strconv.Itoa(port), nil
}

func NormalizeSMTPBool(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "false", nil
	}

	switch strings.ToLower(trimmed) {
	case "true":
		return "true", nil
	case "false":
		return "false", nil
	default:
		return "", fmt.Errorf("must be true or false")
	}
}

func NormalizeSMTPFrom(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return "", fmt.Errorf("must be a valid email address")
	}

	return addr.Address, nil
}

func NormalizeSMTPUsername(raw string) string {
	return strings.TrimSpace(raw)
}

func ResolveSMTPConfig(values map[string]string) (*SMTPConfig, error) {
	host := NormalizeSMTPHost(values[SMTPHostSettingKey])
	if host == "" {
		return nil, fmt.Errorf("请先填写 SMTP 服务器")
	}

	portValue := values[SMTPPortSettingKey]
	if strings.TrimSpace(portValue) == "" {
		portValue = DefaultSMTPPort()
	}
	normalizedPort, err := NormalizeSMTPPort(portValue)
	if err != nil {
		return nil, fmt.Errorf("SMTP 端口%s", err.Error())
	}
	port, _ := strconv.Atoi(normalizedPort)

	tlsValue, err := NormalizeSMTPBool(values[SMTPTLSSettingKey])
	if err != nil {
		return nil, fmt.Errorf("TLS 配置%s", err.Error())
	}

	from, err := NormalizeSMTPFrom(values[SMTPFromSettingKey])
	if err != nil {
		return nil, fmt.Errorf("发件地址%s", err.Error())
	}
	if from == "" {
		return nil, fmt.Errorf("请先填写发件地址")
	}

	username := NormalizeSMTPUsername(values[SMTPUsernameSettingKey])
	password := values[SMTPPasswordSettingKey]
	if username == "" && strings.TrimSpace(password) != "" {
		return nil, fmt.Errorf("填写 SMTP 密码前请先填写用户名")
	}

	return &SMTPConfig{
		Host:     host,
		Port:     port,
		UseTLS:   tlsValue == "true",
		From:     from,
		Username: username,
		Password: password,
	}, nil
}

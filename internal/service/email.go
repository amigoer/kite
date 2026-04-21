package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) SendTestEmail(ctx context.Context, cfg SMTPConfig, to, siteName string) error {
	name := strings.TrimSpace(siteName)
	if name == "" {
		name = "Kite"
	}

	subject := fmt.Sprintf("[%s] SMTP 测试邮件", name)
	body := strings.Join([]string{
		fmt.Sprintf("这是一封来自 %s 的测试邮件。", name),
		"",
		fmt.Sprintf("发送时间：%s", time.Now().Format("2006-01-02 15:04:05 MST")),
		fmt.Sprintf("收件地址：%s", strings.TrimSpace(to)),
		"",
		"如果你成功收到这封邮件，说明当前 SMTP 配置已经可以正常工作。",
	}, "\r\n")

	return s.Send(ctx, cfg, to, subject, body)
}

func (s *EmailService) Send(ctx context.Context, cfg SMTPConfig, to, subject, body string) error {
	recipient, err := normalizeRecipient(to)
	if err != nil {
		return err
	}

	client, err := dialSMTPClient(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	if cfg.Username != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP 认证失败: %w", err)
		}
	}

	if err := client.Mail(cfg.From); err != nil {
		return fmt.Errorf("发件人地址被拒绝: %w", err)
	}
	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("收件人地址被拒绝: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("创建邮件内容失败: %w", err)
	}

	message := buildSMTPMessage(cfg.From, recipient, subject, body)
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("提交邮件内容失败: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("结束 SMTP 会话失败: %w", err)
	}

	return nil
}

func normalizeRecipient(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("当前账号未配置邮箱，无法接收测试邮件")
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return "", fmt.Errorf("当前账号邮箱格式无效")
	}

	return addr.Address, nil
}

func dialSMTPClient(ctx context.Context, cfg SMTPConfig) (*smtp.Client, error) {
	address := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	if cfg.UseTLS && cfg.Port == 465 {
		dialer := &tls.Dialer{
			Config: &tls.Config{
				ServerName: cfg.Host,
				MinVersion: tls.VersionTLS12,
			},
		}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			return nil, fmt.Errorf("连接 SMTP 服务器失败: %w", err)
		}

		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("创建 SMTP 客户端失败: %w", err)
		}
		return client, nil
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("连接 SMTP 服务器失败: %w", err)
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}

	if cfg.UseTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			_ = client.Close()
			return nil, fmt.Errorf("SMTP 服务器不支持 STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{
			ServerName: cfg.Host,
			MinVersion: tls.VersionTLS12,
		}); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("升级 TLS 连接失败: %w", err)
		}
	}

	return client, nil
}

func buildSMTPMessage(from, to, subject, body string) string {
	return strings.Join([]string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
		"",
	}, "\r\n")
}

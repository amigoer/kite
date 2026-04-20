package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// FTPDriver is a storage driver backed by the FTP protocol.
// Each operation opens its own connection so long-lived sessions cannot time out or
// drop between calls.
type FTPDriver struct {
	addr     string
	username string
	password string
	basePath string
	baseURL  string
}

// NewFTPDriver builds an FTPDriver from the given FTPConfig.
// The constructor dials and logs in once up front to validate the configuration.
func NewFTPDriver(cfg FTPConfig) (*FTPDriver, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("ftp driver: host is required")
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("ftp driver: username is required")
	}

	port := cfg.Port
	if port == 0 {
		port = 21
	}

	d := &FTPDriver{
		addr:     fmt.Sprintf("%s:%d", cfg.Host, port),
		username: cfg.Username,
		password: cfg.Password,
		basePath: strings.TrimRight(strings.TrimSpace(cfg.BasePath), "/"),
		baseURL:  strings.TrimRight(cfg.BaseURL, "/"),
	}

	conn, err := d.dial()
	if err != nil {
		return nil, err
	}
	_ = conn.Quit()

	return d, nil
}

// dial opens an FTP connection and performs login.
func (d *FTPDriver) dial() (*ftp.ServerConn, error) {
	conn, err := ftp.Dial(d.addr, ftp.DialWithTimeout(10*time.Second))
	if err != nil {
		return nil, fmt.Errorf("ftp dial %s: %w", d.addr, err)
	}
	if err := conn.Login(d.username, d.password); err != nil {
		_ = conn.Quit()
		return nil, fmt.Errorf("ftp login: %w", err)
	}
	return conn, nil
}

// remotePath joins the relative key onto the base path to produce a remote absolute path.
func (d *FTPDriver) remotePath(key string) string {
	clean := path.Clean("/" + strings.TrimLeft(key, "/"))
	if d.basePath == "" {
		return clean
	}
	return d.basePath + clean
}

// ensureDir creates the remote directory tree one segment at a time.
func (d *FTPDriver) ensureDir(conn *ftp.ServerConn, dir string) error {
	if dir == "" || dir == "/" || dir == "." {
		return nil
	}
	parts := strings.Split(strings.Trim(dir, "/"), "/")
	current := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		current = current + "/" + p
		if err := conn.MakeDir(current); err != nil {
			// FTP returns an error when the directory already exists; treat that as success.
			msg := strings.ToLower(err.Error())
			if !strings.Contains(msg, "exists") && !strings.Contains(msg, "file exists") && !strings.Contains(msg, "already") {
				// Probe the directory: if we can cd into it, it exists.
				if changeErr := conn.ChangeDir(current); changeErr == nil {
					continue
				}
				return fmt.Errorf("ftp mkdir %q: %w", current, err)
			}
		}
	}
	return nil
}

// Put uploads a file to the FTP server.
func (d *FTPDriver) Put(_ context.Context, key string, reader io.Reader, _ int64, _ string) error {
	conn, err := d.dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Quit() }()

	full := d.remotePath(key)
	if err := d.ensureDir(conn, path.Dir(full)); err != nil {
		return err
	}
	if err := conn.Stor(full, reader); err != nil {
		return fmt.Errorf("ftp put %q: %w", key, err)
	}
	return nil
}

// Get reads a file from the FTP server.
// The FTP connection must live as long as the data stream, so a custom ReadCloser is
// returned to keep the two bound together.
func (d *FTPDriver) Get(_ context.Context, key string) (io.ReadCloser, int64, error) {
	conn, err := d.dial()
	if err != nil {
		return nil, 0, err
	}

	full := d.remotePath(key)
	size, err := conn.FileSize(full)
	if err != nil {
		_ = conn.Quit()
		return nil, 0, fmt.Errorf("ftp size %q: %w", key, err)
	}

	resp, err := conn.Retr(full)
	if err != nil {
		_ = conn.Quit()
		return nil, 0, fmt.Errorf("ftp get %q: %w", key, err)
	}

	return &ftpReadCloser{resp: resp, conn: conn}, size, nil
}

// ftpReadCloser binds an FTP response's read cycle to the underlying connection lifetime.
type ftpReadCloser struct {
	resp *ftp.Response
	conn *ftp.ServerConn
}

func (r *ftpReadCloser) Read(p []byte) (int, error) { return r.resp.Read(p) }

func (r *ftpReadCloser) Close() error {
	respErr := r.resp.Close()
	quitErr := r.conn.Quit()
	if respErr != nil {
		return respErr
	}
	return quitErr
}

// Delete removes a file from the FTP server; missing files return nil for idempotence.
func (d *FTPDriver) Delete(_ context.Context, key string) error {
	conn, err := d.dial()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Quit() }()

	full := d.remotePath(key)
	if err := conn.Delete(full); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not found") || strings.Contains(msg, "no such") || strings.Contains(msg, "550") {
			return nil
		}
		return fmt.Errorf("ftp delete %q: %w", key, err)
	}
	return nil
}

// Exists reports whether the given key is present on the FTP server.
func (d *FTPDriver) Exists(_ context.Context, key string) (bool, error) {
	conn, err := d.dial()
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Quit() }()

	full := d.remotePath(key)
	if _, err := conn.FileSize(full); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not found") || strings.Contains(msg, "no such") || strings.Contains(msg, "550") {
			return false, nil
		}
		return false, fmt.Errorf("ftp exists %q: %w", key, err)
	}
	return true, nil
}

// URL returns the public access URL for the file.
// FTP has no native HTTP URL, so the caller must configure baseURL (typically the address
// of a reverse proxy that fronts the FTP directory).
func (d *FTPDriver) URL(key string) string {
	if d.baseURL == "" {
		return "/" + strings.TrimLeft(key, "/")
	}
	return d.baseURL + "/" + strings.TrimLeft(key, "/")
}

// SignedURL is not supported for FTP; it returns the plain URL unchanged.
func (d *FTPDriver) SignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return d.URL(key), nil
}

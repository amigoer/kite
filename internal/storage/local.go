package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalDriver is the storage driver backed by the local filesystem.
// Files live on the server's disk and are served directly over HTTP.
type LocalDriver struct {
	basePath string // absolute root directory for stored files
	baseURL  string // access URL prefix
}

// NewLocalDriver builds a LocalDriver from the given LocalConfig.
func NewLocalDriver(cfg LocalConfig) (*LocalDriver, error) {
	if cfg.BasePath == "" {
		return nil, fmt.Errorf("local driver: base_path is required")
	}
	// Ensure the storage directory exists.
	absPath, err := filepath.Abs(cfg.BasePath)
	if err != nil {
		return nil, fmt.Errorf("local driver: resolve base_path: %w", err)
	}
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("local driver: create base_path %q: %w", absPath, err)
	}

	return &LocalDriver{
		basePath: absPath,
		baseURL:  strings.TrimRight(cfg.BaseURL, "/"),
	}, nil
}

// resolveKey resolves an external key to an absolute path inside basePath.
// Inputs containing "..", absolute paths, or other traversal sequences are rejected.
func (d *LocalDriver) resolveKey(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("local driver: empty key")
	}
	clean := filepath.Clean(filepath.FromSlash(key))
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("local driver: invalid key %q", key)
	}
	fullPath := filepath.Join(d.basePath, clean)
	rel, err := filepath.Rel(d.basePath, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return "", fmt.Errorf("local driver: path traversal detected in key %q", key)
	}
	return fullPath, nil
}

// Put writes a file to the local disk.
// Intermediate directories in the key path are created on demand.
func (d *LocalDriver) Put(_ context.Context, key string, reader io.Reader, _ int64, _ string) error {
	fullPath, err := d.resolveKey(key)
	if err != nil {
		return err
	}

	// Create the parent directory.
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("local put mkdir %q: %w", dir, err)
	}

	// Write to a temp file first and atomically rename so readers never see a partial write.
	tmpFile, err := os.CreateTemp(dir, ".kite-upload-*")
	if err != nil {
		return fmt.Errorf("local put create temp: %w", err)
	}
	tmpPath := tmpFile.Name()

	_, writeErr := io.Copy(tmpFile, reader)
	closeErr := tmpFile.Close()

	if writeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("local put write %q: %w", key, writeErr)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("local put close %q: %w", key, closeErr)
	}

	if err := os.Rename(tmpPath, fullPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("local put rename %q: %w", key, err)
	}

	return nil
}

// Get reads a file from the local disk.
// The caller must close the returned ReadCloser.
func (d *LocalDriver) Get(_ context.Context, key string) (io.ReadCloser, int64, error) {
	fullPath, err := d.resolveKey(key)
	if err != nil {
		return nil, 0, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("local get %q: %w", key, err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, fmt.Errorf("local get stat %q: %w", key, err)
	}

	return file, info.Size(), nil
}

// Delete removes a file from the local disk.
// Missing files return nil so the operation is idempotent.
func (d *LocalDriver) Delete(_ context.Context, key string) error {
	fullPath, err := d.resolveKey(key)
	if err != nil {
		return err
	}

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local delete %q: %w", key, err)
	}
	return nil
}

// Exists reports whether the given key exists on the local disk.
func (d *LocalDriver) Exists(_ context.Context, key string) (bool, error) {
	fullPath, err := d.resolveKey(key)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("local exists %q: %w", key, err)
	}
	return true, nil
}

// URL returns the file's access URL.
// When baseURL is unset it falls back to /uploads/{key}, served by the built-in static route.
func (d *LocalDriver) URL(key string) string {
	if d.baseURL == "" {
		return "/uploads/" + key
	}
	return d.baseURL + "/" + key
}

// SignedURL is not supported on local storage; it returns the plain URL unchanged.
func (d *LocalDriver) SignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return d.URL(key), nil
}

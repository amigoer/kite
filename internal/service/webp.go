package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// WebPService wraps the system `cwebp` binary to transcode JPEG/PNG bytes
// into WebP. We use the binary instead of cgo bindings so the Go binary
// stays statically linked (CGO_ENABLED=0) and cross-compiles cleanly to
// every supported architecture without dragging libwebp into the build
// matrix. The fork/exec overhead is ~5-10ms per image, which is dwarfed
// by the encode itself.
//
// If `cwebp` isn't on PATH at startup, the service runs in a degraded
// mode where Encode always returns ErrWebPUnavailable. The upload
// pipeline is expected to treat that as "skip transcoding, keep
// original".
type WebPService struct {
	cwebpPath string
	available bool
}

// ErrWebPUnavailable signals that cwebp wasn't located at startup. The
// upload path treats it as a no-op rather than a hard failure.
var ErrWebPUnavailable = errors.New("cwebp binary not available")

// webpEncodeTimeout caps a single transcoding run. cwebp typically
// finishes a 5MB PNG in under a second on modern hardware; ten seconds
// is a generous ceiling for pathological inputs without ever blocking
// the upload pipeline indefinitely.
const webpEncodeTimeout = 10 * time.Second

func NewWebPService() *WebPService {
	path, err := exec.LookPath("cwebp")
	if err != nil {
		// Soft-fail: surface the warning once at boot so the operator
		// notices, but don't refuse to run. The setting can still be
		// toggled on in the admin UI; uploads will skip transcoding
		// until the binary is installed.
		log.Printf("webp: cwebp not on PATH — image_auto_webp will be a no-op (install libwebp-tools to enable)")
		return &WebPService{}
	}
	return &WebPService{cwebpPath: path, available: true}
}

// Available reports whether cwebp was found at startup.
func (s *WebPService) Available() bool {
	return s != nil && s.available
}

// Encode transcodes `data` to WebP via cwebp at the given quality (1-100,
// clamped). The input is fed through stdin; the encoded WebP is read from
// stdout. If cwebp errors, times out, or produces empty output, an error
// is returned and the caller should treat the file as un-transcodable.
//
// Quality is the standard cwebp -q parameter: higher = larger + better.
// 75-85 is the typical "perceptually identical to JPEG" range. We don't
// expose the lossless flag in v1 — most uploads are lossy JPEGs already
// and lossless WebP rarely beats lossless PNG for screenshots.
func (s *WebPService) Encode(ctx context.Context, data []byte, quality int) ([]byte, error) {
	if !s.Available() {
		return nil, ErrWebPUnavailable
	}
	if quality < 1 {
		quality = 1
	} else if quality > 100 {
		quality = 100
	}

	runCtx, cancel := context.WithTimeout(ctx, webpEncodeTimeout)
	defer cancel()

	// Args:
	//   -q <n>  : lossy quality 0..100
	//   -quiet  : suppress progress output on stderr
	//   -mt     : enable multi-threaded encode (small but free win)
	//   -o -    : write output to stdout
	//   --      : end-of-options sentinel
	//   -       : read input from stdin
	cmd := exec.CommandContext(runCtx, s.cwebpPath,
		"-q", strconv.Itoa(quality),
		"-quiet",
		"-mt",
		"-o", "-",
		"--",
		"-",
	)
	cmd.Stdin = bytes.NewReader(data)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrTrim := strings.TrimSpace(stderr.String())
		if stderrTrim != "" {
			return nil, fmt.Errorf("cwebp run: %w (stderr: %s)", err, stderrTrim)
		}
		return nil, fmt.Errorf("cwebp run: %w", err)
	}
	if stdout.Len() == 0 {
		return nil, errors.New("cwebp produced empty output")
	}
	return stdout.Bytes(), nil
}

// IsTranscodableMime reports whether a given source mime is in our
// allow-list for WebP transcoding. Restricted to the two everyday raster
// formats that benefit most: PNG (often lossless screenshots that crush
// well at lossy WebP quality 80) and JPEG (gains ~25-35% on average).
// GIF is intentionally excluded — animated GIF -> animated WebP requires
// img2webp/gif2webp and a different command line; static GIF is rare.
// SVG is vector and meaningless to transcode.
func IsTranscodableMime(mime string) bool {
	switch strings.ToLower(strings.TrimSpace(mime)) {
	case "image/png", "image/jpeg", "image/jpg":
		return true
	default:
		return false
	}
}

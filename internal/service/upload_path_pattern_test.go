package service

import (
	"strings"
	"testing"
	"time"
)

func TestRenderUploadPathPattern_DefaultPattern(t *testing.T) {
	got, err := RenderUploadPathPattern("{year}/{month}/{md5_8}/{uuid}.{ext}", UploadPathPatternData{
		Now:      time.Date(2026, time.April, 21, 0, 0, 0, 0, time.UTC),
		UserID:   "user-1",
		FileType: "file",
		HashMD5:  "0123456789abcdef0123456789abcdef",
		FileID:   "123e4567-e89b-12d3-a456-426614174000",
		Ext:      "txt",
	})
	if err != nil {
		t.Fatalf("RenderUploadPathPattern: %v", err)
	}
	want := "2026/04/01234567/123e4567-e89b-12d3-a456-426614174000.txt"
	if got != want {
		t.Fatalf("unexpected rendered path: got=%q want=%q", got, want)
	}
}

func TestRenderUploadPathPattern_CustomVariables(t *testing.T) {
	got, err := RenderUploadPathPattern("{user_id}/{file_type}/{day}/{md5_8}/{uuid}.{ext}", UploadPathPatternData{
		Now:      time.Date(2026, time.April, 21, 0, 0, 0, 0, time.UTC),
		UserID:   "user-42",
		FileType: "image",
		HashMD5:  "fedcba9876543210fedcba9876543210",
		FileID:   "file-42",
		Ext:      "png",
	})
	if err != nil {
		t.Fatalf("RenderUploadPathPattern: %v", err)
	}
	want := "user-42/image/21/fedcba98/file-42.png"
	if got != want {
		t.Fatalf("unexpected rendered path: got=%q want=%q", got, want)
	}
}

func TestRenderUploadPathPattern_ExtFallbackToBin(t *testing.T) {
	got, err := RenderUploadPathPattern("{year}/{uuid}.{ext}", UploadPathPatternData{
		Now:     time.Date(2026, time.April, 21, 0, 0, 0, 0, time.UTC),
		HashMD5: strings.Repeat("a", 32),
		FileID:  "file-1",
	})
	if err != nil {
		t.Fatalf("RenderUploadPathPattern: %v", err)
	}
	want := "2026/file-1.bin"
	if got != want {
		t.Fatalf("unexpected rendered path: got=%q want=%q", got, want)
	}
}

func TestNormalizeUploadPathPattern_UnknownVariable(t *testing.T) {
	if _, err := NormalizeUploadPathPattern("{year}/{unknown}/{uuid}.{ext}"); err == nil {
		t.Fatal("expected unknown variable error")
	}
}

func TestNormalizeUploadPathPattern_InvalidPaths(t *testing.T) {
	cases := []string{
		"/{year}/{uuid}.{ext}",
		"{year}//{uuid}.{ext}",
		"{year}/../{uuid}.{ext}",
	}

	for _, pattern := range cases {
		if _, err := NormalizeUploadPathPattern(pattern); err == nil {
			t.Fatalf("expected invalid pattern error for %q", pattern)
		}
	}
}

func TestNormalizeUploadPathPattern_TrimsTrailingSlash(t *testing.T) {
	got, err := NormalizeUploadPathPattern("{year}/{uuid}.{ext}/")
	if err != nil {
		t.Fatalf("NormalizeUploadPathPattern: %v", err)
	}
	if got != "{year}/{uuid}.{ext}" {
		t.Fatalf("unexpected normalized pattern: %q", got)
	}
}

package service

import (
	"fmt"
	"strings"
	"time"
)

const UploadPathPatternSettingKey = "upload.path_pattern"

type UploadPathPatternData struct {
	Now      time.Time
	UserID   string
	FileType string
	HashMD5  string
	FileID   string
	Ext      string
}

var uploadPathPatternVariables = map[string]func(UploadPathPatternData) string{
	"year": func(data UploadPathPatternData) string {
		return fmt.Sprintf("%04d", data.Now.Year())
	},
	"month": func(data UploadPathPatternData) string {
		return fmt.Sprintf("%02d", int(data.Now.Month()))
	},
	"day": func(data UploadPathPatternData) string {
		return fmt.Sprintf("%02d", data.Now.Day())
	},
	"user_id": func(data UploadPathPatternData) string {
		if strings.TrimSpace(data.UserID) == "" {
			return "guest"
		}
		return data.UserID
	},
	"file_type": func(data UploadPathPatternData) string {
		if strings.TrimSpace(data.FileType) == "" {
			return "file"
		}
		return data.FileType
	},
	"md5": func(data UploadPathPatternData) string {
		return data.HashMD5
	},
	"md5_8": func(data UploadPathPatternData) string {
		if len(data.HashMD5) < 8 {
			return data.HashMD5
		}
		return data.HashMD5[:8]
	},
	"uuid": func(data UploadPathPatternData) string {
		return data.FileID
	},
	"ext": func(data UploadPathPatternData) string {
		return normalizeUploadPathExt(data.Ext)
	},
}

func NormalizeUploadPathPattern(pattern string) (string, error) {
	normalized, err := sanitizeUploadPathPattern(pattern)
	if err != nil {
		return "", err
	}

	rest := normalized
	for key := range uploadPathPatternVariables {
		rest = strings.ReplaceAll(rest, "{"+key+"}", "")
	}
	if strings.ContainsAny(rest, "{}") {
		return "", fmt.Errorf("包含无效占位符，请使用形如 {year} 的变量")
	}

	for _, part := range extractUploadPathPlaceholders(normalized) {
		if _, ok := uploadPathPatternVariables[part]; !ok {
			return "", fmt.Errorf("不支持的路径变量 {%s}", part)
		}
	}

	_, err = RenderUploadPathPattern(normalized, UploadPathPatternData{
		Now:      time.Date(2026, time.April, 21, 0, 0, 0, 0, time.UTC),
		UserID:   "user-demo",
		FileType: "file",
		HashMD5:  "0123456789abcdef0123456789abcdef",
		FileID:   "123e4567-e89b-12d3-a456-426614174000",
		Ext:      "txt",
	})
	if err != nil {
		return "", err
	}

	return normalized, nil
}

func ValidateUploadPathPattern(pattern string) error {
	_, err := NormalizeUploadPathPattern(pattern)
	return err
}

func RenderUploadPathPattern(pattern string, data UploadPathPatternData) (string, error) {
	normalized, err := sanitizeUploadPathPattern(pattern)
	if err != nil {
		return "", err
	}

	replaced := normalized
	for key, resolver := range uploadPathPatternVariables {
		replaced = strings.ReplaceAll(replaced, "{"+key+"}", resolver(data))
	}
	if strings.ContainsAny(replaced, "{}") {
		return "", fmt.Errorf("路径模板包含未解析的占位符")
	}

	key, err := sanitizeResolvedUploadPathKey(replaced)
	if err != nil {
		return "", err
	}
	return key, nil
}

func extractUploadPathPlaceholders(pattern string) []string {
	parts := make([]string, 0)
	start := -1
	for i, ch := range pattern {
		switch ch {
		case '{':
			start = i
		case '}':
			if start >= 0 && start < i {
				parts = append(parts, pattern[start+1:i])
			}
			start = -1
		}
	}
	return parts
}

func sanitizeUploadPathPattern(pattern string) (string, error) {
	normalized := strings.TrimSpace(pattern)
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	if normalized == "" {
		return "", fmt.Errorf("路径模板不能为空")
	}
	if strings.HasPrefix(normalized, "/") {
		return "", fmt.Errorf("路径模板不能以 / 开头")
	}
	normalized = strings.TrimRight(normalized, "/")
	if normalized == "" {
		return "", fmt.Errorf("路径模板不能为空")
	}

	segments := strings.Split(normalized, "/")
	for _, segment := range segments {
		if segment == "" {
			return "", fmt.Errorf("路径模板不能包含空目录层级")
		}
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("路径模板不能包含 . 或 ..")
		}
	}

	return strings.Join(segments, "/"), nil
}

func sanitizeResolvedUploadPathKey(key string) (string, error) {
	normalized := strings.TrimSpace(key)
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	normalized = strings.TrimRight(normalized, "/")
	if normalized == "" {
		return "", fmt.Errorf("渲染后的存储路径不能为空")
	}
	if strings.HasPrefix(normalized, "/") {
		return "", fmt.Errorf("渲染后的存储路径不能以 / 开头")
	}

	segments := strings.Split(normalized, "/")
	for _, segment := range segments {
		if segment == "" {
			return "", fmt.Errorf("渲染后的存储路径不能包含空目录层级")
		}
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("渲染后的存储路径不能包含 . 或 ..")
		}
	}

	return strings.Join(segments, "/"), nil
}

func normalizeUploadPathExt(ext string) string {
	clean := strings.ToLower(strings.TrimSpace(ext))
	clean = strings.TrimPrefix(clean, ".")
	if clean == "" {
		return "bin"
	}
	return clean
}

package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// FaviconHandler 处理网站 favicon 抓取
type FaviconHandler struct{}

func NewFaviconHandler() *FaviconHandler {
	return &FaviconHandler{}
}

// faviconRequest 抓取请求
type faviconRequest struct {
	URL string `json:"url" binding:"required"`
}

// Fetch 抓取指定网站的 favicon URL
func (h *FaviconHandler) Fetch(c *gin.Context) {
	var req faviconRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "url is required")
		return
	}

	siteURL := strings.TrimSpace(req.URL)
	parsed, err := url.ParseRequestURI(siteURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid url")
		return
	}

	faviconURL, err := fetchFaviconURL(siteURL, parsed)
	if err != nil {
		Error(c, http.StatusUnprocessableEntity, http.StatusUnprocessableEntity, fmt.Sprintf("failed to fetch favicon: %v", err))
		return
	}

	Success(c, gin.H{"favicon": faviconURL})
}

// fetchFaviconURL 抓取网站首页 HTML 并解析 favicon URL
func fetchFaviconURL(siteURL string, parsed *url.URL) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		// 不自动跟随重定向超过 5 次
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", siteURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KiteBot/1.0)")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 限制读取体积为 512KB，避免大文件
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}

	html := string(body)
	origin := parsed.Scheme + "://" + parsed.Host

	// 1. 尝试从 <link> 标签中提取 favicon
	// 匹配 <link rel="icon" href="..."> 或 <link rel="shortcut icon" href="...">
	// 也匹配 rel="apple-touch-icon" 作为备选
	patterns := []string{
		`<link[^>]*rel=["'](?:shortcut )?icon["'][^>]*href=["']([^"']+)["']`,
		`<link[^>]*href=["']([^"']+)["'][^>]*rel=["'](?:shortcut )?icon["']`,
		`<link[^>]*rel=["']apple-touch-icon["'][^>]*href=["']([^"']+)["']`,
		`<link[^>]*href=["']([^"']+)["'][^>]*rel=["']apple-touch-icon["']`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile("(?i)" + pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) >= 2 {
			href := strings.TrimSpace(matches[1])
			return resolveHref(href, origin), nil
		}
	}

	// 2. 回退：尝试 /favicon.ico，仅在确认可达时返回
	fallbackURL := origin + "/favicon.ico"

	headReq, err := http.NewRequest("HEAD", fallbackURL, nil)
	if err != nil {
		return "", fmt.Errorf("未找到 favicon")
	}
	headReq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KiteBot/1.0)")

	headResp, err := client.Do(headReq)
	if err != nil {
		return "", fmt.Errorf("未找到 favicon")
	}
	defer headResp.Body.Close()

	if headResp.StatusCode == http.StatusOK {
		return fallbackURL, nil
	}

	return "", fmt.Errorf("未找到 favicon")
}

// resolveHref 将相对路径转换为绝对路径
func resolveHref(href, origin string) string {
	// 已经是绝对 URL
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	// 协议相对 URL
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	// 根路径相对
	if strings.HasPrefix(href, "/") {
		return origin + href
	}
	// 其他相对路径
	return origin + "/" + href
}

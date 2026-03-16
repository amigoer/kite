package service

import (
	"github.com/amigoer/kite-blog/internal/config"
)

// SiteSettings 站点基础设置
type SiteSettings struct {
	SiteName    string `json:"site_name" yaml:"site_name"`
	SiteURL     string `json:"site_url" yaml:"site_url"`
	Description string `json:"description" yaml:"description"`
	Keywords    string `json:"keywords" yaml:"keywords"`
	Favicon     string `json:"favicon" yaml:"favicon"`
	Logo        string `json:"logo" yaml:"logo"`
	ICP         string `json:"icp" yaml:"icp"`
	Footer      string `json:"footer" yaml:"footer"`
}

// PostSettingsResp 文章相关设置
type PostSettingsResp struct {
	PostsPerPage    int    `json:"posts_per_page" yaml:"posts_per_page"`
	EnableComment   bool   `json:"enable_comment" yaml:"enable_comment"`
	EnableToc       bool   `json:"enable_toc" yaml:"enable_toc"`
	SummaryLength   int    `json:"summary_length" yaml:"summary_length"`
	DefaultCoverURL string `json:"default_cover_url" yaml:"default_cover_url"`
}

// RenderSettingsResp 渲染模式设置
type RenderSettingsResp struct {
	RenderMode string `json:"render_mode" yaml:"render_mode"`
	APIPrefix  string `json:"api_prefix" yaml:"api_prefix"`
	EnableCORS bool   `json:"enable_cors" yaml:"enable_cors"`
}

// AISettings AI 集成设置
type AISettings struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Provider    string `json:"provider" yaml:"provider"`
	APIKey      string `json:"api_key" yaml:"api_key"`
	Model       string `json:"model" yaml:"model"`
	AutoSummary bool   `json:"auto_summary" yaml:"auto_summary"`
	AutoTag     bool   `json:"auto_tag" yaml:"auto_tag"`
}

// AllSettings 全部设置聚合
type AllSettings struct {
	Site   SiteSettings       `json:"site" yaml:"site"`
	Post   PostSettingsResp   `json:"post" yaml:"post"`
	Render RenderSettingsResp `json:"render" yaml:"render"`
	AI     AISettings         `json:"ai" yaml:"ai"`
}

// SettingsService 设置服务
type SettingsService struct {
	cfg *config.Config
}

func NewSettingsService(cfg *config.Config) *SettingsService {
	return &SettingsService{cfg: cfg}
}

// Get 获取当前全部设置
func (s *SettingsService) Get() *AllSettings {
	return &AllSettings{
		Site: SiteSettings{
			SiteName:    s.cfg.Site.SiteName,
			SiteURL:     s.cfg.Site.SiteURL,
			Description: s.cfg.Site.Description,
			Keywords:    s.cfg.Site.Keywords,
			Favicon:     s.cfg.Site.Favicon,
			Logo:        s.cfg.Site.Logo,
			ICP:         s.cfg.Site.ICP,
			Footer:      s.cfg.Site.Footer,
		},
		Post: PostSettingsResp{
			PostsPerPage:    s.cfg.Post.PostsPerPage,
			EnableComment:   s.cfg.Post.EnableComment,
			EnableToc:       s.cfg.Post.EnableToc,
			SummaryLength:   s.cfg.Post.SummaryLength,
			DefaultCoverURL: s.cfg.Post.DefaultCoverURL,
		},
		Render: RenderSettingsResp{
			RenderMode: s.cfg.RenderMode,
			APIPrefix:  "/api/v1",
			EnableCORS: true,
		},
		AI: AISettings{
			Enabled:     s.cfg.AI.Enabled,
			Provider:    s.cfg.AI.Provider,
			APIKey:      maskAPIKey(s.cfg.AI.APIKey),
			Model:       s.cfg.AI.Model,
			AutoSummary: s.cfg.AI.AutoSummary,
			AutoTag:     s.cfg.AI.AutoTag,
		},
	}
}

// Update 更新设置（运行时覆盖，不持久化）
func (s *SettingsService) Update(input AllSettings) *AllSettings {
	// 站点设置
	s.cfg.Site.SiteName = input.Site.SiteName
	s.cfg.Site.SiteURL = input.Site.SiteURL
	s.cfg.Site.Description = input.Site.Description
	s.cfg.Site.Keywords = input.Site.Keywords
	s.cfg.Site.Favicon = input.Site.Favicon
	s.cfg.Site.Logo = input.Site.Logo
	s.cfg.Site.ICP = input.Site.ICP
	s.cfg.Site.Footer = input.Site.Footer

	// 文章设置
	if input.Post.PostsPerPage > 0 {
		s.cfg.Post.PostsPerPage = input.Post.PostsPerPage
	}
	s.cfg.Post.EnableComment = input.Post.EnableComment
	s.cfg.Post.EnableToc = input.Post.EnableToc
	if input.Post.SummaryLength > 0 {
		s.cfg.Post.SummaryLength = input.Post.SummaryLength
	}
	s.cfg.Post.DefaultCoverURL = input.Post.DefaultCoverURL

	// 渲染模式
	if input.Render.RenderMode == config.RenderModeClassic || input.Render.RenderMode == config.RenderModeHeadless {
		s.cfg.RenderMode = input.Render.RenderMode
	}

	// AI 设置
	s.cfg.AI.Enabled = input.AI.Enabled
	s.cfg.AI.Provider = input.AI.Provider
	// 仅当收到非掩码值时更新 API Key
	if input.AI.APIKey != "" && input.AI.APIKey != maskAPIKey(s.cfg.AI.APIKey) {
		s.cfg.AI.APIKey = input.AI.APIKey
	}
	s.cfg.AI.Model = input.AI.Model
	s.cfg.AI.AutoSummary = input.AI.AutoSummary
	s.cfg.AI.AutoTag = input.AI.AutoTag

	return s.Get()
}

// maskAPIKey 掩盖 API Key 中间部分
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

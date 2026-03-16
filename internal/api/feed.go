package api

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/amigoer/kite-blog/internal/config"
	"github.com/amigoer/kite-blog/internal/service"
	"github.com/gin-gonic/gin"
)

// RSS 2.0 结构
type rssRoot struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	PubDate     string    `xml:"pubDate,omitempty"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate,omitempty"`
	GUID        string `xml:"guid"`
}

// Sitemap 结构
type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc        string `xml:"loc"`
	Lastmod    string `xml:"lastmod,omitempty"`
	Changefreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// FeedHandler RSS 和 Sitemap 处理器
type FeedHandler struct {
	cfg         *config.Config
	postService *service.PostService
	pageService *service.PageService
}

func NewFeedHandler(cfg *config.Config, postService *service.PostService, pageService *service.PageService) *FeedHandler {
	return &FeedHandler{cfg: cfg, postService: postService, pageService: pageService}
}

// RSS 生成 RSS 2.0 feed
func (h *FeedHandler) RSS(c *gin.Context) {
	siteURL := strings.TrimRight(h.cfg.Site.SiteURL, "/")
	if siteURL == "" {
		siteURL = fmt.Sprintf("http://%s", c.Request.Host)
	}

	// 获取最近 20 篇已发布文章
	result, err := h.postService.ListPublic(service.PostListParams{
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "生成 RSS 失败")
		return
	}

	items := make([]rssItem, 0, len(result.Items))
	var latestPubDate string
	for _, post := range result.Items {
		pubDate := ""
		if post.PublishedAt != nil {
			pubDate = post.PublishedAt.Format(time.RFC1123Z)
			if latestPubDate == "" {
				latestPubDate = pubDate
			}
		}
		desc := post.Summary
		if desc == "" && len(post.ContentHTML) > 300 {
			desc = post.ContentHTML[:300] + "..."
		} else if desc == "" {
			desc = post.ContentHTML
		}
		items = append(items, rssItem{
			Title:       post.Title,
			Link:        siteURL + "/posts/" + post.Slug,
			Description: desc,
			PubDate:     pubDate,
			GUID:        siteURL + "/posts/" + post.Slug,
		})
	}

	feed := rssRoot{
		Version: "2.0",
		Channel: rssChannel{
			Title:       h.cfg.Site.SiteName,
			Link:        siteURL,
			Description: h.cfg.Site.Description,
			Language:    "zh-CN",
			PubDate:     latestPubDate,
			Items:       items,
		},
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=3600")
	output, _ := xml.MarshalIndent(feed, "", "  ")
	c.String(http.StatusOK, xml.Header+string(output))
}

// Sitemap 生成 sitemap.xml
func (h *FeedHandler) Sitemap(c *gin.Context) {
	siteURL := strings.TrimRight(h.cfg.Site.SiteURL, "/")
	if siteURL == "" {
		siteURL = fmt.Sprintf("http://%s", c.Request.Host)
	}

	urls := []sitemapURL{
		{Loc: siteURL + "/", Changefreq: "daily", Priority: "1.0"},
	}

	// 已发布文章
	result, err := h.postService.ListPublic(service.PostListParams{
		Page:     1,
		PageSize: 1000,
	})
	if err == nil {
		for _, post := range result.Items {
			lastmod := post.UpdatedAt.Format("2006-01-02")
			urls = append(urls, sitemapURL{
				Loc:        siteURL + "/posts/" + post.Slug,
				Lastmod:    lastmod,
				Changefreq: "weekly",
				Priority:   "0.8",
			})
		}
	}

	// 已发布页面
	pages, _ := h.pageService.ListPublic()
	if pages != nil {
		for _, page := range pages.Items {
			lastmod := page.UpdatedAt.Format("2006-01-02")
			urls = append(urls, sitemapURL{
				Loc:        siteURL + "/pages/" + page.Slug,
				Lastmod:    lastmod,
				Changefreq: "monthly",
				Priority:   "0.6",
			})
		}
	}

	sitemap := sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=3600")
	output, _ := xml.MarshalIndent(sitemap, "", "  ")
	c.String(http.StatusOK, xml.Header+string(output))
}

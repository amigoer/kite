package service

import (
	"regexp"
	"strings"
)

// defaultSpamKeywords 默认垃圾评论关键词
var defaultSpamKeywords = []string{
	"casino", "viagra", "lottery", "poker", "gambling",
	"crypto trading", "binary options", "forex signal",
	"earn money fast", "make money online", "work from home",
	"click here", "buy now", "free trial", "limited offer",
	"adult", "porn", "sex", "xxx",
}

var urlPattern = regexp.MustCompile(`https?://[^\s]+`)

// SpamChecker 垃圾评论检查器
type SpamChecker struct {
	keywords    []string
	maxURLCount int
}

// NewSpamChecker 创建垃圾评论检查器
func NewSpamChecker() *SpamChecker {
	return &SpamChecker{
		keywords:    defaultSpamKeywords,
		maxURLCount: 3,
	}
}

// IsSpam 检查内容是否为垃圾评论
func (sc *SpamChecker) IsSpam(content string) bool {
	lower := strings.ToLower(content)

	// 关键词检查
	for _, kw := range sc.keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	// URL 数量检查
	urls := urlPattern.FindAllString(content, -1)
	if len(urls) > sc.maxURLCount {
		return true
	}

	return false
}

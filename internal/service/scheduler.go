package service

import (
	"log"
	"time"

	"github.com/amigoer/kite-blog/internal/model"
	"github.com/amigoer/kite-blog/internal/repo"
)

// PostScheduler 定时检查并发布到期的定时文章
type PostScheduler struct {
	postRepo *repo.PostRepository
	interval time.Duration
	stopCh   chan struct{}
}

// NewPostScheduler 创建定时发布调度器
func NewPostScheduler(postRepo *repo.PostRepository, interval time.Duration) *PostScheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &PostScheduler{
		postRepo: postRepo,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动后台调度
func (s *PostScheduler) Start() {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// 启动时立即检查一次
		s.publishDuePosts()

		for {
			select {
			case <-ticker.C:
				s.publishDuePosts()
			case <-s.stopCh:
				return
			}
		}
	}()
	log.Printf("📅 定时发布调度器已启动（每 %v 检查一次）", s.interval)
}

// Stop 停止调度
func (s *PostScheduler) Stop() {
	close(s.stopCh)
}

// publishDuePosts 查找所有到期的 scheduled 文章并发布
func (s *PostScheduler) publishDuePosts() {
	now := time.Now()

	posts, err := s.postRepo.FindScheduledBefore(now)
	if err != nil {
		log.Printf("定时发布：查询失败: %v", err)
		return
	}

	for _, post := range posts {
		post.Status = model.PostStatusPublished
		if err := s.postRepo.Update(&post); err != nil {
			log.Printf("定时发布：发布文章 [%s] 失败: %v", post.Title, err)
		} else {
			log.Printf("📢 定时发布：「%s」已自动发布", post.Title)
		}
	}
}

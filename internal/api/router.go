package api

import (
	"io/fs"
	"net/http"

	"github.com/amigoer/kite-blog/internal/config"
	"github.com/amigoer/kite-blog/internal/repo"
	"github.com/amigoer/kite-blog/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(cfg *config.Config, templateFS fs.FS, db *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), CORSMiddleware())

	registerAPIRoutes(router, cfg, db)
	registerPageRoutes(router, cfg, templateFS, db)

	return router
}

func registerAPIRoutes(router *gin.Engine, cfg *config.Config, db *gorm.DB) {
	systemService := service.NewSystemService(cfg)
	healthHandler := NewHealthHandler(systemService)
	tagRepository := repo.NewTagRepository(db)
	categoryRepository := repo.NewCategoryRepository(db)
	postRepository := repo.NewPostRepository(db)
	postService := service.NewPostService(postRepository, tagRepository, categoryRepository)
	postHandler := NewPostHandler(postService)
	adminSessionRepository := repo.NewAdminSessionRepository(db)
	adminAuthService := service.NewAdminAuthService(cfg, adminSessionRepository)
	adminAuthHandler := NewAdminAuthHandler(adminAuthService)
	adminAuthMiddleware := NewAdminAuthMiddleware(adminAuthService)
	friendLinkRepository := repo.NewFriendLinkRepository(db)
	friendLinkService := service.NewFriendLinkService(friendLinkRepository)
	friendLinkHandler := NewFriendLinkHandler(friendLinkService)
	tagService := service.NewTagService(tagRepository)
	tagHandler := NewTagHandler(tagService)
	categoryService := service.NewCategoryService(categoryRepository)
	categoryHandler := NewCategoryHandler(categoryService)
	commentRepository := repo.NewCommentRepository(db)
	commentService := service.NewCommentService(commentRepository, postRepository)
	commentHandler := NewCommentHandler(commentService)
	pageRepository := repo.NewPageRepository(db)
	pageService := service.NewPageService(pageRepository)
	pageHandler := NewPageHandler(pageService)
	settingsService := service.NewSettingsService(cfg)
	settingsHandler := NewSettingsHandler(settingsService)

	apiV1 := router.Group("/api/v1")
	apiV1.GET("/health", healthHandler.Get)
	apiV1.GET("/posts", postHandler.ListPublic)
	apiV1.GET("/posts/:id", postHandler.GetPublicByID)
	apiV1.GET("/posts/slug/:slug", postHandler.GetPublicBySlug)
	apiV1.GET("/posts/:id/comments", commentHandler.ListByPost)
	apiV1.POST("/posts/:id/comments", commentHandler.Create)
	apiV1.GET("/friend-links", friendLinkHandler.ListPublic)
	apiV1.GET("/friend-links/:id", friendLinkHandler.GetPublicByID)
	apiV1.GET("/tags", tagHandler.List)
	apiV1.GET("/tags/:id", tagHandler.GetByID)
	apiV1.GET("/categories", categoryHandler.List)
	apiV1.GET("/categories/:id", categoryHandler.GetByID)
	apiV1.GET("/pages", pageHandler.ListPublic)
	apiV1.GET("/pages/slug/:slug", pageHandler.GetPublicBySlug)

	adminV1 := apiV1.Group("/admin")
	adminV1.POST("/auth/login", adminAuthHandler.Login)
	adminV1.GET("/auth/me", adminAuthHandler.Me)

	protectedAdminV1 := adminV1.Group("/")
	protectedAdminV1.Use(adminAuthMiddleware.Require())
	protectedAdminV1.POST("/auth/logout", adminAuthHandler.Logout)
	protectedAdminV1.GET("/posts", postHandler.List)
	protectedAdminV1.GET("/posts/:id", postHandler.GetByID)
	protectedAdminV1.GET("/posts/slug/:slug", postHandler.GetBySlug)
	protectedAdminV1.POST("/posts", postHandler.Create)
	protectedAdminV1.PUT("/posts/:id", postHandler.Update)
	protectedAdminV1.PATCH("/posts/:id", postHandler.Patch)
	protectedAdminV1.DELETE("/posts/:id", postHandler.Delete)
	protectedAdminV1.GET("/friend-links", friendLinkHandler.List)
	protectedAdminV1.GET("/friend-links/:id", friendLinkHandler.GetByID)
	protectedAdminV1.POST("/friend-links", friendLinkHandler.Create)
	protectedAdminV1.PUT("/friend-links/:id", friendLinkHandler.Update)
	protectedAdminV1.PATCH("/friend-links/:id", friendLinkHandler.Patch)
	protectedAdminV1.DELETE("/friend-links/:id", friendLinkHandler.Delete)
	protectedAdminV1.GET("/tags", tagHandler.List)
	protectedAdminV1.GET("/tags/:id", tagHandler.GetByID)
	protectedAdminV1.POST("/tags", tagHandler.Create)
	protectedAdminV1.PUT("/tags/:id", tagHandler.Update)
	protectedAdminV1.PATCH("/tags/:id", tagHandler.Patch)
	protectedAdminV1.DELETE("/tags/:id", tagHandler.Delete)
	protectedAdminV1.GET("/categories", categoryHandler.List)
	protectedAdminV1.GET("/categories/:id", categoryHandler.GetByID)
	protectedAdminV1.POST("/categories", categoryHandler.Create)
	protectedAdminV1.PUT("/categories/:id", categoryHandler.Update)
	protectedAdminV1.PATCH("/categories/:id", categoryHandler.Patch)
	protectedAdminV1.DELETE("/categories/:id", categoryHandler.Delete)
	protectedAdminV1.GET("/comments", commentHandler.List)
	protectedAdminV1.GET("/comments/stats", commentHandler.Stats)
	protectedAdminV1.PATCH("/comments/:id", commentHandler.Moderate)
	protectedAdminV1.DELETE("/comments/:id", commentHandler.Delete)
	protectedAdminV1.GET("/pages", pageHandler.List)
	protectedAdminV1.GET("/pages/:id", pageHandler.GetByID)
	protectedAdminV1.POST("/pages", pageHandler.Create)
	protectedAdminV1.PUT("/pages/:id", pageHandler.Update)
	protectedAdminV1.PATCH("/pages/:id", pageHandler.Patch)
	protectedAdminV1.DELETE("/pages/:id", pageHandler.Delete)
	protectedAdminV1.GET("/settings", settingsHandler.Get)
	protectedAdminV1.PUT("/settings", settingsHandler.Update)
}

func registerPageRoutes(router *gin.Engine, cfg *config.Config, templateFS fs.FS, db *gorm.DB) {
	if cfg == nil || cfg.RenderMode != config.RenderModeClassic {
		return
	}

	if templateFS == nil {
		return
	}

	tmpl, err := loadTemplateSet(templateFS)
	if err != nil {
		return
	}

	router.SetHTMLTemplate(tmpl)

	// 静态资源服务（从 templates/static/ 提供）
	staticFS, err := fs.Sub(templateFS, "templates/static")
	if err == nil {
		router.StaticFS("/static", http.FS(staticFS))
	}

	// 构建 SSR 所需的 service/repo 实例
	tagRepo := repo.NewTagRepository(db)
	categoryRepo := repo.NewCategoryRepository(db)
	postRepo := repo.NewPostRepository(db)
	postService := service.NewPostService(postRepo, tagRepo, categoryRepo)
	pageRepo := repo.NewPageRepository(db)
	pageService := service.NewPageService(pageRepo)
	friendLinkRepo := repo.NewFriendLinkRepository(db)
	friendLinkService := service.NewFriendLinkService(friendLinkRepo)

	ssr := NewSSRHandler(cfg, postService, pageService, friendLinkService, categoryRepo, tagRepo, pageRepo)

	// 前台页面路由
	router.GET("/", ssr.Index)
	router.GET("/posts/:slug", ssr.PostDetail)
	router.GET("/categories/:slug", ssr.CategoryArchive)
	router.GET("/tags/:slug", ssr.TagArchive)
	router.GET("/pages/:slug", ssr.PageDetail)
	router.GET("/friends", ssr.Friends)

	// 404 兜底
	router.NoRoute(func(c *gin.Context) {
		ssr.renderError(c, http.StatusNotFound)
	})
}

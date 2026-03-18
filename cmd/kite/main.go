package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	kiteembed "github.com/amigoer/kite-blog"
	"github.com/amigoer/kite-blog/internal/api"
	"github.com/amigoer/kite-blog/internal/build"
	"github.com/amigoer/kite-blog/internal/config"
	"github.com/amigoer/kite-blog/internal/repo"
	"github.com/amigoer/kite-blog/internal/service"
	"gorm.io/gorm"
)

const defaultDBPath = "kite.db"

func main() {
	// 解析子命令
	if len(os.Args) > 1 && os.Args[1] == "build" {
		runBuild()
		return
	}

	// 默认行为：启动 Web 服务
	runServer()
}

// runServer 启动 Kite Web 服务（原有逻辑）
func runServer() {
	// 首次运行：DB 不存在 → 启动安装引导页
	if !fileExists(defaultDBPath) {
		startInstallServer(defaultDBPath)
		return
	}

	cfg := config.Default()
	cfg.Database.Driver = config.DatabaseDriverSQLite
	cfg.Database.Path = defaultDBPath

	db := initDB(cfg)
	templateFS := kiteembed.TemplateFS()
	adminFS := kiteembed.AdminFS()
	router := api.NewRouter(cfg, templateFS, adminFS, db)

	addr := ":8080"
	log.Printf("🪁 Kite 已启动: http://localhost%s", addr)
	if err := router.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("start kite server: %v", err)
	}
}

// runBuild 执行静态站点生成
func runBuild() {
	// 解析 build 子命令的参数
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	outputDir := buildCmd.String("o", "public", "输出目录")
	dbPath := buildCmd.String("db", defaultDBPath, "数据库文件路径")
	buildCmd.Parse(os.Args[2:])

	// 检查数据库文件是否存在
	if !fileExists(*dbPath) {
		fmt.Fprintf(os.Stderr, "❌ 数据库文件不存在: %s\n", *dbPath)
		os.Exit(1)
	}

	// 初始化配置和数据库
	cfg := config.Default()
	cfg.Database.Driver = config.DatabaseDriverSQLite
	cfg.Database.Path = *dbPath

	db := initDB(cfg)

	// 从 DB 加载站点设置到配置
	settingsRepo := repo.NewSettingsRepository(db)
	_ = service.NewSettingsService(cfg, settingsRepo) // 构造时自动从 DB 加载设置

	// 构建所需的 service/repo 实例
	tagRepo := repo.NewTagRepository(db)
	categoryRepo := repo.NewCategoryRepository(db)
	postRepo := repo.NewPostRepository(db)
	postService := service.NewPostService(postRepo, tagRepo, categoryRepo)
	pageRepo := repo.NewPageRepository(db)
	pageService := service.NewPageService(pageRepo)
	friendLinkRepo := repo.NewFriendLinkRepository(db)
	friendLinkService := service.NewFriendLinkService(friendLinkRepo)

	// 使用嵌入的模板文件系统
	templateFS := kiteembed.TemplateFS()

	// 创建并执行静态生成器
	builder := build.New(
		cfg, templateFS, *outputDir,
		postService, pageService, friendLinkService,
		categoryRepo, tagRepo, pageRepo,
	)

	if err := builder.Build(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 静态生成失败: %v\n", err)
		os.Exit(1)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func initDB(cfg *config.Config) *gorm.DB {
	db, err := repo.InitDB(cfg)
	if err != nil {
		log.Fatalf("init database failed: %v", err)
	}
	return db
}

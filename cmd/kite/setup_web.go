package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/amigoer/kite-blog/internal/config"
	"github.com/amigoer/kite-blog/internal/model"
	"github.com/amigoer/kite-blog/internal/repo"
	"github.com/amigoer/kite-blog/internal/service"
	"golang.org/x/crypto/bcrypt"
)

// fileExists 检查文件是否存在
func dbExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// installRequest 安装请求
type installRequest struct {
	SiteName string `json:"site_name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// installPageHTML 安装引导页面模板
const installPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Kite — 安装引导</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <style>
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap');
    body { font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; }
    .step { display: none; }
    .step.active { display: block; }
  </style>
</head>
<body class="min-h-screen bg-zinc-50/50 flex flex-col items-center justify-center p-4">
  
  <div class="w-full max-w-md bg-white border border-zinc-200 rounded-xl shadow-sm text-zinc-950">
    <div class="flex flex-col space-y-1.5 p-6 pb-4">
      <div class="flex items-center justify-center space-x-2 mb-2">
        <span class="text-3xl">🪁</span>
        <h1 class="text-2xl font-bold tracking-tight">Kite</h1>
      </div>
      <p class="text-sm text-zinc-500 text-center text-balance">
        轻量级 AI 原生博客引擎 — 安装引导
      </p>
    </div>

    <!-- 进度条点缀 -->
    <div class="flex justify-center items-center gap-2 mb-6 px-6">
      <div id="dot1" class="h-2 w-2 rounded-full bg-zinc-900 transition-all duration-300"></div>
      <div id="dot2" class="h-2 w-2 rounded-full bg-zinc-200 transition-all duration-300"></div>
      <div id="dot3" class="h-2 w-2 rounded-full bg-zinc-200 transition-all duration-300"></div>
    </div>

    <div class="p-6 pt-0">
      <div id="error" class="hidden mb-4 p-3 rounded-md bg-red-50 text-red-600 border border-red-200 text-sm font-medium"></div>

      <!-- Step 1 -->
      <div class="step active space-y-4" id="step1">
        <div class="space-y-2">
          <label for="siteName" class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">站点名称</label>
          <input type="text" id="siteName" value="Kite" placeholder="给你的博客取个名字" class="flex h-9 w-full rounded-md border border-zinc-200 bg-transparent px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-zinc-500 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-950 disabled:cursor-not-allowed disabled:opacity-50">
        </div>
        <button onclick="goStep(2)" class="w-full inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-950 disabled:pointer-events-none disabled:opacity-50 bg-zinc-900 text-zinc-50 shadow hover:bg-zinc-900/90 h-9 px-4 py-2 mt-2">
          下一步
        </button>
      </div>

      <!-- Step 2 -->
      <div class="step space-y-4" id="step2">
        <div class="space-y-2">
          <label for="username" class="text-sm font-medium leading-none">管理员用户名</label>
          <input type="text" id="username" placeholder="用于登录后台" class="flex h-9 w-full rounded-md border border-zinc-200 bg-transparent px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-zinc-500 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-950">
        </div>
        <div class="space-y-2">
          <label for="password" class="text-sm font-medium leading-none">管理员密码</label>
          <input type="password" id="password" placeholder="至少 6 位" class="flex h-9 w-full rounded-md border border-zinc-200 bg-transparent px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-zinc-500 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-950">
        </div>
        <div class="space-y-2">
          <label for="passwordConfirm" class="text-sm font-medium leading-none">确认密码</label>
          <input type="password" id="passwordConfirm" placeholder="再输入一次密码" class="flex h-9 w-full rounded-md border border-zinc-200 bg-transparent px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-zinc-500 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-950">
        </div>
        <div class="flex gap-3 pt-2">
          <button onclick="goStep(1)" class="inline-flex w-[100px] items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-950 bg-white border border-zinc-200 text-zinc-900 shadow-sm hover:bg-zinc-100 hover:text-zinc-900 h-9 px-4 py-2">
            上一步
          </button>
          <button id="submitBtn" onclick="submitInstall()" class="inline-flex flex-1 items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-950 disabled:pointer-events-none disabled:opacity-50 bg-zinc-900 text-zinc-50 shadow hover:bg-zinc-900/90 h-9 px-4 py-2">
            完成安装
          </button>
        </div>
      </div>

      <!-- Step 3 -->
      <div class="step" id="step3">
        <div class="flex flex-col items-center justify-center space-y-3 py-4">
          <div class="h-12 w-12 rounded-full bg-emerald-100 flex items-center justify-center mb-2">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-emerald-600"><path d="M20 6 9 17l-5-5"/></svg>
          </div>
          <h2 class="text-xl font-semibold tracking-tight">安装完成！</h2>
          <p class="text-sm text-zinc-500 text-center">
            配置已写入数据库，请重启 Kite 以加载配置。
          </p>
        </div>
        
        <div id="installInfo" class="mt-4 rounded-md border border-zinc-200 divide-y divide-zinc-200">
          <!-- Dynamically populated -->
        </div>
      </div>

    </div>
  </div>

  <p class="mt-8 text-center text-sm text-zinc-500">
    &copy; 2026 Kite. Powered by AI.
  </p>

  <script>
    function goStep(n) {
      hideError();
      if (n === 2) {
        var siteName = document.getElementById('siteName').value.trim();
        if (!siteName) { showError('请输入站点名称'); return; }
      }
      document.querySelectorAll('.step').forEach(function(el) { el.classList.remove('active'); });
      document.getElementById('step' + n).classList.add('active');
      
      const dots = [document.getElementById('dot1'), document.getElementById('dot2'), document.getElementById('dot3')];
      dots.forEach((dot, i) => {
        dot.className = 'h-2 w-2 rounded-full transition-all duration-300';
        if (i + 1 < n) dot.classList.add('bg-zinc-900');
        else if (i + 1 === n) dot.classList.add('bg-zinc-900', 'w-6');
        else dot.classList.add('bg-zinc-200');
      });
    }

    function showError(msg) { 
      var e = document.getElementById('error'); 
      e.textContent = msg; 
      e.classList.remove('hidden'); 
    }
    
    function hideError() { 
      document.getElementById('error').classList.add('hidden'); 
    }

    function submitInstall() {
      hideError();
      var u = document.getElementById('username').value.trim();
      var p = document.getElementById('password').value;
      var pc = document.getElementById('passwordConfirm').value;
      
      if (!u) { showError('请输入管理员用户名'); return; }
      if (p.length < 6) { showError('密码至少 6 位'); return; }
      if (p !== pc) { showError('两次输入的密码不一致'); return; }
      
      var btn = document.getElementById('submitBtn');
      btn.disabled = true; 
      btn.textContent = '安装中…';
      
      fetch('/api/install', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          site_name: document.getElementById('siteName').value.trim(), 
          username: u, 
          password: p 
        })
      }).then(function(r) { return r.json(); }).then(function(d) {
        if (d.code !== 200) { 
          showError(d.msg || '安装失败'); 
          btn.disabled = false; 
          btn.textContent = '完成安装'; 
          return; 
        }
        
        document.getElementById('installInfo').innerHTML =
          '<div class="flex justify-between items-center p-3 text-sm"><span class="text-zinc-500">站点名称</span><span class="font-medium">' + d.data.site_name + '</span></div>' +
          '<div class="flex justify-between items-center p-3 text-sm"><span class="text-zinc-500">管理员</span><span class="font-medium">' + d.data.username + '</span></div>' +
          '<div class="flex justify-between items-center p-3 text-sm"><span class="text-zinc-500">数据库</span><span class="font-medium text-emerald-600">kite.db (SQLite)</span></div>';
        
        goStep(3);
      }).catch(function(e) { 
        showError('网络错误: ' + e.message); 
        btn.disabled = false; 
        btn.textContent = '完成安装'; 
      });
    }
  </script>
</body>
</html>`

var installTmpl = template.Must(template.New("install").Parse(installPageHTML))

// startInstallServer 启动安装引导 Web 服务器
func startInstallServer(dbPath string) {
	installed := false

	mux := http.NewServeMux()

	// 所有页面都展示安装引导
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/install" {
			handleInstallAPI(w, r, dbPath, &installed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if installed {
			fmt.Fprint(w, `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Kite</title></head><body style="font-family:sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;"><div style="text-align:center;"><h1>🪁 安装完成</h1><p style="color:#64748b;margin:1rem 0;">请重启 Kite 以加载配置并启动博客。</p></div></body></html>`)
			return
		}
		installTmpl.Execute(w, nil)
	})

	addr := ":8080"
	log.Printf("🪁 Kite 首次运行 — 请访问安装引导页: http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
		log.Fatalf("install server: %v", err)
	}
}

// handleInstallAPI 处理安装表单提交（直接写入 SQLite）
func handleInstallAPI(w http.ResponseWriter, r *http.Request, dbPath string, installed *bool) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		fmt.Fprint(w, `{"code":405,"msg":"method not allowed"}`)
		return
	}
	if *installed || dbExists(dbPath) {
		fmt.Fprint(w, `{"code":400,"msg":"Kite 已安装，请重启"}`)
		return
	}

	var req installRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprint(w, `{"code":400,"msg":"请求格式错误"}`)
		return
	}

	req.SiteName = strings.TrimSpace(req.SiteName)
	req.Username = strings.TrimSpace(req.Username)
	if req.SiteName == "" {
		req.SiteName = "Kite"
	}
	if req.Username == "" {
		fmt.Fprint(w, `{"code":400,"msg":"管理员用户名不能为空"}`)
		return
	}
	if len(req.Password) < 6 {
		fmt.Fprint(w, `{"code":400,"msg":"密码至少 6 位"}`)
		return
	}

	// 生成 bcrypt 哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprint(w, `{"code":500,"msg":"密码哈希生成失败"}`)
		return
	}

	// 生成 session secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		fmt.Fprint(w, `{"code":500,"msg":"session secret 生成失败"}`)
		return
	}
	sessionSecret := hex.EncodeToString(secretBytes)

	// 创建数据库并建表
	db, err := repo.InitSQLiteDB(dbPath)
	if err != nil {
		fmt.Fprintf(w, `{"code":500,"msg":"创建数据库失败: %s"}`, err.Error())
		return
	}

	// 组装配置并写入 DB
	cfg := config.Default()
	cfg.Database.Driver = config.DatabaseDriverSQLite
	cfg.Database.Path = dbPath
	cfg.Admin.Enabled = true
	cfg.Admin.Username = req.Username
	cfg.Admin.PasswordHash = string(hash)
	cfg.Admin.SessionSecret = sessionSecret
	cfg.Admin.SessionTTLHours = 168
	cfg.Site.SiteName = req.SiteName

	settingsRepo := repo.NewSettingsRepository(db)
	settingsService := service.NewSettingsService(cfg, settingsRepo)
	if err := settingsService.SaveInitialSettings(); err != nil {
		// 回滚：删除创建的 DB 文件
		os.Remove(dbPath)
		fmt.Fprintf(w, `{"code":500,"msg":"写入配置失败: %s"}`, err.Error())
		return
	}

	// 创建欢迎文章
	postRepo := repo.NewPostRepository(db)
	welcomeMarkdown := `🎉 恭喜你成功安装了 **Kite** —— 一个轻量级 AI 原生博客引擎！

## ✨ Kite 的特色

- **单二进制部署** — 一个文件就是整个博客，无需 Runtime / Docker
- **AI 原生** — 内置 AI 摘要、标签建议（大模型 API 可选）
- **富文本编辑器** — 基于 Tiptap 的所见即所得编辑，支持 Markdown 源码模式
- **SSR 主题** — Go Template 渲染，PaperMod 风格，SEO 友好
- **密码保护** — 支持全文加密和片段加密
- **友链管理** — 内置友链页面和管理功能

## 🚀 接下来可以做什么？

1. 前往 [后台管理](/admin) 开始写第一篇文章
2. 在 **设置** 中自定义站点名称、描述、备案号
3. 添加分类和标签来组织你的内容
4. 在 **友链** 页面添加好友的博客链接

## 📝 关于 Markdown

Kite 支持完整的 Markdown 语法，包括：

- **代码块** — 支持语法高亮
- **表格** — 标准 Markdown 表格
- **Callout** — 提示、警告等信息块
- **链接和图片** — 拖拽 / 粘贴上传图片

> 这是第一篇文章，你可以编辑或删除它。祝你写作愉快！🪁`

	welcomeHTML := `<p>🎉 恭喜你成功安装了 <strong>Kite</strong> —— 一个轻量级 AI 原生博客引擎！</p>
<h2>✨ Kite 的特色</h2>
<ul>
<li><strong>单二进制部署</strong> — 一个文件就是整个博客，无需 Runtime / Docker</li>
<li><strong>AI 原生</strong> — 内置 AI 摘要、标签建议（大模型 API 可选）</li>
<li><strong>富文本编辑器</strong> — 基于 Tiptap 的所见即所得编辑，支持 Markdown 源码模式</li>
<li><strong>SSR 主题</strong> — Go Template 渲染，PaperMod 风格，SEO 友好</li>
<li><strong>密码保护</strong> — 支持全文加密和片段加密</li>
<li><strong>友链管理</strong> — 内置友链页面和管理功能</li>
</ul>
<h2>🚀 接下来可以做什么？</h2>
<ol>
<li>前往 <a href="/admin">后台管理</a> 开始写第一篇文章</li>
<li>在 <strong>设置</strong> 中自定义站点名称、描述、备案号</li>
<li>添加分类和标签来组织你的内容</li>
<li>在 <strong>友链</strong> 页面添加好友的博客链接</li>
</ol>
<h2>📝 关于 Markdown</h2>
<p>Kite 支持完整的 Markdown 语法，包括：</p>
<ul>
<li><strong>代码块</strong> — 支持语法高亮</li>
<li><strong>表格</strong> — 标准 Markdown 表格</li>
<li><strong>Callout</strong> — 提示、警告等信息块</li>
<li><strong>链接和图片</strong> — 拖拽 / 粘贴上传图片</li>
</ul>
<blockquote><p>这是第一篇文章，你可以编辑或删除它。祝你写作愉快！🪁</p></blockquote>`

	now := time.Now().UTC()
	welcomePost := &model.Post{
		Title:           "欢迎使用 Kite 🪁",
		Slug:            "hello-kite",
		Summary:         "恭喜你成功安装了 Kite —— 一个轻量级 AI 原生博客引擎！了解 Kite 的特色功能和使用方法。",
		ContentMarkdown: welcomeMarkdown,
		ContentHTML:     welcomeHTML,
		Status:          model.PostStatusPublished,
		PublishedAt:     &now,
		ShowComments:    true,
	}
	if err := postRepo.Create(welcomePost); err != nil {
		log.Printf("⚠ 创建欢迎文章失败: %v", err)
	}

	*installed = true
	log.Printf("✓ 安装完成: site=%s, admin=%s, db=%s", req.SiteName, req.Username, dbPath)

	fmt.Fprintf(w, `{"code":200,"data":{"site_name":"%s","username":"%s"},"msg":"ok"}`, req.SiteName, req.Username)
}

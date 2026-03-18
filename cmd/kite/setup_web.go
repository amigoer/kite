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
  <style>
    :root {
      --c-bg: #f8fafc;
      --c-surface: #ffffff;
      --c-text: #0f172a;
      --c-text-secondary: #64748b;
      --c-accent: #2563eb;
      --c-accent-hover: #1d4ed8;
      --c-border: #e2e8f0;
      --c-error: #ef4444;
      --c-success: #22c55e;
      --radius: 8px;
      --font: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    }
    @media (prefers-color-scheme: dark) {
      :root {
        --c-bg: #0f172a;
        --c-surface: #1e293b;
        --c-text: #e2e8f0;
        --c-text-secondary: #94a3b8;
        --c-accent: #60a5fa;
        --c-accent-hover: #93bbfc;
        --c-border: #334155;
      }
    }
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: var(--font);
      background: var(--c-bg);
      color: var(--c-text);
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 2rem;
    }
    .install-card {
      background: var(--c-surface);
      border: 1px solid var(--c-border);
      border-radius: 12px;
      padding: 2.5rem;
      width: 100%;
      max-width: 440px;
    }
    .install-logo {
      text-align: center;
      margin-bottom: 0.5rem;
    }
    .install-logo h1 {
      font-size: 2rem;
      font-weight: 800;
      letter-spacing: -0.03em;
    }
    .install-logo h1 span { color: var(--c-accent); }
    .install-desc {
      text-align: center;
      color: var(--c-text-secondary);
      font-size: 0.875rem;
      margin-bottom: 2rem;
    }
    .step-indicator {
      display: flex;
      justify-content: center;
      gap: 0.5rem;
      margin-bottom: 2rem;
    }
    .step-dot {
      width: 8px; height: 8px;
      border-radius: 50%;
      background: var(--c-border);
      transition: all .2s;
    }
    .step-dot.active { background: var(--c-accent); width: 24px; border-radius: 4px; }
    .step-dot.done { background: var(--c-success); }
    .form-group { margin-bottom: 1.25rem; }
    .form-group label {
      display: block;
      font-size: 0.8125rem;
      font-weight: 600;
      margin-bottom: 0.375rem;
      color: var(--c-text);
    }
    .form-group input {
      width: 100%;
      padding: 0.625rem 0.875rem;
      border: 1px solid var(--c-border);
      border-radius: var(--radius);
      font-size: 0.9375rem;
      background: var(--c-bg);
      color: var(--c-text);
      outline: none;
      transition: border-color .15s;
    }
    .form-group input:focus { border-color: var(--c-accent); }
    .btn {
      display: block;
      width: 100%;
      padding: 0.75rem;
      border: none;
      border-radius: var(--radius);
      font-size: 0.9375rem;
      font-weight: 600;
      cursor: pointer;
      transition: background .15s;
    }
    .btn-primary { background: var(--c-accent); color: #fff; }
    .btn-primary:hover { background: var(--c-accent-hover); }
    .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
    .btn-row { display: flex; gap: 0.75rem; margin-top: 1.5rem; }
    .btn-secondary {
      background: var(--c-bg);
      color: var(--c-text);
      border: 1px solid var(--c-border);
      flex: 1;
    }
    .btn-row .btn-primary { flex: 2; }
    .error-msg {
      background: #fef2f2;
      border: 1px solid #fecaca;
      color: var(--c-error);
      padding: 0.625rem 0.875rem;
      border-radius: var(--radius);
      font-size: 0.8125rem;
      margin-bottom: 1rem;
      display: none;
    }
    @media (prefers-color-scheme: dark) {
      .error-msg { background: #450a0a; border-color: #7f1d1d; }
    }
    .step { display: none; }
    .step.active { display: block; }
    .success-icon { text-align: center; font-size: 3rem; margin-bottom: 1rem; }
    .success-text { text-align: center; margin-bottom: 1.5rem; }
    .success-text h2 { font-size: 1.25rem; margin-bottom: 0.5rem; }
    .success-text p { color: var(--c-text-secondary); font-size: 0.875rem; }
    .info-row {
      display: flex;
      justify-content: space-between;
      padding: 0.5rem 0;
      border-bottom: 1px solid var(--c-border);
      font-size: 0.875rem;
    }
    .info-row:last-child { border-bottom: none; }
    .info-row .label { color: var(--c-text-secondary); }
  </style>
</head>
<body>
  <div class="install-card">
    <div class="install-logo"><h1>🪁 <span>Kite</span></h1></div>
    <p class="install-desc">轻量级 AI 原生博客引擎 — 安装引导</p>

    <div class="step-indicator">
      <div class="step-dot active" id="dot1"></div>
      <div class="step-dot" id="dot2"></div>
      <div class="step-dot" id="dot3"></div>
    </div>

    <div id="error" class="error-msg"></div>

    <div class="step active" id="step1">
      <div class="form-group">
        <label for="siteName">站点名称</label>
        <input type="text" id="siteName" value="Kite" placeholder="给你的博客取个名字">
      </div>
      <div class="btn-row">
        <button class="btn btn-primary" onclick="goStep(2)">下一步</button>
      </div>
    </div>

    <div class="step" id="step2">
      <div class="form-group">
        <label for="username">管理员用户名</label>
        <input type="text" id="username" placeholder="用于登录后台" autocomplete="username">
      </div>
      <div class="form-group">
        <label for="password">管理员密码</label>
        <input type="password" id="password" placeholder="至少 6 位" autocomplete="new-password">
      </div>
      <div class="form-group">
        <label for="passwordConfirm">确认密码</label>
        <input type="password" id="passwordConfirm" placeholder="再输入一次密码" autocomplete="new-password">
      </div>
      <div class="btn-row">
        <button class="btn btn-secondary" onclick="goStep(1)">上一步</button>
        <button class="btn btn-primary" id="submitBtn" onclick="submitInstall()">完成安装</button>
      </div>
    </div>

    <div class="step" id="step3">
      <div class="success-icon">🎉</div>
      <div class="success-text">
        <h2>安装完成！</h2>
        <p>配置已写入数据库，请重启 Kite 以加载配置。</p>
      </div>
      <div id="installInfo"></div>
    </div>
  </div>

  <script>
    function goStep(n) {
      hideError();
      if (n === 2) {
        var siteName = document.getElementById('siteName').value.trim();
        if (!siteName) { showError('请输入站点名称'); return; }
      }
      document.querySelectorAll('.step').forEach(function(el) { el.classList.remove('active'); });
      document.getElementById('step' + n).classList.add('active');
      document.querySelectorAll('.step-dot').forEach(function(dot, i) {
        dot.classList.remove('active', 'done');
        if (i + 1 < n) dot.classList.add('done');
        if (i + 1 === n) dot.classList.add('active');
      });
    }
    function showError(msg) { var e = document.getElementById('error'); e.textContent = msg; e.style.display = 'block'; }
    function hideError() { document.getElementById('error').style.display = 'none'; }
    function submitInstall() {
      hideError();
      var u = document.getElementById('username').value.trim();
      var p = document.getElementById('password').value;
      var pc = document.getElementById('passwordConfirm').value;
      if (!u) { showError('请输入管理员用户名'); return; }
      if (p.length < 6) { showError('密码至少 6 位'); return; }
      if (p !== pc) { showError('两次输入的密码不一致'); return; }
      var btn = document.getElementById('submitBtn');
      btn.disabled = true; btn.textContent = '安装中…';
      fetch('/api/install', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ site_name: document.getElementById('siteName').value.trim(), username: u, password: p })
      }).then(function(r) { return r.json(); }).then(function(d) {
        if (d.code !== 200) { showError(d.msg || '安装失败'); btn.disabled = false; btn.textContent = '完成安装'; return; }
        document.getElementById('installInfo').innerHTML =
          '<div class="info-row"><span class="label">站点名称</span><span>' + d.data.site_name + '</span></div>' +
          '<div class="info-row"><span class="label">管理员</span><span>' + d.data.username + '</span></div>' +
          '<div class="info-row"><span class="label">数据库</span><span>kite.db (SQLite)</span></div>';
        goStep(3);
      }).catch(function(e) { showError('网络错误: ' + e.message); btn.disabled = false; btn.textContent = '完成安装'; });
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

# 🪁 Kite Blog

轻量级 AI 原生博客引擎，Go + React + SQLite 单二进制部署。

## 特性

- 🚀 **单二进制部署** — 无需外部依赖，`CGO_ENABLED=0` 纯静态编译
- 🎨 **双渲染模式** — SSR（经典）/ Headless（纯 API）+ 主题系统
- 🤖 **AI 原生** — 集成 AI 自动摘要、自动标签
- ✍️ **富文本编辑** — 基于 Tiptap 的现代编辑器
- 📦 **内嵌 SQLite** — 零配置数据库，纯 Go 实现
- 🌐 **Web 安装引导** — 首次运行浏览器交互式设置

## 快速开始

### 1. 编译

```bash
go build -o kite ./cmd/kite
```

### 2. 运行

```bash
./kite
```

首次运行（无 `kite.yaml`）会自动启动安装引导页：

```
🪁 Kite 首次运行 — 请访问安装引导页: http://localhost:8080
```

打开浏览器，填写站点名称和管理员账号即可完成安装。

### 3. 配置

安装引导自动生成 `kite.yaml`，完整配置项参考 [`kite.example.yaml`](kite.example.yaml)。

## 开发

### 后端

```bash
# 运行后端（需有 kite.yaml）
go run ./cmd/kite
```

### 前端 Admin

```bash
cd ui/admin
npm install
npm run dev          # 启动开发服务器 (Vite)
```

Vite 已配置代理，开发时 `/api` 请求自动转发到 `http://localhost:8080`。

### SSR 主题

模板目录 `templates/`，参考 [`docs/theme-dev.md`](docs/theme-dev.md)。

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go, Gin, GORM, SQLite (modernc) |
| 前端 | React 19, TypeScript, Vite, Semi Design |
| 编辑器 | Tiptap |
| 模板 | html/template |

## 文档

- [API 文档](docs/api.md)
- [SSR 接口文档](docs/ssr.md)
- [主题开发指南](docs/theme-dev.md)

## License

MIT

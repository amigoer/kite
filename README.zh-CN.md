<div align="center">
  <img src="web/admin/public/logo.png" alt="Kite logo" width="88" height="88" />
  <h1>Kite 静态资源托管平台</h1>
  <p>一个轻量、快速、现代化的静态资源托管平台。</p>
  <p>
    <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" alt="Go 1.25+" />
    <img src="https://img.shields.io/badge/Gin-1.12-008ECF?logo=gin&logoColor=white" alt="Gin 1.12" />
    <img src="https://img.shields.io/badge/React-19-20232A?logo=react&logoColor=61DAFB" alt="React 19" />
    <img src="https://img.shields.io/badge/Vite-8-646CFF?logo=vite&logoColor=white" alt="Vite 8" />
    <img src="https://img.shields.io/badge/License-MIT-16A34A" alt="MIT License" />
    <img src="https://img.shields.io/badge/Storage-Local%20%7C%20S3%20%7C%20FTP-111827" alt="Storage Drivers" />
  </p>
  <p>
    简体中文 | <a href="README.md">English</a>
  </p>
</div>

## 🪁 简介

Kite 是一个轻量、快速、现代化的静态资源托管平台。它不仅仅是一个图床，更是一个涵盖了图片、音视频以及各类静态文件托管的综合性解决方案。就像天空中翱翔的风筝一样，Kite 旨在为您提供无缝、轻松的全类型媒体资源管理和分发体验。

Kite 由 Go 后端、内嵌式 React 管理后台和可扩展的存储抽象层组成，支持本地磁盘、FTP 以及主流对象存储服务。它适合自托管图床、静态资源分发、内部媒体库和轻量团队部署等场景。

## ✨ 特性

- **全格式支持**：轻松托管和分发图片、音频、视频以及其他标准的静态文件格式。
- **轻量且快速**：采用现代 Web 技术构建，提供卓越的性能与极致的资源加载速度。
- **简单易用**：简洁直观的用户界面，方便您上传、管理和查看各类资源。
- **中英双语界面**：内置简体中文与英文两种语言，公共站点、管理后台与后端响应消息会随同一个语言开关同步切换。
- **首次安装向导**：全新部署会引导您通过 `/setup` 完成管理员账号、站点 URL 与默认存储后端的初始化，全程不会暴露任何临时凭据。
- **安全可靠**：强大的安全措施确保您的数据和资产安全。
- **开源免费**：自由使用、修改和分发。

## 🧱 技术栈

- **后端**：Go、Gin、GORM、SQLite / MySQL / PostgreSQL
- **前端**：React、TypeScript、Vite、TanStack Query、Radix UI
- **存储**：本地、S3、FTP
- **媒体能力**：缩略图、静态文件分发、公开上传接口

## 🚀 快速开始

### 开发模式

```bash
make dev
```

该命令会同时启动 Go 后端和管理后台前端开发服务。

首次克隆仓库后，建议执行一次下面的命令启用 Git Hook：

```bash
make hooks-install
```

`pre-commit` 负责提交前的快速修正：自动格式化已暂存的 Go 与管理后台代码，并执行 `go mod tidy`。

`pre-push` 会在一个隔离的 `HEAD` 临时检出目录里执行较慢的完整校验，包括 `go test`、Go 构建校验，以及管理后台构建校验。

### 生产构建

```bash
make build
./build/kite
```

生产构建会先编译前端，并将其嵌入到 Go 二进制中。

### 首次安装向导

全新部署（settings 表中没有 `is_installed` 标记）时，Kite 不会再自动创建默认管理员，而是引导操作者通过 `/setup` 完成首次配置。向导会依次询问：

- **站点信息**：站点名称与对外可访问的 URL（用于 OAuth 回调与生成绝对链接）。
- **管理员账号**：用户名、邮箱与密码（密码至少 6 位）。账号会被赋予管理员角色，整个流程不会向数据库或日志写入任何临时凭据。
- **默认存储后端**：可在本地磁盘、S3 兼容对象存储、FTP 以及主流云厂商之间选择，并直接填写连接参数；保存后向导会更新数据库并热加载 storage manager。

如果在安装过程中需要切换数据库驱动 / DSN，向导也会把新的 `database` 配置写入磁盘上的配置文件（参见 [`KITE_CONFIG_FILE`](#环境变量)），下一次启动会按新的连接信息恢复服务。

提交成功后向导会写入 `is_installed=true`，路由也会停止响应，避免后来者再次提交表单。

#### 跳过向导

对于自动化脚本或容器化部署等不方便交互的场景，可以设置 `KITE_LEGACY_BOOTSTRAP=1` 回退到旧的引导逻辑：首次启动时仍然自动创建 `admin / admin` 账号，并打上 `password_must_change` 标记，首次登录会强制跳转重置页面。出于安全考虑，这组密码不会写入应用日志，请在首次登录后立刻修改，并将新凭据妥善保存（切勿提交进版本库）。

## ⚙️ 配置

Kite 的运行时配置由三层叠加而成，后一层会覆盖前一层：

1. **编译期默认值**：适合单机开发的合理初值。
2. **配置文件**：安装向导写入的可选 JSON 文件，默认路径为数据库目录下的 `data/config.json`，可通过 `KITE_CONFIG_FILE` 修改；如果文件不存在或字段缺失，会静默退化为「沿用上一层」。
3. **环境变量**：优先级最高，常用于临时覆盖（端口转发、CI 中替换 DSN、临时调高日志等级等）。

部分运行时可调字段（站点名称、注册策略、JWT 秘钥等）还会保存在 `settings` 表中并由管理后台编辑——这些字段会覆盖静态配置文件，但仍可被环境变量再覆盖。

### 环境变量

| 变量 | 作用 | 示例 |
| --- | --- | --- |
| `KITE_PORT` | HTTP 监听端口。 | `8080` |
| `KITE_HOST` | 监听地址。容器中通常使用 `0.0.0.0` 以接受外部请求。 | `0.0.0.0` |
| `KITE_DSN` | 数据库连接串，格式取决于 `KITE_DB_DRIVER`。 | `data/kite.db` |
| `KITE_DB_DRIVER` | 数据库驱动，可选 `sqlite`、`mysql`、`postgres`。 | `sqlite` |
| `KITE_SITE_URL` | 站点对外 URL，用于生成 OAuth 回调与绝对链接。 | `https://kite.example.com` |
| `KITE_LOG_LEVEL` | 最低日志等级，可选 `debug`、`info`、`warn`、`error`，默认 `info`。 | `info` |
| `KITE_LOG_FORMAT` | 日志格式，可选 `text`、`json`，默认 `text`；接入日志收集时建议切换为 `json`。 | `json` |
| `KITE_CORS_ORIGINS` | 静态资源接口（`/i/`、`/v/`、`/a/`、`/f/`）允许跨域的来源列表，使用英文逗号分隔；未设置时只接受同源请求。 | `https://kite.example.com,https://app.example.com` |
| `KITE_LEGACY_BOOTSTRAP` | 设为 `1` 时跳过安装向导，首启动按旧逻辑自动创建 `admin / admin` 账号，方便脚本化部署。 | `1` |
| `KITE_CONFIG_FILE` | 安装向导维护的 JSON 配置文件路径，默认为 `data/config.json`。 | `/etc/kite/config.json` |

环境变量会同时覆盖编译期默认值与配置文件，配合 `make dev` 等本地脚本即可灵活控制。

### 国际化

后端消息目录、公开落地页与管理后台 SPA 共用同一个 locale 解析链：

1. `kite_locale` Cookie（由公共头部的语言切换按钮与 SPA 的 `setLocale` 写入）。
2. `Accept-Language` 请求头。
3. 静态兜底——英文。

任意一个界面（公开站点、管理后台、服务端渲染的上传页）切换语言后都会更新 Cookie，下次渲染其他界面时就会看到一致的语言。当前内置英文与简体中文，新增第三种语言只需在 `internal/i18n` 中扩展后端目录，并在 `web/admin/src/i18n/locales` 添加同名文件。

## 📦 项目结构

```text
cmd/kite            应用入口
internal/           后端处理器、服务、仓储与存储驱动
web/admin/          React 管理后台
template/           落地页与上传页模板
deploy/             docker-compose 与 nginx 示例
```

## 🐳 部署

- 使用 [deploy/docker-compose.yml](deploy/docker-compose.yml) 快速容器化部署。
- 使用 [Dockerfile](Dockerfile) 构建独立镜像。
- 使用 [deploy/nginx/conf.d/www.kite.plus.conf](deploy/nginx/conf.d/www.kite.plus.conf) 作为反向代理参考。

## 🤝 参与贡献

欢迎任何形式的贡献！

## 📄 开源协议

本项目采用 MIT 协议 - 详情请见 [LICENSE](LICENSE) 文件。

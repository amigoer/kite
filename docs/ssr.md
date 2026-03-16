# Kite SSR 接口文档

## 概述

当 `config.render_mode=classic` 时，Kite 后端通过 SSR（服务端渲染）直接输出 HTML 页面。SSR 与 JSON API 共存，分别服务前台访客和后台管理端。

## SSR 路由表

| 路由 | 模板 | 数据源 | 描述 |
|------|------|--------|------|
| `GET /` | `index.html` | Post 列表 + 分页 | 首页 |
| `GET /posts/:slug` | `post.html` | Post 详情 | 文章页 |
| `GET /categories/:slug` | `archive.html` | Category + Post 列表 | 分类归档 |
| `GET /tags/:slug` | `archive.html` | Tag + Post 列表 | 标签归档 |
| `GET /pages/:slug` | `pages/{template}.html` | Page 详情 + Config | 独立页面 |
| `GET /friends` | `friends.html` | FriendLink 列表 | 友链页 |
| `GET /static/*` | — | 嵌入式文件服务 | 静态资源 |
| 其他路由 | `404.html` | — | 404 兜底 |

## 模板上下文数据结构

### 公共数据（所有页面共享）

每个页面模板都能访问以下字段：

| 字段 | 类型 | 来源 | 描述 |
|------|------|------|------|
| `.SiteName` | string | Config.Site | 站点名称 |
| `.Description` | string | Config.Site | 站点描述 |
| `.Keywords` | string | Config.Site | SEO 关键词 |
| `.Favicon` | string | Config.Site | Favicon URL |
| `.Logo` | string | Config.Site | Logo URL |
| `.ICP` | string | Config.Site | 备案号 |
| `.Footer` | string | Config.Site | 页脚文本 |
| `.PageTitle` | string | 动态 | 当前页标题（用于 `<title>` 标签） |
| `.NavPages` | []Page | DB 查询 | 导航栏页面（`show_in_nav=true` 的已发布页面） |

### 首页（index.html）

| 字段 | 类型 | 描述 |
|------|------|------|
| `.Posts` | []Post | 文章列表 |
| `.Pagination` | PaginationData | 分页数据 |

### 文章详情（post.html）

| 字段 | 类型 | 描述 |
|------|------|------|
| `.Post` | Post | 文章对象（含 Category、Tags） |

### 归档页（archive.html）

| 字段 | 类型 | 描述 |
|------|------|------|
| `.Posts` | []Post | 文章列表 |
| `.ArchiveType` | string | `"分类"` 或 `"标签"` |
| `.ArchiveName` | string | 分类/标签名称 |
| `.Pagination` | PaginationData | 分页数据 |

### 独立页面（pages/*.html）

| 字段 | 类型 | 描述 |
|------|------|------|
| `.Page` | Page | 页面对象 |
| `.PageConfig` | map[string]any | Page.Config JSON 解析后的键值对 |

### 友链页（friends.html）

| 字段 | 类型 | 描述 |
|------|------|------|
| `.FriendLinks` | []FriendLink | 友情链接列表 |

### PaginationData 结构

| 字段 | 类型 | 描述 |
|------|------|------|
| `.CurrentPage` | int | 当前页码 |
| `.TotalPages` | int | 总页数 |
| `.Total` | int64 | 总记录数 |
| `.PageSize` | int | 每页数量 |
| `.BasePath` | string | 基础路径（用于分页链接拼接） |
| `.PageNumbers` | []int | 页码列表（最多 7 项） |

## 自定义模板函数

模板中可使用以下函数：

| 函数 | 签名 | 描述 |
|------|------|------|
| `formatDate` | `formatDate(time) → string` | 格式化时间为 `2006-01-02` |
| `add` | `add(a, b int) → int` | 加法 |
| `subtract` | `subtract(a, b int) → int` | 减法 |

使用示例：

```html
<time>{{ formatDate .Post.PublishedAt }}</time>

{{ if gt .CurrentPage 1 }}
  <a href="{{ .BasePath }}?page={{ subtract .CurrentPage 1 }}">上一页</a>
{{ end }}
```

## 静态资源服务

静态资源位于 `templates/static/` 目录，通过 `/static/` 路径访问：

```
templates/static/css/kite.css  →  GET /static/css/kite.css
templates/static/js/app.js     →  GET /static/js/app.js
templates/static/img/logo.png  →  GET /static/img/logo.png
```

静态文件嵌入到二进制中，部署时不需要额外的文件依赖。

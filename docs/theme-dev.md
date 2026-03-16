# Kite 主题模板开发指南

## 概述

Kite 使用 Go `html/template` 作为模板引擎。一个完整的主题就是一个 `templates/` 目录，包含 HTML 模板、组件片段和静态资源。

## 目录结构

```
templates/
├── static/                # 静态资源（CSS/JS/图片）
│   ├── css/
│   │   └── kite.css       # 主题样式
│   ├── js/                # JavaScript 文件
│   └── img/               # 图片资源
├── partials/              # 可复用的模板片段
│   ├── head.html          # <head> 标签（引入 CSS、meta 等）
│   ├── header.html        # 站点头部和导航栏
│   ├── footer.html        # 站点页脚
│   ├── pagination.html    # 分页组件
│   └── post_card.html     # 文章列表卡片
├── pages/                 # 独立页面模板
│   ├── default.html       # 默认页面模板
│   ├── default.json       # 默认模板清单
│   ├── github.html        # GitHub 仓库展示
│   └── github.json        # GitHub 模板清单
├── index.html             # 首页
├── post.html              # 文章详情页
├── archive.html           # 分类/标签归档页
├── friends.html           # 友情链接页
└── 404.html               # 404 错误页
```

## 快速入门

### 1. 页面模板基本结构

每个页面模板必须用 `{{ define "文件名" }}...{{ end }}` 包裹：

```html
{{ define "index.html" }}
<!DOCTYPE html>
<html lang="zh-CN">
{{ template "head" . }}
<body>
  {{ template "header" . }}

  <main>
    <div class="container">
      <!-- 页面内容 -->
    </div>
  </main>

  {{ template "footer" . }}
</body>
</html>
{{ end }}
```

要点：
- `define` 名称必须与文件名一致
- 用 `{{ template "head" . }}` 引入片段，`.` 表示传递当前上下文
- 所有页面都能访问公共数据（站点名、导航等）

### 2. 片段模板（Partials）

片段用 `{{ define "名称" }}...{{ end }}` 定义，在页面中用 `{{ template "名称" . }}` 引入：

```html
<!-- partials/header.html -->
{{ define "header" }}
<header>
  <a href="/">{{ .SiteName }}</a>
  <nav>
    {{ range .NavPages }}
      <a href="/pages/{{ .Slug }}">{{ .Title }}</a>
    {{ end }}
  </nav>
</header>
{{ end }}
```

### 3. 公共数据

所有页面模板中都可以使用以下变量：

| 变量 | 描述 |
|------|------|
| `{{ .SiteName }}` | 站点名称 |
| `{{ .Description }}` | 站点描述 |
| `{{ .Favicon }}` | Favicon URL |
| `{{ .ICP }}` | 备案号 |
| `{{ .Footer }}` | 页脚自定义文本 |
| `{{ .PageTitle }}` | 当前页标题 |
| `{{ .NavPages }}` | 导航栏页面列表 |

## 各页面模板开发

### 首页（index.html）

可用数据：

```html
<!-- 文章列表 -->
{{ range .Posts }}
  <h2><a href="/posts/{{ .Slug }}">{{ .Title }}</a></h2>
  <time>{{ formatDate .CreatedAt }}</time>
  <p>{{ .Summary }}</p>

  <!-- 分类 -->
  {{ if .Category }}
    <a href="/categories/{{ .Category.Slug }}">{{ .Category.Name }}</a>
  {{ end }}

  <!-- 标签 -->
  {{ range .Tags }}
    <a href="/tags/{{ .Slug }}">{{ .Name }}</a>
  {{ end }}
{{ end }}

<!-- 分页 -->
{{ template "pagination" .Pagination }}
```

### 文章详情（post.html）

可用数据：

```html
<h1>{{ .Post.Title }}</h1>
<time>{{ formatDate .Post.PublishedAt }}</time>

<!-- 分类和标签 -->
{{ if .Post.Category }}
  <a href="/categories/{{ .Post.Category.Slug }}">{{ .Post.Category.Name }}</a>
{{ end }}
{{ range .Post.Tags }}
  <a href="/tags/{{ .Slug }}">{{ .Name }}</a>
{{ end }}

<!-- 文章正文（已渲染的 HTML） -->
<div class="post-content">
  {{ .Post.ContentHTML }}
</div>
```

### 归档页（archive.html）

用于展示某个分类或标签下的文章列表：

```html
<h1>{{ .ArchiveType }}：{{ .ArchiveName }}</h1>
<!-- .ArchiveType 为 "分类" 或 "标签" -->

{{ range .Posts }}
  ...
{{ end }}

{{ template "pagination" .Pagination }}
```

### 友链页（friends.html）

```html
{{ range .FriendLinks }}
  <a href="{{ .URL }}">
    <img src="{{ .Logo }}" alt="{{ .Name }}">
    <h3>{{ .Name }}</h3>
    <p>{{ .Description }}</p>
  </a>
{{ end }}
```

## 自定义页面模板

独立页面支持选择不同的模板。模板作者可以创建专门的展示页面。

### 创建自定义页面模板

1. 在 `templates/pages/` 下创建模板文件和清单文件：

```
templates/pages/
├── douban.html     # 模板文件
└── douban.json     # 清单文件
```

2. 模板文件的 `define` 名称必须是 `pages/文件名`：

```html
{{ define "pages/douban.html" }}
<!DOCTYPE html>
<html lang="zh-CN">
{{ template "head" . }}
<body>
  {{ template "header" . }}
  <main>
    <div class="container">
      <h1>{{ .Page.Title }}</h1>

      <!-- 读取页面配置参数 -->
      <div data-user-id="{{ index .PageConfig "user_id" }}">
        ...
      </div>
    </div>
  </main>
  {{ template "footer" . }}
</body>
</html>
{{ end }}
```

3. 清单文件声明模板所需的参数：

```json
{
  "name": "豆瓣书影展示",
  "description": "展示指定用户的豆瓣观影和阅读记录",
  "fields": [
    {
      "key": "user_id",
      "label": "豆瓣用户 ID",
      "type": "string",
      "required": true,
      "placeholder": "例如：123456789"
    },
    {
      "key": "show_books",
      "label": "显示书籍",
      "type": "boolean",
      "default": true
    },
    {
      "key": "show_movies",
      "label": "显示电影",
      "type": "boolean",
      "default": true
    }
  ]
}
```

### 清单文件规范

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `name` | string | ✅ | 模板显示名称 |
| `description` | string | ✅ | 模板描述 |
| `fields` | array | ✅ | 配置参数列表 |
| `fields[].key` | string | ✅ | 参数键名 |
| `fields[].label` | string | ✅ | 显示标签 |
| `fields[].type` | string | ✅ | 类型：`string` / `number` / `boolean` |
| `fields[].required` | boolean | — | 是否必填 |
| `fields[].default` | any | — | 默认值 |
| `fields[].placeholder` | string | — | 输入提示 |

### 自定义页面模板中的数据

页面模板除了公共数据外，还有：

| 变量 | 描述 |
|------|------|
| `{{ .Page }}` | Page 对象（title, slug, content_html 等） |
| `{{ .PageConfig }}` | Config JSON 解析后的 map |
| `{{ index .PageConfig "key" }}` | 读取指定配置参数 |

## 可用的模板函数

| 函数 | 用法 | 描述 |
|------|------|------|
| `formatDate` | `{{ formatDate .Post.PublishedAt }}` | 输出 `2006-01-02` 格式日期 |
| `add` | `{{ add 1 2 }}` | 加法，返回 `3` |
| `subtract` | `{{ subtract 5 2 }}` | 减法，返回 `3` |

## 静态资源

主题静态资源放在 `templates/static/` 目录下，通过 `/static/` URL 路径访问：

```html
<link rel="stylesheet" href="/static/css/kite.css">
<script src="/static/js/app.js"></script>
<img src="/static/img/logo.png">
```

## 发布主题

将整个 `templates/` 目录打包发布即可。用户只需替换 Kite 安装目录下的 `templates/` 文件夹。

主题包建议结构：

```
kite-theme-xxx/
├── README.md           # 主题说明和预览截图
├── templates/          # 完整的 templates 目录
│   ├── static/
│   ├── partials/
│   ├── pages/
│   ├── index.html
│   ├── post.html
│   ├── archive.html
│   ├── friends.html
│   └── 404.html
└── screenshot.png      # 预览截图
```

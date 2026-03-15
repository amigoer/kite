# Kite

> **Flight on Code, Focused on Words.**
> Kite 是一款由 Go 驱动、极致扁平化的 AI 原生博客引擎。

[![Go Report Card](https://goreportcard.com/badge/github.com/amigoer/kite-blog)](https://goreportcard.com/report/github.com/amigoer/kite-blog)
[![License](https://img.shields.io/badge/license-MIT-black.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.25%2B-blue.svg)](https://golang.org)

## 项目介绍

Kite 是一个轻量的 Go 博客引擎，围绕极致扁平化的视觉风格与 AI 原生内容工作流设计。它希望在保持架构简洁、便于部署和易于扩展的前提下，提供干净纯粹的写作与阅读体验。

Kite 同时支持经典 SSR 渲染与 Headless API 输出，既适合传统博客，也适合与自定义前端配合使用。

## 特性

- 极致扁平化设计，零圆角、零阴影
- 面向未来 AI 内容工作流的原生架构
- 支持 Classic SSR 与 Headless 双模输出
- 同时兼容 SQLite 与 PostgreSQL
- 基于 UUID 的数据模型设计
- 基于嵌入资源的单文件部署能力

## 目录结构

```text
├── cmd/kite/main.go          # 程序入口
├── internal/                 # 私有后端核心逻辑
├── templates/                # SSR 模板
├── ui/admin/                 # 管理后台源码
├── embed.go                  # 嵌入静态资源
└── Makefile                  # 构建脚本
```

## 设计理念

Kite 遵循严格的 Flat Design：

- 零圆角
- 零阴影
- 高对比度硬边框
- 回归内容本身的阅读体验

## 开源协议

项目基于 MIT License 开源，详见 `LICENSE`。

# Kite

> **Flight on Code, Focused on Words.**
> A flat-designed, AI-native blog engine powered by Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/amigoer/kite-blog)](https://goreportcard.com/report/github.com/amigoer/kite-blog)
[![License](https://img.shields.io/badge/license-MIT-black.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.25%2B-blue.svg)](https://golang.org)

## Introduction

Kite is a lightweight blog engine built with Go, designed around a strict flat visual language and an AI-native content workflow. It aims to provide a clean writing and reading experience while keeping the architecture simple, portable, and extensible.

Kite supports both classic server-side rendering and headless API delivery, making it suitable for traditional blogs as well as custom frontends.

## Features

- Flat design with zero radius and zero shadow
- AI-native blog workflow for future summarization and content assistance
- Classic SSR and headless rendering modes
- SQLite and PostgreSQL support
- UUID-based data model design
- Single-binary deployment with embedded assets

## Project Structure

```text
├── cmd/kite/main.go          # Application entry point
├── internal/                 # Private backend logic
├── templates/                # SSR templates
├── ui/admin/                 # Admin frontend source
├── embed.go                  # Embedded assets
└── Makefile                  # Build automation
```

## Design Philosophy

Kite follows a strict flat design language:

- Zero border radius
- Zero box shadow
- Hard borders and strong contrast
- Minimal, content-first reading experience

## License

Distributed under the MIT License. See `LICENSE` for details.

<div align="center">
  <img src="web/admin/public/logo.png" alt="Kite logo" width="88" height="88" />
  <h1>Kite Static Asset Hosting</h1>
  <p>A lightweight, fast, and modern static asset hosting platform.</p>
  <p>
    <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" alt="Go 1.25+" />
    <img src="https://img.shields.io/badge/Gin-1.12-008ECF?logo=gin&logoColor=white" alt="Gin 1.12" />
    <img src="https://img.shields.io/badge/React-19-20232A?logo=react&logoColor=61DAFB" alt="React 19" />
    <img src="https://img.shields.io/badge/Vite-8-646CFF?logo=vite&logoColor=white" alt="Vite 8" />
    <img src="https://img.shields.io/badge/License-MIT-16A34A" alt="MIT License" />
    <img src="https://img.shields.io/badge/Storage-Local%20%7C%20S3%20%7C%20FTP-111827" alt="Storage Drivers" />
  </p>
  <p>
    <a href="README.zh-CN.md">简体中文</a> | English
  </p>
</div>

## 🪁 Introduction

Kite is a lightweight, fast, and modern static asset hosting platform. Much more than just an image host, Kite is designed to handle images, audio, video, and general static files. Just like a kite soaring in the sky, Kite aims to provide a seamless and effortless experience for managing and serving all your media assets.

Kite ships with a Go backend, an embedded React admin panel, and a storage abstraction layer that supports local disk, FTP, and mainstream object storage providers. It is suitable for self-hosted image hosting, static asset delivery, internal media libraries, and lightweight team deployments.

## ✨ Features

- **Comprehensive Support**: Host and serve images, audio, video, and standard static files with ease.
- **Lightweight & Fast**: Built with modern web technologies for optimal performance and rapid delivery.
- **Easy to Use**: A clean and intuitive user interface for uploading, managing, and viewing your files.
- **Bilingual UI**: English and Simplified Chinese ship out of the box, with a language switcher that flips the public site, admin console and backend response messages in lockstep.
- **First-run install wizard**: A guided `/setup` flow on a fresh deployment provisions the admin account, the public site URL and the default storage backend without ever exposing temporary credentials.
- **Secure**: Robust security measures to keep your data and assets safe.
- **Open Source**: Free to use and modify.

## 🧱 Tech Stack

- **Backend**: Go, Gin, GORM, SQLite/MySQL/PostgreSQL
- **Frontend**: React, TypeScript, Vite, TanStack Query, Radix UI
- **Storage**: Local, S3, FTP
- **Media**: Image thumbnailing, static file serving, public upload endpoints

## 🚀 Quick Start

### Development

```bash
make dev
```

This starts the Go backend and the admin frontend development server.

Enable the repository Git hooks once after cloning:

```bash
make hooks-install
```

The `pre-commit` hook focuses on quick fixes before a commit: it formats staged Go and admin frontend files and runs `go mod tidy`.

The `pre-push` hook runs the slower validation steps in an isolated checkout of `HEAD`: `go test`, Go build verification, and the admin frontend build check.

### Production Build

```bash
make build
./build/kite
```

The production build compiles the frontend and embeds it into the Go binary.

### First-run install wizard

On a fresh deployment (with no `is_installed` flag in the settings table) Kite refuses to bootstrap a default admin and instead routes the operator through the guided wizard at `/setup`. The wizard asks for:

- **Site identity** — site name and public URL (used for OAuth callbacks and absolute links).
- **Admin account** — username, email and password (minimum 6 characters). The account is created with the administrator role; no temporary credentials are ever written to the database or the logs.
- **Default storage backend** — pick from local disk, S3-compatible object storage, FTP and the supported cloud providers, then either keep the suggested defaults or fill in the connection parameters. The wizard saves the choice into the database and reloads the storage manager in-place.

If the database driver / DSN need to change as part of the install, the wizard can also persist a new `database` block to the on-disk config file (see [`KITE_CONFIG_FILE`](#environment-variables)). The next request restarts with the new connection.

After a successful submit the wizard sets `is_installed=true` and the route stops responding so the form can't be re-driven by a later visitor.

#### Skipping the wizard

For scripted or container-based deployments where running an interactive wizard isn't practical, set `KITE_LEGACY_BOOTSTRAP=1` to fall back to the legacy behavior: an `admin / admin` account is seeded on first boot and flagged `password_must_change`, so the first login still redirects to a mandatory reset page. The password is never written to the application logs — change it immediately and keep the new credentials out of version control.

## ⚙️ Configuration

Kite reads configuration from three layers, applied in order so each layer can override the previous one:

1. **Compiled-in defaults** — sensible values for a single-host development run.
2. **Config file** — an optional JSON file persisted by the install wizard. Its path defaults to `data/config.json` next to the database, and can be overridden with `KITE_CONFIG_FILE`. Missing or partial files degrade silently to "use whatever the previous layer had".
3. **Environment variables** — the highest-priority layer, primarily used for ephemeral overrides (port forwarding, swapping the DSN in a CI run, capping the log level on a noisy host).

A subset of runtime-tunable settings (site name, registration policy, JWT secret, etc.) also lives in the database `settings` table and is editable from the admin console — those override the static config file but are themselves overridden by environment variables.

### Environment variables

| Variable | Purpose | Example |
| --- | --- | --- |
| `KITE_PORT` | HTTP listener port. | `8080` |
| `KITE_HOST` | Bind address. Use `0.0.0.0` to accept connections from outside the host. | `0.0.0.0` |
| `KITE_DSN` | Database connection string. The format depends on `KITE_DB_DRIVER`. | `data/kite.db` |
| `KITE_DB_DRIVER` | Database driver. One of `sqlite`, `mysql`, `postgres`. | `sqlite` |
| `KITE_SITE_URL` | Public site URL — used to generate OAuth callbacks and absolute links. | `https://kite.example.com` |
| `KITE_LOG_LEVEL` | Minimum log level. One of `debug`, `info`, `warn`, `error`. Default `info`. | `info` |
| `KITE_LOG_FORMAT` | Log format. One of `text`, `json`. Default `text` for development; pick `json` for log shippers. | `json` |
| `KITE_CORS_ORIGINS` | Comma-separated origin allowlist for the static asset endpoints (`/i/`, `/v/`, `/a/`, `/f/`). When unset only same-origin requests are allowed. | `https://kite.example.com,https://app.example.com` |
| `KITE_LEGACY_BOOTSTRAP` | Set to `1` to skip the install wizard and seed the `admin / admin` account on first boot. Useful for scripted deployments. | `1` |
| `KITE_CONFIG_FILE` | Path to the on-disk JSON config file managed by the install wizard. Defaults to `data/config.json`. | `/etc/kite/config.json` |

Environment variables override both the compiled-in defaults and the contents of the config file, so the file → env-var precedence matches the rest of the stack and the `make dev` helpers.

### Internationalization

The backend message catalogue, the public landing pages and the admin SPA all share a single locale resolution chain:

1. The `kite_locale` cookie (written by the language switcher in the public header and by the SPA's `setLocale`).
2. The `Accept-Language` request header.
3. The static fallback — English.

Changing the locale on any one surface (public site, admin console, server-rendered upload page) updates the cookie, so the next render of every other surface picks up the same language. The catalogue currently ships English and Simplified Chinese — adding a third locale is a code change in `internal/i18n` plus a sibling locale file in `web/admin/src/i18n/locales`.

## 📦 Project Structure

```text
cmd/kite            application entrypoint
internal/           backend handlers, services, repos, storage drivers
web/admin/          React admin panel
template/           landing and upload page templates
deploy/             docker-compose and nginx examples
```

## 🐳 Deployment

- Use [deploy/docker-compose.yml](deploy/docker-compose.yml) for a containerized setup.
- Use [Dockerfile](Dockerfile) to build a standalone image.
- Use [deploy/nginx/conf.d/www.kite.plus.conf](deploy/nginx/conf.d/www.kite.plus.conf) as a reverse-proxy reference.

## 🤝 Contributing

We welcome contributions! 

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

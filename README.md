<div align="center">
  <img src="https://user-images.githubusercontent.com/8456633/59972109-8e9c8480-95cc-11e9-8350-38f7f86ba76d.png" width="120" />
  <h1>dcli</h1>
  <p><strong>The Lazier Way to Manage Everything Docker</strong></p>
  <p>
    <a href="https://golang.org/"><img src="https://img.shields.io/badge/Language-Go-blue.svg" alt="Language" /></a>
    <a href="https://github.com/yaogh99123/dcli/releases"><img src="https://img.shields.io/github/v/release/yaogh99123/dcli" alt="Release" /></a>
    <a href="https://goreportcard.com/report/github.com/yaogh99123/dcli"><img src="https://goreportcard.com/badge/github.com/yaogh99123/dcli" alt="Go Report Card" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/yaogh99123/dcli" alt="License" /></a>
  </p>
  <p>An efficient, intuitive, and "lazy" CLI manager for Docker and Docker-Compose.</p>
</div>

---

## 📖 Introduction

**dcli** is a high-efficiency Docker management tool built for developers. Evolved from the acclaimed `lazydocker`, it moves away from a heavy TUI (Terminal UI) towards a **pure CLI interactive mode** that better aligns with terminal philosophy.

It automatically detects your project environment, merges fragmented Compose files, and frees you from typing long `docker-compose -f ... -f ... up` commands.

---

## ✨ Key Features

- 📂 **Smart Context Detection**
  Automatically identifies `.git` or project roots, deep scans, and merges all `docker-compose*.yml` configuration files.
- 🔍 **Interactive Fuzzy Search**
  Built-in `fzf`-like search experience. Simply press `s` to instantly locate containers, images, or volumes.
- 🛠️ **Unified Management Workflow**
  Seamlessly manages both local Compose projects and standalone Docker containers. Automatically switches to Fallback mode for external containers to ensure core management is never interrupted.
- ⚡ **One-Key Actions**
  Supports numeric shortcuts (1-17) and shorthand commands. Start, restart, or view logs in an instant.
- 🧹 **System Deep Cleanup**
  Built-in Prune commands to quickly reclaim disk space from unused networks, volumes, images, and build caches.
- 📦 **Predefined Stack Templates**
  One-click deployment for common development environments like ELK log monitoring stacks, MySQL/Redis database stacks, and more.

---

## 🚀 Quick Start

### Installation

**Via Go:**
```bash
go install github.com/yaogh99123/dcli@latest
```

**Or download binary from the release page:**
[GitHub Releases](https://github.com/yaogh99123/dcli/releases)

---

## ⌨️ Usage

Simply run `dcli` in your terminal to enter interactive mode:

```bash
dcli
```

### Command Flags

| Flag              | Description                                      |
| :---------------- | :----------------------------------------------- |
| `-a`, `--arun`    | Show all services (including non-running ones)   |
| `-n`, `--nrun`    | Show only non-running services                   |
| `-f`, `--file`    | Manually specify custom Compose files            |
| `-p`, `--project` | Specify a custom project name                    |
| `-c`, `--config`  | Print the current default configuration          |

---

## 🎮 Interactive Guide

Once inside `dcli`, you can perform operations using the following shortcuts:

### Core Operations
- `1 - 3`: Up / Stop / Restart services
- `4`: View logs (type `exit` to return)
- `7`: Enter container (auto-detects `bash` or `sh`)
- `8`: Build and update services
- `s`: Enable fuzzy search mode
- `menu`: Show full functionality menu
- `0`: Exit program

### Quick Filtering
- `a`: Switch to "Show All"
- `r`: Switch to "Running Only"
- `l`: Switch to list mode

---

## 🛠️ Customization

Configuration files are typically located at:

- macOS: `~/Library/Application Support/yaogh99123/dcli/config.yml`
- Linux: `~/.config/yaogh99123/dcli/config.yml`

You can view your current default configuration and make adjustments by running `dcli -c`.

---

## 📜 Credits

The birth of this project was inspired by [lazydocker](https://github.com/yaogh99123/lazydocker). We have refactored and optimized its powerful Docker management logic into a CLI-focused tool for heavy terminal users.

---

<div align="center">
  <p>Built with ❤️ by the open-source community.</p>
</div>

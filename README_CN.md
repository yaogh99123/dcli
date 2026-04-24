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
  <p>高效、直观且“懒惰”的 Docker / Docker-Compose 命令行管理专家。</p>
</div>

---

## 📖 简介 / Introduction

**dcli** 是一款专为开发者打造的高效率 Docker 管理工具。它从备受好评的 `lazydocker` 演进而来，摒弃了沉重的 TUI 界面，转向更符合终端哲学的 **纯 CLI 交互模式**。

它能自动感知你的项目环境，合并碎片化的 Compose 文件，让你告别冗长的 `docker-compose -f ... -f ... up` 指令。

---

## ✨ 核心特性 / Key Features

- 📂 **智能项目探测 (Smart Context)**
  自动识别 `.git` 或项目根目录，深度扫描并自动合并 `docker-compose*.yml` 配置文件。
- 🔍 **交互式模糊搜索 (Fuzzy Search)**
  内置类似 `fzf` 的搜索体验，输入 `s` 即可在海量容器、镜像、卷中秒速定位。
- 🛠️ **统一管理体验 (Unified Workflow)**
  无缝衔接本地 Compose 项目与独立 Docker 容器。对外部容器自动切换至 Fallback 模式，确保基础管理不断档。
- ⚡ **一键极速操作 (One-Key Actions)**
  支持数字快捷键（1-17）和简写指令，启动、重启、查看日志均在瞬息之间。
- 🧹 **系统深度清理 (System Cleanup)**
  内置 Prune 指令，快速回收网络、卷、镜像及构建缓存所占用的磁盘空间。
- 📦 **预设服务栈 (Stack Templates)**
  支持一键部署 ELK 日志监控栈、MySQL/Redis 数据库栈等常用开发环境。

---

## 🚀 快速开始 / Quick Start

### 安装 / Installation

**使用 Go 安装：**

```bash
go install github.com/yaogh99123/dcli@latest
```

**或者通过 Release 页面下载二进制文件：**
[GitHub Releases](https://github.com/yaogh99123/dcli/releases)

---

## ⌨️ 常用指令 / Usage

直接运行 `dcli` 进入交互模式：

```bash
dcli
```

### 启动参数 / Command Flags

| 参数              | 说明                         |
| :---------------- | :--------------------------- |
| `-a`, `--arun`    | 显示所有服务（包括未运行的） |
| `-n`, `--nrun`    | 仅显示当前未运行的服务       |
| `-f`, `--file`    | 手动指定特定的 Compose 文件  |
| `-p`, `--project` | 指定特定的项目名称           |
| `-c`, `--config`  | 打印当前默认配置             |

---

## 🎮 交互指南 / Interactive Guide

进入 `dcli` 后，你可以通过以下快捷键进行操作：

### 核心操作

- `1 - 3`: 启动 / 停止 / 重启服务
- `4`: 查看日志 (输入 `exit` 返回)
- `7`: 进入容器 (自动探测 `bash` 或 `sh`)
- `8`: 编译并更新服务
- `s`: 开启模糊搜索模式
- `menu`: 显示完整功能菜单
- `0`: 退出程序

### 快捷筛选

- `a`: 切换至“显示全部”
- `r`: 切换至“仅显示运行中”
- `l`: 切换至列表模式

---

## 🛠️ 配置自定义 / Customization

配置文件通常位于：

- macOS: `~/Library/Application Support/yaogh99123/dcli/config.yml`
- Linux: `~/.config/yaogh99123/dcli/config.yml`

你可以通过 `dcli -c` 查看当前的默认配置并进行按需修改。

---

## 📜 致谢 / Credits

本项目的诞生离不开 [lazydocker](https://github.com/jesseduffield/lazydocker) 的启发。我们在其强大的 Docker 管理逻辑基础上，针对重度终端用户进行了 CLI 化的重构与优化。

---

<div align="center">
  <p>Built with ❤️ by the open-source community.</p>
</div>

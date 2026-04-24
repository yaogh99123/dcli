# Dcli 重构进度记录 (Progress.md)

## 当前重点
- 已完成从 TUI 到纯 CLI 的核心重写。
- 正在优化项目文档，建立全新的品牌视觉。
- [x] **所有权变更**: 已将 `jesseduffield/dcli` 统一重命名为 `yaogh99123/dcli`。
- [x] **全量文件探测**: 修改 `main.go` 探测逻辑，支持自动加载并合并所有 `docker-compose*.yml` 文件。
- [x] **外部服务兼容 (Docker Fallback)**:
    - 针对非本地服务（`IsLocal=false`），自动回退到原生 `docker` 命令进行管理。
    - 修复了菜单项 5, 6, 8, 100 在处理外部服务时的报错逻辑。
    - 实现了对外部服务的智能引导：仅允许查看状态和进入容器，屏蔽需要本地配置文件的“查看配置”和“编译”操作。

## 核心决策
1. **彻底移除 GUI**: 删除 `pkg/gui` 目录及相关库依赖。
2. **命令对象解耦**: 弃用 `getComposeCommandWithFiles` 的直接调用，统一通过 `NewCommandObject` 动态生成。
3. **交互式引导**: 对非本地服务提供友好的提示，告知为何某些操作（如查看配置）不可用。

## 已完成事项
- [x] 重新编写 README.md (英文版) 和 README_cn.md (中文版)。
- [x] 分析 `devos` 模式与 GUI 的耦合度。
- [x] 修改 `pkg/app/app.go` 解耦 GUI 依赖。
- [x] 修改 `main.go` 简化入口点并增强文件探测。
- [x] 实现服务过滤功能：支持 `-a` (全显) 和 `-n` (仅显未运行)。
- [x] **项目重命名**: 全量替换 `lazydocker` 关键字。
- [x] **项目隔离修复**: 通过 `IsLocal` 标记彻底解决 `no such service` 报错。
- [x] **全能管理模式**: 完美支持本地项目的高级 Compose 操作和外部容器的基础管理。

## 待办事项
- [ ] 优化 CLI 模式下的错误处理提示。
- [ ] 清理 `config.yml` 中已废弃的 GUI 相关字段。

# M0 阶段（Engine SDK + CLI 兼容）User Stories

> M0 目标是把现有 CLI（`cs` TUI）核心逻辑抽离为可复用的 **Engine SDK**，同时保证终端体验 100 % 兼容。主要利益相关者是 *CLI 用户* 与 *二次开发者*。

---

## 1. CLI 用户视角

| # | User Story（作为…我想…以便…） | Done Criteria |
|---|--------------------------------|---------------|
| **C‑1** | **作为现有 `claude-squad` CLI 用户**，我想升级到新版本后 **命令参数与键位保持原样**，以便无需学习新用法。 | 1. `cs` 子命令、flags 与 < v1.0.8 完全一致；<br>2. Bubbletea TUI 显示和交互行为相同。 |
| **C‑2** | 作为 CLI 用户，我希望可以继续 **并行创建 / 暂停 / 恢复 / 终止** 会话，确保当前工作流不受影响。 | `n / N` 新建、`c` Checkout(Pause)、`r` Resume、`x` Kill 功能均可用；状态在 UI 实时刷新。 |
| **C‑3** | 作为 CLI 用户，我希望在升级后仍能 **恢复历史 Session**（state.json），避免数据丢失。 | 启动新版本时自动读取旧 `state.json`；会话列表正确加载。 |

---

## 2. 二次开发者（SDK 消费者）视角

| # | User Story | Done Criteria |
|---|------------|---------------|
| **D‑1** | **作为一名 Go 开发者**，我想通过 `go get` 引入 `github.com/smtg-ai/claude-squad/pkg/engine`，以便在我的脚本或服务里复用 Session 管理能力。 | `go get` 成功；能够 `engine.New()` → `Start` 会话 → `List`；事件流通过 channel 获得。 |
| **D‑2** | 作为开发者，我想在单元测试里使用 **Fake Git / Fake tmux** 注入，避免真实依赖。 | Engine 构造函数支持传入 `GitBackend` / `TmuxBackend` 接口；官方示例和测试中提供假实现。 |
| **D‑3** | 作为开发者，我希望 `Engine` 的公有 API 文档清晰，可在 pkg.go.dev 浏览。 | Godoc 注释齐全；`ENGINE_SDK.md` 提供快速示例；覆盖率 > 80 %。 |

---

## 3. DevOps / 维护者视角

| # | User Story | Done Criteria |
|---|------------|---------------|
| **M‑1** | **作为维护者**，我想在 CI 中运行单元 + 竞态 + 集成测试，确保 SDK 重构不破坏 CLI 兼容。 | GitHub Actions Workflow：`go test ./...`、`go test -race`, `integration` tag；所有测试通过即绿灯。 |
| **M‑2** | 作为维护者，我想在发布后通过 `CHANGELOG.md` 了解所有破坏性或新增改动。 | Release 说明列出：新 Engine SDK、无 CLI 行为变化、迁移步骤。 |

---

> 🔑 **备注**：M0 主要面向“后台／开发”场景，用户不是终端产品用户而是 CLI 和 SDK 使用者。编写这些轻量级 User Story 有利于确保功能重构以 *用户视角* 出发，而非仅技术驱动。

# 一、改造目标与范围

| 目标 | 说明 |
|------|------|
| 面向 Web 的交互体验 | 让非终端用户通过浏览器即可创建 / 暂停 / 恢复 / 回收 AI Session，并实时查看终端输出与 git diff。 |
| 保持现有核心能力 | 继续依赖 tmux + git worktree 隔离机制；CLI 仍可单机使用。 |
| 支持多用户 | 初期最简「账号-Session」映射，后续可接 GitHub OAuth & RBAC。 |
| 安全与可扩展 | 为长时间运行的 Agent 引入进程、资源、文件沙箱，并为 SaaS 化预留横向扩容空间。 |

---

## 二、总体架构

```
┌─────────────┐        REST/WS          ┌────────────────────┐
│  Browser    │◀──────────────────────▶│  Web API (Go)      │
│  React SPA  │                        │  Gin + Gorilla WS  │
└─────────────┘        EventBus         │  ┌──────────────┐  │
        ▲                               │  │ Engine SDK   │  │
        │                               │  │ (core logic) │  │
        │ WebSocket stdout/diff         │  └──────────────┘  │
        │                               │        │           │
        │                               │   tmux / git       │
        ▼                               └────────┬──────────┘
┌──────────────────┐         PTY               ┌─────────────┐
│  tmux Session N  │◀─────────────────────────▶│ Git worktree│
└──────────────────┘                            └─────────────┘
```

- **Engine SDK** = 把 app, session, git, tmux 等包抽离为纯 Go 库，供 CLI & Web 共用。
- **Web API** = 轻量服务层，负责鉴权、HTTP 路由、WebSocket 广播与持久化。
- **React SPA** = 前端三栏「任务列表 – 终端/Diff – 元数据」工作台。
- **EventBus** = 内存 channel + fan-out goroutine，把 Engine 产生的 stdout/stderr/diff 事件统一推给订阅端。

---

## 三、代码层面重构

| 步骤 | 关键改动 | 备注 |
|------|----------|------|
| 1. 提炼 Engine 包 | 新建 pkg/engine；封装 Start/Stop/Pause/Resume/GetDiff/StreamLogs 等 API。 | CLI 调用路径保持不变，只是内部指向新包。 |
| 2. Session 实体升级 | 在 Instance 结构体中新增 OwnerID string、EventCh chan Event 字段，并把 Started bool 等私有状态保持不变。 | 源码里已有字段定义，直接扩展即可。 |
| 3. 事件流 | 统一 type Event struct { Kind string; Payload any; Ts time.Time }；stdout/stderr/diff 均走该通道。 | 方便 WebSocket 与 CLI TUI 复用。 |
| 4. 状态持久化 | 保留 state.json 方案（见 config 包），引入文件锁；后续可换 SQLite。 | |
| 5. Daemon 调度器 | 把 daemon 包注册为 Engine 的一个可选 goroutine，支持 Auto-Yes 与心跳。 | |

---

## 四、服务层（Web API）

### 4.1 技术栈选型
- **Gin** （或 Chi）— 路由 & 中间件
- **Gorilla/WebSocket** — 双向实时通信
- **JWT + Cookie** — 最简鉴权；可对接 GitHub OAuth
- **Zap / slog** — 结构化日志
- **Viper** — 统一配置管理

### 4.2 REST & WS 端点

| Method / Path | 功能 |
|---------------|------|
| POST /api/sessions | 创建 Session（body: title, program, autoYes） |
| GET  /api/sessions | 列表 / 分页 |
| GET  /api/sessions/{id} | 查询详情 |
| PATCH /api/sessions/{id} | {action: pause \| resume \| kill} |
| WS   /ws/sessions/{id} | 订阅事件流（stdout/stderr/diff） |
| WS   /ws/dashboard | 后台总览：推送全局状态 & 统计 |

### 4.3 中间件
1. **Auth**：校验 JWT，注入 UserID。
2. **Rate-Limit**：IP + User 级 QPS 控制。
3. **CORS**：允许前端域名。
4. **Recovery**：panic 捕获并写日志。

---

## 五、前端（React + Vite）

| 页面 | 组件 | 说明 |
|------|------|------|
| Dashboard | `<SessionList>` | 左侧列表，展示 Title / 状态 Tag / 分支名 |
| | `<TerminalPane>` | 使用 xterm.js 渲染 stdout/stderr；支持 ctrl+click 跳转文件 |
| | `<DiffPane>` | diff2html 或自研彩色行号视图 |
| | `<MetaSidebar>` | Branch、计时器、按钮（Resume/Pause/Commit/Push） |
| Auth | `<LoginWithGitHub>` | OAuth 回调写回 JWT Cookie |

- **状态管理**：Zustand（简洁）或 Redux Toolkit。
- **通信**：自封装 useWebSocket(url)，自动重连、心跳。
- **样式**：Tailwind；暗色 / 亮色主题切换。

---

## 六、安全与隔离

| 维度 | 实施 |
|------|------|
| 进程 | 每 Session 启动 tmux new-session -s {UserID}-{SessionID}；可选 cgroup 限制 CPU/Memory。 |
| 文件 | git worktree 已隔离；如需进一步沙箱，整套 Engine + repo 运行在独立 Docker 容器。 |
| 网络 | 非白名单域名拒绝外连；必要时启用 eBPF 网络策略。 |
| 加密 | 所有 WS / HTTP 强制 TLS；配置文件仅存加密后的 token。 |

---

## 七、部署与运维

### 1. 单机 MVP
- systemd 启动 Web API；Nginx 反向代理；SQLite 或 state.json 本地持久化。

### 2. SaaS 阶段
- API + Engine 运行在 Kubernetes；Per-user repo 与 tmux 容器使用 PVC；事件经 NATS Streaming。

### 3. CI/CD
- GitHub Actions：go test + go vet + docker build & push。
- 前端 pnpm build 产物上传到 S3 + CloudFront。

---

## 八、渐进式里程碑

| 里程碑 | 交付物 | 估时（周） | 设计文档 | 用户故事文档 |
|--------|--------|------------|----------|-------------|
| M0 | Engine SDK + CLI 兼容 | 1 | webui-detailed-design-M0.md | webui-user-story-M0.md |
| M1 | 单用户 Web GUI（本地） | 2 | webui-detailed-design-M1.md | webui-user-story-M1.md |
| M2 | 多用户登录 & JWT 鉴权 | 1 | webui-detailed-design-M2.md | webui-user-story-M2.md |

---

## 九、风险与对策

| 风险 | 对策 |
|------|------|
| tmux / git CLI 行为变更 | 为 tmux、git 调用封装适配层，并加单元测试。 |
| 长连接雪崩 | WS 使用心跳 + back-off；NATS 缓冲高峰。 |
| Agent 无限循环写盘 | 设置 worktree quota；启用文件改动阈值警报。 |

---

## 结语

该设计最大化复用了 claude-squad 现有「tmux + git worktree」能力，同时通过 Engine–Service–SPA 分层实现了从终端工具到多人 Web 协作平台的平滑演进，为未来接入 Verdent 的 Plan-Code-Verify 日志、代码安全扫描或 CI/CD 流程留足了扩展空间。
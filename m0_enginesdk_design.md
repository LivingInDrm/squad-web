# M0 阶段详细设计 - Engine SDK + CLI 兼容

以下内容聚焦 M0 阶段 –「Engine SDK ＋ CLI 兼容」 的详细技术设计。目标是在不破坏现有终端体验的前提下，把核心业务抽象为一套可复用的 Go SDK，为后续 Web/API 层奠定基础。

---

## 1. 设计目标

| 目标 | 约束 | 结果 |
|------|------|------|
| 抽离核心逻辑 | 不引入 Web 依赖；保持包级 API 简洁 | 新增 pkg/engine 作为唯一外部入口 |
| CLI 100% 兼容 | cs 的所有命令行为、键位、状态保持不变 | TUI 仅调用 Engine，无直接 session.* 操作 |
| 无破坏式重构 | 原包 (session/, git/, tmux/ 等) 保留路径 | 风险最低，可渐进迁移 |
| 可观测、可测试 | 单元测试覆盖核心 API；日志沿用 log 包 | 为后续 API / WebSocket 做好事件流准备 |

---

## 2. 新目录与模块边界

```
.
├── app/             # 仅保留 Bubbletea TUI，改为依赖 engine
├── cmd/             # Cobra commands
├── pkg/
│   └── engine/
│       ├── engine.go     # Facade API
│       ├── manager.go    # Session registry & lifecycle
│       ├── event.go      # 统一事件定义
│       └── storage.go    # 包装原 session.Storage
└── session/         # 原有实现，包内可见性适配
```

### 2.1 Engine Facade（engine.go）

```go
type Engine struct {
    mu       sync.RWMutex
    mgr      *manager           // 见 2.2
    store    *storage           // 见 2.4
    cfg      *config.Config
    logger   *log.Logger
}

func New(cfg *config.Config, appState config.AppState) (*Engine, error)
func (e *Engine) Start(ctx context.Context, opts SessionOpts) (string, error)
func (e *Engine) Pause(id string) error
func (e *Engine) Resume(id string) error
func (e *Engine) Kill(id string) error
func (e *Engine) List() []SessionInfo
func (e *Engine) Events(id string) (<-chan Event, error)
```

返回 SessionID（uuid.NewString()）避免与旧的 Title/Branch 冲突。

### 2.2 Session Manager（manager.go）
- **职责**：持有 map[string]*session.Instance；封装并发安全访问。
- **互斥**：sync.RWMutex; 每个 Instance 仍负责自身内部状态机（Running/Paused…）
- **事件派发**：实例通过 chan Event 上报 stdout/stderr/diff；Manager 负责 fan-out 到 Engine 的订阅者。

### 2.3 Event 统一结构（event.go）

```go
type Kind string
const (
    EventStdout Kind = "stdout"
    EventStderr Kind = "stderr"
    EventDiff   Kind = "diff"
    EventState  Kind = "state"   // Running → Ready 等
)

type Event struct {
    ID        string    // SessionID
    Kind      Kind
    Payload   any
    Timestamp time.Time
}
```

### 2.4 Storage 适配层（storage.go）
- 包装原 session.Storage，隐藏磁盘格式差异（state.json）
- 新增 LoadAll(), SaveAll()，在 Engine 初始化/关闭时自动调用。

---

## 3. CLI 改动（向下兼容）

| 文件 | 关键改动 |
|------|----------|
| main.go | app.Run() → engine.New()；将 programFlag / autoYesFlag 传入 Engine；保持 Cobra 命令不变 |
| app/home.go | 去除对 session.NewStorage, session.Instance 的直接调用，改用 Engine API；事件流与 UI 的绑定保持原样，逻辑集中于 instanceChanged() 等方法 |

因 Bubbletea 模型需要频繁查询实例状态，可在 home 内维护一个只读缓存，通过 Engine 订阅事件实时刷新。

---

## 4. Engine 内部实现细节

### 4.1 SessionOpts 与现有结构映射

```go
type SessionOpts struct {
    Title   string
    Path    string
    Program string
    AutoYes bool
    Prompt  string
}
```

- 映射到原 session.InstanceOptions；Title 仍用于 tmux session 名称
- 启动流程保持：git.NewGitWorktree → tmux.NewTmuxSession → instance.Start(true)。

### 4.2 事件采集

```go
func (m *manager) watch(instance *session.Instance, out chan<- Event) {
    go func() {
        for {
            updated, prompt := instance.HasUpdated()
            if updated {
                out <- Event{ID: id, Kind: EventStdout, Payload: instance.Preview(), Timestamp: time.Now()}
            }
            if err := instance.UpdateDiffStats(); err == nil {
                out <- Event{ID: id, Kind: EventDiff, Payload: instance.GetDiffStats(), Timestamp: time.Now()}
            }
            if prompt && instance.AutoYes {
                instance.TapEnter()
            }
            time.Sleep(500 * time.Millisecond)
        }
    }()
}
```

- 周期同现有 tickUpdateMetadataCmd 逻辑，避免 UI 回归时重写。

### 4.3 并发模型

| 资源 | 粒度 | 锁策略 |
|------|------|--------|
| Engine.mgr.sessions | map[string]*Instance | 全局 RWMutex |
| 单个 Instance | tmuxSession, gitWorktree | 已封装于 session.* |

全局锁只持有极短时间（增删实例），高频读取事件走 channel，避免阻塞。

### 4.4 错误与日志
- 沿用 log.{Info,Warning,Error}Log；Engine 对外返回语义化错误。
- CLI/TUI 收到错误后，保持现有 errBox 展示逻辑。

---

## 5. 测试策略

| 层级 | 覆盖点 | 方法 |
|------|--------|------|
| 单元 | manager CRUD / 锁 | go test ./pkg/engine -run TestManager* |
| 集成 | Engine.Start → Pause → Resume → Kill | 临时 clone public git repo，mock tmux with pty stub |
| 回归 | CLI (cs) | expect 脚本自动按键；确保 UI 行为不变 |

---

## 6. 迁移步骤（实施顺序）
1. 复制 session/, git/, tmux/, config/, log/ 到 pkg/engine/internal（保持包路径不变）。
2. 实现 pkg/engine Facade；在 Facade 内引用内部包。
3. 改造 CLI：main.go & Bubbletea home 只使用 Facade。
4. 跑测试；确保 cs TUI 功能完全一致。
5. go mod tidy；Tag v1.1.0‐sdk.
6. 文档：更新 README 与 CHANGELOG，标注 SDK 用法示例。

---

## 7. 风险评估与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| 循环依赖 (app → engine → session ↔ ui) | 构建失败 | 使用 internal/ 限制可见性，分层引用 |
| 锁粒度不当 | UI 卡顿 / 数据竞态 | 单实例无锁，全局读多写少；Benchmark 调整 |
| 事件风暴 | 高 CPU / 内存 | 帧率限制（≥ 200 ms）；Diff 内容仅在改变时发送 |
| 隐藏 bug 回归 | CLI 行为异常 | 自动化 expect 脚本 + GitHub Actions CI |

---

## 8. 交付物

| 文件/目录 | 描述 |
|-----------|------|
| pkg/engine/*.go | Engine SDK 源码 |
| examples/sdk_demo.go | 30 行示例：创建 Session、打印 diff、关闭 |
| docs/ENGINE_SDK.md | API 文档（godoc 链接 + 用例） |
| CHANGELOG.md | v1.1.0 新增：Engine SDK |

---

## 结语

实现 M0 后，你将拥有 **稳定 CLI + 清晰 Engine SDK** 的双栈结构：
- **现有用户** 无感升级，仍用 cs TUI。
- **内部开发** 可直接 go get github.com/smtg-ai/claude-squad/pkg/engine，在单元测试、脚本或未来 Web 服务中调用一致的核心能力，真正做到"一处实现，多端复用"。
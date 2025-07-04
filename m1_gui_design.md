# M1 阶段：本地单用户 Web GUI 技术设计

> 在 **M0（Engine SDK + CLI 兼容）** 的基础上，为 `claude-squad` 引入“仅限本机、单用户”版本的 Web 界面。CLI 完全保留，浏览器用户无需安装 tmux。

---

## 1 总体架构

```text
Browser (React SPA) ────HTTP/WS──▶ squadd (Go 服务 + Engine SDK) ──tmux/git──▶ Worktree
```

- **squadd**：新增守护进程，暴露 REST + WebSocket，内部复用 M0 Engine。
- **React SPA**：单页前端，默认访问 `http://localhost:7999`。
- **单用户**：无登录体系，服务仅监听 `127.0.0.1`。

---

## 2 后端（squadd）

### 2.1 目录结构

```
cmd/
  squadd/           # 入口
internal/
  api/http/         # Gin 路由
  api/ws/           # WebSocket Hub
```

### 2.2 端口

| 协议 | 端口 | 说明 |
|------|------|------|
| HTTP | 7999 | REST、WS 复用同端口 |

### 2.3 REST 接口

| 方法 & 路径           | 主要字段                     | 功能 |
|-----------------------|------------------------------|------|
| `POST /api/session`   | `title, program, autoYes`    | 新建会话 |
| `GET  /api/sessions`  | —                            | 列出会话 |
| `GET  /api/session/:id` | —                          | 会话详情 |
| `PATCH /api/session/:id` | `action: pause\|resume\|kill` | 控制会话 |
| `POST /api/session/:id/commit` | `message`           | Commit 并 Push |
| `GET  /api/version`   | —                            | 版本信息 |

错误统一返回 `{"error":"…"}`。

### 2.4 WebSocket

```
GET /ws/session/:id
```

返回 `engine.Event` 的 JSON 编码：

```json
{ "kind":"stdout", "payload":"Compiling…", "ts":"2025-07-04T12:00:00Z" }
{ "kind":"diff",   "payload":{"insert":23,"delete":4}, "ts":"…" }
{ "kind":"state",  "payload":"Paused", "ts":"…" }
```

### 2.5 启动流程

1. 读取 `~/.claude-squad/config.json`。  
2. `engine.New()` 创建单例。  
3. 建立 WS Hub，`hub.Publish(id, event)` 广播。  
4. Gin 注册 REST/WS，监听 `127.0.0.1:7999`。  

---

## 3 前端（React + Vite）

### 3.1 依赖

| 库 | 用途 |
|----|------|
| `react`, `vite` | SPA 框架 |
| `xterm.js` | 终端渲染 |
| `diff2html` | 彩色 Diff |
| `zustand` | 全局状态 |
| `tailwindcss` | 样式 |

### 3.2 布局

```
┌ 会话列表 ┬ 终端 / Diff (Tab) ┬ 元数据/操作卡片 ┐
```

组件：

- `<SessionList>`：title + 状态 Tag + 快捷按钮。  
- `<TerminalPane>`：xterm 绑定 WS。  
- `<DiffPane>`：彩色行号视图。  
- `<MetaCard>`：分支、计时器、Commit 输入框与 Push。  

### 3.3 通信

```ts
fetch('/api/sessions')
connect(`ws://localhost:7999/ws/session/${id}`)
```

WS 自动重连（指数退避）。

### 3.4 打包与嵌入

```bash
pnpm build         # 生成 dist/
go build -o squadd # 静态资源可用 embed.FS 嵌入
```

---

## 4 部署流程

1. 机器（或容器）内预装 **tmux** 与 **git**。  
2. 运行 `./squadd`。  
3. 浏览器打开 `http://localhost:7999`。  
4. CLI 用户继续使用 `./cs`，体验不变。

---

## 5 测试方案

| 层级 | 覆盖 |
|------|------|
| **后端单元** | REST 返回码、WS Hub 广播 |
| **后端集成** | `POST → GET → WS` 全流程 |
| **前端单元** | Zustand store、组件渲染 |
| **E2E** | Playwright：创建→输出→暂停→杀死 |

---

## 6 风险与对策

| 风险 | 对策 |
|------|------|
| tmux/git 路径差异 | `--tmux-bin`、`--git-bin` flag |
| 多标签冲突 | WS Hub 广播到每个连接 |
| 大 Diff 性能 | 后端行数上限 + 前端虚拟滚动 |

---

## 7 交付物

| 路径 | 内容 |
|------|------|
| `cmd/squadd/` | 后端主程序 |
| `web/`        | React 源码 |
| `Dockerfile`  | 完整镜像 |
| `docs/m1_gui.md` | 使用与 API 文档 |

---

本阶段实现后，`claude-squad` 将在本机提供现代浏览器 GUI，终端恐惧症用户无需接触 tmux，即可管理 AI 会话；同时为未来多用户、远程部署奠定事件流与组件基础。

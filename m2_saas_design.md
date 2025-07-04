# M2 阶段：SaaS 雏形（多用户 + 容器隔离 + 作业队列）技术设计

> 在 **M1（本地单用户 Web GUI）** 基础上，将 `claude-squad` 升级为支持多用户、云端部署、安全隔离的 SaaS。CLI 本地模式保持兼容。

---

## 1 目标与范围

| 目标 | 说明 |
|------|------|
| **多用户** | GitHub OAuth 登录；API 需 JWT |
| **资源隔离** | 每 Session 独立 Docker 容器，限制 CPU/内存/磁盘 |
| **可扩展** | Redis 队列调度，Worker 横向扩容 |
| **持久化** | 元数据存 PostgreSQL；仓库数据挂卷 |
| **向下兼容** | CLI & M1 squadd 本地模式继续可用 |

---

## 2 整体架构

```text
Browser SPA──HTTP/WS──▶ API Gateway ──Redis─┐
                        (Auth/JWT)          │
                                            ▼
                                Worker (Go, Docker SDK)
                                │        │
                                ▼        ▼
                     Session-Container A   Session-Container B
                     (tmux + Engine)       (tmux + Engine)
```

---

## 3 组件设计

### 3.1 API Gateway

| 功能 | 技术 |
|------|------|
| OAuth 登录 | GitHub OAuth2 |
| 用户 & Session REST | Gin |
| GraphQL(可选) | gqlgen |
| WS 转发 | Gorilla + Redis Pub/Sub |

**数据库表**

```sql
users(id, github_id, login, avatar_url, created_at)
repos(id, user_id, name, ssh_url, created_at)
sessions(id, user_id, repo_id, branch, status, container_id, created_at)
```

### 3.2 队列

* **Redis Stream** `session_jobs`

### 3.3 Worker

1. Redis 取 Job  
2. Docker SDK 创建 `squad-session` 容器  
3. Attach 日志 → Redis 发布  
4. 更新 `sessions.status`

容器资源：Memory 2 GiB、CPU 1 vCore。

### 3.4 Session 容器

`engine-svc` 启动后 clone 仓库并调用 Engine SDK，stdout 输出 `engine.Event`(JSONL)。

### 3.5 事件流

```
Container stdout → Worker → Redis Pub/Sub → Gateway → Browser WS
```

---

## 4 前端改造

| 变更 | 说明 |
|------|------|
| 登录页 | GitHub OAuth → JWT 存 `localStorage` |
| Dashboard | 多用户会话视图 |
| Repo 选择 | 新 API `/v1/repos` |
| WS 连接 | `/ws?session=<id>&token=<jwt>` |

---

## 5 安全与隔离

| 维度 | 措施 |
|------|------|
| 容器 | seccomp、cap-drop all |
| 资源 | CPU/Memory 限额 |
| 文件 | 独立 `docker volume`，定期 GC |
| 网络 | `--network none` or 内网代理 |
| 凭证 | Deploy Key 存 Vault |

---

## 6 部署

### 6.1 Docker Compose (PoC)

```yaml
services:
  api:
    build: cmd/squad-api
  worker:
    build: cmd/squad-worker
  redis:
    image: redis:7
  pg:
    image: postgres:16
  docker:
    image: docker:dind
    privileged: true
```

### 6.2 Kubernetes

- `api` Deployment + HPA  
- `worker` Deployment  
- Redis & Postgres StatefulSet  
- Session 容器由 Worker 动态创建

---

## 7 代码改动

| 位置 | 说明 |
|------|------|
| `pkg/engine` | 无改动 |
| `cmd/squadd` | 拆分为 `squad-api` / `squad-worker` |
| `internal/auth` | JWT & OAuth |
| `internal/store` | GORM 模型 |
| `web/` | Auth 流、Dashboard、Repo 选择 |

---

## 8 测试策略

| 层级 | 场景 | 工具 |
|------|------|------|
| 单元 | Auth、Queue | Go test |
| 集成 | Worker→容器→事件 | DinD in CI |
| E2E | OAuth→创建→输出 | Playwright |

---

## 9 工期预估

| 子任务 | 周期 |
|--------|------|
| OAuth & JWT | 1 周 |
| DB & GORM | 1 周 |
| 队列 + Worker | 1.5 周 |
| Docker 镜像 | 1 周 |
| 前端改造 | 1 周 |
| CI/CD & 测试 | 0.5 周 |
| **合计** | **≈5 周** |

---

## 10 风险与对策

| 风险 | 对策 |
|------|------|
| 容器资源爆量 | 限额 + 监控 |
| Redis 单点 | Redis Cluster |
| OAuth 滥用 | 限额 / 白名单 |
| 代码执行安全 | Read‑only FS, egress ACL |

---

完成 M2 后，`claude-squad` 变为可多用户在线使用的云端工作区，为后续 AI 回放、PR 合并等功能奠定分布式基础。

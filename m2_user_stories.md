# M2 阶段主要功能 — User Stories

以下以 **“作为 … 我想 … 以便 …”** 模式描述关键需求，并附 **验收要点**（Done Criteria）。

---

## 1. 登录 / 鉴权

| # | User Story | Done Criteria |
|---|------------|---------------|
| **A‑1** | **作为一名开发者**，我想使用自己的 GitHub 账户一键登录平台，以便不必创建新账号并让系统自动关联我的仓库权限。 | 1. 点击“使用 GitHub 登录”后跳转 GitHub OAuth 授权页；<br>2. 授权成功返回 `/oauth-callback?token=`；<br>3. JWT 被写入 `localStorage` 并随请求携带；<br>4. 授权失败时前端提示错误并支持重试。 |

---

## 2. 仪表盘 Dashboard

| # | User Story | Done Criteria |
|---|------------|---------------|
| **B‑1** | **作为已登录用户**，我想在仪表盘看到自己所有 AI Session 的列表、状态与资源用量，便于快速了解当前工作负载。 | 列表含 Repo 名、分支、状态 Tag、创建时间；WS 推动状态与资源实时刷新；空态显示“New Session”。 |
| **B‑2** | 作为用户，我想点“+ New Session”创建新会话，让 AI 代理开始工作。 | 表单校验 Repo/Branch 等必填；创建成功 201 并跳转 `/session/:id`。 |
| **B‑3** | 作为用户，我想在列表中直接 Resume / Pause / Kill 会话，无需进入详情页。 | 按钮两秒内改变状态；Kill 后卡片淡出。 |

---

## 3. 仓库管理 Repos

| # | User Story | Done Criteria |
|---|------------|---------------|
| **C‑1** | 作为用户，我想绑定 GitHub 仓库，以便 AI 会话能克隆并提交代码。 | 向导生成 Deploy Key → GitHub 设置 → “Test Connection” 成功后状态为 Connected。 |
| **C‑2** | 作为用户，我想删除或重新同步仓库，以防泄露或换分支。 | 操作菜单含 Re‑sync / Delete；删除需确认，之后不可再用创建 Session。 |

---

## 4. Session 工作台

| # | User Story | Done Criteria |
|---|------------|---------------|
| **D‑1** | 作为用户，我想实时查看 AI 代理终端输出，掌握执行进度。 | 进入页即建立 WS；xterm.js 滚动；刷新可重连并回放 1 000 行。 |
| **D‑2** | 作为用户，我想查看一次提交前的代码 Diff，以决定是否 Push。 | “Diff” 标签显示 diff2html；支持大文件虚拟滚动。 |
| **D‑3** | 作为用户，我想输入 Commit Message 并一键 Commit & Push。 | Message 必填 ≤ 72 字；成功 toast，失败展示错误详情。 |
| **D‑4** | 作为用户，我想监控 Session CPU/内存占用，以判断是否暂停或升级套餐。 | 资源条 2 s 更新；超限变红并提示 Upgrade。 |

---

## 5. 个人设置 Settings

| # | User Story | Done Criteria |
|---|------------|---------------|
| **E‑1** | 作为用户，我希望查看头像、昵称、注册日期、本月资源消耗。 | 卡片列出基本信息与 CPU 秒、Mem‑GB‑Hr 统计。 |
| **E‑2** | 作为用户，我想生成或撤销 API Token，供 CLI/脚本调用。 | “Generate” 仅显示一次；Token 以 SHA‑256 存储；可 “Revoke”。 |
| **E‑3** | （计费可选）作为用户，我想绑定信用卡并选择套餐，防止超额停机。 | Stripe Elements 流程；套餐升级写回后端。 |

---

## 6. 后台运维 Admin（仅管理员）

| # | User Story | Done Criteria |
|---|------------|---------------|
| **F‑1** | 作为管理员，我想查看所有用户、会话与节点资源，快速发现异常负载。 | Admin 路由 RBAC；表格支持排序过滤。 |
| **F‑2** | 作为管理员，我想强制终止异常 Session 容器，防止资源被耗尽。 | “Force Kill” 调 API，容器停止、DB 标记 `killed_by_admin`。 |

---

### MVP 首批交付建议

1. A‑1 登录  
2. B‑1 ~ B‑3 Dashboard  
3. D‑1 ~ D‑3 Session 工作台  

其余 Story 可在后续迭代逐步实现。

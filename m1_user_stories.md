# M1 阶段（本地单用户 Web GUI）User Stories

> 无登录、权限和账单概念；用户本机运行 `squadd`，浏览器访问 `http://localhost:7999`。

| # | User Story（作为…我想…以便…） | Done Criteria（验收要点） |
|---|--------------------------------|---------------------------|
| **A‑1** | **作为本机用户**，我想在浏览器输入 `localhost:7999` 就能打开工作区首页，以便不再使用终端命令行。 | 1. 首页正常加载 React SPA；<br>2. 后端未运行时显示「后端未启动」提示。 |
| **B‑1** | 作为用户，我想在首页 **列表** 中看到所有会话的标题、分支和状态，便于快速了解进度。 | 请求 `/api/sessions`；卡片显示 Title、Branch、Running/Paused/Finished；WS 实时刷新。 |
| **B‑2** | 作为用户，我想点击「+ New Session」并填写信息，创建新的 AI 会话。 | 表单必填 Title、Branch；创建成功返回 201 并跳转 `/session/:id`。 |
| **B‑3** | 作为用户，我想在列表卡片直接 **Pause / Resume / Kill** 会话，无需进入详情页。 | Running 状态显示 Pause；Paused 显示 Resume；Kill 后卡片淡出。 |
| **C‑1** | 作为用户，我想在工作台实时查看 AI 代理终端输出，掌握执行进度。 | 进入 `/session/:id` 建立 WS；xterm.js 滚动；刷新可回放 1 000 行。 |
| **C‑2** | 作为用户，我想查看 **Diff**，决定是否提交。 | 点击 Diff tab 加载 diff2html；支持大文件虚拟滚动。 |
| **C‑3** | 作为用户，我想输入 Commit Message 并 **Commit & Push**。 | Message ≤ 72 字；成功 toast，失败显示错误。 |
| **C‑4** | 作为用户，我想查看运行时长、分支名、工作目录，方便定位。 | 右栏显示 Branch、Elapsed Time、Worktree；时间每秒刷新。 |
| **D‑1** | 作为用户，我想使用浏览器“返回”键在页面间切换而不丢失状态。 | React Router 处理历史；WS 可自动重连。 |
| **E‑1** | 作为用户，我想用快捷键提高效率：<br>`Ctrl+\`` 聚焦 Terminal；`Ctrl+K` 清屏；`Tab` 切换 Terminal/Diff。 | 快捷键生效且不与浏览器默认冲突；Terminal 焦点输入正常。 |

**MVP 建议先交付**：A‑1，B‑1 ~ B‑3，C‑1 ~ C‑3。

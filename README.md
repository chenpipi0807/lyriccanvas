# 🎵 LyricCanvas — AI 驱动的歌词画布创作工具

> 一款基于 **Vue Flow 可视化画布** + **DeepSeek AI** 的中文歌词创作助手。  
> 在可拖拽的无限画布上自由编排歌词段落，结合 AI 智能补全、润色与押韵联想，让歌词创作如绘画般直观流畅。

---

<img width="1920" height="919" alt="ScreenShot_2026-05-16_194641_572" src="https://github.com/user-attachments/assets/41565fc2-c0c7-4c6a-a706-f218138f5382" />


## ✨ 功能特性

### 🎨 可视化歌词画布
- 基于 **Vue Flow** 的无限画布，支持拖拽、缩放、分组管理歌词节点
- 歌词段落可视化编排，直观查看歌曲结构（主歌 → 副歌 → 尾声）
- 支持锁定段落、调整顺序、一键生成草稿区

### 🤖 DeepSeek AI 智能创作
- **AI 续写**：根据已有歌词自动续写下一段
- **AI 润色**：优化用词、韵律与情感表达
- **AI 重写**：以不同风格重新演绎歌词
- **一键生成**：从风格描述直接生成完整歌词
- 支持流式（SSE）输出，实时查看 AI 创作过程

### 📝 中文押韵联想
- 内置 **拼音字典**，支持单字 / 批量押韵词查询
- 自动提取韵母，匹配同韵脚汉字
- 在画布节点上实时显示押韵提示

### 💬 对话式创作
- 侧边栏 AI 对话面板，支持多轮对话协作创作
- 快捷操作模板（续写 / 润色 / 重写 / 生成）
- 对话历史自动保存，上下文连贯

### 📁 项目管理
- 多项目支持，本地 JSON 文件存储
- 项目自动保存画布状态（视口、节点、分组）
- 项目列表按最近编辑时间排序

### 🚀 单文件部署
- Go 后端编译为**单一可执行文件**，内嵌前端静态资源
- 无需安装 Node.js、无需配置 Web 服务器
- 开箱即用，一个 exe 即整个应用

---

## 🖼️ 界面预览

```
┌──────────────────────────────────────────────────────┐
│  🎵 LyricCanvas           [项目列表] [新建] [设置]    │
├────────────┬───────────────────────────┬──────────────┤
│            │                          │              │
│  📁 项目   │    🎨 歌词画布           │  💬 AI 对话  │
│  侧边栏    │                          │              │
│            │   [前奏段]               │  助手: ...   │
│  · 歌曲A   │      ↓                  │  你: ...     │
│  · 歌曲B   │   [主歌A]               │              │
│  · 歌曲C   │      ↓                  │  [快捷操作]  │
│            │   [主歌B]               │  续写 润色   │
│  [+新建]   │      ↓                  │  重写 生成   │
│            │   [副歌] ★              │              │
│            │      ↓                  │  [输入框]    │
│            │   [尾声]                │              │
│            │                          │              │
└────────────┴──────────────────────────┴──────────────┘
```

---

## 🛠️ 技术栈

| 层级 | 技术 |
|------|------|
| **前端** | Vue 3 + TypeScript + Vite |
| **画布** | Vue Flow（基于 React Flow 的 Vue 版） |
| **状态管理** | Pinia |
| **HTTP 客户端** | Axios |
| **后端** | Go 1.23 + Gin |
| **AI 接口** | DeepSeek API（OpenAI 兼容） |
| **押韵引擎** | 自建拼音字典（extractFinal 韵母提取） |
| **存储** | 本地 JSON 文件 |
| **部署** | Go `embed` 内嵌前端 dist，单文件分发 |

---

## 📂 项目结构

```
lyriccanvas/
├── backend/                     # Go 后端
│   ├── main.go                  # 入口（内嵌前端 dist）
│   ├── go.mod / go.sum          # Go 模块
│   ├── build.bat                # Windows 构建脚本
│   ├── config/
│   │   └── config.go            # 环境变量配置加载
│   ├── handlers/
│   │   ├── chat.go              # AI 对话 API（含 SSE 流式）
│   │   ├── health.go            # 健康检查
│   │   ├── project.go           # 项目 CRUD
│   │   └── rhyme.go             # 押韵查询 API
│   ├── models/
│   │   └── project.go           # 数据模型定义
│   ├── services/
│   │   ├── deepseek_client.go   # DeepSeek API 客户端
│   │   ├── prompt_templates.go  # Prompt 模板引擎
│   │   ├── project_store.go     # 本地 JSON 存储
│   │   ├── rhyme_service.go     # 押韵引擎
│   │   └── pinyin_dict.json     # 拼音字典（embed）
│   ├── data/                    # 项目数据（gitignore）
│   └── frontend-dist/           # 构建产物（gitignore）
├── frontend/                    # Vue 3 前端
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── main.ts              # Vue 入口
│       ├── App.vue              # 根组件
│       ├── style.css            # 全局样式
│       ├── api/                 # API 封装
│       │   ├── client.ts        # Axios 实例
│       │   ├── chat.ts          # 对话 API
│       │   ├── projects.ts      # 项目 API
│       │   └── rhyme.ts         # 押韵 API
│       ├── components/
│       │   ├── canvas/          # 画布组件
│       │   │   ├── LyricCanvas.vue
│       │   │   ├── LyricNode.vue
│       │   │   └── GroupLabel.vue
│       │   ├── chat/            # 对话组件
│       │   │   ├── ChatPanel.vue
│       │   │   ├── ChatInput.vue
│       │   │   ├── ChatMessage.vue
│       │   │   ├── QuickActions.vue
│       │   │   └── TemplatePanel.vue
│       │   ├── layout/          # 布局组件
│       │   │   ├── AppLayout.vue
│       │   │   └── Modal.vue
│       │   └── sidebar/         # 侧边栏组件
│       │       ├── ProjectSidebar.vue
│       │       └── ProjectCard.vue
│       ├── stores/              # Pinia 状态管理
│       │   ├── canvas.ts
│       │   ├── chat.ts
│       │   └── project.ts
│       └── types/
│           └── index.ts         # TypeScript 类型定义
├── .env.example                 # 环境变量模板
├── .gitignore
└── README.md
```

---

## 🚀 快速开始

### 前提条件

- **Go** 1.23+
- **Node.js** 18+（仅开发前端时需要）
- **DeepSeek API Key** → [获取地址](https://platform.deepseek.com/)

### 1. 克隆项目

```bash
git clone https://github.com/chenpipi0807/lyriccanvas.git
cd lyriccanvas
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的 DeepSeek API Key：

```env
DEEPSEEK_API_KEY=sk-your-api-key-here
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_DEFAULT_MODEL=deepseek-chat
PORT=8080
DATA_DIR=./data
ALLOWED_ORIGINS=http://localhost:5173
```

### 3. 构建前端（可选，仅开发时）

```bash
cd frontend
npm install
npm run build      # 产物输出到 ../backend/frontend-dist/
```

### 4. 构建并启动后端

```bash
cd backend
go build -o lyriccanvas.exe .
./lyriccanvas.exe
```

或直接运行：

```bash
cd backend
go run .
```

启动后访问 **http://localhost:8080**

### 5. 前端开发模式（可选）

如需热更新开发前端：

```bash
cd frontend
npm run dev        # 启动 Vite 开发服务器 → http://localhost:8848
```

---

## 🔧 环境变量说明

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `DEEPSEEK_API_KEY` | DeepSeek API 密钥（**必填**） | - |
| `DEEPSEEK_BASE_URL` | API 地址 | `https://api.deepseek.com` |
| `DEEPSEEK_DEFAULT_MODEL` | 默认模型 | `deepseek-chat` |
| `PORT` | 后端监听端口 | `8080` |
| `DATA_DIR` | 项目数据存储目录 | `./data` |
| `ALLOWED_ORIGINS` | CORS 允许的前端地址 | `http://localhost:5173` |

---

## 📡 API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/health` | 健康检查 + DeepSeek 连通性 |
| `GET` | `/api/projects` | 获取项目列表 |
| `POST` | `/api/projects` | 创建新项目 |
| `GET` | `/api/projects/:id` | 获取项目详情 |
| `PUT` | `/api/projects/:id` | 更新项目 |
| `DELETE` | `/api/projects/:id` | 删除项目 |
| `POST` | `/api/chat` | 发送 AI 对话（支持 SSE 流式） |
| `GET` | `/api/chat/history/:projectId` | 获取对话历史 |
| `DELETE` | `/api/chat/history/:projectId` | 清空对话历史 |
| `GET` | `/api/chat/system-prompt` | 获取 System Prompt |
| `GET` | `/api/chat/templates` | 获取快捷模板列表 |
| `GET` | `/api/chat/actions` | 获取操作类型列表 |
| `POST` | `/api/lyrics/parse` | 智能解析歌词为节点 |
| `GET` | `/api/rhyme?char=花` | 单字押韵查询 |
| `POST` | `/api/rhyme/batch` | 批量押韵查询 |

---

## 🎯 创作法则

AI 生成歌词时遵循以下专业创作法则：

1. **结构完整**：前奏 → 主歌A → 主歌B → 副歌 → 间奏 → 主歌B' → 副歌 → 尾声
2. **副歌传唱**：2-4 句核心重复句式，朗朗上口
3. **段落对仗**：相邻主歌句数相近、句长对称
4. **标点必加**：每句末尾必须有标点，确保断句清晰
5. **尾韵优先**：尽量押韵但自然流畅优先
6. **文辞质量**：有文学感和画面感，避免口语大白话

---

## 🧩 设计理念

- **画布即结构**：歌词段落不再是线性文本，而是在画布上自由排列的节点，直观呈现歌曲结构
- **AI 即搭档**：AI 不替代创作者，而是作为协作伙伴提供灵感、润色和补全
- **单文件即应用**：Go embed 技术将所有前端资源打包进一个可执行文件，部署零依赖

---

## 📄 开源协议

MIT License

---

## 👤 作者

**Chen Pipi** — [GitHub](https://github.com/chenpipi0807)

---

<p align="center">
  <sub>Made with 🎵 and ☕ by Chen Pipi</sub>
</p>

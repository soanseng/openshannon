<p align="center">
  <img src="shannon.jpg" alt="OpenShannon" width="200">
</p>

<h1 align="center">OpenShannon</h1>

<p align="center">
  <a href="https://github.com/soanseng/openshannon/actions/workflows/deploy.yml"><img src="https://github.com/soanseng/openshannon/actions/workflows/deploy.yml/badge.svg" alt="Deploy"></a>
  <a href="https://github.com/soanseng/openshannon/releases"><img src="https://img.shields.io/github/v/tag/soanseng/openshannon?label=version" alt="Version"></a>
  <a href="https://goreportcard.com/report/github.com/soanseng/openshannon"><img src="https://goreportcard.com/badge/github.com/soanseng/openshannon" alt="Go Report Card"></a>
  <a href="https://github.com/soanseng/openshannon/blob/main/LICENSE"><img src="https://img.shields.io/github/license/soanseng/openshannon" alt="License"></a>
  <a href="https://openshannon.org"><img src="https://img.shields.io/badge/docs-openshannon.org-blue" alt="Docs"></a>
</p>

<p align="center">
  <a href="README.md">English</a> | 繁體中文
</p>

<p align="center">
  一個 Go daemon，將 Telegram 串接到 <a href="https://docs.anthropic.com/en/docs/claude-code">Claude Code</a> 或 Codex CLI，讓你從手機遠端操控 coding agent。隨時隨地用 Telegram 訊息指揮你的 coding assistant。
</p>

每個 Telegram Forum Topic 對應一個獨立的 agent session，擁有各自的工作目錄。
Claude Code session 會用 `--resume` 保留對話上下文；Codex 目前會在該
session workdir 和設定的 `add_dirs` 中逐次執行 prompt。

## 功能特色

- **文字、語音、圖片、檔案** — 支援各種訊息類型
- **Agent 切換** — 每個 session 可用 `/agent claude` 或 `/agent codex`
- **串流回覆** — 透過 `editMessageText` 看到 agent 的回應
- **Session 管理** — 多個獨立 session 對應到 Forum Topics
- **安全過濾** — 雙層防護（Go blocklist + Claude Code deny list）
- **直接執行 Shell** — `/shell` 指令快速跑系統命令
- **ntfy 通知** — daemon 事件的推播通知
- **systemd 服務** — 自動重啟、journald 日誌
- **單一執行檔** — 除了選定的 agent CLI 外無其他依賴

## 前置需求

- **Go 1.22+**
- **Claude Code CLI** 已安裝並認證（`claude --version`）
- （選用）**Codex CLI** 已安裝並認證（`codex --version`）
- **Telegram Bot** — 透過 [@BotFather](https://t.me/BotFather) 建立
- **你的 Telegram User ID** — 從 [@userinfobot](https://t.me/userinfobot) 取得
- （選用）**Groq API key** 用於語音轉文字

## 快速開始

### 1. 建立 Telegram Bot

1. 在 Telegram 開啟 [@BotFather](https://t.me/BotFather)
2. 發送 `/newbot` 並依照提示操作
3. 複製 bot token（格式如 `7123456789:AAH...`）
4. 發送 `/setprivacy` → 選擇你的 bot → `Disable`（讓 bot 能讀取群組訊息）

### 2. 設定 Forum 群組（建議）

1. 建立新的 Telegram 群組
2. 前往群組設定 → Topics → 啟用
3. 將 bot 加入群組
4. 將 bot 設為管理員（需要才能存取 topic）
5. 為你的專案建立 topics：「infra」、「feedbot」等

每個 topic 就是一個獨立的 agent session。Claude Code 會跨 prompt 保留
對話上下文；Codex 會在該 topic workdir 中逐次執行 prompt。

### 3. 安裝

```bash
git clone https://github.com/soanseng/openshannon.git ~/infra/openshannon
cd ~/infra/openshannon

# 互動式安裝精靈（推薦）
bash install.sh

# 或非互動式：make setup
```

安裝精靈會引導你完成：

1. **建置** — 編譯 Go 執行檔
2. **Telegram** — bot token + user ID 設定
3. **Gemini** — （選用）API key，用於 `/imagine` 圖片生成
4. **Google 服務** — （選用）gog CLI 認證，支援 Gmail、Calendar、Drive、Tasks、Contacts
5. **設定檔** — 寫入設定檔並設定正確權限
6. **工作區** — 建立 `~/OpenShannon/` 含 CLAUDE.md 和 systemd service

### 4. 設定

編輯 `~/.config/openshannon/config.yaml`：

```yaml
telegram:
  token: "${TELEGRAM_BOT_TOKEN}"
  allowed_users:
    - 你的_TELEGRAM_USER_ID    # ← 替換這裡
```

編輯 `~/.config/openshannon/env`：

```bash
TELEGRAM_BOT_TOKEN=7123456789:AAHxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
# 選用：
GEMINI_API_KEY=your_google_ai_api_key
GOG_KEYRING_PASSWORD=your_gog_keyring_password
GOG_ACCOUNT=your@gmail.com
```

### 5. 啟動服務

```bash
make start    # 啟用並啟動 systemd service
make status   # 檢查狀態
make logs     # 看即時 log
```

## 使用方式

### 基本互動

直接發送文字訊息給 bot — 它會直接變成目前選定 agent 的 prompt：

```
你: 幫我找出程式碼中所有的 TODO 註解
Bot: ⚡（處理中...）
Bot: 我在 5 個檔案中找到了 12 個 TODO 註解...
```

### 指令列表

| 指令 | 說明 | 範例 |
|---|---|---|
| `/new [workdir]` | 建立新 session | `/new ~/infra` |
| `/resume [id]` | 恢復閒置的 session | `/resume` |
| `/sessions` | 列出所有 session | `/sessions` |
| `/clear` | 清除 agent 上下文，保留 workdir | `/clear` |
| `/kill [id]` | 完全刪除 session | `/kill` |
| `/cd <path>` | 切換工作目錄 | `/cd ~/apps/feedbot` |
| `/status` | Daemon 狀態與統計 | `/status` |
| `/cancel` | 取消執行中的指令 | `/cancel` |
| `/shell <cmd>` | 直接執行 shell 指令 | `/shell git status` |
| `/long <prompt>` | 延長 timeout 至 30m | `/long 重構整個模組` |
| `/agent [name]` | 切換 coding agent | `/agent codex` |
| `/model [name]` | 切換模型 | `/model haiku` |
| `/imagine <desc>` | 生成圖片（Gemini） | `/imagine 太空貓` |
| `/gog <cmd>` | Google 服務 | `/gog gmail search newer_than:1d` |
| `/help` | 顯示所有指令 | `/help` |

### Forum Topics = Sessions

在啟用 Forum 的群組中，每個 topic 是一個獨立 session：

```
Topic: "infra"       → workdir: ~/infra
Topic: "feedbot"     → workdir: ~/apps/feedbot
Topic: "openshannon" → workdir: ~/infra/openshannon
```

### 模型切換

每個 topic/session 可以使用不同的模型：

```
/model haiku       # Claude Haiku 4.5（快速、便宜）
/model sonnet      # Claude Sonnet 4.6（平衡）
/model opus        # Claude Opus 4.6（最強）
/model gemini      # Gemini 2.5 Flash
/model gemini-pro  # Gemini 2.5 Pro
/model default     # 重設為預設
```

### Agent 切換

每個 topic/session 可以使用 Claude Code 或 Codex CLI：

```
/agent claude   # 使用 Claude Code
/agent codex    # 使用 Codex CLI
/agent default  # 重設為 Claude Code
```

Codex 會以 `codex exec --cd <workdir>` 執行。設定 `codex.default_workdir`
和 `codex.add_dirs` 可以讓 Codex 使用完整 OpenShannon repo，例如
`~/infra/openshannon`。

### 圖片生成

使用 Claude 優化 prompt，再用 Gemini 3.1 Flash 生成圖片：

```
/imagine 一隻穿太空衣的貓在畫蒙娜麗莎
```

需要在 env 檔中設定 `GEMINI_API_KEY`。從 [Google AI Studio](https://aistudio.google.com/apikey) 取得。

### Google 服務

透過 [gog CLI](https://github.com/AarynSmith/gog) 整合 Google Workspace：

```
/gog gmail search newer_than:1d          # 最近的郵件
/gog gmail send --to x@y.com --subject "嗨" --body "你好"
/gog calendar events                     # 今天的行程
/gog drive ls                            # Drive 檔案列表
/gog tasks lists list                    # 任務清單
/gog contacts search "名字"              # 搜尋聯絡人
```

### 安全機制

雙層防護：

**第一層 — Go daemon blocklist**（在選定 agent 看到 prompt 之前）：
- 攔截危險指令：`rm -rf /`、`mkfs`、`dd if=`、`curl | sh`
- 攔截危險 shell：`sudo`、`shutdown`、`git push --force`
- 保護敏感路徑：`/etc/`、`/boot/`、`~/.ssh/authorized_keys`

**第二層 — Claude Code deny list**（你的 `settings.json`）：
- 攔截工具執行：`Bash(sudo *)`、`Bash(rm -rf /*)` 等

## 法律聲明

OpenShannon 是一個獨立的開源專案，**與 Anthropic 無任何關聯、背書或贊助關係**。

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) 是 [Anthropic](https://www.anthropic.com) 的產品。「Claude」和「Anthropic」是 Anthropic, PBC 的商標。
- OpenShannon 使用 Claude Code 官方 CLI（`claude -p`）的 print mode — 一個 Anthropic 公開提供的程式化介面。
- 本專案不修改、不逆向工程、不重新發布 Claude Code 本身。
- 使用者必須擁有自己的 Anthropic 帳號，並遵守 Anthropic 的[消費者服務條款](https://www.anthropic.com/legal/consumer-terms)、[商業條款](https://www.anthropic.com/legal/commercial-terms)和[使用政策](https://www.anthropic.com/legal/aup)。

**使用者有責任確保使用本工具時符合所有適用的 Anthropic 條款和政策。**

## 授權

MIT

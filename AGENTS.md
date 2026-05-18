# AGENTS.md — EzeTranslate

> 本文件面向 AI Agent，描述项目结构、开发规范和常用命令。

## 项目概览

| 字段 | 值 |
|------|-----|
| 项目名 | EzeTranslate |
| 语言 | Go 1.22 |
| GUI 框架 | Fyne v2.7.3 |
| 模块路径 | `github.com/Ericwyn/EzeTranslate` |
| 当前版本 | V1.8-Release |
| 支持平台 | Linux、Windows |

一款纯 Go 实现的桌面翻译工具，支持 Google / Baidu / Youdao / OpenAI 四个翻译引擎，具备划词翻译（Linux xclip）和 OCR 截图翻译（Ubuntu GNOME）功能。

---

## 常用命令

| 用途 | 命令 |
|------|------|
| 安装依赖 | `go mod download` |
| 编译验证 | `go build ./...` |
| 完整构建（Linux） | `bash build.sh` |
| 完整构建（Windows） | `build.bat` |
| 打 deb 包 | `bash build-deb.sh` |
| 运行测试 | `go test -v ./...` |
| 格式化代码 | `gofmt -w .` |

> **重要**：每次修改代码后，必须执行 `go build ./...` 确认可以正常编译。

构建产物输出到 `./build-target/` 目录。构建命令统一使用 `-trimpath -ldflags="-s -w"` 缩减二进制体积。

---

## 目录结构

```
EzeTranslate.go          # 入口，解析 -x / -ocr / -v 参数
conf/                    # 配置管理（viper + YAML）
  config.go              # 配置键常量、InitConfig、SaveConfig
trans/                   # 翻译引擎
  trans.go               # 公共类型：TransResCallback、GetSelection（xclip）
  google/                # Google 翻译，支持自定义 URL 和 HTTP 代理
  baidu/                 # 百度翻译，MD5 签名
  youdao/                # 有道翻译，SHA256 签名
  openai/                # OpenAI Chat API（默认引擎），返回 JSON
ui/                      # Fyne GUI 层
  ui.go                  # StartApp、startTrans、IPC 消息处理
  home.go                # 完整模式窗口（450×600），EzeInputEntry 自定义输入框
  mini.go                # 迷你模式窗口（400×300），无输入框
  set.go                 # 设置窗口，配置各 API Key/URL
  assembly.go            # 公共 UI 组件：格式化 CheckBox、翻译引擎选择、目标语言下拉
  menu.go                # 菜单栏（设置 / 关于）
  log.go                 # 日志查看窗口
  about.go               # 关于窗口
  resource/
    resource.go          # 图标、字体静态资源（内嵌）
    theme.go             # 自定义主题（强制亮色 + NotoSansSC 字体）
    cusWidget/
      checkGroup.go      # CheckGroup、CreateCheckGroup、CreateDropDown 自定义组件
ipc/                     # 进程间通信（仅 Linux）
  server.go              # IPC 消息类型定义、StartUnixSocketListener
  client.go              # SendMessage
  socket.go              # Unix Domain Socket 实现（/tmp/trans_utils.socket）
ocr/                     # OCR 识别（仅 Ubuntu GNOME）
  ocrrr.go               # gnome-screenshot → mogrify → tesseract 流程
strutils/                # 字符串工具
  strutils.go            # FormatInputBoxText、FormatCamelCaseText、DetectLanguage
log/                     # 日志模块
ajax/                    # HTTP 工具
res-static/              # 嵌入式静态资源（字体、图标）
  embed.go               # go:embed 声明
```

---

## 架构要点

### 翻译流程

```
用户输入 / 划词 / OCR
  → strutils.FormatInputBoxText()   # 输入优化（去注释/空格/回车/驼峰拆解）
  → 检测语言（DetectLanguage）
  → 选择引擎（viper 配置 translateSelect）
  → 异步调用对应 trans/xxx.Translate()
  → 回调更新 UI（resBox.SetText）
```

### IPC 多进程机制（Linux）

第二个进程启动时，通过 Unix Socket 向已有进程发送消息后退出，不重复开窗口。消息类型：

| 消息 | 含义 |
|------|------|
| `PING` | 探测是否有进程在运行 |
| `NEW_SELECTION` | 获取 xclip 划词并翻译 |
| `OCR` | 仅 OCR 识别，不翻译 |
| `OCR_AND_TRANS` | OCR 识别后翻译 |

### 窗口模式切换

- 完整模式 ↔ 迷你模式可互相切换，切换时保留当前翻译结果
- 模式偏好持久化到配置文件（`miniMode` 键）

### OCR 流程

截图前隐藏窗口（避免遮挡），截图后恢复：
```
Hide() → sleep 300ms → gnome-screenshot -a → mogrify 放大400%+灰度 → tesseract → Show()
```

---

## 配置文件

路径：`~/.config/EzeTranslate/config.yaml`（由 `conf.GetConfigFileDirPath()` 确定）

| 配置键 | 说明 | 默认值 |
|--------|------|--------|
| `translateSelect` | 翻译引擎 | `openai` |
| `miniMode` | 启动迷你模式 | `false` |
| `baiduTransAppId` / `baiduTransAppSecret` | 百度 API | 占位符 |
| `youdaoTransAppId` / `youdaoTransAppSecret` | 有道 API | 占位符 |
| `googleTranslateUrl` | Google 翻译地址 | `https://translate.googleapis.com` |
| `googleTranslateProxy` | Google HTTP 代理 | 空 |
| `openAiApiUrl` | OpenAI 接口地址 | `https://api.openai.com/v1/chat/completions` |
| `openAiKey` | OpenAI Key | 占位符 |
| `openAiModel` | OpenAI 模型 | 空 |
| `formatAnnotation/Space/CarriageReturn/CamelCase` | 输入优化开关 | `false` |

---

## 开发规范

- **修改代码时做局部修改**，不要重写整个文件
- **保留原有注释**，除非注释本身有错误
- **每次改动后必须编译验证**：`go build ./...`
- 新增翻译引擎需实现签名 `func Translate(str string, toLang strutils.Lang, callback trans.TransResCallback)`，并在 `ui/ui.go` 的 `startTrans()` 中注册
- UI 组件遵循 Fyne 规范，复用 `ui/resource/cusWidget/` 中已有组件
- 不要在代码中硬编码 API Key 或敏感信息，统一通过 viper 读取配置

---

## CI

`.github/workflows/build.yml` 在 push/PR/release 时触发：

1. **test** job：`go test -v ./...`（ubuntu-latest）
2. **build-binaries** job：编译 linux-amd64 / darwin-amd64 / windows-amd64，使用 `-trimpath -ldflags="-s -w"`
3. **build-deb** job：打 `.deb` 安装包

Release 时自动上传产物到 GitHub Release。

---

## 平台限制

| 功能 | 限制 |
|------|------|
| 划词翻译（xclip） | 仅 Linux，需安装 `xclip` |
| IPC 多进程 | 仅 Linux（Unix Socket） |
| OCR 截图翻译 | 仅 Ubuntu + GNOME，需安装 `gnome-screenshot`、`tesseract-ocr`、`imagemagick`（mogrify） |
| Windows 剪贴板监听 | 待实现 |

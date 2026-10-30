# AGENTS.md — EzeTranslate

> 本文件面向 AI Agent，描述项目结构、开发规范和常用命令。

## 项目概览

| 字段 | 值 |
|------|-----|
| 项目名 | EzeTranslate |
| 语言 | Go 1.26+（go.mod 声明 1.26，Viva 要求） |
| GUI 框架 | Viva v0.2.0（`github.com/vivaui/viva`，GPU 原生，无 WebView） |
| 模块路径 | `github.com/Ericwyn/EzeTranslate` |
| 当前版本 | V1.9-Release |
| 支持平台 | Linux（X11）、Windows |

一款纯 Go 实现的桌面翻译工具，支持 Google / Baidu / Youdao / OpenAI 四个翻译引擎，具备划词翻译（Linux xclip）和 OCR 截图翻译（Ubuntu GNOME）功能。

> GUI 已于 V1.9 从 Fyne 迁移到 Viva。Viva 直接调用系统窗口与 GPU（Linux=XCB/Vulkan，Windows=D3D11，macOS=Cocoa/Metal），不打包渲染运行时。

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

> **Go 工具链**：go.mod 要求 Go 1.26+。若系统 `go` 版本较旧，构建脚本已设置
> `GOTOOLCHAIN=auto`，`go` 命令会自动下载匹配工具链（可用 `VIVA_GO_DOWNLOAD=0`
> 或离线环境预装 Go 1.26+ 规避）。

> **CGO**：Linux / Windows 后端在运行时 dlopen 系统图形库，构建用 `CGO_ENABLED=0`
> （脚本已设置），无需 `xorg-dev`/`libgl1-mesa-dev`。macOS 后端经 cgo 调 Cocoa/Metal，
> 必须在 macOS 上带 Xcode 命令行工具原生构建（`CGO_ENABLED=1`），不能跨平台交叉编译。

构建产物输出到 `./build-target/` 目录。构建命令统一使用 `-trimpath -ldflags="-s -w"` 缩减二进制体积。

---

## 目录结构

```
EzeTranslate.go          # 入口，解析 -x / -ocr / -v 参数
conf/                    # 配置管理（viper + YAML）
  config.go              # 配置键常量、InitConfig、SaveConfig
trans/                   # 翻译引擎（与 GUI 框架无关，迁移时未改动）
  trans.go               # 公共类型：TransResCallback、GetSelection（xclip）
  google/                # Google 翻译，支持自定义 URL 和 HTTP 代理
  baidu/                 # 百度翻译，MD5 签名
  youdao/                # 有道翻译，SHA256 签名
  openai/                # OpenAI Chat API（默认引擎），返回 JSON
ui/                      # Viva GUI 层（事件驱动 Paint，从 model 读状态）
  state.go               # ezeApp 应用结构、Tree/控件标识、StartApp、IPC 接线、翻译调度、OCR/划词
  transstate.go          # transState：home/mini 共享的翻译数据（输入框/结果框/note/loading/下拉状态）
  editbox.go             # 多行可编辑框助手（封装 text.Editor + TextareaLayout + 按键/滚动/IME）
  assembly.go            # 公共控件：输入优化勾选、翻译引擎 Select、目标语言 Select
  home.go                # 完整模式窗口（450×600），含输入框+结果框
  mini.go                # 迷你模式窗口（400×300），无输入框
  set.go                 # 设置窗口，9 个 flex.Field 表单 + 保存
  log.go                 # 日志查看窗口
  about.go               # 关于窗口（app.OpenURL 打开 GitHub 链接）
  resource/
    resource.go          # WindowIcon()：把内嵌 icon.png 解码成 Viva 需要的 RGBA 图标
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
viva-spike/              # Viva 技术验证 demo（独立 module，replace 指向主模块，可独立 viva dev）
```

---

## 架构要点

### Viva 编程模型（与 Fyne 的本质差异）

- **事件驱动 Paint**：`Config.Paint(canvas, size)` 在每次可能改变画面的事件后被调用，
  它只从 model 读状态来绘制，**不得修改状态**（按钮点击是唯一例外，在声明处处理）。
  空闲时不绘制，没有帧循环。
- **状态集中在 model**：可变状态放在 `ezeApp` / `transState` 等结构体，Paint 只读。
- **跨线程回传**：后台 goroutine（翻译/IPC/OCR）**不能直接碰窗口**。它先在加锁区写入
  共享状态，再调 `Invalidator.Invalidate()`；窗口线程的 `OnWake` 在重绘前把结果搬进 model。
  这取代了 Fyne 的 `fyne.Do`。
- **多窗口**：`app.RunApplication` 托管事件循环，`application.OpenWindow(config, canvas)`
  开每个窗口，各自独立。最后一个窗口关闭后事件循环退出。
- **多行可编辑文本**：Viva 原生文本服务只认单行，多行换行/按视觉行移动/滚动靠
  `components.TextareaLayout` 协助，`Config.TextInput` 把原生输入接到 `text.Editor`。
  统一封装在 `ui/editbox.go`。

### 翻译流程

```
用户输入 / 划词 / OCR
  → strutils.FormatInputBoxText()   # 输入优化（去注释/空格/回车/驼峰拆解）
  → 检测语言（DetectLanguage）
  → 选择引擎（viper 配置 translateSelect）
  → 异步调用对应 trans/xxx.Translate()（后台 goroutine）
  → transState.setResult() 写入结果 + Invalidate()
  → 窗口 OnWake 前重绘，结果框显示
```

### IPC 多进程机制（Linux）

第二个进程启动时，通过 Unix Socket 向已有进程发送消息后退出，不重复开窗口。消息类型：

| 消息 | 含义 |
|------|------|
| `PING` | 探测是否有进程在运行 |
| `NEW_SELECTION` | 获取 xclip 划词并翻译 |
| `OCR` | 仅 OCR 识别，不翻译 |
| `OCR_AND_TRANS` | OCR 识别后翻译 |

IPC 回调在后台 goroutine 执行，开窗/隐藏/聚焦等窗口操作通过 `postHome`/`postMini`
投递到窗口线程（Invalidate + OnWake 里 drain 执行）。

### 窗口模式切换

- 完整模式 ↔ 迷你模式可互相切换，切换时保留当前翻译结果（共享同一份 `transState`）
- 切换时置 `switching=true`，使被关闭窗口的 `OnClosed` 不退出整个应用
- 模式偏好持久化到配置文件（`miniMode` 键）

### OCR 流程

截图前隐藏窗口（避免遮挡），截图后恢复：
```
postActive(Hide) → sleep 300ms → gnome-screenshot -a → mogrify 放大400%+灰度
  → tesseract → postActive(Show + 填入输入框 + 可选翻译)
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
- 新增翻译引擎需实现签名 `func Translate(str string, toLang strutils.Lang, callback trans.TransResCallback)`，并在 `ui/state.go` 的 `startTrans()` 中注册
- UI 控件遵循 Viva `flex` 规范：容器 `Row/Column`、叶子 `Text/Button/Checkbox/Field/Select/Textarea`；
  控件需用 `tree.ID(...)` 稳定标识命名；新增多行输入复用 `ui/editbox.go`
- Paint 函数只读状态、不得分配大对象、不得启动 goroutine；跨线程改 UI 走 Invalidate + OnWake
- 不要在代码中硬编码 API Key 或敏感信息，统一通过 viper 读取配置
- 业务层 `trans/`、`ipc/`、`ocr/`、`conf/`、`strutils/` 与 GUI 框架无关，改 UI 时不要动它们

---

## CI

`.github/workflows/build.yml` 在 push/PR/release 时触发：

1. **build-binaries** job：matrix 编译 linux-amd64（CGO=0）/ darwin-amd64（macOS runner，CGO=1）/
   windows-amd64（CGO=0），Go 版本取自 `go.mod`，使用 `-trimpath -ldflags="-s -w"`
2. **build-deb** job：打 `.deb` 安装包（仅需 `dpkg-dev`，不再需要 xorg-dev/mesa 开发包）

Release 时自动上传产物到 GitHub Release。

---

## 平台限制

| 功能 | 限制 |
|------|------|
| GUI 渲染 | Linux 仅 X11（Viva Linux 后端为 X11/Vulkan，暂无原生 Wayland；Wayland 会话经 XWayland 运行） |
| 运行时依赖（Linux） | 需系统 `libvulkan1`（+ 显卡驱动或 `mesa-vulkan-drivers` 软件渲染）、`libxcb1`、`libpangocairo-1.0-0`（deb 已声明 Depends） |
| 划词翻译（xclip） | 仅 Linux，需安装 `xclip` |
| IPC 多进程 | 仅 Linux（Unix Socket） |
| OCR 截图翻译 | 仅 Ubuntu + GNOME，需安装 `gnome-screenshot`、`tesseract-ocr`、`imagemagick`（mogrify） |
| 中文输入法 | Linux 经 IBus / Fcitx5 的 IBus 前端（Viva 实验阶段，预编辑下划线位置可能偏移） |
| 输入法中英切换 | Viva v0.2 X11 IME 后端尚未打通切换键（Shift / Ctrl+Space 等），应用层无法处理，需用系统输入法自带全局切换键 |
| Ctrl+Enter 触发翻译 | 多行输入框下 Enter（含带修饰键）被 Viva 文本层当换行吃掉，不进 OnKey/OnShortcut，故不提供回车类翻译快捷键，统一用「执行翻译」按钮 |
| Ctrl+A 全选 | Viva Linux 文本层未内建 select-all；应用对自管 editor 的输入框（home 输入框、设置窗字段）在 OnShortcut 里手动实现 Ctrl+A 全选 |
| Windows 剪贴板监听 | 待实现 |

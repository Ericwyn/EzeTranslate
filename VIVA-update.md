# 从 Fyne 迁移到 Viva 实战指南

> 本文记录 EzeTranslate 把桌面 GUI 从 **Fyne v2.7.3** 全量迁移到 **Viva v0.2.0** 的
> 完整过程、踩到的坑和解决方案。目标读者：后续想把类似的 Fyne 简易桌面程序迁到 Viva 的人。
>
> 本次迁移对应提交：`feat(ui): 将 GUI 从 Fyne 全量迁移到 Viva，并升级至 v2.0.0`

---

## 0. 一句话结论

Viva 在 **Linux X11 + GNOME** 上**技术可行**，桌面观感和组件比 Fyne 现代，二进制更小
（本项目 Fyne 版 cgo 链接 → Viva 版 CGO=0，Linux 产物约 10MB）。但 Viva 处于
**v0.2.0 + Linux 后端 experimental** 阶段，中文 IME、键盘快捷键等有明确限制（见第 5 节），
生产项目迁移前务必先做可行性 spike 验证这些点。

---

## 1. Viva 替代 Fyne 的可行性

### 1.1 验证过可行的能力

| 能力 | Viva 表现 |
|------|-----------|
| 窗口 / 多窗口 | ✅ `app.RunApplication` + `OpenWindow`，每窗口独立 |
| 中文显示 | ✅ CoreText/Pango 系统字体回退，中英文混排正常 |
| 中文输入（IME） | ✅ Linux 走 IBus/Fcitx5 的 IBus 前端，拼音候选词、上屏可用 |
| 多行文本编辑 | ✅ 但需手工接线（见 3.3），不像 Fyne 一个 Entry 搞定 |
| 单行输入 / 表单 | ✅ `flex.Field` 或自管 editor + `f.Input` |
| 勾选框 / 下拉 / 按钮 | ✅ `flex.Checkbox / Select / Button` |
| 滚动区 | ✅ `flex.ScrollArea` + `View.Wheel` 路由 |
| 系统托盘 / 弹出面板 | ✅ `AddStatusItem` + `WindowPanel`（本项目未用，但 API 存在，比 Fyne 原生） |
| 打开 URL | ✅ `app.OpenURL`（仅 http/https，拒绝 file://） |
| 后台→UI 跨线程更新 | ✅ `Invalidator.Invalidate()` + `OnWake`（替代 `fyne.Do`） |
| 二进制体积 / 空闲占用 | ✅ 更小、事件驱动无空转帧循环 |

### 1.2 关键前置条件

- **Go 版本**：Viva 要求 **Go 1.26+**。系统 go 较旧时，`GOTOOLCHAIN=auto` 会自动下载
  匹配工具链到用户缓存（`~/.cache/viva/go` 或 go 默认缓存）。离线环境需预装。
- **Linux 运行时库**：Viva 在运行时 `dlopen` 系统 `libvulkan1`、`libxcb1`、
  `libpangocairo-1.0-0`，**不是 cgo 静态链接**。因此：
  - 构建用 `CGO_ENABLED=0`（Linux/Windows），**不需要** `xorg-dev`/`libgl1-mesa-dev`
  - 但**运行机器**需装上述 .so（deb/rpm 要声明 Depends）
  - 无 GPU 的机器需 `mesa-vulkan-drivers`（lavapipe 软件渲染）
- **会话类型**：Viva Linux 后端目前**仅 X11**。Wayland 会话要靠 XWayland 跑，无原生 Wayland。
- **平台构建差异**：
  | 平台 | CGO | 说明 |
  |------|-----|------|
  | Linux | 0 | 任意机器可编，运行时加载 XCB/Vulkan |
  | Windows | 0 | 可从 Linux 交叉编译 |
  | macOS | **1** | 必须在 macOS 带 Xcode 原生构建（cgo 调 Cocoa/Metal），**不能交叉编译** |

---

## 2. 迁移前的准备（spike 验证）

生产项目不要盲目全量替换。先花最小成本验证三个"生死问题"：

1. **环境就绪**：`go install github.com/vivaui/viva/cmd/viva@latest`，然后 `viva doctor`
   （检查 libxcb/libvulkan/工具链）。`viva init demo` 生成可跑模板。
2. **中文输入**：做一个最小窗口，放一个 `flex.Field`，真机用拼音输入法打字，确认候选词能上屏。
3. **业务复用**：用 `replace` 指回主模块，在 spike 里真实调用一次你现有的业务函数，确认
   GUI 层和业务层能解耦。

本项目 spike 一次性验证通过了这三点，才决定全量迁移。

---

## 3. 编程模型差异（这是迁移的核心认知）

### 3.1 保留式 widget 树 → 事件驱动 Paint

- **Fyne**：创建 `widget.Entry`/`Label` 等对象持有引用，随时 `.SetText()` 改它，框架自动重绘。
- **Viva**：`Config.Paint(canvas, size)` 每次事件后被调用，**从 model 读状态重新描述整个界面**。
  - Paint **只能读状态，不能改状态**（唯一例外：按钮点击在声明处处理，如 `if f.Button(id,"x").Clicked() {...}`）。
  - Paint **不得有副作用**：不打日志、不读写文件、不发网络、不开 goroutine、不每帧 `fmt.Sprintf` 大对象。
  - 空闲时不绘制，没有帧循环。

**迁移心法**：把所有"可变状态"收拢进一个 model 结构体（本项目是 `ezeApp` + `transState`），
UI 代码变成"纯函数：model → 界面描述"。

### 3.2 跨线程更新（替代 fyne.Do）

后台 goroutine（翻译/IPC/OCR）**绝不能直接碰窗口或 widget**。标准模式：

```go
// Config 里：
config.Invalidator = inv          // inv := &app.Invalidator{}
config.OnWake = func() { /* 在窗口线程把后台结果搬进 model */ }

// 后台 goroutine：
result := doWork()
mu.Lock(); sharedState = result; mu.Unlock()
inv.Invalidate()                  // 唤醒窗口线程，触发一次重绘；OnWake 先跑
```

本项目还封装了 `postHome/postMini`（投递闭包到窗口线程）处理 IPC 回调里要开窗/隐藏窗的场景。

### 3.3 多行可编辑文本（最难的部分）

Viva 原生文本服务**只认识单行**。多行框要手工接三样东西（本项目封装在 `ui/editbox.go`）：

- `text.Editor`：持有文本内容，`Config.TextInput` 把原生输入接到它
- `components.TextareaLayout`：负责软换行、按视觉行移动光标、滚动
- `components.TextareaSnapshot`：每帧把 editor 的值/选区/IME 组合态喂给 `f.Textarea()`

关键接线：
```go
// TextInput provider：告诉平台"聚焦的这个控件用哪个 editor"
config.TextInput = func(ctx app.TextInputContext) *app.TextInput {
    if ctx.ID == myBoxID { return &app.TextInput{Editor: myEditor} }
    return nil
}
// 多行的上下键/翻页要自己用 layout.Move 算（单行不用）
// 滚轮走 View.Wheel 路由到 ScrollArea
```

单行输入（如设置项）更简单：`flex.Field`（框架自带 editor）或自管 editor + `f.Input`。
本项目设置窗为了支持 Ctrl+A 全选，用的是**自管 editor + f.Input**（见 5.3）。

### 3.4 布局：flex 声明式

`f.Column()/Row()/Stack()` 容器 + `Text/Button/Checkbox/Field/Select/Textarea` 叶子，
链式设尺寸 `.Fill()/.Grow(1)/.Width()/.MinWidth()/.Pad()/.Gap()`。
每个交互控件要用 `tree.ID(scope,"name")` 稳定标识命名。

---

## 4. 本次具体改了哪些东西

### 4.1 代码结构（ui/ 层整体重写，业务层零改动）

| 文件 | 变化 |
|------|------|
| `ui/state.go` | 🆕 ezeApp 多窗口托管、Tree 标识、StartApp、IPC 接线、翻译调度、OCR/划词 |
| `ui/transstate.go` | 🆕 home/mini 共享翻译状态，含 Invalidator、Select 开合态 |
| `ui/editbox.go` | 🆕 多行/单行可编辑框封装（text.Editor + TextareaLayout + Ctrl+A/滚动/IME） |
| `ui/screen.go` | 🆕 xrandr/xdotool 查屏幕几何，窗口按光标所在屏居中（多屏适配） |
| `ui/assembly.go` | 公共控件：输入优化勾选、引擎 Select、目标语言 Select |
| `ui/home.go` `mini.go` `set.go` `log.go` `about.go` | 全部重写为 Viva Paint |
| `ui/resource/resource.go` | 改为 `WindowIcon()`：PNG→RGBA |
| `ui/ui.go` `menu.go` `resource/theme.go` `loader.go` `cusWidget/` | ❌ 删除（Fyne 专属） |
| `ui/home_test.go` `set_test.go` | 🆕 headless 测试（渲染、中文输入、Ctrl+A 全选） |
| `trans/` `ipc/` `ocr/` `conf/` `strutils/` | ✅ **零改动**（仅 conf 改版本号常量） |

### 4.2 构建 / 依赖 / CI

- `go.mod`：`go 1.22 → 1.26`，移除整套 Fyne 依赖，引入 `github.com/vivaui/viva`
- `build.sh` / `build.bat` / `build-deb.sh`：加 `GOTOOLCHAIN=auto` + `CGO_ENABLED=0`，
  去掉 Fyne 时代的 `xorg-dev`/`libgl1-mesa-dev` 依赖
- `.github/workflows/build.yml`：matrix 按平台设 CGO（Linux/Windows=0，macOS=1），
  Go 版本改用 `go-version-file: go.mod`
- `deb-build-tpl/DEBIAN/control`：Depends 增加 `libvulkan1, libxcb1, libpangocairo-1.0-0`，
  Recommends `mesa-vulkan-drivers`

---

## 5. 踩到的坑 & 解决方案

### 5.1 窗口不居中（多屏）

- **现象**：Viva `Config.Origin` 为 nil 时交给窗管摆放，GNOME/Mutter 默认不居中。
- **原因**：Viva v0.2 **没有暴露查询屏幕/显示器尺寸的 API**。
- **解决**：用 `xrandr --listmonitors` 拿各屏几何 + `xdotool getmouselocation` 拿光标位置，
  自己算出光标所在屏的居中坐标，填进 `Config.Origin`（见 `ui/screen.go`）。拿不到信息时
  返回 nil 退化为窗管默认（不报错）。

### 5.2 设置窗滚动时疯狂刷日志

- **现象**：滚动设置窗，每帧打一条"找到配置文件"日志。
- **原因**：`Paint` 里每帧调了 `conf.GetConfigFilePath()`，该函数有 `os.Stat` + 打日志的副作用。
  **违反 Viva "Paint 不得有副作用" 的铁律**。
- **解决**：打开窗口时取一次路径缓存进 model，Paint 只读缓存。
- **通用教训**：迁移后务必审查所有 Paint 路径，把任何文件/网络/日志/大对象分配移出去。

### 5.3 Ctrl+A 全选在 Linux 不工作

- **原因**：Viva Linux 文本层**没内建 select-all**，Ctrl+A 不会被原生处理。
- **解决**：对**自管 editor** 的输入框，在 `Config.OnShortcut` 里拦截 Ctrl+A（`e.Text=="a"` +
  Ctrl/Meta 修饰），手动 `editor.SetSelection({0, RuneCount})` 并返回 `Handled`。
  - ⚠️ `flex.Field` 的 editor 是**私有**的，无法全选。本项目因此把设置窗的 9 个字段
    从 `flex.Field` 改成**自管 editor + f.Input**（`newSingleLineEditBox`）才能支持 Ctrl+A。

### 5.4 Ctrl+Enter 触发动作 —— 做不到（多行框）

- **实证**：多行框聚焦时按 Ctrl+Enter，文本直接变 `"...\n"`（换行），`OnKey`/`OnShortcut`
  都收到 **0 次**。Viva 文本层在更底层就把多行框的所有 Enter（含带修饰键）当换行吞掉。
- **结论**：保留多行原生编辑的前提下，**任何含 Enter 的快捷键都无法用于触发动作**。
  用按钮代替（本项目保留「执行翻译」按钮）。单行 `Field` 的 Enter 可以走 `Submitted()`。

### 5.5 输入法中英切换无法响应

- **原因**：切换键（单独 Shift / Ctrl+Space）的处理在 Viva v0.2 的 X11 IME 后端内部，
  还没打通，**应用层触达不到**。
- **解决**：无应用层解法。用系统输入法自己的全局切换键（fcitx/ibus 配置），不经过应用。
- 另：IME 预编辑下划线在 Linux 下位置可能画偏（Viva 实验阶段渲染细节），不影响功能。

### 5.6 CGO / 交叉编译

- Linux CGO=0 即可（运行时 dlopen）。但 **macOS 必须 cgo + 在 Mac 上原生编**，
  从 Linux `GOOS=darwin` 交叉编译会报 `undefined: nativeApplication`。CI 用 macos runner 解决。

---

## 6. 给后续迁移的 checklist

- [ ] `viva doctor` 过；确认目标机是 X11（或接受 XWayland）
- [ ] go.mod 升 1.26；移除 fyne；`go mod tidy`
- [ ] 把可变状态收进 model 结构体，UI 改为 `Paint(model)→界面`
- [ ] 后台→UI 一律走 `Invalidator.Invalidate()` + `OnWake`，删掉所有 `fyne.Do`
- [ ] 多行输入用 `editbox.go` 这套（text.Editor + TextareaLayout + TextInput provider）
- [ ] 单行若要 Ctrl+A，用自管 editor + f.Input，别用 flex.Field
- [ ] 审查每个 Paint：无日志/文件/网络/goroutine/每帧大分配
- [ ] 窗口居中自己用 xrandr/xdotool 算 Origin（没有现成 API）
- [ ] 快捷键避开"多行框 + Enter"；触发动作用按钮
- [ ] 构建脚本设 GOTOOLCHAIN=auto + CGO_ENABLED=0（Linux/Win）；macOS 原生编
- [ ] 打包 Depends 补 libvulkan1 / libxcb1 / libpangocairo-1.0-0
- [ ] 业务层（翻译/IPC/OCR/配置等与 GUI 无关的）保持不动，这是迁移成本可控的根本
- [ ] 加 headless 测试：`app.NewDriver(config, canvas)` 无窗口驱动，验证渲染/输入/快捷键

---

## 7. 参考

- Viva 仓库：https://github.com/vivaui/viva
- Viva 文档：https://vivaui.dev/
- 每个项目生成的 `AGENTS.md`（`viva init` 产出）里有当前版本的编程模型速查，迁移前通读一遍

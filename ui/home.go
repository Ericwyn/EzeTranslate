package ui

import (
	"runtime"

	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/spf13/viper"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// homeView 持有完整模式窗口每帧复用的 flex 视图。
type homeView struct {
	view flex.View
}

var home = &homeView{}

// homeConfig 构造完整模式窗口的 Viva 配置。抽出来是为了 openHomeWindow
// 与 headless 测试都能复用同一套 Paint/TextInput/事件接线。
func (a *ezeApp) homeConfig(theme ui.Theme) app.Config {
	config := app.Config{
		Tree:        &tree,
		Title:       "EzeTranslate",
		Icon:        iconOrEmpty(),
		Width:       450,
		Height:      600,
		Origin:      centeredOrigin(450, 600),
		ClearColor:  clearColor(theme),
		Invalidator: a.trans.homeInv,
		TextInput: func(ctx app.TextInputContext) *app.TextInput {
			var holder app.TextInput
			return a.trans.input.textInput(ctx, &holder)
		},
		OnWake:     a.drainHome,
		OnKey:      a.homeOnKey,
		OnShortcut: a.homeOnShortcut,
		OnScroll: func(e ui.ScrollEvent) ui.EventResult {
			// 结果框通常更长、更可能溢出，滚轮优先作用于它；
			// 输入框较短，可用方向键按视觉行移动光标。
			return a.trans.result.handleScroll(e)
		},
	}
	config.Paint = func(canvas *ui.Canvas, size ui.Size) {
		a.paintHome(canvas, size, theme)
	}
	return config
}

// openHomeWindow 打开完整模式窗口。
func (a *ezeApp) openHomeWindow() error {
	theme := ui.LightTheme()

	config := a.homeConfig(theme)
	config.OnClosed = func() {
		if a.switching {
			a.switching = false
			return
		}
		a.application.Quit()
	}

	win, err := a.application.OpenWindow(config, new(ui.Canvas))
	if err != nil {
		return err
	}
	a.homeWin = win
	return nil
}

// homeOnKey 处理完整模式下的按键：Ctrl/Alt+Enter 触发翻译，其余交给输入框。
func (a *ezeApp) homeOnKey(e ui.KeyEvent) ui.EventResult {
	if e.Key == ui.KeyEnter && (e.Modifiers&ui.ModifierControl != 0 || e.Modifiers&ui.ModifierAlt != 0) {
		a.startTrans()
		return ui.EventResult{Handled: true, Repaint: true}
	}
	return a.trans.input.handleKey(e)
}

// homeOnShortcut 处理 Ctrl/Meta 修饰的可打印快捷键：
// Ctrl+A 在输入框聚焦时全选输入框内容。
func (a *ezeApp) homeOnShortcut(e ui.KeyEvent) ui.EventResult {
	if (e.Text == "a" || e.Text == "A") && (e.Modifiers&ui.ModifierControl != 0 || e.Modifiers&ui.ModifierMeta != 0) {
		if e.Target == homeInputID {
			a.trans.input.selectAll()
			return ui.EventResult{Handled: true, Repaint: true}
		}
	}
	return ui.EventResult{}
}

// paintHome 绘制完整模式窗口。
func (a *ezeApp) paintHome(canvas *ui.Canvas, size ui.Size, theme ui.Theme) {
	bounds := ui.Rect{Width: size.Width, Height: size.Height}
	canvas.Fill(bounds, theme.Background)

	transLabel := "执行翻译"
	if a.trans.isLoading() {
		transLabel = "翻译中…"
	}

	home.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
		f.Column().Fill().Pad(14).Gap(10).Children(func() {
			// 顶部工具行：设置 / 日志 / 关于
			f.Row().Gap(8).Items(flex.Center).Children(func() {
				if f.Button(homeSetBtnID, "参数设置").Outline().Small().Clicked() {
					a.openSetWindow()
				}
				if f.Button(homeLogBtnID, "运行日志").Outline().Small().Clicked() {
					a.openLogWindow()
				}
				if f.Button(homeAboutBtnID, "关于").Outline().Small().Clicked() {
					a.openAboutWindow()
				}
				f.Spacer()
			})

			// 输入优化勾选
			a.buildFormatChecks(f)

			// 输入框
			a.trans.input.paint(f, canvas, "输入文本", "请输入需要翻译的文字", 150).FillWidth()

			// 翻译结果
			a.buildEngineSelect(f)
			a.trans.result.paint(f, canvas, "翻译结果", "等待翻译中...", 150).FillWidth().Grow(1)

			if note := a.trans.getNote(); note != "" {
				f.Text(note).FontSize(12).Muted()
			}

			// 底部操作行
			f.Row().Gap(8).Items(flex.Center).Children(func() {
				if f.Button(homeTransBtnID, transLabel).Clicked() && !a.trans.isLoading() {
					a.startTrans()
				}
				if runtime.GOOS == "linux" {
					if f.Button(homeMiniBtnID, "迷你模式").Outline().Clicked() {
						a.switchToMini()
					}
					if f.Button(homeOcrBtnID, "OCR 翻译").Outline().Clicked() {
						go a.runOcr(true)
					}
				}
				f.Spacer()
				a.buildLangSelect(f)
			})
		})
	})
}

// switchToMini 从完整模式切到迷你模式，保留当前翻译结果。
func (a *ezeApp) switchToMini() {
	a.trans.selectNow = a.trans.input.Text()
	viper.Set(conf.ConfigKeyMiniMode, true)
	conf.SaveConfig()

	if err := a.openMiniWindow(); err != nil {
		return
	}
	// 关闭 home 窗口但不退出应用
	a.switching = true
	if a.homeWin != nil {
		a.homeWin.Close()
		a.homeWin = nil
	}
}

func iconOrEmpty() app.WindowIcon {
	return appIcon
}

func clearColor(theme ui.Theme) app.Color {
	return app.Color{
		Red:   float64(theme.Background.Red),
		Green: float64(theme.Background.Green),
		Blue:  float64(theme.Background.Blue),
		Alpha: 1,
	}
}

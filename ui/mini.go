package ui

import (
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/spf13/viper"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// miniView 持有迷你模式窗口每帧复用的 flex 视图。
type miniView struct {
	view flex.View
}

var mini = &miniView{}

// miniConfig 与 headless 测试共享迷你模式的事件接线。
func (a *ezeApp) miniConfig(theme ui.Theme) app.Config {
	config := app.Config{
		Tree:        &tree,
		Title:       "EzeTranslate",
		Icon:        appIcon,
		Width:       400,
		Height:      300,
		Origin:      centeredOrigin(400, 300),
		ClearColor:  clearColor(theme),
		Invalidator: a.trans.miniInv,
		OnWake:      a.drainMini,
		OnPointer:   a.trans.result.handlePointer,
		OnKey:       a.trans.result.handleKey,
		OnShortcut:  a.trans.result.handleShortcut,
		OnScroll: func(e ui.ScrollEvent) ui.EventResult {
			return a.trans.result.handleScroll(e)
		},
	}

	config.Paint = func(canvas *ui.Canvas, size ui.Size) {
		a.paintMini(canvas, size, theme)
	}

	return config
}

// openMiniWindow 打开迷你模式窗口（无输入框，只翻译划词/结果）。
func (a *ezeApp) openMiniWindow() error {
	config := a.miniConfig(ui.LightTheme())
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
	a.miniWin = win
	return nil
}

// paintMini 绘制迷你模式窗口。
func (a *ezeApp) paintMini(canvas *ui.Canvas, size ui.Size, theme ui.Theme) {
	bounds := ui.Rect{Width: size.Width, Height: size.Height}
	canvas.Fill(bounds, theme.Background)

	transLabel := "翻译选中文字"
	if a.trans.isLoading() {
		transLabel = "翻译中…"
	}

	mini.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
		f.Column().Fill().Pad(12).Gap(8).Children(func() {
			a.buildFormatChecks(f)
			a.buildEngineSelect(f)

			f.Text("翻译结果").Muted()
			a.trans.result.paint(f, canvas, "翻译结果", "等待翻译中...", 120).FillWidth().Grow(1)

			if note := a.trans.getNote(); note != "" {
				f.Text(note).FontSize(12).Muted()
			}

			f.Row().Gap(8).Items(flex.Center).Children(func() {
				if f.Button(miniTransBtnID, transLabel).Clicked() && !a.trans.isLoading() {
					a.startTrans()
				}
				if f.Button(miniFullBtnID, "完整模式").Outline().Clicked() {
					a.switchToHome()
				}
				if f.Button(miniOcrBtnID, "OCR").Outline().Clicked() {
					go a.runOcr(true)
				}
				f.Spacer()
				a.buildLangSelect(f)
			})
		})
	})
}

// switchToHome 从迷你模式切回完整模式，保留结果并回填输入框。
func (a *ezeApp) switchToHome() {
	viper.Set(conf.ConfigKeyMiniMode, false)
	conf.SaveConfig()

	if err := a.openHomeWindow(); err != nil {
		return
	}
	// 回填上次划词/OCR 文本到输入框
	a.trans.input.SetText(a.trans.selectNow)

	a.switching = true
	if a.miniWin != nil {
		a.miniWin.Close()
		a.miniWin = nil
	}
}

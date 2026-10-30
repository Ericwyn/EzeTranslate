package ui

import (
	"github.com/Ericwyn/EzeTranslate/log"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// logState 持有日志窗口状态。
type logState struct {
	view flex.View
	inv  *app.Invalidator
	text string
}

var logS *logState

var logScrollID = tree.ID(logFeat, "scroll")

// openLogWindow 打开运行日志窗口；已打开则前置。
func (a *ezeApp) openLogWindow() {
	if a.logWin != nil && !a.logWin.Closed() {
		a.logWin.Show()
		return
	}

	logS = &logState{inv: &app.Invalidator{}, text: log.GetLog1000()}

	theme := ui.LightTheme()
	config := app.Config{
		Tree:        &tree,
		Title:       "运行日志",
		Icon:        appIcon,
		Width:       600,
		Height:      600,
		Origin:      centeredOrigin(600, 600),
		ClearColor:  clearColor(theme),
		Invalidator: logS.inv,
		OnScroll: func(e ui.ScrollEvent) ui.EventResult {
			return logS.view.Wheel(e)
		},
	}
	config.Paint = func(canvas *ui.Canvas, size ui.Size) {
		a.paintLog(canvas, size, theme)
	}
	config.OnClosed = func() {
		a.logWin = nil
	}

	win, err := a.application.OpenWindow(config, new(ui.Canvas))
	if err != nil {
		log.E("打开日志窗口失败: ", err)
		return
	}
	a.logWin = win
}

func (a *ezeApp) paintLog(canvas *ui.Canvas, size ui.Size, theme ui.Theme) {
	bounds := ui.Rect{Width: size.Width, Height: size.Height}
	canvas.Fill(bounds, theme.Background)

	logS.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
		f.Column().Fill().Pad(14).Gap(10).Children(func() {
			f.Row().Gap(8).Items(flex.Center).Children(func() {
				f.Text("运行日志").FontSize(20).Bold()
				f.Spacer()
				if f.Button(logRefreshID, "刷新日志").Outline().Small().Clicked() {
					logS.text = log.GetLog1000()
				}
				if f.Button(logClearID, "清除日志").Outline().Small().Clicked() {
					log.ClearLogBuff()
					logS.text = ""
				}
			})

			f.ScrollArea(logScrollID).Grow(1).Children(func() {
				content := logS.text
				if content == "" {
					content = "暂无日志信息"
				}
				f.Wrap(content).FontSize(12).Mono()
			})
		})
	})
}

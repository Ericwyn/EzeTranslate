package ui

import (
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// aboutState 持有关于窗口状态。
type aboutState struct {
	view flex.View
	inv  *app.Invalidator
}

var about = &aboutState{}

// openAboutWindow 打开关于窗口；已打开则前置。
func (a *ezeApp) openAboutWindow() {
	if a.aboutWin != nil && !a.aboutWin.Closed() {
		a.aboutWin.Show()
		return
	}

	about = &aboutState{inv: &app.Invalidator{}}

	theme := ui.LightTheme()
	config := app.Config{
		Tree:        &tree,
		Title:       "关于",
		Icon:        appIcon,
		Width:       320,
		Height:      300,
		Origin:      centeredOrigin(320, 300),
		ClearColor:  clearColor(theme),
		Invalidator: about.inv,
	}
	config.Paint = func(canvas *ui.Canvas, size ui.Size) {
		a.paintAbout(canvas, size, theme)
	}
	config.OnClosed = func() {
		a.aboutWin = nil
	}

	win, err := a.application.OpenWindow(config, new(ui.Canvas))
	if err != nil {
		log.E("打开关于窗口失败: ", err)
		return
	}
	a.aboutWin = win
}

func (a *ezeApp) paintAbout(canvas *ui.Canvas, size ui.Size, theme ui.Theme) {
	bounds := ui.Rect{Width: size.Width, Height: size.Height}
	canvas.Fill(bounds, theme.Background)

	about.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
		f.Column().Fill().Pad(20).Gap(10).Items(flex.Center).Children(func() {
			f.Text("EzeTranslate").FontSize(22).Bold()
			f.Text("当前版本: " + conf.Version + "-" + conf.ReleaseDate)
			f.Text("软件作者: github.com/Ericwyn").Muted()

			f.Spacer()

			f.Row().Gap(8).Justify(flex.Center).Children(func() {
				if f.Button(aboutGithubID, "Github").Outline().Clicked() {
					openURL("https://github.com/Ericwyn/EzeTranslate")
				}
				if f.Button(aboutIssueID, "问题反馈").Outline().Clicked() {
					openURL("https://github.com/Ericwyn/EzeTranslate/issues")
				}
			})
			if f.Button(aboutReleaseID, "新版下载").Clicked() {
				openURL("https://github.com/Ericwyn/EzeTranslate/releases")
			}
		})
	})
}

func openURL(target string) {
	if err := app.OpenURL(target); err != nil {
		log.E("打开链接失败: ", err)
	}
}

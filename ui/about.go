package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/ui/resource"
	"net/url"
)

var aboutWindow fyne.Window
var aboutWindowsOpening = false

func showAboutUi() {
	if aboutWindowsOpening {
		aboutWindow.RequestFocus()
		return
	}

	aboutWindow = mainApp.NewWindow("关于")

	aboutWindow.Resize(fyne.Size{
		Width: 200,
		//Height: 200,
	})
	aboutWindow.CenterOnScreen()

	icon := canvas.NewImageFromResource(resource.ResourceIcon())
	icon.SetMinSize(fyne.Size{
		Width:  60,
		Height: 60,
	})

	githubBtn := widget.NewButton("Github", func() {
		u, _ := url.Parse("https://github.com/Ericwyn/EzeTranslate")
		_ = mainApp.OpenURL(u)
	})

	issueBtn := widget.NewButton("问题反馈", func() {
		u, _ := url.Parse("https://github.com/Ericwyn/EzeTranslate/issues")
		_ = mainApp.OpenURL(u)
	})

	releaseBtn := widget.NewButton("          新版下载          ", func() {
		u, _ := url.Parse("https://github.com/Ericwyn/EzeTranslate/releases")
		_ = mainApp.OpenURL(u)
	})

	aboutWindow.SetContent(container.NewVBox(
		container.NewHBox(widget.NewLabel("")),
		container.NewCenter(icon),
		container.NewHBox(widget.NewLabel("")),
		container.NewHBox(widget.NewLabel("当前版本:"), widget.NewLabel(conf.Version+"-"+conf.ReleaseDate)),
		container.NewHBox(widget.NewLabel("软件作者:"), widget.NewLabel("github.com/Ericwyn")),
		container.NewHBox(widget.NewLabel("")),
		container.NewCenter(container.NewHBox(githubBtn, issueBtn)),
		container.NewCenter(releaseBtn),
		container.NewHBox(widget.NewLabel("")),
	))

	aboutWindow.SetOnClosed(func() {
		aboutWindowsOpening = false
	})

	aboutWindowsOpening = true
	aboutWindow.Show()
}

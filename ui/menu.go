package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"net/url"
)

var logWindow fyne.Window
var logEntryBox *widget.Entry
var logWindowsOpening = false

/*
菜单结构是这样的

- MainMenu
	Menu1
		MenuItem-1
		MenuItem-2
		MenuItem-3
	Menu2
	Menu3

在 Menu1 中, 如果又没有个 MenuItem 的 IsQuit 为 True 的话, 那么会加一个 Quit 的 MenuItem

*/
func createAppMenu() *fyne.MainMenu {

	openLogMenuItem := fyne.NewMenuItem("运行日志", func() {
		showLogUi()
	})
	openLogMenuItem.IsQuit = true

	logMenu := fyne.NewMenu("日志", openLogMenuItem)

	mainMenu := fyne.NewMainMenu(
		logMenu,
		fyne.NewMenu("关于", fyne.NewMenuItem("Github ", func() {
			u, _ := url.Parse("https://github.com/Ericwyn/EzeTranslate")
			_ = mainApp.OpenURL(u)
		})),
	)

	return mainMenu
}

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Ericwyn/TransUtils/log"
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

func showLogUi() {

	if logWindowsOpening {
		return
	}

	logWindow = mainApp.NewWindow("运行日志")
	logWindow.Resize(fyne.Size{
		Width:  600,
		Height: 600,
	})

	logEntryBox = widget.NewMultiLineEntry()

	logEntryBox.SetPlaceHolder(`暂无日志信息`)

	bottomPanel := container.NewHBox(
		widget.NewButton("刷新日志", func() {
			logEntryBox.SetText(log.GetLog1000())
		}),
		widget.NewButton("清除日志", func() {
			log.ClearLogBuff()
			logEntryBox.SetText("")
		}),
	)

	logPanel := container.NewBorder(nil, bottomPanel, nil, nil,
		container.NewGridWithColumns(1, logEntryBox))
	logWindow.SetContent(logPanel)
	logEntryBox.SetText(log.GetLog1000())

	logWindow.SetOnClosed(func() {
		logWindowsOpening = false
	})

	logWindowsOpening = true
	logWindow.Show()
}

package ui

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Ericwyn/EzeTranslate/log"
)

var logWindow fyne.Window
var logEntryBox *widget.Entry
var logWindowsOpening = false

func showLogUi() {

	if logWindowsOpening {
		logWindow.RequestFocus()
		return
	}

	logWindow = mainApp.NewWindow("运行日志")

	logWindow.Resize(fyne.Size{
		Width:  600,
		Height: 600,
	})
	logWindow.CenterOnScreen()

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

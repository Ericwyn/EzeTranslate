package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/spf13/viper"
	"os"
	"runtime"
)

// 翻译入口页面

var homeWindow fyne.Window

var homeInputBoxPanel *fyne.Container
var homeInputBox *widget.Entry

var homeTransResBoxPanel *fyne.Container
var homeTransResBox *widget.Entry
var homeNoteLabel *widget.Label

func showHomeUi(showAndRun bool) {

	homeWindow = mainApp.NewWindow("EzeTranslate")

	homeWindow.SetMainMenu(createAppMenu())

	homeWindow.Resize(fyne.Size{
		Width:  400,
		Height: 600,
	})
	homeWindow.CenterOnScreen()

	inputBoxPanelTitle := buildFormatCheckBox()

	homeInputBox = widget.NewMultiLineEntry()
	homeInputBox.SetPlaceHolder(`请输入需要翻译的文字`)
	homeInputBox.Wrapping = fyne.TextWrapBreak

	homeInputBoxPanel = container.NewBorder(inputBoxPanelTitle, nil, nil, nil,
		container.NewGridWithColumns(1, homeInputBox))

	transResBoxPanelTitle := buildTransApiCheckBox()

	homeTransResBox = widget.NewMultiLineEntry()
	homeTransResBox.SetPlaceHolder(`等待翻译中...`)
	homeTransResBox.Wrapping = fyne.TextWrapBreak

	homeTransResBoxPanel = container.NewBorder(transResBoxPanelTitle, nil, nil, nil,
		container.NewGridWithColumns(1, homeTransResBox))

	homeNoteLabel = widget.NewLabel("")

	bottomPanel := container.NewHBox(
		widget.NewButton("翻译当前文字", func() {
			startTrans()
		}),
	)
	if runtime.GOOS == "linux" {
		bottomPanel.Add(
			widget.NewButton("迷你模式", func() {
				// 断开 homeUi 的 Closed 回调, 不关闭 app
				homeWindow.SetOnClosed(func() {})

				resBoxText := homeTransResBox.Text
				noteText := homeNoteLabel.Text
				// 先展示，再关闭
				showMiniUi(false)
				closeHomeUi()
				miniTransResBox.SetText(resBoxText)
				miniNoteLabel.SetText(noteText)

				viper.Set(conf.ConfigKeyMiniMode, true)
				conf.SaveConfig()
			}),
		)
	}
	bottomPanel.Add(homeNoteLabel)

	mainPanel := container.NewBorder(nil, bottomPanel, nil, nil,
		container.NewGridWithColumns(1, homeInputBoxPanel, homeTransResBoxPanel))

	homeWindow.SetContent(mainPanel)

	homeWindow.SetOnClosed(func() {
		os.Exit(0)
	})

	if showAndRun {
		homeWindow.ShowAndRun()
	} else {
		homeWindow.Show()
	}
}

func closeHomeUi() {
	homeWindow.Close()

	homeWindow = nil
	homeInputBoxPanel = nil
	homeInputBox = nil
	homeTransResBoxPanel = nil
	homeTransResBox = nil
	homeNoteLabel = nil
}

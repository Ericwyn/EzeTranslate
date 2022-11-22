package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/ipc"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/spf13/viper"
	"os"
	"runtime"
)

// 翻译入口页面

var homeWindow fyne.Window

var homeInputBoxPanel *fyne.Container
var homeInputBox *EzeInputEntry

var homeTransResBoxPanel *fyne.Container
var homeTransResBox *widget.Entry
var homeNoteLabel *widget.Label

type EzeInputEntry struct {
	widget.Entry
}

func (m *EzeInputEntry) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*desktop.CustomShortcut); !ok {
		m.Entry.TypedShortcut(s)
		return
	}
	shortcut := s.(*desktop.CustomShortcut)
	// 如果是 alt + 回车 / ctrl + 回车，就直接触发翻译
	if shortcut.KeyName == fyne.KeyReturn &&
		(shortcut.Modifier == fyne.KeyModifierControl || shortcut.Modifier == fyne.KeyModifierAlt) {
		startTrans()
	}
}

func showHomeUi(showAndRun bool) {

	homeWindow = mainApp.NewWindow("EzeTranslate")

	homeWindow.SetMainMenu(createAppMenu())

	homeWindow.Resize(fyne.Size{
		Width:  400,
		Height: 600,
	})
	homeWindow.CenterOnScreen()

	inputBoxPanelTitle := buildFormatCheckBox()

	homeInputBox = &EzeInputEntry{}
	homeInputBox.MultiLine = true
	homeInputBox.Wrapping = fyne.TextTruncate
	homeInputBox.ExtendBaseWidget(homeInputBox)
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
		bottomPanel.Add(
			widget.NewButton("OCR 翻译", func() {
				trySendMessage(ipc.IpcMessageOcrAndTrans)
				//log.D("ocr 识别成功")
				//ocrRes, successFlag := ocr.RunOcr()
				//if successFlag {
				//	homeInputBox.SetText(ocrRes)
				//	startTrans()
				//}
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

	altTab := desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierAlt}
	homeWindow.Canvas().AddShortcut(&altTab, func(shortcut fyne.Shortcut) {
		log.I("alt + tab")
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

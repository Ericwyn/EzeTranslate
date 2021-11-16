package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/resource/cusWidget"
	"github.com/spf13/viper"
	"os"
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

	inputBoxPanelTitle := container.NewHBox(
		widget.NewLabel("翻译设置        "),
		cusWidget.CreateCheckGroup(
			[]cusWidget.LabelAndInit{
				{"注释优化", viper.GetBool(conf.ConfigKeyFormatAnnotation)},
				{"空格优化", viper.GetBool(conf.ConfigKeyFormatSpace)},
				{"回车优化", viper.GetBool(conf.ConfigKeyFormatCarriageReturn)},
			},
			true,  // 横向
			false, // 单选
			func(label string, checked bool) {
				if label == "注释优化" {
					viper.Set(conf.ConfigKeyFormatAnnotation, checked)
				} else if label == "空格优化" {
					viper.Set(conf.ConfigKeyFormatSpace, checked)
				} else if label == "回车优化" {
					viper.Set(conf.ConfigKeyFormatCarriageReturn, checked)
				}
				e := viper.WriteConfig()
				if e != nil {
					log.E("配置文件保存失败")
					log.E(e)
				}
			},
		),
	)

	homeInputBox = widget.NewMultiLineEntry()
	homeInputBox.SetPlaceHolder(`请输入需要翻译的文字`)
	homeInputBox.Wrapping = fyne.TextWrapBreak

	homeInputBoxPanel = container.NewBorder(inputBoxPanelTitle, nil, nil, nil,
		container.NewGridWithColumns(1, homeInputBox))

	transResBoxPanelTitle := container.NewHBox(
		widget.NewLabel("翻译结果        "),
		cusWidget.CreateCheckGroup(
			[]cusWidget.LabelAndInit{
				{"Google", viper.GetString(conf.ConfigKeyTranslateSelect) == "google"},
				{"Baidu", viper.GetString(conf.ConfigKeyTranslateSelect) == "baidu"},
				{"Youdao", viper.GetString(conf.ConfigKeyTranslateSelect) == "youdao"},
			},
			true, // 横向
			true, // 单选
			func(label string, checked bool) {
				if label == "Google" {
					viper.Set(conf.ConfigKeyTranslateSelect, "google")
				} else if label == "Baidu" {
					viper.Set(conf.ConfigKeyTranslateSelect, "baidu")
				} else if label == "Youdao" {
					viper.Set(conf.ConfigKeyTranslateSelect, "youdao")
				}
				e := viper.WriteConfig()
				if e != nil {
					log.E("配置文件保存失败")
					log.E(e)
				}
			},
		),
	)

	homeTransResBox = widget.NewMultiLineEntry()
	homeTransResBox.SetPlaceHolder(`等待翻译中...`)
	homeTransResBox.Wrapping = fyne.TextWrapBreak

	homeTransResBoxPanel = container.NewBorder(transResBoxPanelTitle, nil, nil, nil,
		container.NewGridWithColumns(1, homeTransResBox))

	homeNoteLabel = widget.NewLabel("")

	bottomPanel := container.NewHBox(
		container.NewHBox(
			widget.NewButton("翻译当前文字", func() {
				startTrans()
			}),
		),
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
		}),
		homeNoteLabel,
	)

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

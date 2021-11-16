package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/resource/cusWidget"
	"github.com/spf13/viper"
	"os"

	"github.com/Ericwyn/EzeTranslate/log"
)

// 迷你翻译页面
// 对比起 home ui, 去除了输入窗口

var miniWindow fyne.Window
var miniTransResBoxPanel *fyne.Container
var miniTransResBox *widget.Entry
var miniNoteLabel *widget.Label

var miniSelectTextNow = ""

func showMiniUi(showAndRun bool) {

	miniWindow = mainApp.NewWindow("EzeTranslate")

	miniWindow.Resize(fyne.Size{
		Width:  400,
		Height: 300,
	})
	miniWindow.CenterOnScreen()

	miniInputBoxPanelTitle := container.NewHBox(
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

	miniTransResBoxPanelTitle := container.NewHBox(
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

	miniTransResBox = widget.NewMultiLineEntry()
	miniTransResBox.SetPlaceHolder(`等待翻译中...`)
	miniTransResBox.Wrapping = fyne.TextWrapBreak

	miniTransResBoxPanel = container.NewBorder(
		container.NewVBox(miniInputBoxPanelTitle, miniTransResBoxPanelTitle),
		nil, nil, nil,
		container.NewGridWithColumns(1, miniTransResBox))

	miniNoteLabel = widget.NewLabel("")

	miniBottomPanel := container.NewHBox(
		widget.NewButton("翻译选中文字", func() {
			startTrans()
		}),
		widget.NewButton("完整模式", func() {
			// 断开 homeUi 的 Closed 回调, 不关闭 app
			miniWindow.SetOnClosed(func() {})

			resBoxText := miniTransResBox.Text
			noteText := miniNoteLabel.Text
			showHomeUi(false)
			closeMiniUi()
			homeTransResBox.SetText(resBoxText)
			homeNoteLabel.SetText(noteText)
			homeInputBox.SetText(miniSelectTextNow)
			conf.SaveConfig()
		}),
		miniNoteLabel,
	)

	miniPanel := container.NewBorder(nil, miniBottomPanel, nil, nil,
		container.NewGridWithColumns(1, miniTransResBoxPanel))

	miniWindow.SetContent(miniPanel)

	miniWindow.SetOnClosed(func() {
		os.Exit(0)
	})

	if showAndRun {
		miniWindow.ShowAndRun()
	} else {
		miniWindow.Show()
	}
}

func closeMiniUi() {
	miniWindow.Close()

	miniWindow = nil
	miniTransResBoxPanel = nil
	miniTransResBox = nil
	miniNoteLabel = nil
}

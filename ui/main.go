package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/resource/cusWidget"
	"github.com/Ericwyn/EzeTranslate/strutils"
	"github.com/Ericwyn/EzeTranslate/trans/baidu"
	"github.com/Ericwyn/EzeTranslate/trans/google"
	"github.com/Ericwyn/EzeTranslate/trans/youdao"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"

	//"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	//"github.com/Ericwyn/GoTools/file"
	"github.com/Ericwyn/EzeTranslate/ipc"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/resource"
	"github.com/Ericwyn/EzeTranslate/trans"
	"github.com/Ericwyn/GoTools/shell"
	//"os"
	"runtime"
)

func StartApp(xclip bool) {

	shell.Debug(true)

	conf.InitConfig()

	if trySendMessage() {
		// 如果已经有其他翻译进程的话, 就直接退出
		return
	}

	// 开启 server 监听来自其他进程的翻译请求
	startUnixSocketServer()

	// 如果启动时候带有参数的话，那么就应该直接获取一遍选中的文字并进行翻译
	if xclip {
		go func() {
			time.Sleep(time.Millisecond * 500)
			trySendMessage()
		}()
	}

	showMainUi()
}

var mainApp fyne.App
var mainWindow fyne.Window

var inputBoxPanel *fyne.Container
var inputBox *widget.Entry

var transResBoxPanel *fyne.Container
var transResBox *widget.Entry
var noteLabel *widget.Label

func showMainUi() {

	mainApp = app.New()
	mainApp.SetIcon(resource.ResourceIcon)
	mainApp.Settings().SetTheme(&resource.CustomerTheme{})

	mainWindow = mainApp.NewWindow("EzeTranslate")

	mainWindow.SetMainMenu(createAppMenu())

	mainWindow.Resize(fyne.Size{
		Width:  400,
		Height: 600,
	})
	mainWindow.CenterOnScreen()

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

	inputBox = widget.NewMultiLineEntry()
	inputBox.SetPlaceHolder(`请输入需要翻译的文字`)
	inputBox.Wrapping = fyne.TextWrapBreak

	inputBoxPanel = container.NewBorder(inputBoxPanelTitle, nil, nil, nil,
		container.NewGridWithColumns(1, inputBox))

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

	transResBox = widget.NewMultiLineEntry()
	transResBox.SetPlaceHolder(`等待翻译中...`)
	transResBox.Wrapping = fyne.TextWrapBreak

	transResBoxPanel = container.NewBorder(transResBoxPanelTitle, nil, nil, nil,
		container.NewGridWithColumns(1, transResBox))

	noteLabel = widget.NewLabel("")

	bottomPanel := container.NewHBox(
		container.NewHBox(
			widget.NewButton("翻译当前文字", func() {
				startTrans()
			}),
		),
		noteLabel,
	)

	if runtime.GOOS == "linux" {
		bottomPanel.Add(
			widget.NewButton("获取选词", func() {
				text := trans.GetSelection()
				inputBox.SetText(text)
			}),
		)
	}

	mainPanel := container.NewBorder(nil, bottomPanel, nil, nil,
		container.NewGridWithColumns(1, inputBoxPanel, transResBoxPanel))

	mainWindow.SetContent(mainPanel)

	mainWindow.SetOnClosed(func() {
		os.Exit(0)
	})

	mainWindow.ShowAndRun()
}

func startTrans() {
	formatText := strutils.FormatInputBoxText(inputBox.Text)

	if strings.Trim(formatText, " ") == "" {
		transResBox.SetPlaceHolder("请输入需要翻译的内容")
		return
	}

	transResBox.SetText("")
	transResBox.SetPlaceHolder("正在翻译..........")

	handleTransResult := func(result string, note string) {
		fmt.Println("翻译结果:", result)
		transResBox.SetText(result)
		noteLabel.SetText(note)
	}

	if viper.GetString(conf.ConfigKeyTranslateSelect) == "google" {
		go google.Translate(formatText, handleTransResult)
	} else if viper.GetString(conf.ConfigKeyTranslateSelect) == "baidu" {
		go baidu.Translate(formatText, handleTransResult)
	} else if viper.GetString(conf.ConfigKeyTranslateSelect) == "youdao" {
		go youdao.Translate(formatText, handleTransResult)
	}

}

func trySendMessage() bool {
	if runtime.GOOS != "linux" {
		log.D("not linux, don't send socket msg")
		return false
	}
	err := ipc.SendMessage(ipc.MessageNewSelection)
	if err == nil {
		fmt.Println("已发送给其他翻译进程")
		return true
	}
	return false
}

// 开启一个 UnixSocketServer, 接收 IPC 消息
func startUnixSocketServer() {
	if runtime.GOOS != "linux" {
		log.D("not linux, don't start socket server")
		return
	}
	go ipc.StartUnixSocketListener(func(message string) {
		log.D("接收到 IPC 消息")
		switch message {
		case ipc.MessageNewSelection:

			// 请求焦点
			mainWindow.RequestFocus()

			selectText := trans.GetSelection()
			fmt.Println("获取的划词:", selectText)

			if strings.Trim(inputBox.Text, " ") ==
				strings.Trim(selectText, " ") {

				// 如果翻译框有数据的话，就不进行翻译
				if strings.Trim(transResBox.Text, " ") != "" {
					log.D("获取的划词与当前 inputBox 中文字一致，不进行翻译")
					return
				}

			}

			// 刷新当前数据
			inputBox.SetText(selectText)
			startTrans()

			break
		}
	})
}

package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/resource/cusWeight"
	"github.com/Ericwyn/EzeTranslate/strutils"
	"github.com/spf13/viper"
	"os"
	"strings"

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

func StartApp() {

	shell.Debug(true)

	conf.InitConfig()

	if trySendMessage() {
		// 如果已经有其他翻译进程的话, 就直接退出
		return
	}

	// 开启 server 监听来自其他进程的翻译请求
	startUnixSocketServer()

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

	mainWindow = mainApp.NewWindow("划词翻译")

	mainWindow.SetMainMenu(createAppMenu())

	mainWindow.Resize(fyne.Size{
		Width:  400,
		Height: 600,
	})

	inputBoxPanelTitle := container.NewHBox(
		widget.NewLabel("翻译设置        "),
		cusWeight.CreateCheckGroup(
			[]cusWeight.LabelAndInit{
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

	inputBoxPanel = container.NewBorder(inputBoxPanelTitle, nil, nil, nil,
		container.NewGridWithColumns(1, inputBox))

	transResBox = widget.NewMultiLineEntry()
	transResBox.SetPlaceHolder(`等待翻译中...`)
	//transResBox.Enable()
	transResBox.Wrapping = fyne.TextWrapBreak

	transResBoxPanel = container.NewBorder(widget.NewLabel("翻译结果"), nil, nil, nil,
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

	transResBox.SetPlaceHolder("正在翻译..........")
	//inputBox.SetText(formatText)
	go trans.BaiduTrans(formatText, func(result string, note string) {
		fmt.Println("翻译结果:", result)
		transResBox.SetText(result)
		noteLabel.SetText(note)
	})
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
			// 刷新当前数据
			inputBox.SetText(selectText)

			startTrans()

			break
		}
	})
}

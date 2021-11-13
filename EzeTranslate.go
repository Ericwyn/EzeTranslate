package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/Ericwyn/TransUtils/conf"
	"github.com/spf13/viper"

	//"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	//"github.com/Ericwyn/GoTools/file"
	"github.com/Ericwyn/GoTools/shell"
	"github.com/Ericwyn/TransUtils/ipc"
	"github.com/Ericwyn/TransUtils/log"
	"github.com/Ericwyn/TransUtils/resource"
	"github.com/Ericwyn/TransUtils/trans"
	//"os"
	"runtime"
)

// 请安装

func main() {
	shell.Debug(true)

	conf.InitConfig()

	fmt.Println(conf.ConfigKeyBaiduTransAppId, viper.GetString(conf.ConfigKeyBaiduTransAppId))
	fmt.Println(conf.ConfigKeyBaiduTransAppSecret, viper.GetString(conf.ConfigKeyBaiduTransAppSecret))

	if runtime.GOOS == "linux" {
		err := ipc.SendMessage(ipc.MessageNewSelection)
		if err == nil {
			fmt.Println("发送给 IPC 进程执行")
			return
		}
	}

	showMainUi()
}

var inputBox *widget.Entry
var transResBox *widget.Entry
var noteLabel *widget.Label

func showMainUi() {
	// 中文支持
	// 参考 https://www.wangfeng.me/article/Go-fyne-ui-kuang-jia-she-zhi-zhong-wen-bing-da-bao-dao-er-jin-zhi-wen-jian

	//fontFile := file.OpenFile("./resource/fonts/Alibaba-PuHuiTi-Regular.ttf")
	//if fontFile.IsFile() {
	//	log.D("set FYNE_FONT env:", fontFile.AbsPath())
	//	os.Setenv("FYNE_FONT", fontFile.AbsPath())
	//}

	a := app.New()

	a.SetIcon(resource.ResourceIcon)

	a.Settings().SetTheme(&resource.CustomerTheme{})

	w := a.NewWindow("划词翻译")

	w.Resize(fyne.Size{
		Width:  400,
		Height: 400,
	})

	inputBox := widget.NewMultiLineEntry()
	inputBox.SetPlaceHolder(`请输入需要翻译的文字`)
	inputBox.Wrapping = fyne.TextWrapWord

	transResBox := widget.NewMultiLineEntry()
	transResBox.SetPlaceHolder(`等待翻译中...`)
	//transResBox.Enable()
	transResBox.Wrapping = fyne.TextWrapWord

	noteLabel := widget.NewLabel("")

	bottomPanel := container.NewHBox(

		container.NewHBox(
			widget.NewButton("翻译当前文字", func() {
				go trans.BaiduTrans(inputBox.Text, func(result string, note string) {
					fmt.Println("翻译结果:", result)
					transResBox.SetText(result)
					noteLabel.SetText(note)
				})
				transResBox.SetPlaceHolder("正在翻译..........")
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

	c := container.NewBorder(nil, bottomPanel, nil, nil,
		container.NewGridWithColumns(1, inputBox, transResBox))

	w.SetContent(c)

	if runtime.GOOS == "linux" {
		go ipc.StartUnixSocketListener(func(message string) {
			log.D("接收到 IPC 消息")
			switch message {
			case ipc.MessageNewSelection:
				selectText := trans.GetSelection()
				fmt.Println("获取的划词:", selectText)
				// 刷新当前数据
				inputBox.SetText(selectText)

				go trans.BaiduTrans(inputBox.Text, func(result string, note string) {
					fmt.Println("翻译结果:", result)
					transResBox.SetText(result)
					noteLabel.SetText(note)
				})

				transResBox.SetPlaceHolder("正在翻译..........")

				break
			}
		})
	}

	w.ShowAndRun()
}

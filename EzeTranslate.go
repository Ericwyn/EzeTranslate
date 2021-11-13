package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/Ericwyn/TransUtils/conf"
	"github.com/Ericwyn/TransUtils/resource/cusWeight"
	"github.com/spf13/viper"
	"strings"

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

// 1. 安装 app
// 并且启动 app

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

type StartTransFunction func()

var startTrans StartTransFunction

func showMainUi() {

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

	transResBox := widget.NewMultiLineEntry()
	transResBox.SetPlaceHolder(`等待翻译中...`)
	//transResBox.Enable()
	transResBox.Wrapping = fyne.TextWrapWord

	noteLabel := widget.NewLabel("")

	topPanel := container.NewVBox(
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

	bottomPanel := container.NewHBox(
		container.NewHBox(
			widget.NewButton("翻译当前文字", func() {
				transResBox.SetPlaceHolder("正在翻译..........")
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

	c := container.NewBorder(topPanel, bottomPanel, nil, nil,
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
				transResBox.SetPlaceHolder("正在翻译..........")

				startTrans()

				break
			}
		})
	}

	// 要在这里面初始化 startTrans 函数, inputBox 和 transResBox 才不会为 nil
	startTrans = func() {
		formatText := formatInputBoxText(inputBox.Text)
		inputBox.SetText(formatText)
		go trans.BaiduTrans(formatText, func(result string, note string) {
			fmt.Println("翻译结果:", result)
			transResBox.SetText(result)
			noteLabel.SetText(note)
		})
	}

	w.ShowAndRun()
}

// 去除 startStr 开头, 如果 startStr 前面有空格, 也会一起去除
func replaceIfStartWith(str string, startStr string) string {
	if strings.HasPrefix(strings.Trim(str, " "), startStr) {
		if strings.HasPrefix(str, "  ") {
			// 此处要去除的是两个空格, 因为去除一个空格的话会影响 ' */' 和 ' *'
			str = strings.Trim(str, "  ")
		}
		return strings.Replace(str, startStr, "", 1)
	}
	return str
}

func formatInputBoxText(formatText string) string {
	if viper.GetBool(conf.ConfigKeyFormatAnnotation) {
		// 去除注释, 因为 //, /**, /*, *, # 这些总是在每一行的最前面, 所以我们只需要去除最前面的就可以了
		newFormatText := ""
		for _, line := range strings.Split(formatText, "\n") {
			// 将 /t 变成四个空格
			line = strings.ReplaceAll(line, "\t", "    ")

			line = replaceIfStartWith(line, "//")
			line = replaceIfStartWith(line, "/**")
			line = replaceIfStartWith(line, "/*")
			// */ 这个可能出现在前面也可能出现在后面, 反正就是直接去掉就是了
			line = strings.Replace(line, "*/", "", -1)
			line = replaceIfStartWith(line, " *")
			line = replaceIfStartWith(line, "*")
			line = replaceIfStartWith(line, "#")

			newFormatText += line + "\n"
		}
		formatText = newFormatText
	}
	if viper.GetBool(conf.ConfigKeyFormatCarriageReturn) {
		// 去除回车, 回车变成空格
		formatText = strings.ReplaceAll(formatText, "\n", " ")
	}
	if viper.GetBool(conf.ConfigKeyFormatSpace) {
		// 去除多余空格
		formatText = strings.ReplaceAll(formatText, "  ", " ")
		formatText = strings.Trim(formatText, " ")
	}

	return formatText
}

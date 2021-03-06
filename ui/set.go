package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/spf13/viper"
)

var setWindow fyne.Window
var setWindowsOpening = false

func showSetUi() {
	if setWindowsOpening {
		setWindow.RequestFocus()
		return
	}

	setWindow = mainApp.NewWindow("程序设置")
	setWindow.Resize(fyne.Size{
		Width: 500,
		//Height: 600,
	})
	setWindow.CenterOnScreen()

	baiduAppIdEntry := widget.NewEntry()
	baiduAppIdEntry.SetPlaceHolder("百度翻译 AppId")
	baiduAppIdEntry.SetText(viper.GetString(conf.ConfigKeyBaiduTransAppId))

	baiduAppSecretEntry := widget.NewEntry()
	baiduAppSecretEntry.SetPlaceHolder("百度翻译 AppSecret")
	baiduAppSecretEntry.SetText(viper.GetString(conf.ConfigKeyBaiduTransAppSecret))

	youdaoAppIdEntry := widget.NewEntry()
	youdaoAppIdEntry.SetPlaceHolder("有道翻译 AppId")
	youdaoAppIdEntry.SetText(viper.GetString(conf.ConfigKeyYouDaoTransAppId))

	youdaoAppSecretEntry := widget.NewEntry()
	youdaoAppSecretEntry.SetPlaceHolder("有道翻译 AppSecret")
	youdaoAppSecretEntry.SetText(viper.GetString(conf.ConfigKeyYouDaoTransAppSecret))

	googleTranslateProxyEntry := widget.NewEntry()
	googleTranslateProxyEntry.SetPlaceHolder("Google 翻译代理地址")
	googleTranslateProxyEntry.SetText(viper.GetString(conf.ConfigKeyGoogleTranslateProxy))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "百度 AppId    ", Widget: baiduAppIdEntry, HintText: "请填写从百度处申请到的 AppId"},
			{Text: "百度 AppSecret", Widget: baiduAppSecretEntry, HintText: "请填写从百度处申请到的 AppSecret"},
			{Text: "", Widget: widget.NewLabel(""), HintText: ""},
			{Text: "有道 AppId    ", Widget: youdaoAppIdEntry, HintText: "请填写从有道处申请到的 AppId"},
			{Text: "有道 AppSecret", Widget: youdaoAppSecretEntry, HintText: "请填写从有道处申请到的 AppSecret"},
			{Text: "", Widget: widget.NewLabel(""), HintText: ""},
			{Text: "Google Host", Widget: googleTranslateProxyEntry, HintText: "google 翻译的代理地址, 填写 translate.xxxxx.xxx 后面部分"},
			{Text: "", Widget: widget.NewLabel(""), HintText: ""},
		},

		SubmitText: "保存设置",
		OnSubmit: func() {
			log.D("保存设置")
			viper.Set(conf.ConfigKeyBaiduTransAppId, baiduAppIdEntry.Text)
			viper.Set(conf.ConfigKeyBaiduTransAppSecret, baiduAppSecretEntry.Text)

			viper.Set(conf.ConfigKeyYouDaoTransAppId, youdaoAppIdEntry.Text)
			viper.Set(conf.ConfigKeyYouDaoTransAppSecret, youdaoAppSecretEntry.Text)

			viper.Set(conf.ConfigKeyGoogleTranslateProxy, googleTranslateProxyEntry.Text)

			err := viper.WriteConfig()
			if err != nil {
				log.E("保存设置时候发生错误")
			}
			// 点击保存之后返回
			setWindowsOpening = false
			setWindow.Close()
		},
	}

	setWindow.SetContent(form)

	setWindow.SetOnClosed(func() {
		setWindowsOpening = false
	})

	setWindowsOpening = true
	setWindow.Show()
}

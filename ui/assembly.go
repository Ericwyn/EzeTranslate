package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/ui/resource/cusWidget"
	"github.com/spf13/viper"
)

func buildFormatCheckBox() *fyne.Container {
	return container.NewHBox(
		widget.NewLabel("输入优化    "),
		cusWidget.CreateCheckGroup(
			[]cusWidget.LabelAndInit{
				{"注释", viper.GetBool(conf.ConfigKeyFormatAnnotation)},
				{"空格", viper.GetBool(conf.ConfigKeyFormatSpace)},
				{"回车", viper.GetBool(conf.ConfigKeyFormatCarriageReturn)},
				{"驼峰", viper.GetBool(conf.ConfigKeyFormatCamelCase)},
			},
			true,  // 横向
			false, // 单选
			func(label string, checked bool) {
				if label == "注释" {
					viper.Set(conf.ConfigKeyFormatAnnotation, checked)
					log.I("输入优化: 注释: ", checked)
				} else if label == "空格" {
					viper.Set(conf.ConfigKeyFormatSpace, checked)
					log.I("输入优化: 空格: ", checked)
				} else if label == "回车" {
					viper.Set(conf.ConfigKeyFormatCarriageReturn, checked)
					log.I("输入优化: 回车: ", checked)
				} else if label == "驼峰" {
					viper.Set(conf.ConfigKeyFormatCamelCase, checked)
					log.I("输入优化: 驼峰: ", checked)
				}
				conf.SaveConfig()
			},
		),
	)
}

func buildTransApiCheckBox() *fyne.Container {
	return container.NewHBox(
		widget.NewLabel("翻译结果    "),
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
}

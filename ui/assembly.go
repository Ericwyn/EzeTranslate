package ui

import (
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/strutils"
	"github.com/spf13/viper"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// formatCheck 描述一个输入优化勾选项。
type formatCheck struct {
	id    ui.ID
	label string
	key   string
}

// buildFormatChecks 画出“输入优化”的四个勾选框（注释/空格/回车/驼峰）。
// 每组“勾选框 + 文字”紧挨在一起（小 gap），组与组之间留较大间距，
// 这样 label 的归属一目了然，不会与相邻项混淆。
func (a *ezeApp) buildFormatChecks(f *flex.Frame) {
	checks := []formatCheck{
		{ckAnnotationID, "注释", conf.ConfigKeyFormatAnnotation},
		{ckSpaceID, "空格", conf.ConfigKeyFormatSpace},
		{ckReturnID, "回车", conf.ConfigKeyFormatCarriageReturn},
		{ckCamelID, "驼峰", conf.ConfigKeyFormatCamelCase},
	}

	f.Row().Gap(16).Items(flex.Center).Children(func() {
		f.Text("输入优化").Muted()
		for _, c := range checks {
			// 闭包捕获
			cur := viper.GetBool(c.key)
			key := c.key
			label := c.label
			// 每组自己一个紧凑 Row：勾选框与文字 gap=4，整组可点
			f.Row().Gap(4).Items(flex.Center).Children(func() {
				if f.Checkbox(c.id, cur).Label(label).Clicked() {
					viper.Set(key, !cur)
					log.I("输入优化: "+label+": ", !cur)
					conf.SaveConfig()
				}
				f.Text(label)
			})
		}
	})
}

var engineLabels = []string{"Google", "Baidu", "Youdao", "OpenAI"}
var engineKeys = []string{"google", "baidu", "youdao", "openai"}

// buildEngineSelect 画出翻译引擎选择下拉。
func (a *ezeApp) buildEngineSelect(f *flex.Frame) {
	current := viper.GetString(conf.ConfigKeyTranslateSelect)
	selected := -1
	for i, k := range engineKeys {
		if k == current {
			selected = i
			break
		}
	}
	f.Row().Gap(8).Items(flex.Center).Children(func() {
		f.Text("翻译引擎").Muted()
		// 给下拉足够宽度，避免 "OpenAI"/"Youdao" 文字被截断
		if i, ok := f.Select(engineScope, &a.trans.engineSel, engineLabels, selected).
			Placeholder("选择引擎").MinWidth(120).Picked(); ok {
			viper.Set(conf.ConfigKeyTranslateSelect, engineKeys[i])
			if e := viper.WriteConfig(); e != nil {
				log.E("配置文件保存失败")
				log.E(e)
			}
		}
	})
}

var langLabels = []string{"自动", "中文", "英文"}

// buildLangSelect 画出目标语言选择下拉。
func (a *ezeApp) buildLangSelect(f *flex.Frame) {
	selected := 0
	switch conf.ToLang {
	case string(strutils.Chinese):
		selected = 1
	case string(strutils.English):
		selected = 2
	}
	f.Row().Gap(8).Items(flex.Center).Children(func() {
		f.Text("目标语言").Muted()
		if i, ok := f.Select(langScope, &a.trans.langSel, langLabels, selected).
			Placeholder("目标语言").MinWidth(90).Picked(); ok {
			switch i {
			case 0:
				conf.ToLang = ""
			case 1:
				conf.ToLang = string(strutils.Chinese)
			case 2:
				conf.ToLang = string(strutils.English)
			}
		}
	})
}

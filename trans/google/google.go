package google

import (
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/trans"
	"github.com/bregydoc/gtranslate"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"unicode"
)

func Translate(str string, transCallback trans.TransResCallback) {

	log.D("Google 翻译文字:", str)

	// 判断 str 是否包含中文
	strLen := 0.0
	hanLen := 0.0

	for _, c := range str {
		strLen += 1
		if unicode.Is(unicode.Han, c) {
			hanLen += 1
		}
	}
	log.D("总长度:", len(str), "汉字长度:", hanLen)

	percentHan := hanLen / strLen

	transRes := ""

	// 有中文的时候，就翻译成英文

	// 纯中文
	// 中英文
	// 翻译成英语

	// 纯英文
	// 翻译成

	note := ""
	var err error
	if percentHan > 0.5 {
		// 中文较多的时候，都会翻译成英文句子
		log.D("翻译中文句子为英文")
		note = "zh -> en"
		transRes, err = gtranslate.Translate(str, language.Chinese, language.English,
			viper.GetString(conf.ConfigKeyGoogleTranslateProxy))
	} else {
		note = "en -> zh"
		transRes, err = gtranslate.Translate(str, language.English, language.Chinese,
			viper.GetString(conf.ConfigKeyGoogleTranslateProxy))
	}

	if err != nil {
		//err.Error()
		log.E(err.Error())
		transCallback("翻译错误, 请查看日志", note)
		return
	}

	transCallback(transRes, note)
}

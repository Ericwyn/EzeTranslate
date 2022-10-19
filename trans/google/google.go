package google

import (
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/trans"
	translator "github.com/Ericwyn/go-googletrans"
	"github.com/spf13/viper"
	"unicode"
)

var translateApi *translator.TranslateApi

func generalTransApi() *translator.TranslateApi {
	var translatorConfig = translator.TranslateConfig{}

	if viper.GetString(conf.ConfigKeyGoogleTranslateUrl) != "" {
		var url = viper.GetString(conf.ConfigKeyGoogleTranslateUrl)
		translatorConfig.ServiceUrls = []string{url}
		log.I("为 google 翻译设置 URL:" + url)
	}

	if viper.GetString(conf.ConfigKeyGoogleTranslateProxy) != "" {
		var proxy = viper.GetString(conf.ConfigKeyGoogleTranslateProxy)
		translatorConfig.Proxy = proxy
		log.I("为 google 翻译设置代理:" + proxy)
	}

	return translator.New(translatorConfig)
}

func Translate(str string, transCallback trans.TransResCallback) {
	if translateApi == nil {
		translateApi = generalTransApi()
	}

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
	//var err error
	if percentHan > 0.5 {
		// 中文较多的时候，都会翻译成英文句子
		log.D("翻译中文句子为英文")
		note = "zh -> en"
		result, err := translateApi.Translate(str, "auto", "en")
		if err != nil {
			//err.Error()
			log.E(err.Error())
			transCallback("翻译错误, 请查看日志", note)
			return
		}
		transRes = result.Text
		note = result.Src + "->" + result.Dest
	} else {
		note = "en -> zh"
		result, err := translateApi.Translate(str, "auto", "zh-cn")
		if err != nil {
			log.E(err.Error())
			transCallback("翻译错误, 请查看日志", note)
			return
		}
		transRes = result.Text
		note = result.Src + "->" + result.Dest
	}
	transCallback(transRes, note)
}

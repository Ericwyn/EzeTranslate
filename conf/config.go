package conf

import (
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/ui/resource"
	"github.com/spf13/viper"
)

const Version = "V1.1"
const ReleaseDate = "2021.11.17"

const ConfigKeyMiniMode = "miniMode"
const ConfigKeyBaiduTransAppId = "baiduTransAppId"
const ConfigKeyBaiduTransAppSecret = "baiduTransAppSecret"

const ConfigKeyYouDaoTransAppId = "youdaoTransAppId"
const ConfigKeyYouDaoTransAppSecret = "youdaoTransAppSecret"

const ConfigKeyFormatSpace = "formatSpace"
const ConfigKeyFormatCarriageReturn = "formatCarriageReturn"
const ConfigKeyFormatAnnotation = "formatAnnotation"
const ConfigKeyFormatCamelCase = "formatCamelCase"

// google, baidu,
const ConfigKeyTranslateSelect = "translateSelect"
const ConfigKeyGoogleTranslateProxy = "googleTranslateProxy"

func InitConfig() {
	viper.SetDefault(ConfigKeyMiniMode, false)

	viper.SetDefault(ConfigKeyBaiduTransAppId, "baiduTransAppId-xxxxxxxxxxxxxxx")
	viper.SetDefault(ConfigKeyBaiduTransAppSecret, "baiduTransAppSecret-xxxxxxxxxxxxxxx")

	viper.SetDefault(ConfigKeyYouDaoTransAppId, "youdaoTransAppId-xxxxxxxxxxxxxxx")
	viper.SetDefault(ConfigKeyYouDaoTransAppSecret, "youdaoTransAppSecret-xxxxxxxxxxxxxxx")

	viper.SetDefault(ConfigKeyGoogleTranslateProxy, "google.com")
	viper.SetDefault(ConfigKeyTranslateSelect, "google")

	viper.SetDefault(ConfigKeyFormatSpace, false)
	viper.SetDefault(ConfigKeyFormatCarriageReturn, false)
	viper.SetDefault(ConfigKeyFormatAnnotation, false)
	viper.SetDefault(ConfigKeyFormatCamelCase, false)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(resource.GetRunnerPath() + "/.conf")
	viper.AddConfigPath(resource.GetRunnerPath())
	err := viper.ReadInConfig()

	if err != nil {
		log.E("载入配置时候出错")
		panic(err)
	}
	printConfigs()
}

func printConfigs() {
	configList := []string{
		ConfigKeyBaiduTransAppId,
		ConfigKeyBaiduTransAppSecret,
		ConfigKeyYouDaoTransAppId,
		ConfigKeyYouDaoTransAppSecret,
		ConfigKeyFormatSpace,
		ConfigKeyFormatCarriageReturn,
		ConfigKeyFormatAnnotation,
		ConfigKeyTranslateSelect,
		ConfigKeyGoogleTranslateProxy,
	}
	for _, key := range configList {
		log.D("config " + key + "  :  " + viper.GetString(key))
	}
}

// 返回百度翻译 api 的 appId 和 appSecret
func GetBaiduTransApiMsg() (string, string) {
	return "", ""
}

func SaveConfig() {
	e := viper.WriteConfig()
	if e != nil {
		log.E("配置文件保存失败")
		log.E(e)
	}
}

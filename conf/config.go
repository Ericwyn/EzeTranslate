package conf

import (
	"os"
	"path"

	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/spf13/viper"
)

const Version = "V1.8-Release"
const ReleaseDate = "2026.03.25"
const FyneVersion = "v2.7.3"

const ConfigKeyMiniMode = "miniMode"
const ConfigKeyBaiduTransAppId = "baiduTransAppId"
const ConfigKeyBaiduTransAppSecret = "baiduTransAppSecret"

const ConfigKeyYouDaoTransAppId = "youdaoTransAppId"
const ConfigKeyYouDaoTransAppSecret = "youdaoTransAppSecret"

const ConfigKeyGoogleTranslateProxy = "googleTranslateProxy"
const ConfigKeyGoogleTranslateUrl = "googleTranslateUrl"

const ConfigKeyOpenAIApiUrl = "openAiApiUrl"
const ConfigKeyOpenAiKey = "openAiKey"
const ConfigKeyOpenAiModel = "openAiModel"

const ConfigKeyFormatSpace = "formatSpace"
const ConfigKeyFormatCarriageReturn = "formatCarriageReturn"
const ConfigKeyFormatAnnotation = "formatAnnotation"
const ConfigKeyFormatCamelCase = "formatCamelCase"

const ConfigKeyTranslateSelect = "translateSelect"

var ToLang = ""

const configFileDirName = "EzeTranslate"
const configFileName = "config"
const configFileType = "yaml"

func InitConfig() {
	viper.SetDefault(ConfigKeyMiniMode, false)

	viper.SetDefault(ConfigKeyBaiduTransAppId, "baiduTransAppId-xxxxxxxxxxxxxxx")
	viper.SetDefault(ConfigKeyBaiduTransAppSecret, "baiduTransAppSecret-xxxxxxxxxxxxxxx")

	viper.SetDefault(ConfigKeyYouDaoTransAppId, "youdaoTransAppId-xxxxxxxxxxxxxxx")
	viper.SetDefault(ConfigKeyYouDaoTransAppSecret, "youdaoTransAppSecret-xxxxxxxxxxxxxxx")

	viper.SetDefault(ConfigKeyGoogleTranslateProxy, "")
	viper.SetDefault(ConfigKeyGoogleTranslateUrl, "https://translate.googleapis.com")

	viper.SetDefault(ConfigKeyOpenAIApiUrl, "https://api.openai.com/v1/chat/completions")
	viper.SetDefault(ConfigKeyOpenAiKey, "openAiKey-xxxxxxxxxxxxxxx")

	viper.SetDefault(ConfigKeyTranslateSelect, "openai")

	viper.SetDefault(ConfigKeyFormatSpace, false)
	viper.SetDefault(ConfigKeyFormatCarriageReturn, false)
	viper.SetDefault(ConfigKeyFormatAnnotation, false)
	viper.SetDefault(ConfigKeyFormatCamelCase, false)

	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)

	GetConfigFilePath()

	viper.AddConfigPath(GetConfigFileDirPath())
	err := viper.ReadInConfig()

	if err != nil {
		log.E("载入配置时候出错")
		panic(err)
	}
}

func GetConfigFilePath() string {
	dir := GetConfigFileDirPath()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	configFilePath := path.Join(dir, configFileName+"."+configFileType)

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		log.I("创建配置文件:" + configFilePath)
		f, err := os.Create(configFilePath)
		if err != nil {
			log.E("无法创建配置文件")
			log.E(err)
			os.Exit(-1)
		}
		f.Close()
		return configFilePath
	} else {
		log.I("找到配置文件:" + configFilePath)
		return configFilePath
	}
}

func GetConfigFileDirPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.E("无法获取配置目录", err)
		os.Exit(-1)
	}
	return path.Join(configDir, configFileDirName)
}

func printConfigs() {
	configList := []string{
		ConfigKeyBaiduTransAppId,
		ConfigKeyBaiduTransAppSecret,
		ConfigKeyYouDaoTransAppId,
		ConfigKeyYouDaoTransAppSecret,
		ConfigKeyGoogleTranslateUrl,
		ConfigKeyGoogleTranslateProxy,
		ConfigKeyFormatSpace,
		ConfigKeyFormatCarriageReturn,
		ConfigKeyFormatAnnotation,
		ConfigKeyTranslateSelect,
	}
	for _, key := range configList {
		log.D("config " + key + "  :  " + viper.GetString(key))
	}
}

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

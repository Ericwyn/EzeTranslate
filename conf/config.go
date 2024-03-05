package conf

import (
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/spf13/viper"
	"os"
	"path"
)

const Version = "V1.6-Release"
const ReleaseDate = "2024.03.01"

const ConfigKeyMiniMode = "miniMode"
const ConfigKeyBaiduTransAppId = "baiduTransAppId"
const ConfigKeyBaiduTransAppSecret = "baiduTransAppSecret"

const ConfigKeyYouDaoTransAppId = "youdaoTransAppId"
const ConfigKeyYouDaoTransAppSecret = "youdaoTransAppSecret"

const ConfigKeyGoogleTranslateProxy = "googleTranslateProxy"
const ConfigKeyGoogleTranslateUrl = "googleTranslateUrl"

const ConfigKeyFormatSpace = "formatSpace"
const ConfigKeyFormatCarriageReturn = "formatCarriageReturn"
const ConfigKeyFormatAnnotation = "formatAnnotation"
const ConfigKeyFormatCamelCase = "formatCamelCase"

// ConfigKeyTranslateSelect 选择哪个翻译
const ConfigKeyTranslateSelect = "translateSelect"

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
	viper.SetDefault(ConfigKeyGoogleTranslateUrl, "translate.google.com")
	viper.SetDefault(ConfigKeyTranslateSelect, "google")

	viper.SetDefault(ConfigKeyFormatSpace, false)
	viper.SetDefault(ConfigKeyFormatCarriageReturn, false)
	viper.SetDefault(ConfigKeyFormatAnnotation, false)
	viper.SetDefault(ConfigKeyFormatCamelCase, false)

	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)
	//viper.AddConfigPath(resource.GetRunnerPath() + "/.conf")
	//viper.AddConfigPath(resource.GetRunnerPath())

	// 初始化配置文件
	GetConfigFilePath()

	viper.AddConfigPath(GetConfigFileDirPath())
	err := viper.ReadInConfig()

	if err != nil {
		log.E("载入配置时候出错")
		panic(err)
	}
	//printConfigs()
}

// GetConfigFilePath
// 获取配置文件路径
// linux 下在 ～/.config/EzeTranslate/config.yaml
func GetConfigFilePath() string {
	dir := GetConfigFileDirPath()

	// 如果 dir 不存在的话, 我们就创建
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	configFilePath := path.Join(dir, configFileName+"."+configFileType)

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		log.I("创建配置文件:" + configFilePath)
		// 如果文件不存在的话, 我们就创建
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

// GetConfigFileDirPath
// 获取配置文件目录
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

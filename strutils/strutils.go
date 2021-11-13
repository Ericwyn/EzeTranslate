package strutils

import (
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/spf13/viper"
	"strings"
)

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

func FormatInputBoxText(formatText string) string {
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
		formatText = strings.ReplaceAll(formatText, "\r\n", " ")
		formatText = strings.ReplaceAll(formatText, "\r", " ")
		formatText = strings.ReplaceAll(formatText, "\n", " ")
	}
	if viper.GetBool(conf.ConfigKeyFormatSpace) {
		// 去除多余空格
		formatText = strings.ReplaceAll(formatText, "  ", " ")
		formatText = strings.Trim(formatText, " ")
	}

	return formatText
}

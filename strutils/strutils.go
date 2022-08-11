package strutils

import (
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
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
	if viper.GetBool(conf.ConfigKeyFormatCamelCase) {
		// 因为 FormatCamelCaseText 会要求输入的 formatText 符合函数名格式才可以
		formatText = strings.Replace(formatText, "\r\n", " ", 1)
		formatText = strings.Replace(formatText, "\r", " ", 1)
		formatText = strings.Replace(formatText, "\n", " ", 1)
		formatText = FormatCamelCaseText(formatText)
	}

	return formatText
}

// FormatCamelCaseText
// 将一个驼峰命名的函数名拆开来
// 将第二个开始的大写字母拆成小写 + 空格
// 还得判断是否全为大写
func FormatCamelCaseText(str string) string {

	// 判断是否含有空格(有空格或者是标点符号的话是一个句子，应该直接返回)
	if strings.Contains(strings.Trim(str, " "), " ") ||
		strings.Contains(str, ",") ||
		strings.Contains(str, ".") {
		return str
	}

	runeStr := []rune(strings.Trim(str, " "))

	// 判断是否全为大写，如果是的话就直接返回，因为有可能是静态变量名
	upperCount := 0
	charCount := 0

	for _, c := range runeStr {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			charCount++
			if (c >= 'A' && c <= 'Z') || c == '_' {
				upperCount++
			}
		}
	}

	if upperCount == charCount {
		// 全部都是大写字母和下划线，有可能是静态变量的命名
		// 我们可以把大写字母全部转成小写，之后再拆开
		runeStr = []rune(strings.ToLower(strings.Trim(str, " ")))
	}

	if charCount < len(runeStr) {
		// 除了大小写和下划线之外还有别的字符，不做驼峰命名拆解，直接返回
		return str
	}

	res := ""
	for i, c := range runeStr {
		if i == 0 {
			res += string(c)
			continue
		}

		if c >= 'A' && c <= 'Z' {
			res += " " + strings.ToLower(string(c))
		} else if c == '_' {
			res += " "
		} else {
			res += string(c)
		}
	}

	log.D("驼峰优化", str, "->", res)

	return res
}

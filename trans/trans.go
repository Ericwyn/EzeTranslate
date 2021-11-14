package trans

import (
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/GoTools/shell"
	"strings"
	"unicode"
)

type TransResCallback func(result string, note string)

// 获取选择的文字
func GetSelection() string {
	return shell.RunShellRes("xclip", "-out")
}

// 翻译文字
// 使用 translate-shell 进行翻译
func ShellTranslateStr(str string) string {
	log.D("翻译文字:", str)

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

	if percentHan > 0.5 {
		// 中文较多的时候，都会翻译成英文句子
		log.D("翻译中文句子为英文")
		transRes = tranToEn(str, true)
	} else {
		// 翻译成中文
		if len(strings.Split(strings.Trim(str, " "), " ")) > 1 {
			log.D("翻译英文语句为中文")
			transRes = tranToZh(str, true)
		} else {
			log.D("翻译英文单词为中文")
			transRes = tranToZh(str, false)
		}

	}

	// 有 ,，。. 的时候，翻译成语句
	//fmt.Println(transRes)
	return transRes
}

// 翻译至英文，参数代表是否翻译为句子
func tranToEn(str string, isLongText bool) string {
	if isLongText {
		return shell.RunShellRes("trans", "-no-ansi", "-b", ":en", "\""+str+"\"")
	} else {
		return shell.RunShellRes("trans", "-no-ansi", ":en", "\""+str+"\"")
	}
}

func tranToZh(str string, isLongText bool) string {
	if isLongText {
		return shell.RunShellRes("trans", "-no-ansi", "-b", ":zh", "\""+str+"\"")
	} else {
		return shell.RunShellRes("trans", "-no-ansi", ":zh", "\""+str+"\"")
	}
}

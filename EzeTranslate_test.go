package main

import (
	"fmt"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/strutils"
	"github.com/Ericwyn/EzeTranslate/trans/google"
	"github.com/Ericwyn/EzeTranslate/trans/youdao"
	"testing"
)

func TestFormatAnnotation(t *testing.T) {
	conf.InitConfig()

	fmt.Println(strutils.FormatInputBoxText(`/**
 * The service class that manages LocationProviders and issues location
 * updates and alerts.
 */
`))
}

func TestFormatCamelCaseText(t *testing.T) {
	a := "hello friends"
	fmt.Println(a, "->", strutils.FormatCamelCaseText(a))

	b := "requestLocationUpdate"
	fmt.Println(b, "->", strutils.FormatCamelCaseText(b))

	c := "format_linux_function"
	fmt.Println(c, "->", strutils.FormatCamelCaseText(c))

	d := "CONFIG_KEY_TRANSLATE"
	fmt.Println(d, "->", strutils.FormatCamelCaseText(d))

	e := "Con_F_Key"
	fmt.Println(e, "->", strutils.FormatCamelCaseText(e))

}

func TestGoogleTrans(t *testing.T) {
	conf.InitConfig()

	google.Translate("你好啊", func(result string, note string) {
		fmt.Println("你好啊 -> " + result)
	})

	google.Translate("hello", func(result string, note string) {
		fmt.Println("hello -> " + result)
	})
}

func TestYoudaoTrans(t *testing.T) {
	conf.InitConfig()

	youdao.Translate("你好啊", func(result string, note string) {
		fmt.Println("你好啊 -> " + result)
	})

	youdao.Translate("hello", func(result string, note string) {
		fmt.Println("hello -> " + result)
	})
}

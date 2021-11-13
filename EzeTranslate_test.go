package main

import (
	"fmt"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/strutils"
	"github.com/Ericwyn/EzeTranslate/trans"
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

func TestGoogleTrans(t *testing.T) {
	conf.InitConfig()

	trans.GoogleTrans("你好啊", func(result string, note string) {
		fmt.Println("你好啊 -> " + result)
	})

	trans.GoogleTrans("hello", func(result string, note string) {
		fmt.Println("hello -> " + result)
	})
}

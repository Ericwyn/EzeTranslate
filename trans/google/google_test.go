package google

import (
	"fmt"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/strutils"
	"github.com/spf13/viper"
	"testing"
)

func TestTranslate(t *testing.T) {
	translator := NewTranslatorWithConfig(
		"https://translate.googleapis.com",
		"",
	)
	text := "Hello"
	mdText := "你好"

	translated, err := translator.Translate(
		text, "en", "zh")
	if err != nil {
		t.Skipf("skip integration test because translate api is unavailable: %v", err)
	}
	fmt.Println(translated)
	if translated != mdText {
		t.Logf("expected: %s", mdText)
		t.Logf("given: %s", translated)
		t.Fail()
	}
}

func TestGoogleTranslate(t *testing.T) {
	conf.InitConfig()

	if viper.GetString(conf.ConfigKeyGoogleTranslateProxy) != "" {
		t.Skip("skip integration test because google proxy is configured externally")
	}

	Translate("你好啊", strutils.English, func(result string, note string) {
		fmt.Println("你好啊 -> " + result)
		fmt.Println(note)
	})

	Translate("hello", strutils.Chinese, func(result string, note string) {
		fmt.Println("hello -> " + result)
		fmt.Println(note)
	})
}

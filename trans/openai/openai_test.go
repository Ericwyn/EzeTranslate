package openai

import (
	"fmt"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/spf13/viper"
	"testing"
)

func TestNormalizeJSONContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain json",
			input:    "{\n\t\"result\":\"hello\",\n\t\"from\":\"en\",\n\t\"to\":\"zh\"\n}",
			expected: "{\n\t\"result\":\"hello\",\n\t\"from\":\"en\",\n\t\"to\":\"zh\"\n}",
		},
		{
			name:     "json fenced by markdown",
			input:    "```json\n{\n  \"result\": \"hello\",\n  \"from\": \"en\",\n  \"to\": \"zh\"\n}\n```",
			expected: "{\n  \"result\": \"hello\",\n  \"from\": \"en\",\n  \"to\": \"zh\"\n}",
		},
		{
			name:     "json fenced with surrounding text",
			input:    "translation result:\n```json\n{\n  \"result\": \"hello\",\n  \"from\": \"en\",\n  \"to\": \"zh\"\n}\n```\ncompleted",
			expected: "{\n  \"result\": \"hello\",\n  \"from\": \"en\",\n  \"to\": \"zh\"\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeJSONContent(tt.input); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestDoTranslateReq(t *testing.T) {
	conf.InitConfig()
	log.D("openai url:", viper.GetString(conf.ConfigKeyOpenAIApiUrl))
	log.D("openai key:", viper.GetString(conf.ConfigKeyOpenAiKey))

	if viper.GetString(conf.ConfigKeyOpenAiKey) == "" || viper.GetString(conf.ConfigKeyOpenAiKey) == "openAiKey-xxxxxxxxxxxxxxx" {
		t.Skip("skip integration test without valid OpenAI key")
	}

	Translate("你好啊", "", func(result string, note string) {
		fmt.Println("你好啊 -> " + result)
		fmt.Println(note)
	})

	fmt.Println("========================================")

	Translate("You are a real, professional translation engine, please follow the following process "+
		"step by step to determine and translate the input content\\n1."+
		" Determine what language the input is in\\n2. if the input language is Chinese"+
		", then translate it into English", "", func(result string, note string) {
		fmt.Println("hello -> " + result)
		fmt.Println(note)
	})
}

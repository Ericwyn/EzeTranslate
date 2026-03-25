package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/strutils"
	"github.com/Ericwyn/EzeTranslate/trans"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model          string         `json:"model"`
	Stream         bool           `json:"stream"`
	ResponseFormat ResponseFormat `json:"response_format"`
	Messages       []Message      `json:"messages"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type SpecialResponseBody struct {
	Success bool         `json:"success"`
	Code    string       `json:"code"`
	Data    ResponseBody `json:"data"`
}

type ResponseBody struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type TransApiRes struct {
	Result string `json:"result"`
	From   string `json:"from"`
	To     string `json:"to"`
}

func GetChatCompletion(messages []Message, openAiApiUrl string, apiKey string, model string) (string, error) {
	requestBody := RequestBody{
		Model:    model,
		Messages: messages,
		Stream:   false,
		ResponseFormat: ResponseFormat{
			Type: "json_object",
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", openAiApiUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	log.D("Sending translation request to OpenAI API")
	log.D("openai url: ", openAiApiUrl)
	log.D("prompt: ", strutils.ToJson(requestBody))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var responseBody ResponseBody

	respJson := string(respByte)
	if strings.Contains(respJson, "code") &&
		strings.Contains(respJson, "success") &&
		strings.Contains(respJson, "data") {
		var specialResponseBody SpecialResponseBody
		err = json.Unmarshal(respByte, &specialResponseBody)
		if err != nil {
			return "", err
		}
		responseBody = specialResponseBody.Data
	} else {
		err = json.Unmarshal(respByte, &responseBody)
		if err != nil {
			log.E("open ai req error: ", err.Error())
			return "", err
		}
	}

	if len(responseBody.Choices) > 0 {
		return responseBody.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("error response: \n" + string(respByte))
}

func buildPromptWithLang(toLang string, inputText string) []Message {
	return []Message{
		{
			Role:    "system",
			Content: `You are a real, professional translation engine`,
		},
		{
			Role: "user",
			Content: `Please translate the contents of <translate_input> into ` + toLang + `.
1. you need to follow the format of the original text as closely as possible, preserving line breaks, spaces and so on
2. output the translation result and translation language in json format, for example

{
    "result": "translate result"
    "from": "zh",
    "to": "en"
}

Here is the content
<translate_input>` + inputText + `</translate_input>`,
		},
	}
}

func normalizeJSONContent(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end >= start {
		content = content[start : end+1]
	}

	return strings.TrimSpace(content)
}

func doTranslateRequest(inputText string, toLang string) (string, string) {
	openaiUrl := viper.GetString(conf.ConfigKeyOpenAIApiUrl)

	content, err := GetChatCompletion(buildPromptWithLang(toLang, inputText),
		openaiUrl,
		viper.GetString(conf.ConfigKeyOpenAiKey),
		viper.GetString(conf.ConfigKeyOpenAiModel),
	)
	if err != nil {
		return "openai 翻译异常: " + err.Error(), ""
	}

	normalizedContent := normalizeJSONContent(content)

	var transResult TransApiRes
	err = json.Unmarshal([]byte(normalizedContent), &transResult)
	if err != nil {
		return "openai 翻译异常, 数据解析失败: " + content, ""
	}

	return transResult.Result, transResult.From + "->" + transResult.To
}

func Translate(str string, toLang strutils.Lang, transCallback trans.TransResCallback) {
	fromLang := strutils.DetectLanguage(str)
	if toLang == "" {
		if fromLang == strutils.Chinese {
			toLang = strutils.English
		} else {
			toLang = strutils.Chinese
		}
	}
	if fromLang == toLang {
		transCallback(str, string(fromLang+"->"+toLang))
		return
	}
	transCallback(doTranslateRequest(str, string(toLang)))
}

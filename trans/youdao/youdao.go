package youdao

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/trans"
	"github.com/spf13/viper"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unicode"
)

const (
	host        = "https://openapi.youdao.com/api"
	lAuto       = "auto"
	lChinese    = "zh-CHS"
	lJapanese   = "ja"
	lEnglish    = "en"
	lKorean     = "ko"
	lFrance     = "fr"
	lRussian    = "ru"
	lPortuguese = "pt" // 葡萄牙语
	lEspanol    = "es" // 西班牙语
)

type basicField struct {
	Phonetic   string   `json:"phonetic"`
	UkPhonetic string   `json:"uk-phonetic"`
	UsPhonetic string   `json:"us-phonetic"`
	UkSpeech   string   `json:"uk-speech"`
	UsSpeech   string   `json:"us-speech"`
	Explains   []string `json:"explains"`
}

type webField struct {
	Value []string `json:"value"`
	Key   string   `json:"key"`
}

type dict map[string]string

// Result 有道字典查询结果
type youDaoTransResult struct {
	ErrorCode   string       `json:"errorCode"`
	Query       string       `json:"queryYoudaoApi"`
	Translation *[]string    `json:"translation"`
	Basic       *basicField  `json:"basic"`
	Web         *[]*webField `json:"web"`
	Dict        *dict        `json:"dict"`
	WebDict     *dict        `json:"webdict"`
	L           string       `json:"l"`
	TSpeakURL   string       `json:"tSpeakUrl"`
	SpeakURL    string       `json:"speakUrl"`
}

var (
	allLanguages = [9]string{lAuto, lChinese, lJapanese, lEnglish, lKorean, lFrance, lRussian, lPortuguese, lEspanol}
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

func randString(n int) string {
	b := make([]byte, n)

	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// signType=v3；
// sign=sha256(应用ID+input+salt+curtime+应用密钥)；
// 其中，input的计算方式为：input=q前10个字符 + q长度 + q后10个字符（当q长度大于20）或 input=q字符串（当q长度小于等于20）；
func signNew(appId string, appSecret string, q string, salt string, curTime string) string {
	input := q
	runeStr := []rune(input)

	if len(runeStr) >= 20 {
		// 不能直接做 str[start:end]
		// 比如 "今天天气不错啊"
		// utf8 编码的汉字会出错, 参考 https://juejin.cn/post/6954545886286331917
		// 导致接口返回 202 错误
		input = string(runeStr[0:10]) + strconv.Itoa(len(runeStr)) + string(runeStr[len(runeStr)-10:])
	}
	before := appId + input + salt + curTime + appSecret
	sign := sha256.Sum256([]byte(before))
	return fmt.Sprintf("%x", sign)
}

func queryYoudaoApi(from, to, q string) (*youDaoTransResult, error) {

	appId := viper.GetString(conf.ConfigKeyYouDaoTransAppId)
	appSecret := viper.GetString(conf.ConfigKeyYouDaoTransAppSecret)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resource, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	salt := randString(6)
	query := url.Values{}
	query.Set("q", q)

	//from := "auto"
	//to := "auto"
	//if len(c.to) > 0 {
	//	to = c.to
	//}

	query.Set("from", from)
	query.Set("to", to)

	query.Set("appKey", appId)
	query.Set("salt", salt)

	curTime := fmt.Sprint(time.Now().Unix())

	query.Set("curtime", curTime)
	query.Set("sign", signNew(appId, appSecret, q, salt, curTime))
	query.Set("signType", "v3")

	resource.RawQuery = query.Encode()

	response, err := client.Get(resource.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var r youDaoTransResult
	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, err
	}
	if r.ErrorCode != "0" {
		return nil, getError(r.ErrorCode)
	}
	return &r, nil
}

var errCodeMessageMap = map[string]string{
	"101": "缺少必填的参数，出现这个情况还可能是et的值和实际加密方式不对应",
	"102": "不支持的语言类型",
	"103": "翻译文本过长",
	"104": "不支持的API类型",
	"105": "不支持的签名类型",
	"106": "不支持的响应类型",
	"107": "不支持的传输加密类型",
	"108": "appKey无效，注册账号， 登录后台创建应用和实例并完成绑定， 可获得应用ID和密钥等信息，其中应用ID就是appKey（ 注意不是应用密钥）",
	"109": "batchLog格式不正确",
	"110": "无相关服务的有效实例",
	"111": "开发者账号无效，可能是账号为欠费状态",
	"201": "解密失败，可能为DES,BASE64,URLDecode的错误",
	"202": "签名检验失败",
	"203": "访问IP地址不在可访问IP列表",
	"301": "辞典查询失败",
	"302": "翻译查询失败",
	"303": "服务端的其它异常",
	"401": "账户已经欠费停",
}

func getError(errCode string) error {
	s, ok := errCodeMessageMap[errCode]
	if !ok {
		s = "未知的错误"
	}
	return errors.New(fmt.Sprintf("[%s] %s", errCode, s))
}

func Translate(str string, transCallback trans.TransResCallback) {
	log.D("youdao 翻译文字:", str)

	from := lAuto
	to := lAuto

	note := ""

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
	if percentHan > 0.5 {
		from = lChinese
		to = lEnglish

		note = "auto -> en"
	} else {
		from = lEnglish
		to = lChinese
		note = "auto -> zh"
	}

	resJson, e := queryYoudaoApi(from, to, str)

	if e != nil {
		transCallback("翻译失败: "+e.Error(), "翻译失败")
		return
	}

	transRes := ""
	for _, line := range *(resJson.Translation) {
		transRes += line + "\n"
	}

	transCallback(transRes, note)
}

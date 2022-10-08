package log

import (
	"fmt"
	"github.com/Ericwyn/GoTools/date"
	"time"
)

var logBuff []string = make([]string, 0)

func getLogTimeStr() string {
	return date.Format(time.Now(), "MM-dd HH:mm:ss")
}

func E(out ...interface{}) {
	log := "[" + getLogTimeStr() + "] [E] " + fmt.Sprint(out...)
	logBuff = append(logBuff, log)

	fmt.Println(log)
}

func D(out ...interface{}) {
	log := "[" + getLogTimeStr() + "] [D] " + fmt.Sprint(out...)
	logBuff = append(logBuff, log)

	fmt.Println(log)
}

func I(out ...interface{}) {
	log := "[" + getLogTimeStr() + "] [I] " + fmt.Sprint(out...)
	logBuff = append(logBuff, log)

	fmt.Println(log)
}

// 获取前 1000 行 log
func GetLog1000() string {
	res := ""
	for i := len(logBuff) - 1; i >= 0; i-- {
		res += logBuff[i] + "\n"
	}
	return res
}

func ClearLogBuff() {
	logBuff = make([]string, 0)
}

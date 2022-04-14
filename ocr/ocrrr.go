package ocr

import (
	"fmt"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/GoTools/file"
	"github.com/Ericwyn/GoTools/shell"
	"math"
	"os"
	"os/user"
	"strings"
	"time"
)

var OCRTextTempPath = ""
var ScrPng = OCRTextTempPath + ".png"

var initFlag = false

func getOcrTextTempPath() string {
	if OCRTextTempPath != "" {
		return OCRTextTempPath
	}
	current, err := user.Current()
	if err != nil {
		return ""
	}
	homePath := current.HomeDir
	log.D("homePath :", homePath)

	//"/home/mi/Pictures/ocr-temp"
	picPath := homePath + "/Pictures/"
	log.D("picPath :", picPath)

	dir := file.OpenFile(picPath)
	if !dir.IsDir() {
		dir.Mkdirs()
	}

	OCRTextTempPath = picPath + "/ocr-temp"
	log.D("OCRTextTempPath :", OCRTextTempPath)

	return OCRTextTempPath
}

func RunOcr() (string, bool) {
	if !initFlag {
		ocrResTempFile := file.OpenFile(getOcrTextTempPath())
		if !ocrResTempFile.IsFile() {
			ocrResTempFile.CreateFile()
			_ = ocrResTempFile.Close()
		}
		initFlag = true
	}

	shell.Debug(true)

	shell.RunShellRes("gnome-screenshot", "-a", "-f", ScrPng)

	// 获取文件修改时间，看看是不是1s内修改的
	modTime := GetFileModTime(ScrPng)

	if math.Abs((float64)(modTime-time.Now().Unix())) > 10 {
		fmt.Println("文件未修改，不做识别")
		return "", false
	}

	shell.RunShellRes("mogrify", "-modulate", "100,0", "-resize", "400%", ScrPng)

	tranRes := shell.RunShellRes("tesseract", ScrPng, "stdout", "-l", "eng")

	tranRes = strings.Trim(tranRes, " ")
	tranRes = strings.Trim(tranRes, "\n")
	tranRes = strings.Trim(tranRes, "\r")
	tranRes = strings.Trim(tranRes, "\n\f")

	//fmt.Println(tranRes)
	//
	//// 复制内容到剪切板
	//err := clipboard.WriteAll(tranRes)
	//
	//if err != nil {
	//	shell.RunShellRes("notify-send", "设置剪贴板错误")
	//} else {
	//	shell.RunShellRes("notify-send", "OCRTextTempPath 成功")
	//}

	return tranRes, true
}

//获取文件修改时间 返回unix时间戳
func GetFileModTime(path string) int64 {
	f, err := os.Open(path)
	if err != nil {
		log.E("open file error")
		return time.Now().Unix()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.E("stat fileinfo error")
		return time.Now().Unix()
	}

	return fi.ModTime().Unix()
}

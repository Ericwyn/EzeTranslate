package resource

import (
	"bufio"
	"fyne.io/fyne/v2"
	"github.com/Ericwyn/EzeTranslate/log"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var resourceBuf = make(map[string]*fyne.StaticResource)
var runnerPath = ""

func GetRunnerPath() string {
	if runnerPath == "" {
		//返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.E("无法获取程序运行目录")
			log.E(err)
		}

		//将\替换成/
		runnerPath = strings.Replace(dir, "\\", "/", -1)
	}

	return runnerPath
}

func GetResource(resourcePath string) *fyne.StaticResource {

	if resourceBuf[resourcePath] != nil {
		log.D("load resource :" + resourcePath + " from buf map")
		return resourceBuf[resourcePath]
	}

	finalByte := make([]byte, 0)

	fi, err := os.Open(resourcePath)

	if err != nil {
		//fmt.Println("read file Error")
		//fmt.Println(err.Error())
		//return
		panic(err)
	}

	defer fi.Close()
	r := bufio.NewReader(fi)

	readBuf := make([]byte, 1024)
	for {
		n, err := r.Read(readBuf)
		if err != nil && err != io.EOF {
			panic(err)
			//return
		}
		if 0 == n {
			break
		} else {
			// 将读取到的数据交给 callback 处理
			//readFileCb(string(readBuf[:n]))
			finalByte = append(finalByte, readBuf[:n]...)
		}
	}

	resourceBuf[resourcePath] = &fyne.StaticResource{
		StaticName:    fi.Name(),
		StaticContent: finalByte,
	}

	log.D("load resource :" + resourcePath + " success !")

	return resourceBuf[resourcePath]
}

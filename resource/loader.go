package resource

import (
	"bufio"
	"fyne.io/fyne/v2"
	"github.com/Ericwyn/TransUtils/log"
	"io"
	"os"
)

var resourceBuf = make(map[string]*fyne.StaticResource)

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

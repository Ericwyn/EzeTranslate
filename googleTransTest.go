package main

import (
	"fmt"
	"github.com/Ericwyn/TransUtils/trans"
)

func main() {
	trans.BaiduTrans("apple", func(result string, note string) {
		fmt.Println(result)
	})
}

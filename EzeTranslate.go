package main

import (
	"flag"
	"fmt"
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/ui"
	"os"
)

var xClipFlag = flag.Bool("x", false, "use xclip to get selected text after boot")
var ocrFlag = flag.Bool("ocr", false, "just ocr only")

var versionFlag = flag.Bool("v", false, "show version")

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println(conf.Version)
		os.Exit(0)
	} else {
		ui.StartApp(*xClipFlag, *ocrFlag)
	}

}

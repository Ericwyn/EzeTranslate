package main

import (
	"flag"
	"github.com/Ericwyn/EzeTranslate/ui"
)

var xClipFlag = flag.Bool("x", false, "use xclip to get selected text after boot")
var ocrFlag = flag.Bool("ocr", false, "just ocr only")

func main() {
	flag.Parse()
	ui.StartApp(*xClipFlag, *ocrFlag)
}

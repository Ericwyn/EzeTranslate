package main

import (
	"flag"
	"github.com/Ericwyn/EzeTranslate/ui"
)

var xClipFlag = flag.Bool("x", false, "use xclip to get selected text after boot")

func main() {
	flag.Parse()
	ui.StartApp(*xClipFlag)
}

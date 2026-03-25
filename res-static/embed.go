package staticres

import _ "embed"

var (
	//go:embed icon/icon.png
	IconPNG []byte

	//go:embed fonts/NotoSansSC-Medium.ttf
	NotoSansSCMediumTTF []byte
)

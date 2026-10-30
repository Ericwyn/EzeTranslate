package resource

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/png"

	"github.com/Ericwyn/EzeTranslate/log"
	staticres "github.com/Ericwyn/EzeTranslate/res-static"
	"github.com/vivaui/viva/app"
)

var windowIconCache *app.WindowIcon

// WindowIcon 把内嵌的 icon.png 解码成 Viva 需要的 RGBA WindowIcon。
func WindowIcon() app.WindowIcon {
	if windowIconCache != nil {
		return *windowIconCache
	}

	icon := app.WindowIcon{}
	img, _, err := image.Decode(bytes.NewReader(staticres.IconPNG))
	if err != nil {
		log.E("解码应用图标失败: ", err)
		windowIconCache = &icon
		return icon
	}

	b := img.Bounds()
	// 统一转成正方形 RGBA（取宽高中较大者，居中贴图）
	side := b.Dx()
	if b.Dy() > side {
		side = b.Dy()
	}
	rgba := image.NewRGBA(image.Rect(0, 0, side, side))
	offset := image.Pt((side-b.Dx())/2, (side-b.Dy())/2)
	draw.Draw(rgba, b.Add(offset), img, b.Min, draw.Src)

	icon = app.WindowIcon{Size: side, RGBA: rgba.Pix}
	windowIconCache = &icon
	return icon
}

package resource

import (
	"fyne.io/fyne/v2"
	staticres "github.com/Ericwyn/EzeTranslate/res-static"
)

var resourceIconCache *fyne.StaticResource

func ResourceIcon() *fyne.StaticResource {
	if resourceIconCache != nil {
		return resourceIconCache
	}
	resourceIconCache = fyne.NewStaticResource("icon.png", staticres.IconPNG)
	return resourceIconCache
}

var resourceFontCache *fyne.StaticResource

func ResourceFontNoto() *fyne.StaticResource {
	if resourceFontCache != nil {
		return resourceFontCache
	}
	resourceFontCache = fyne.NewStaticResource("NotoSansSC-Medium.ttf", staticres.NotoSansSCMediumTTF)
	return resourceFontCache
}

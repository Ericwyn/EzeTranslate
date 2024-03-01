package resource

import "fyne.io/fyne/v2"

var resourceIconCache *fyne.StaticResource = nil

func ResourceIcon() *fyne.StaticResource {
	if resourceIconCache != nil {
		return resourceIconCache
	}
	return GetResource(GetRunnerPath() + "/res-static/icon/icon.png")
}

var resourceFontCache *fyne.StaticResource = nil

func ResourceFontNoto() *fyne.StaticResource {
	if resourceFontCache != nil {
		return resourceFontCache
	}
	return GetResource(GetRunnerPath() + "/res-static/fonts/NotoSansSC-Medium.ttf")
}

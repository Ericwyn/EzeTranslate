package resource

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

type CustomerTheme struct{}

func (t *CustomerTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.LightTheme().Color(name, variant)
}

func (t *CustomerTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.LightTheme().Icon(name)
}

func (t *CustomerTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.LightTheme().Size(name)
}

func (*CustomerTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		return theme.LightTheme().Font(s)
	}
	if s.Bold {
		if s.Italic {
			return theme.LightTheme().Font(s)
		}
		return ResourceFont
	}
	if s.Italic {
		return theme.LightTheme().Font(s)
	}
	return ResourceFont
}

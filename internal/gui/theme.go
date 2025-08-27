package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type CyberpunkTheme struct{}

func (c *CyberpunkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{0x0a, 0x0a, 0x0f, 0xff} // Dark blue-black
	case theme.ColorNameForeground:
		return color.NRGBA{0x00, 0xff, 0xff, 0xff} // Cyan
	case theme.ColorNamePrimary:
		return color.NRGBA{0xff, 0x00, 0x80, 0xff} // Neon pink
	case theme.ColorNameFocus:
		return color.NRGBA{0x00, 0xff, 0x00, 0xff} // Neon green
	case theme.ColorNameButton:
		return color.NRGBA{0x1a, 0x1a, 0x2e, 0xff} // Dark purple
	case theme.ColorNameDisabledButton:
		return color.NRGBA{0x3a, 0x3a, 0x3a, 0xff} // Gray
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{0x80, 0x80, 0x80, 0xff} // Light gray
	case theme.ColorNamePressed:
		return color.NRGBA{0xff, 0x00, 0x80, 0x80} // Semi-transparent pink
	case theme.ColorNameSelection:
		return color.NRGBA{0x00, 0xff, 0xff, 0x40} // Semi-transparent cyan
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (c *CyberpunkTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	return theme.DefaultTheme().Font(style)
}

func (c *CyberpunkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (c *CyberpunkTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNameSubHeadingText:
		return 16
	default:
		return theme.DefaultTheme().Size(name)
	}
}

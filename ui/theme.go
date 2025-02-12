package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type RoyalDarkTheme struct{}

var (
	// Primary colors (Readability focused)
	royalBlack = color.RGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xff} // Base Background
	royalGrey  = color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff} // Slightly lighter background
	royalGold  = color.RGBA{R: 0xff, G: 0xd7, B: 0x00, A: 0xff} // Keep, good accent
	royalWhite = color.RGBA{R: 0xee, G: 0xee, B: 0xee, A: 0xff} // Light grey for text - improved readability

	// Accent colors (Complementary and more subdued)
	accentGold    = color.RGBA{R: 0xB3, G: 0x96, B: 0x05, A: 0xff} // More dull gold color
	accentPurple  = color.RGBA{R: 0x69, G: 0x00, B: 0xff, A: 0xff} // Keep
	accentBlue    = color.RGBA{R: 0x00, G: 0x7A, B: 0xCC, A: 0xff} // Less Intense
	cardBgColor   = color.RGBA{R: 0x2c, G: 0x2c, B: 0x2c, A: 0xff} // Darker but still readable
	hoverColor    = color.RGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xff} // Slightly lighter hover
	disabledColor = color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff} // Dimmed color for disabled controls.
	successColor  = color.RGBA{R: 0x4c, G: 0xaf, B: 0x50, A: 0xff} // Green for success
	errorColor    = color.RGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff} // Red for error

)

const (
	SizeNameInlinePadding fyne.ThemeSizeName = "inline-padding"
)

func (t RoyalDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return royalBlack
	case theme.ColorNameForeground:
		return royalWhite
	case theme.ColorNamePrimary:
		return royalGold // Gold, good for primary action buttons
	case theme.ColorNameButton:
		return accentGold // Accent gold for buttons - keep the contrast
	case theme.ColorNameHover:
		return hoverColor
	case theme.ColorNameInputBackground:
		return royalGrey // The input fields
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff} // Same as before
	case theme.ColorNamePressed:
		return accentPurple // Keep for now
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xff}
	case theme.ColorNameDisabled:
		return disabledColor
	case theme.ColorNameError:
		return errorColor
	case theme.ColorNameSuccess:
		return successColor
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t RoyalDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t RoyalDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t RoyalDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 12
	case SizeNameInlinePadding:
		return 8
	case theme.SizeNameScrollBar:
		return 8
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameCaptionText:
		return 11
	default:
		return theme.DefaultTheme().Size(name)
	}
}

package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	PastelPink     = lipgloss.Color("#FFB3BA")
	PastelPeach    = lipgloss.Color("#FFDFBA")
	PastelYellow   = lipgloss.Color("#FFFFBA")
	PastelMint     = lipgloss.Color("#BAFFC9")
	PastelSky      = lipgloss.Color("#BAE1FF")
	PastelLavender = lipgloss.Color("#D4BAFF")
	PastelRose     = lipgloss.Color("#FFBAE8")

	PastelCream    = lipgloss.Color("#FFF8E7")
	PastelGray     = lipgloss.Color("#E8E8E8")
	PastelDarkGray = lipgloss.Color("#A8A8A8")

	TextDark  = lipgloss.Color("#4A4A4A")
	TextMuted = lipgloss.Color("#8A8A8A")
	TextLight = lipgloss.Color("#FFFFFF")

	SuccessGreen  = lipgloss.Color("#C1E1C1")
	ErrorRed      = lipgloss.Color("#FFB3BA")
	WarningOrange = lipgloss.Color("#FFD4BA")
	InfoBlue      = lipgloss.Color("#B3D9FF")
)

var MoodColors = []lipgloss.Color{
	lipgloss.Color("#FFB3BA"),
	lipgloss.Color("#FFBFC9"),
	lipgloss.Color("#FFCBD6"),
	lipgloss.Color("#FFD7E1"),
	lipgloss.Color("#FFE3EC"),
	lipgloss.Color("#FFF0F5"),
	lipgloss.Color("#E6F5FF"),
	lipgloss.Color("#CCE9FF"),
	lipgloss.Color("#B3DDFF"),
	lipgloss.Color("#99D1FF"),
	lipgloss.Color("#80C5FF"),
}

var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(PastelLavender).
			Bold(true).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PastelSky).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	SelectedBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PastelLavender).
				Padding(1, 2).
				MarginTop(1).
				MarginBottom(1).
				Bold(true)

	ButtonStyle = lipgloss.NewStyle().
			Foreground(TextLight).
			Background(PastelSky).
			Padding(0, 2).
			MarginRight(1).
			Bold(true)

	SelectedButtonStyle = lipgloss.NewStyle().
				Foreground(TextLight).
				Background(PastelLavender).
				Padding(0, 2).
				MarginRight(1).
				Bold(true)

	DisabledButtonStyle = lipgloss.NewStyle().
				Foreground(TextMuted).
				Background(PastelGray).
				Padding(0, 2).
				MarginRight(1)

	ListItemStyle = lipgloss.NewStyle().
			Foreground(TextDark).
			Padding(0, 2)

	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(TextLight).
				Background(PastelLavender).
				Padding(0, 2).
				Bold(true)

	InputStyle = lipgloss.NewStyle().
			Foreground(TextDark).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PastelSky).
			Padding(0, 1)

	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(TextDark).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PastelLavender).
				Padding(0, 1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(TextDark).
			Background(SuccessGreen).
			Padding(0, 2).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(TextDark).
			Background(ErrorRed).
			Padding(0, 2).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(TextDark).
			Background(InfoBlue).
			Padding(0, 2)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(PastelLavender).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderBottom(true).
			BorderForeground(PastelSky).
			Padding(1, 2).
			MarginBottom(1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderTop(true).
			BorderForeground(PastelGray).
			Padding(1, 2).
			MarginTop(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true)
)

func GetMoodColor(level int) lipgloss.Color {
	if level < 0 {
		level = 0
	}
	if level > 10 {
		level = 10
	}
	return MoodColors[level]
}

func MoodStyle(level int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextLight).
		Background(GetMoodColor(level)).
		Bold(true).
		Padding(0, 2)
}

func ProgressBar(value, max int, width int) string {
	if max == 0 {
		return ""
	}

	filled := int(float64(value) / float64(max) * float64(width))
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return lipgloss.NewStyle().
		Foreground(PastelLavender).
		Render(bar)
}

func Sparkline(values []float64, width int) string {
	if len(values) == 0 {
		return lipgloss.NewStyle().
			Foreground(TextMuted).
			Render("(нет данных)")
	}

	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	blocks := "▁▂▃▄▅▆▇█"
	result := ""

	for _, v := range values {
		if max == min {

			result += string(blocks[4])
		} else {

			normalized := (v - min) / (max - min)
			index := int(normalized * 7)
			if index > 7 {
				index = 7
			}

			level := int(v)
			color := GetMoodColor(level)

			styledBlock := lipgloss.NewStyle().
				Foreground(color).
				Bold(true).
				Render(string(blocks[index]))

			result += styledBlock
		}
	}

	return result
}

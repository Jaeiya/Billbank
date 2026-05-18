package utils

import "fmt"

type BorderStyle int

const (
	SingleLine BorderStyle = iota
	Rounded
)

type borderComponents struct {
	TopLeft, TopRight, BottomLeft, BottomRight rune
	Horizontal, Vertical                       rune
	TopT, BottomT, LeftT, RightT, Cross        rune
}

var borders = map[BorderStyle]borderComponents{
	SingleLine: {
		'┌', '┐', '└', '┘',
		'─', '│',
		'┬', '┴', '├', '┤', '┼',
	},
	Rounded: {
		'╭', '╮', '╰', '╯',
		'─', '│',
		'┬', '┴', '├', '┤', '┼',
	},
}

func GetBorderStyle(style BorderStyle) borderComponents {
	if v, exists := borders[style]; exists {
		return v
	}
	// there should never be a reason to pass the wrong style
	panic(fmt.Errorf("border style '%d' does not exist", style))
}

type byteFormat struct {
	amount uint64
	suffix string
}

var byteMap = []byteFormat{
	{1 << 10, "Bytes"},
	{1 << 20, "KiB"},
	{1 << 30, "MiB"},
	{1 << 40, "GiB"},
	{1 << 50, "TiB"},
}

func FormatBytes(bytes uint64) string {
	for i, v := range byteMap {
		if bytes < v.amount {
			if i == 0 {
				return fmt.Sprintf("%d %s", bytes, v.suffix)
			}
			factor := float64(byteMap[i-1].amount)
			return fmt.Sprintf("%.2f %s", float64(bytes)/factor, v.suffix)
		}
	}
	return "unsupported size"
}

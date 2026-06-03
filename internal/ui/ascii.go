package ui

import (
	"charm.land/lipgloss/v2"
)

func ToAsciiFont(s string, style FontStyle) string {
	switch style {
	case CoderMini:
		return newAsciiStr(s, coderMiniFont[:])

	case BigMoney:
		return newAsciiStr(s, bigMoneyFont[:])

	default:
		return ""
	}
}

func newAsciiStr(s string, font []string) string {
	// if len(font) != 27 {
	// 	panic("font arrays must contain exactly 27 values")
	// }

	letters := make([]string, len(s))
	for i, r := range s {
		if r >= 'a' && r <= 'z' {
			letters[i] = font[r-'a']
		} else if r >= 'A' && r <= 'Z' {
			letters[i] = font[r-'A'+26]
		} else {
			letters[i] = font[52] // Always the space char
		}
	}
	return JoinHorizontal(lipgloss.Left, letters...)
}

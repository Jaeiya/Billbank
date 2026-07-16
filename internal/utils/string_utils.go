package utils

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	repeatedHBorder = "" +
		"────────────────────────────────────────────────────────────────" +
		"────────────────────────────────────────────────────────────────"
)

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

func RepeatRune(r rune, count int) string {
	if count <= 0 {
		return ""
	}

	byteCount := utf8.RuneLen(r) * count

	if r == '─' && byteCount < len(repeatedHBorder) {
		return repeatedHBorder[:byteCount]
	}

	sb := strings.Builder{}
	sb.Grow(byteCount)

	for range count {
		sb.WriteRune(r)
	}
	return sb.String()
}

func RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}

// TruncateStr truncates a string to the specified width
// which includes the '…' char
//
// Example:
//
//	TruncateStr("hello", 4) => "hel…"
//	TruncateStr("hello", 5) => "hello"
func TruncateStr(s string, width int) string {
	switch {
	case width == 1:
		return "…"
	case width <= 0:
		return ""
	}

	var truncateAt int
	runes := 0

	for i := range s {
		if runes == width-1 {
			truncateAt = i
		}

		if runes == width {
			return s[:truncateAt] + "…"
		}
		runes++
	}

	return s
}

func ToISO8601(timeStr string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return t, err
	}
	return t, nil
}

// IsNumber returns true if rune is in the range: 0-9
func IsNumber(r rune) bool {
	return r >= 48 && r <= 57
}

// IsAlpha returns true if rune is in the range: a-zA-Z
func IsAlpha(r rune) bool {
	return (r >= 65 && r <= 90) || (r >= 97 && r <= 122)
}

// IsSpecial returns true if rune is any of the following:
// ! " # $ % & ' ( ) * + , - . /
// : ; < = > ? @
// [ \ ] ^ _ `
// { | } ~
func IsSpecial(r rune) bool {
	return (r >= 33 && r <= 47) ||
		(r >= 58 && r <= 64) ||
		(r >= 91 && r <= 96) ||
		(r >= 123 && r <= 126)
}

// IsAlphaNum returns true if a rune is in the range: a-zA-Z0-9
func IsAlphaNum(r rune) bool {
	return IsAlpha(r) || IsNumber(r)
}

// IsAlphaStr returns true if string runes are in the range: a-zA-Z0-9
//
// Or contains: space
func IsAlphaStr(s string) bool {
	for _, r := range s {
		if !IsAlphaNum(r) && r != 32 {
			return false
		}
	}
	return true
}

// IsIntStr returns true if the string can be
// parsed as a uint64.
func IsIntStr(s string) bool {
	if len(s) == 0 || len(s) > 19 {
		return false
	}
	for _, r := range s {
		if !IsNumber(r) {
			return false
		}
	}
	return true
}

// IsFloatStr returns true if the string can
// be parsed as a precise unsigned float64.
func IsFloatStr(s string) bool {
	// After 15, float will round
	if len(s) < 2 || len(s) > 15 {
		return false
	}

	var decimalCount int

	for _, r := range s {
		if r != '.' && !IsNumber(r) {
			return false
		}
		if r == '.' {
			decimalCount++
			if decimalCount > 1 {
				return false
			}
		}
	}

	return decimalCount != 0
}

func GenDashedBorder(length int) string {
	var sb strings.Builder
	sb.Grow(length)

	b := GetBorderStyle(SingleLine)
	for i := range length {
		if i%2 == 0 {
			sb.WriteRune(b.Horizontal)
		} else {
			sb.WriteByte(' ')
		}
	}

	return sb.String()
}

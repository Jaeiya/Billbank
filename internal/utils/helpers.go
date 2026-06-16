package utils

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
)

var workingDir string

var NewID = func() func() int {
	var id int64
	return func() int {
		return int(atomic.AddInt64(&id, 1))
	}
}()

func GetWorkingDir() string {
	if workingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		workingDir = wd
	}

	return workingDir
}

func IsString(v any) bool {
	if _, ok := v.(string); ok {
		return true
	}
	return false
}

func IsInt(v any) bool {
	if _, ok := v.(int); ok {
		return true
	}
	return false
}

// IsNumber returns true if key is in the range: 0-9
func IsNumber(key rune) bool {
	return key >= 48 && key <= 57
}

// IsAlpha returns true if key is in the range: a-zA-Z
func IsAlpha(key rune) bool {
	return (key >= 65 && key <= 90) || (key >= 97 && key <= 122)
}

// IsSpecial returns true if key is any of the following:
// ! " # $ % & ' ( ) * + , - . /
// : ; < = > ? @
// [ \ ] ^ _ `
// { | } ~
func IsSpecial(key rune) bool {
	return (key >= 33 && key <= 47) ||
		(key >= 58 && key <= 64) ||
		(key >= 91 && key <= 96) ||
		(key >= 123 && key <= 126)
}

func ParseInt(s string) (int, error) {
	var newInt int64
	var err error

	if newInt, err = strconv.ParseInt(s, 10, 0); err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			return 0, fmt.Errorf("'%s' is not a number", s)
		}

		if errors.Is(err, strconv.ErrRange) {
			return 0, fmt.Errorf("'%s' is too large or too small", s)
		}

		return 0, fmt.Errorf("unexpected parse error: %w", err)
	}

	return int(newInt), nil
}

func ToAnySlice[T any](s []T) []any {
	newSlice := make([]any, len(s))
	for i, v := range s {
		newSlice[i] = v
	}
	return newSlice
}

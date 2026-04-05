package utils

import (
	"os"
	"strconv"
)

var workingDir string

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

/*
TryDeref tries to dereference a pointer to type T. If it cannot,
then it returns nil.
*/
func TryDeref[T any](p *T) any /*nil|T*/ {
	if p == nil {
		return nil
	}
	return *p
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

func ParseInt(s string) (int, error) {
	newInt, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return int(newInt), err
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

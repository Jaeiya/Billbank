package utils

import (
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

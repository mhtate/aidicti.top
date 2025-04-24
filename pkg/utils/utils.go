package utils

import (
	"fmt"
	"reflect"
	"regexp"
)

func Assert(b bool, s string) {
	// func Assert(b bool, s ...any) {
	if !b {
		panic(fmt.Sprintf("violated! %s", s))
	}
}

func Contract(b bool) {
	if !b {
		panic("Contract violated!")
	}
}

func GetTypeName(tp interface{}) string {
	return reflect.TypeOf(tp).Name()
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("violated!"))
	}

	return v
}

func Sanitize(input string) string {
	re := regexp.MustCompile(`[;|&$><\\'"\` + "`" + `?%*#{}()[\]=]`)

	sanitized := re.ReplaceAllString(input, "_")

	return sanitized
}

func Exhaust[T any](ch <-chan T) {

	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

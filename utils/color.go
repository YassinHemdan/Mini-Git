package utils

import "fmt"

var SGR_CODES = map[string]int{
	"red":   31,
	"green": 32,
}

func Format(style, str string) string {
	code, ok := SGR_CODES[style]
	if !ok {
		return str
	}
	return fmt.Sprintf("\033[%dm%s\033[0m", code, str)
}

package utils

import "fmt"

var SGR_CODES = map[string]int{
	"red":   31,
	"green": 32,
	"cyan":  36,
	"bold":  1,
}

func Format(style, str string) string {
	code, ok := SGR_CODES[style]
	if !ok {
		return str
	}
	return fmt.Sprintf("\033[%dm%s\033[0m", code, str)
}

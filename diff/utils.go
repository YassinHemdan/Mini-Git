package diff

import "strings"

func lines(str string) []string {
	result := strings.SplitAfter(str, "\n")
	if len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return result
}

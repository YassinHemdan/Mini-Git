package diff

import "strings"

func toLines(str string) []*line {
	lines := strings.SplitAfter(str, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	result := make([]*line, 0)
	for i, text := range lines {
		result = append(result, &line{lineNumber: i + 1, text: text})
	}

	return result
}

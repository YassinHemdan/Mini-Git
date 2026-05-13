package diff

import (
	"JIT/utils"
	"fmt"
)

// var typeColorMap map[byte]string

type line struct {
	lineNumber int
	text       string
}

type edit struct {
	Type  byte
	aLine *line
	bLine *line
}

func (e *edit) getValidLine() *line {
	if e.aLine != nil {
		return e.aLine
	}
	return e.bLine
}
func (e *edit) checkNewLine() string {
	text := e.getValidLine().text
	if text[len(text)-1] != '\n' {
		return "\n\\ No newline at end of file\n"
	}

	return ""
}
func (e *edit) toString() string {
	text := e.getValidLine().text

	color := "white"
	switch e.Type {
	case '+':
		color = "green"
	case '-':
		color = "red"
	}
	return utils.Format(color, fmt.Sprintf("%c%s", e.Type, text)) + e.checkNewLine()
}

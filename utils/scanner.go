package utils

import (
	"bufio"
	"bytes"
	"io"
)

type SmartScanner struct {
	scanner      *bufio.Scanner
	currentSplit bufio.SplitFunc
}

func NewObjectScanner(r io.Reader) *SmartScanner {
	ss := &SmartScanner{}
	ss.NewReader(r)
	return ss
}

func (ss *SmartScanner) NewReader(r io.Reader) {
	ss.currentSplit = bufio.ScanLines
	ss.scanner = bufio.NewScanner(r)
	ss.scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		return ss.currentSplit(data, atEOF)
	})
}

func (ss *SmartScanner) SetSplit(split bufio.SplitFunc) {
	ss.currentSplit = split
}

func (ss *SmartScanner) Scan() bool {
	return ss.scanner.Scan()
}
func (ss *SmartScanner) ScanRest() bool {
	ss.SetSplit(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if atEOF {
			return len(data), bytes.TrimRight(data, "\n"), nil
		}
		return 0, nil, nil
	})

	return ss.scanner.Scan()
}
func (ss *SmartScanner) Text() string {
	return ss.scanner.Text()
}

func (ss *SmartScanner) SplitByDelim(delim byte, trim bool) {
	ss.SetSplit(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		start := 0
		if trim {
			for start < len(data) && data[start] == ' ' {
				start++
			}
		}
		if start < len(data) && data[start] == '\n' {
			return start + 1, data[start : start+1], nil
		}
		i := bytes.IndexByte(data, delim)
		if i == -1 {
			i = bytes.IndexByte(data, '\n')
		}
		if i >= 0 {
			return start + i + 1, data[start : start+i], nil
		}

		if atEOF {
			return len(data), bytes.TrimSpace(data[start:]), nil
		}
		return 0, nil, nil
	})
}

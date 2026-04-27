package internals

import (
	"JIT/utils"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Author struct {
	name  string
	email string
	time  time.Time
}

func (a *Author) New(name, email string, time time.Time) error {
	a.name = name
	a.email = email
	a.time = time

	return nil
}

func ParseAuthor(scanner *utils.SmartScanner) *Author {
	state := 0
	scanner.SetSplit(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		start := 0

		for start < len(data) && data[start] == ' ' {
			start++
		}
		if start < len(data) && data[start] == '\n' {
			return start + 1, data[start : start+1], nil
		}

		switch state {
		case 0:
			// extract author name
			if i := bytes.IndexByte(data[start:], '<'); i >= 0 {
				state = 1
				return start + i + 1, bytes.TrimSpace(data[start : start+i]), nil
			}
		case 1:
			// extract author email
			if i := bytes.IndexByte(data[start:], '>'); i >= 0 {
				state = 2
				return start + i + 1, data[start : start+i], nil
			}
		case 2:
			// extract time
			if i := bytes.IndexByte(data[start:], '\n'); i >= 0 {
				state = 3
				return start + i + 1, data[start : start+i], nil
			}
		}
		if atEOF {
			return len(data), bytes.TrimSpace(data), nil
		}

		return 0, nil, nil
	})

	author := Author{}
	scanner.Scan()
	authorName := scanner.Text()

	scanner.Scan()
	authorEmail := scanner.Text()

	scanner.Scan()
	timezone := scanner.Text()
	t, _ := parseTime(timezone)

	author.name = authorName
	author.email = authorEmail
	author.time = t

	return &author
}

func (a *Author) ToString() string {
	timestamp := a.time.Unix()

	_, offsetSeconds := a.time.Zone()
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}
	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60
	timezone := fmt.Sprintf("%s%02d%02d", sign, hours, minutes)

	return fmt.Sprintf("%s <%s> %d %s", a.name, a.email, timestamp, timezone)
}

func parseTime(raw string) (time.Time, error) {
	parts := strings.SplitN(raw, " ", 2)
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format: %s", raw)
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp: %s", parts[0])
	}

	tz := parts[1]
	sign := 1
	if tz[0] == '-' {
		sign = -1
	}

	hours, err := strconv.Atoi(tz[1:3])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone hours: %s", tz)
	}

	minutes, err := strconv.Atoi(tz[3:5])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone minutes: %s", tz)
	}

	offsetSeconds := sign * (hours*3600 + minutes*60)

	location := time.FixedZone(tz, offsetSeconds)
	t := time.Unix(timestamp, 0).In(location)

	return t, nil
}

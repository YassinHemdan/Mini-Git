package internals

import (
	"fmt"
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

package internals

import "time"

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
	return a.name + " <" + a.email + "> " + a.time.Format(time.RFC3339)
}

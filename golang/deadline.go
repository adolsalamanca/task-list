package main

import (
	"fmt"
	"time"
)

var timeFormat = time.DateOnly

type deadline struct {
	date time.Time
}

func NewDeadline(deadlineString string) (deadline, error) {
	date, err := time.Parse(timeFormat, deadlineString)
	if err != nil {
		return deadline{}, err
	}

	return deadline{
		date: date,
	}, nil
}

func (d *deadline) String() string {
	return fmt.Sprintf(" (%s)", d.date.Format(timeFormat))
}

func (d *deadline) IsEmpty() bool {
	return d.date.IsZero()
}

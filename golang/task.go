package main

import (
	"fmt"
	"strconv"
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

type identifier int64

func NewIdentifier(idString string) (identifier, error) {
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return -1, err
	}
	return identifier(id), nil
}

// Task describes an elementary task.
type Task struct {
	id          identifier
	description string
	done        bool
	deadline    deadline
}

// NewTask initializes a Task with the given ID, description and completion status.
func NewTask(id int64, description string, done bool) *Task {
	return &Task{
		id:          identifier(id),
		description: description,
		done:        done,
	}
}

// GetID returns the task ID.
func (t *Task) GetID() identifier {
	return t.id
}

// GetDescription returns the task description.
func (t *Task) GetDescription() string {
	return t.description
}

// IsDone returns whether the task is taskDone or not.
func (t *Task) IsDone() bool {
	return t.done
}

// SetDone changes the completion status of the task.
func (t *Task) SetDone(done bool) {
	t.done = done
}

func (t *Task) SetDeadline(d deadline) {
	t.deadline = d
}

func (t *Task) GetDeadline() string {
	if t.deadline.IsEmpty() {
		return ""
	}

	return t.deadline.String()
}

func (t *Task) IsPreviousToCurrentDate() bool {
	return t.IsDue(time.Now())
}

func (t *Task) IsDue(d time.Time) bool {
	if t.deadline.date.After(d) {
		return false
	}

	return true
}

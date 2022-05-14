package main

import (
	"fmt"
	"strconv"
	"time"
)

type deadline struct {
	value int64
	date  string
}

func NewDeadline(deadlineString string) (deadline, error) {
	value, err := strconv.ParseInt(deadlineString, 10, 64)
	return deadline{
		value: value,
		date:  deadlineString,
	}, err
}

func (d *deadline) String() string {
	return fmt.Sprintf(" (%v)", d.value)
}

func (d *deadline) IsEmpty() bool {
	if d.value == 0 {
		return true
	}
	return false
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
	return t.IsPreviousTo(time.Now().Year(), int(time.Now().Month()), time.Now().Day())
}

func (t *Task) IsPreviousTo(year, month, day int) bool {
	if t.deadline.date <= fmt.Sprintf("%04d%02d%02d\n", year, month, day) {
		return true
	}
	return false
}

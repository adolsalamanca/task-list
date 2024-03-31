package main

import (
	"testing"
	"time"
)

func TestIsPreviousToCurrentDate(t *testing.T) {
	type testData struct {
		name string
		task Task
		want bool
	}

	tests := []testData{
		{
			name: "should return true as input was a valid dateString",
			task: Task{
				deadline: deadline{
					date: parseSafeTime("2020-07-21"),
				},
			},
			want: true,
		},
		{
			name: "also_valid_date",
			task: Task{
				deadline: deadline{
					date: parseSafeTime("2020-07-30"),
				},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.task.IsPreviousToCurrentDate()
			if tc.want != got {
				t.Errorf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestTask_IsDue(t *testing.T) {
	type taskFields struct {
		id          identifier
		description string
		taskDone    bool
		deadline    deadline
	}

	type testData struct {
		name       string
		taskFields taskFields
		date       time.Time
		want       bool
	}

	tests := []testData{
		{
			name: "should return true as task deadline is previous to specified dateString",
			taskFields: taskFields{
				id:          "0",
				description: "",
				taskDone:    false,
				deadline: deadline{
					date: parseSafeTime("2021-11-29"),
				},
			},
			date: parseSafeTime("2021-11-30"),
			want: true,
		},
		{
			name: "should return false as task deadline is not previous to specified dateString",
			taskFields: taskFields{
				id:          "0",
				description: "",
				taskDone:    false,
				deadline: deadline{
					date: parseSafeTime("2050-01-01"),
				},
			},
			date: parseSafeTime("2021-11-30"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			task := &Task{
				id:          tt.taskFields.id,
				description: tt.taskFields.description,
				done:        tt.taskFields.taskDone,
				deadline:    tt.taskFields.deadline,
			}
			if got := task.IsDue(tt.date); got != tt.want {
				t1.Errorf("IsDue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func parseSafeTime(timeString string) time.Time {
	t, err := time.Parse(timeFormat, timeString)
	if err != nil {
		panic(err)
	}

	return t
}

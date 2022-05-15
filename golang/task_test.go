package main

import "testing"

func TestIsPreviousToCurrentDate(t *testing.T) {
	type testData struct {
		name string
		task Task
		want bool
	}

	tests := []testData{
		{
			name: "should return true as input was a valid date",
			task: Task{
				deadline: deadline{
					date: "20200721",
				},
			},
			want: true,
		},
		{
			name: "also_valid_date",
			task: Task{
				deadline: deadline{
					date: "20200730",
				},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.task.IsPreviousToCurrentDate()
			if tc.want != got {
				t.Fatalf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestTask_IsPreviousTo(t1 *testing.T) {
	type taskFields struct {
		id          identifier
		description string
		taskDone    bool
		deadline    deadline
	}
	type date struct {
		year  int
		month int
		day   int
	}

	type testData struct {
		name       string
		taskFields taskFields
		date       date
		want       bool
	}

	tests := []testData{
		{
			name: "should return true as task deadline is previous to specified date",
			taskFields: taskFields{
				id:          0,
				description: "",
				taskDone:    false,
				deadline: deadline{
					value: 0,
					date:  "20211129",
				},
			},
			date: date{
				year:  2021,
				month: 11,
				day:   30,
			},
			want: true,
		},
		{
			name: "should return false as task deadline is not previous to specified date",
			taskFields: taskFields{
				id:          0,
				description: "",
				taskDone:    false,
				deadline: deadline{
					value: 0,
					date:  "20500101",
				},
			},
			date: date{
				year:  2021,
				month: 11,
				day:   30,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Task{
				id:          tt.taskFields.id,
				description: tt.taskFields.description,
				done:        tt.taskFields.taskDone,
				deadline:    tt.taskFields.deadline,
			}
			if got := t.IsPreviousTo(tt.date.year, tt.date.month, tt.date.day); got != tt.want {
				t1.Errorf("IsPreviousTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

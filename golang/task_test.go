package main

import "testing"

func TestIsValidDate(t *testing.T) {
	//
	type tt struct {
		name string
		task Task
		want bool
	}

	tests := []tt{
		{
			name: "valid_date",
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
			got := tc.task.IsDueToday()
			if tc.want != got {
				t.Fatalf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}

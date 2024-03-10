package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"sync"
	"testing"
)

type scenarioTester struct {
	*testing.T

	inWriter   io.Writer
	outReader  io.Reader
	outScanner *bufio.Scanner
}

// TODO: Make use of this struct in all tests
type TaskListRunParams struct {
	wg           sync.WaitGroup
	inPR         *io.PipeReader
	inPW         *io.PipeWriter
	outPW        *io.PipeWriter
	outPR        *io.PipeReader
	errorsChan   chan error
	shutdownChan chan bool
}

func NewTaskListRunParams() TaskListRunParams {
	inPR, inPW := io.Pipe()
	outPR, outPW := io.Pipe()
	return TaskListRunParams{
		wg:           sync.WaitGroup{},
		inPR:         inPR,
		inPW:         inPW,
		outPW:        outPW,
		outPR:        outPR,
		errorsChan:   make(chan error),
		shutdownChan: make(chan bool),
	}
}

func TestTaskList_executeWithErrors(t *testing.T) {
	type args struct {
		cmdCommands []string
	}

	type testData struct {
		name    string
		args    args
		wantErr bool
	}

	tests := []testData{
		{
			name: "test deadline without more parameters returns an error",
			args: args{
				cmdCommands: []string{"deadline"},
			},
			wantErr: true,
		},
		{
			name: "test add without more parameters returns an error",
			args: args{
				cmdCommands: []string{"add"},
			},
			wantErr: true,
		},
		{
			name: "test add without two parameters returns an error",
			args: args{
				cmdCommands: []string{"add foo"},
			},
			wantErr: true,
		},
		{
			name: "test check without more parameters returns an error",
			args: args{
				cmdCommands: []string{"check"},
			},
			wantErr: true,
		},
		{
			name: "test delete without more parameters returns an error",
			args: args{
				cmdCommands: []string{"delete"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runParams := NewTaskListRunParams()
			defer runParams.inPR.Close()
			defer runParams.outPR.Close()
			defer close(runParams.shutdownChan)
			defer close(runParams.errorsChan)

			tester := &scenarioTester{
				T:          t,
				inWriter:   runParams.inPW,
				outReader:  runParams.outPR,
				outScanner: bufio.NewScanner(runParams.outPR),
			}
			initTaskListAndRun(runParams.wg, runParams.inPR, runParams.outPW, runParams.errorsChan, runParams.shutdownChan)

			for _, command := range tt.args.cmdCommands {
				log.Println(tt.name)
				tester.execute(command)
			}

			runParams.inPW.Close()
			runParams.wg.Wait()

			var err error
			select {
			case err = <-runParams.errorsChan:
			case <-runParams.shutdownChan:
			}

			if tt.wantErr && err == nil {
				t.Fail()
			}

			if !tt.wantErr && err != nil {
				t.Fail()
			}
		})
	}
}

func TestTaskList_executeWithReadLines(t *testing.T) {
	type args struct {
		cmdCommands []string
	}

	type testData struct {
		name      string
		args      args
		readLines []string
	}
	tests := []testData{
		{
			name: "after executing run, check and show commands, list of both checked and pending tasks is returned",
			args: args{
				cmdCommands: []string{"add project secrets", "add task secrets Eat more donuts.", "add task secrets Destroy all humans.", "add project training", "add task training Four Elements of Simple Design", "add task training SOLID", "add task training Coupling and Cohesion", "add task training Primitive Obsession", "add task training Outside-In TDD", "add task training Interaction-Driven Design", "check 1", "check 3", "check 5", "check 6", "show"},
			},
			readLines: []string{
				"secrets",
				"    [X] 1: Eat more donuts.",
				"    [ ] 2: Destroy all humans.",
				"",
				"training",
				"    [X] 3: Four Elements of Simple Design",
				"    [ ] 4: SOLID",
				"    [X] 5: Coupling and Cohesion",
				"    [X] 6: Primitive Obsession",
				"    [ ] 7: Outside-In TDD",
				"    [ ] 8: Interaction-Driven Design",
				"",
			},
		},
		{
			name: "after executing run, deadline and show commands, list of pending tasks with deadlines is returned",
			args: args{
				cmdCommands: []string{"add project secrets", "add task secrets Eat more donuts.", "add task secrets Destroy all humans.", "deadline 1 2020-07-21", "deadline 2 2020-07-30", "show"},
			},
			readLines: []string{
				"secrets",
				"    [ ] 1: (2020-07-21) Eat more donuts.",
				"    [ ] 2: (2020-07-30) Destroy all humans.",
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runParams := NewTaskListRunParams()
			defer runParams.inPR.Close()
			defer runParams.outPR.Close()
			defer close(runParams.shutdownChan)
			defer close(runParams.errorsChan)

			tester := &scenarioTester{
				T:          t,
				inWriter:   runParams.inPW,
				outReader:  runParams.outPR,
				outScanner: bufio.NewScanner(runParams.outPR),
			}
			initTaskListAndRun(runParams.wg, runParams.inPR, runParams.outPW, runParams.errorsChan, runParams.shutdownChan)

			for _, command := range tt.args.cmdCommands {
				log.Println(command)
				tester.execute(command)
			}

			tester.readLines(tt.readLines)
			tester.execute("quit")

			runParams.inPW.Close()
			runParams.wg.Wait()

			var err error
			select {
			case err = <-runParams.errorsChan:
				log.Printf("program failed, %s", err)
			case <-runParams.shutdownChan:
				log.Println("finished")
			}

			if err != nil {
				t.Fail()
			}
		})
	}
}

func initTaskListAndRun(wg sync.WaitGroup, inPR *io.PipeReader, outPW *io.PipeWriter, errorsChan chan error, shutdownChan chan bool) {
	go func() {
		wg.Add(1)
		NewTaskList(inPR, outPW).Run(errorsChan, shutdownChan)
		outPW.Close()
		wg.Done()
	}()
}

// execute calls a command, by writing it into the scenario writer.
// It first reads the command prompt, then sends the command.
func (t *scenarioTester) execute(cmd string) error {
	p := make([]byte, len(prompt))
	_, err := t.outReader.Read(p)
	if err != nil {
		return fmt.Errorf("prompt could not be read: %v", err)
	}
	if string(p) != prompt {
		t.Errorf("Invalid prompt, expected \"%s\", got \"%s\"", prompt, string(p))
		return fmt.Errorf("invalid prompt")
	}
	// send command
	fmt.Fprintln(t.inWriter, cmd)
	return nil
}

// readLines reads lines from the scenario scanner, making sure they match
// the expected given lines.
// In case it fails or does not match, makes the calling test fail.
func (t *scenarioTester) readLines(lines []string) {
	for _, expected := range lines {
		if !t.outScanner.Scan() {
			t.Errorf("Expected \"%s\", no input found", expected)
			break
		}
		actual := t.outScanner.Text()
		if actual != expected {
			t.Errorf("Expected \"%s\", got \"%s\"", expected, actual)
		}
	}
	if err := t.outScanner.Err(); err != nil {
		t.Errorf("Could not read input: %v", err)
	}
}

// discardLines reads lines from the scenario scanner, and drops them.
// Used to empty buffers.
func (t *scenarioTester) discardLines(n int) {
	for i := 0; i < n; i++ {
		if !t.outScanner.Scan() {
			t.Error("Expected a line, no input found")
			break
		}
	}
	if err := t.outScanner.Err(); err != nil {
		t.Errorf("Could not read input: %v", err)
	}
}

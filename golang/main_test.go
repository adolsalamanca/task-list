package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"sync"
	"testing"
	"time"
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
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
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
				log.Printf("program failed, %s", err)
			case <-runParams.shutdownChan:
				log.Println("finished")
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

func TestRunToday(t *testing.T) {
	runParams := NewTaskListRunParams()

	log.SetOutput(io.Discard)
	defer runParams.inPR.Close()
	defer runParams.outPR.Close()
	tester := &scenarioTester{
		T:          t,
		inWriter:   runParams.inPW,
		outReader:  runParams.outPR,
		outScanner: bufio.NewScanner(runParams.outPR),
	}

	initTaskListAndRun(runParams.wg, runParams.inPR, runParams.outPW, runParams.errorsChan, runParams.shutdownChan)

	// run command-line scenario
	log.Println("(show empty)")
	tester.execute("show")

	log.Println("(add project)")
	tester.execute("add project secrets")
	log.Println("(add tasks)")
	tester.execute("add task secrets Eat more donuts.")
	tester.execute("add task secrets Destroy all humans.")

	log.Println("(deadline inclusion)")
	tester.execute("deadline 1 20200721")
	tester.execute("deadline 2 20200730")

	log.Println("(today)")
	tester.execute("today")
	tester.readLines([]string{
		"secrets",
		"    [ ] 1: (20200721) Eat more donuts.",
		"    [ ] 2: (20200730) Destroy all humans.",
		"",
	})

	time.Now()

	log.Println("(quit)")
	tester.execute("quit")

	// make sure main program has quit
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
}

func TestRunWithDeadline(t *testing.T) {
	// setup input/output
	log.SetOutput(io.Discard)
	inPR, inPW := io.Pipe()
	defer inPR.Close()
	outPR, outPW := io.Pipe()
	defer outPR.Close()
	tester := &scenarioTester{
		T:          t,
		inWriter:   inPW,
		outReader:  outPR,
		outScanner: bufio.NewScanner(outPR),
	}

	// run main program
	var wg sync.WaitGroup
	shutdownChan := make(chan bool)
	errorsChan := make(chan error)
	initTaskListAndRun(wg, inPR, outPW, errorsChan, shutdownChan)
	// run command-line scenario
	log.Println("(show empty)")
	tester.execute("show")

	log.Println("(add project)")
	tester.execute("add project secrets")
	log.Println("(add tasks)")
	tester.execute("add task secrets Eat more donuts.")
	tester.execute("add task secrets Destroy all humans.")

	log.Println("(deadline inclusion)")
	tester.execute("deadline 1 1595352997")
	tester.execute("deadline 2 1595352922")

	log.Println("(show tasks)")
	tester.execute("show")
	tester.readLines([]string{
		"secrets",
		"    [ ] 1: (1595352997) Eat more donuts.",
		"    [ ] 2: (1595352922) Destroy all humans.",
		"",
	})

	log.Println("(quit)")
	tester.execute("quit")

	// make sure main program has quit
	inPW.Close()
	wg.Wait()

	var err error
	select {
	case err = <-errorsChan:
		log.Printf("program failed, %s", err)
	case <-shutdownChan:
		log.Println("finished")
	}

	if err != nil {
		t.Fail()
	}
}

func TestRun(t *testing.T) {
	// setup input/output
	log.SetOutput(io.Discard)
	inPR, inPW := io.Pipe()
	defer inPR.Close()
	outPR, outPW := io.Pipe()
	defer outPR.Close()
	tester := &scenarioTester{
		T:          t,
		inWriter:   inPW,
		outReader:  outPR,
		outScanner: bufio.NewScanner(outPR),
	}

	// run main program
	var wg sync.WaitGroup
	shutdownChan := make(chan bool)
	errorsChan := make(chan error)
	initTaskListAndRun(wg, inPR, outPW, errorsChan, shutdownChan)

	// run command-line scenario
	log.Println("(show empty)")
	tester.execute("show")

	log.Println("(add project)")
	tester.execute("add project secrets")
	log.Println("(add tasks)")
	tester.execute("add task secrets Eat more donuts.")
	tester.execute("add task secrets Destroy all humans.")

	log.Println("(show tasks)")
	tester.execute("show")
	tester.readLines([]string{
		"secrets",
		"    [ ] 1: Eat more donuts.",
		"    [ ] 2: Destroy all humans.",
		"",
	})

	log.Println("(add second project)")
	tester.execute("add project training")
	log.Println("(add more tasks)")
	tester.execute("add task training Four Elements of Simple Design")
	tester.execute("add task training SOLID")
	tester.execute("add task training Coupling and Cohesion")
	tester.execute("add task training Primitive Obsession")
	tester.execute("add task training Outside-In TDD")
	tester.execute("add task training Interaction-Driven Design")

	log.Println("(check tasks)")
	tester.execute("check 1")
	tester.execute("check 3")
	tester.execute("check 5")
	tester.execute("check 6")

	log.Println("(show completed tasks)")
	tester.execute("show")
	tester.readLines([]string{
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
	})

	log.Println("(quit)")
	tester.execute("quit")

	// make sure main program has quit
	inPW.Close()
	wg.Wait()

	var err error
	select {
	case err = <-errorsChan:
		log.Printf("program failed, %s", err)
	case <-shutdownChan:
		log.Println("finished")
	}

	if err != nil {
		t.Fail()
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
		t.Fatalf("Could not read input: %v", err)
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
		t.Fatalf("Could not read input: %v", err)
	}
}

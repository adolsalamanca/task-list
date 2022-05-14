// Package main implements a command-line task manager.
// A manager is a TaskList object, which is started with the Run() function
// and then scans and executes user commands.
package main

import (
	"os"
)

func main() {
	taskList := NewTaskList(os.Stdin, os.Stdout)
	shutdownChan := make(chan bool)
	errorsChan := make(chan error)

	go func() {
		taskList.Run(errorsChan, shutdownChan)
	}()

	select {
	case err := <-errorsChan:
		println(err)
		os.Exit(1)
	case <-shutdownChan:
		println("finished")
		os.Exit(0)
	}

}

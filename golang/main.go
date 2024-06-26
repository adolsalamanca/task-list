// Package main implements a command-line task manager.
// A manager is a TaskList object, which is started with the Run() function
// and then scans and executes user commands.
package main

import (
	"log"
	"os"

	"github.com/google/uuid"
)

func main() {
	idGenerator := func(_ int64) string {
		return uuid.New().String()
	}
	taskList := NewTaskListReaderWriter(os.Stdin, os.Stdout, idGenerator)
	shutdownChan := make(chan bool)
	errorsChan := make(chan error)

	go func() {
		taskList.Run(errorsChan, shutdownChan)
	}()

	select {
	case <-errorsChan:
		os.Exit(1)
	case <-shutdownChan:
		log.Println("finished")
		os.Exit(0)
	}

}

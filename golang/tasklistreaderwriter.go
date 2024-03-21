package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
)

/*
 * Features to add
 *
 * 1. Deadlines
 *    (i)   Give each task an optional deadline with the 'deadline <ID> <dateString>' command.
 *    (ii)  Show all tasks due today with the 'today' command.
 * 2. Customisable IDs
 *    (i)   Allow the user to specify an identifier that's not a number.
 *    (ii)  Disallow spaces and special characters from the ID.
 * 3. Deletion
 *    (i)   Allow users to delete tasks with the 'delete <ID>' command.
 * 4. Views
 *    (i)   View tasks by dateString with the 'view by dateString' command.
 *    (ii)  View tasks by deadline with the 'view by deadline' command.
 *    (iii) Don't remove the functionality that allows users to view tasks by project,
 *          but change the command to 'view by project'
 */

const (
	// Quit is the text command used to quit the task manager.
	taskNotFoundErr        = Error("Task not found")
	quit            string = "quit"
	prompt          string = "> "

	showCommand     = "show"
	addCommand      = "add"
	checkCommand    = "check"
	uncheckCommand  = "uncheck"
	helpCommand     = "help"
	deadlineCommand = "deadline"
	todayCommand    = "today"
	deleteCommand   = "delete"
)

var (
	invalidParamsDeadline = errors.New("could not execute deadline. Usage: deadline <taskId> <dateAsString>")
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type projectName string

// TaskListReaderWriter wraps a TaskList with read and write capabilities.
type TaskListReaderWriter struct {
	r        io.Reader
	w        io.Writer
	taskList *TaskList
}

// NewTaskListReaderWriter initializes a TaskList on the given reader and writer.
func NewTaskListReaderWriter(r io.Reader, w io.Writer) *TaskListReaderWriter {
	return &TaskListReaderWriter{
		r:        r,
		w:        w,
		taskList: NewTaskList(),
	}
}

// Run runs the command loop of the task manager.
// Sequentially executes any given command, until the user types the Quit message.
func (l *TaskListReaderWriter) Run(errorsChan chan<- error, shutdownChan chan bool) {
	scanner := bufio.NewScanner(l.r)

	fmt.Fprint(l.w, prompt)
	for scanner.Scan() {
		cmdLine := scanner.Text()
		if cmdLine == quit {
			shutdownChan <- true
			return
		}

		err := l.execute(cmdLine)
		if err != nil {
			log.Printf("program exited, %v", err)
			errorsChan <- err
		}
		fmt.Fprint(l.w, prompt)
	}
}

func (l *TaskListReaderWriter) execute(cmdLine string) error {
	args := strings.Split(cmdLine, " ")

	switch command := args[0]; command {
	case showCommand:
		l.show()
	case addCommand:
		if len(args) < 3 {
			return fmt.Errorf("could not execute %s.\nUsage: %s project <project name>\nor\nadd task <project name> <task description>", command, command)
		}
		l.add(args[1:])
	case checkCommand:
		if len(args) < 2 {
			return fmt.Errorf("could not execute %s.\n Usage: %s <taskId> ", command, command)
		}
		l.check(args[1])
	case uncheckCommand:
		l.uncheck(args[1])
	case helpCommand:
		l.help()
	case deadlineCommand:
		if len(args) < 2 {
			return fmt.Errorf("could not execute %s.\n Usage: %s <taskId> <dateAsString>", command, command)
		}
		l.deadline(args[1], args[2])
	case todayCommand:
		l.today()
	case deleteCommand:
		if len(args) < 2 {
			return fmt.Errorf("could not execute %s.\n Usage: %s <taskId>", command, command)
		}
		l.delete()
	default:
		l.error(command)
	}
	return nil
}

func (l *TaskListReaderWriter) help() {
	fmt.Fprintln(l.w, l.taskList.help())
}

func (l *TaskListReaderWriter) error(command string) {
	fmt.Fprintln(l.w, l.taskList.errorMessage(command))
}

func (l *TaskListReaderWriter) today() {
	projectsWithTasks := l.taskList.getProjectWithTasksDueToday()
	for _, projectWithTasks := range projectsWithTasks {
		fmt.Fprintf(l.w, "%s\n", projectWithTasks.projectName)
		for _, task := range projectWithTasks.tasks {
			task.write(l.w)
		}
		fmt.Fprintln(l.w)
	}
}

func (l *TaskListReaderWriter) show() {
	projectsWithTasks := l.taskList.getProjectWithTasks()
	for _, projectWithTasks := range projectsWithTasks {
		fmt.Fprintf(l.w, "%s\n", projectWithTasks.projectName)
		for _, task := range projectWithTasks.tasks {
			task.write(l.w)
		}
		fmt.Fprintln(l.w)
	}
}

// getSortedProjectNames returns all project names sorted, given a map m
// of (key)projectName and (values) slice of tasks
func getSortedProjectNames(projectTasks map[projectName][]*Task) []string {
	projectNames := convertMapOfProjectNamesToSliceOfProjectNames(projectTasks)
	sort.Sort(sort.StringSlice(projectNames))

	return projectNames
}

func convertMapOfProjectNamesToSliceOfProjectNames(projectTasks map[projectName][]*Task) []string {
	projectNames := make([]string, 0, len(projectTasks))
	for projectName := range projectTasks {
		projectNames = append(projectNames, string(projectName))
	}
	return projectNames
}

func (l *TaskListReaderWriter) add(args []string) {
	projectName := args[1]
	if args[0] == "project" {
		l.taskList.addProject(projectName)
		return
	}
	if args[0] == "task" {
		description := strings.Join(args[2:], " ")
		err := l.taskList.addTaskToProject(projectName, description)
		if err != nil {
			fmt.Fprintln(l.w, err)
		}
		return
	}
	command := "add"
	fmt.Fprintf(l.w, "could not execute %s.\nUsage: %s project <project name>\nor\nadd task <project name> <task description>", command, command)
}

func (l *TaskListReaderWriter) check(idString string) {
	err := l.taskList.check(idString)
	if err != nil {
		fmt.Fprintln(l.w, err)
	}
}

func (l *TaskListReaderWriter) uncheck(idString string) {
	err := l.taskList.uncheck(idString)
	if err != nil {
		fmt.Fprintln(l.w, err)
	}
}

func (l *TaskListReaderWriter) deadline(id string, deadlineString string) {
	deadline, err := NewDeadline(deadlineString)
	if err != nil {
		return
	}

	task, err := l.taskList.getTaskBy(id)
	if err != nil {
		fmt.Fprintln(l.w, err)
		return
	}

	task.deadline = deadline
}

func (l *TaskListReaderWriter) delete() {

}

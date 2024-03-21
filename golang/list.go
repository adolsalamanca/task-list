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
	helpMessage            = `Commands:
show
add project <project name>
add task <project name> <task description>
check <task ID>
uncheck <task ID>
deadline <task ID> <date>
today
quit`

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

// TaskList is a set of tasks, grouped by project.
type TaskList struct {
	r io.Reader
	w io.Writer

	projectTasks map[projectName][]*Task
	lastID       int64
}

// NewTaskList initializes a TaskList on the given I/O descriptors.
func NewTaskList(r io.Reader, w io.Writer) *TaskList {
	return &TaskList{
		r:            r,
		w:            w,
		projectTasks: make(map[projectName][]*Task),
		lastID:       0,
	}
}

// Run runs the command loop of the task manager.
// Sequentially executes any given command, until the user types the Quit message.
func (l *TaskList) Run(errorsChan chan<- error, shutdownChan chan bool) {
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

func (l *TaskList) execute(cmdLine string) error {
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

func (l *TaskList) help() {
	fmt.Fprintln(l.w, helpMessage)
}

func (l *TaskList) error(command string) {
	fmt.Fprintf(l.w, "Unknown command \"%s\".\n", command)
}

func (l *TaskList) today() {
	sortedProjects := getSortedProjectNames(l.projectTasks)

	// show projects sequentially
	for _, projectNameStr := range sortedProjects {
		projectName := projectName(projectNameStr)
		tasksOfProject := l.projectTasks[projectName]

		fmt.Fprintf(l.w, "%s\n", projectNameStr)
		for _, task := range tasksOfProject {
			if task.IsPreviousToCurrentDate() {
				task.write(l.w)
			}
		}
		fmt.Fprintln(l.w)
	}
}

func (l *TaskList) show() {
	sortedProjectNames := getSortedProjectNames(l.projectTasks)

	// show projects sequentially
	for _, project := range sortedProjectNames {
		pName := projectName(project)
		tasks := l.projectTasks[pName]

		fmt.Fprintf(l.w, "%s\n", project)
		for _, task := range tasks {
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

func (l *TaskList) add(args []string) {
	projectName := args[1]
	if args[0] == "project" {
		l.addProject(projectName)
		return
	}
	if args[0] == "task" {
		description := strings.Join(args[2:], " ")
		l.addTaskToProject(projectName, description)
		return
	}
	command := "add"
	fmt.Fprintf(l.w, "could not execute %s.\nUsage: %s project <project name>\nor\nadd task <project name> <task description>", command, command)
}

func (l *TaskList) addProject(name string) {
	pName := projectName(name)
	l.projectTasks[pName] = make([]*Task, 0)
}

func (l *TaskList) addTaskToProject(projectNameStr, newTaskDescription string) {
	pName := projectName(projectNameStr)
	tasks, ok := l.projectTasks[pName]

	if !ok {
		fmt.Fprintf(l.w, "Could not find a project with the name \"%s\".\n", projectNameStr)
		return
	}
	l.projectTasks[pName] = append(tasks, NewTask(l.nextID(), newTaskDescription, false))
}

func (l *TaskList) check(idString string) {
	l.setDone(idString, true)
}

func (l *TaskList) uncheck(idString string) {
	l.setDone(idString, false)
}

func (l *TaskList) setDone(idString string, done bool) {
	task, err := l.getTaskBy(idString)
	if err != nil {
		return
	}
	task.done = done
}

func (l *TaskList) getTaskBy(idString string) (*Task, error) {
	id, err := NewIdentifier(idString)
	if err != nil {
		fmt.Fprintf(l.w, "Invalid ID \"%s\".\n", idString)
		return nil, err
	}

	for _, tasks := range l.projectTasks {
		for _, task := range tasks {
			if task.GetID() == id {
				return task, nil
			}
		}
	}

	fmt.Fprintf(l.w, "Task with ID \"%d\" not found.\n", id)
	return nil, taskNotFoundErr
}

func (l *TaskList) nextID() int64 {
	l.lastID++
	return l.lastID
}

func (l *TaskList) deadline(id string, deadlineString string) {
	deadline, err := NewDeadline(deadlineString)
	if err != nil {
		return
	}

	task, err := l.getTaskBy(id)
	if err != nil {
		return
	}

	task.deadline = deadline
}

func (l *TaskList) delete() {

}

package main

import (
	"fmt"
	"sort"
)

const (
	helpMessage = `Commands:
show
add project <project name>
add task <project name> <task description>
check <task ID>
uncheck <task ID>
deadline <task ID> <date>
today
quit`
)

type TaskList struct {
	projectTasks map[projectName][]*Task
	lastID       int64
}

func NewTaskList() *TaskList {
	return &TaskList{
		projectTasks: make(map[projectName][]*Task),
		lastID:       0,
	}
}

func (l *TaskList) help() string {
	return helpMessage
}

func (l *TaskList) errorMessage(command string) string {
	return fmt.Sprintf("Unknown command \"%s\".\n", command)
}

// ProjectWithTasks contains a project name and the associated tasks.
type ProjectWithTasks struct {
	projectName projectName
	tasks       []*Task
}

// getProjectWithTasksDueToday returns the Projects sorted alphabetically
// with the associated tasks that are due today.
func (l *TaskList) getProjectWithTasksDueToday() []ProjectWithTasks {
	var projectstWithTasks []ProjectWithTasks

	sortedProjects := getSortedProjectNames(l.projectTasks)
	for _, projectNameStr := range sortedProjects {
		projectName := projectName(projectNameStr)
		tasksOfProject := l.projectTasks[projectName]

		var tasks []*Task
		for _, task := range tasksOfProject {
			if task.IsPreviousToCurrentDate() {
				tasks = append(tasks, task)
			}
		}

		projectWithTasks := ProjectWithTasks{
			projectName: projectName,
			tasks:       tasks,
		}
		projectstWithTasks = append(projectstWithTasks, projectWithTasks)
	}

	return projectstWithTasks
}

// getProjectWithTasks returns the Projects sorted alphabetically
// with the associated tasks.
func (l *TaskList) getProjectWithTasks() []ProjectWithTasks {
	var projectstWithTasks []ProjectWithTasks

	sortedProjectNames := getSortedProjectNames(l.projectTasks)
	for _, projectNameStr := range sortedProjectNames {
		projectName := projectName(projectNameStr)
		projectWithTasks := ProjectWithTasks{
			projectName: projectName,
			tasks:       l.projectTasks[projectName],
		}
		projectstWithTasks = append(projectstWithTasks, projectWithTasks)
	}

	return projectstWithTasks

}

func (l *TaskList) addProject(name string) {
	pName := projectName(name)
	l.projectTasks[pName] = make([]*Task, 0)
}

func (l *TaskList) addTaskToProject(projectNameStr, newTaskDescription string) error {
	pName := projectName(projectNameStr)
	tasks, ok := l.projectTasks[pName]
	if !ok {
		return fmt.Errorf("could not find a project with the name \"%s\".\n", projectNameStr)
	}

	newTask := NewTask(l.nextID(), newTaskDescription, false)
	l.projectTasks[pName] = append(tasks, newTask)
	return nil
}

func (l *TaskList) check(idString string) error {
	return l.setDone(idString, true)
}

func (l *TaskList) uncheck(idString string) error {
	return l.setDone(idString, false)
}

func (l *TaskList) setDone(idString string, done bool) error {
	task, err := l.getTaskBy(idString)
	if err != nil {
		return err
	}
	task.done = done
	return nil
}

func (l *TaskList) getTaskBy(idString string) (*Task, error) {
	id, err := NewIdentifier(idString)
	if err != nil {
		return nil, fmt.Errorf("invalid ID \"%s\".\n", idString)
	}

	for _, tasks := range l.projectTasks {
		for _, task := range tasks {
			if task.GetID() == id {
				return task, nil
			}
		}
	}

	return nil, fmt.Errorf("task with ID \"%d\" not found.\n", id)
}

func (l *TaskList) nextID() int64 {
	l.lastID++
	return l.lastID
}

func (l *TaskList) deadline(id string, deadlineString string) error {
	deadline, err := NewDeadline(deadlineString)
	if err != nil {
		return err
	}

	task, err := l.getTaskBy(id)
	if err != nil {
		return err
	}

	task.deadline = deadline

	return nil
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

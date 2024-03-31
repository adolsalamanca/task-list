package main

import (
	"reflect"
	"testing"
)

func TestGetProjectWithTasksNoError(t *testing.T) {
	taskList := NewTaskList()

	projectName := "secrets"
	taskList.addProject(projectName)
	taskList.addTaskToProject(projectName, "Eat more donuts")
	taskList.addTaskToProject(projectName, "Destroy all human")

	projectName = "amazing project"
	taskList.addProject(projectName)
	taskList.addTaskToProject(projectName, "Something really amazing")

	projectName = "training"
	taskList.addProject(projectName)
	taskList.addTaskToProject(projectName, "SOLID")
	taskList.addTaskToProject(projectName, "Four Elements of Simple Design")
	taskList.addTaskToProject(projectName, "Coupling and Cohesion")

	taskList.check("1")
	taskList.check("3")
	taskList.check("5")

	projectsWithTasks := taskList.getProjectWithTasks()
	expectedProjectWithTasks := []ProjectWithTasks{
		{
			projectName: "amazing project",
			tasks: []*Task{
				{
					id:          identifier("3"),
					description: "Something really amazing",
					done:        true,
				},
			},
		},
		{
			projectName: "secrets",
			tasks: []*Task{
				{
					id:          identifier("1"),
					description: "Eat more donuts",
					done:        true,
				},
				{
					id:          identifier("2"),
					description: "Destroy all human",
					done:        false,
				},
			},
		},
		{
			projectName: "training",
			tasks: []*Task{
				{
					id:          identifier("4"),
					description: "SOLID",
					done:        false,
				},
				{
					id:          identifier("5"),
					description: "Four Elements of Simple Design",
					done:        true,
				},
				{
					id:          identifier("6"),
					description: "Coupling and Cohesion",
					done:        false,
				},
			},
		},
	}
	if !reflect.DeepEqual(expectedProjectWithTasks, projectsWithTasks) {
		t.Fatalf("expectedProjectWithTasks is not equal to projectsWithTasks:\n%+v, %+v", expectedProjectWithTasks, projectsWithTasks)
	}
}

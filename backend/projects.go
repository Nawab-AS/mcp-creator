package backend

import (
	"context"
	"fmt"
	"time"
)

// Projects struct
type Projects struct{ ctx context.Context }

func (a *Projects) Startup(ctx context.Context) { a.ctx = ctx }

// Status enums
type Status int

const (
	StatusOnline Status = iota
	StatusOffline
	StatusStarting
	StatusStopping
	StatusUnknown
)

type Project struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Star         bool   `json:"star"`
	LastModified string `json:"lastModified"`
	Status       Status `json:"status"`
	ModelUsed    string `json:"modelUsed"`
}

var projects = []Project{
	{
		Name:         "Project 1",
		Path:         "/path/to/project1",
		Star:         true,
		LastModified: "2026-02-01T03:12:00Z",
		Status:       StatusOnline,
		ModelUsed:    "Model A",
	},
	{
		Name:         "Project 2",
		Path:         "/path/to/project2",
		Star:         false,
		LastModified: "2026-04-02T14:45:00Z",
		Status:       StatusOffline,
		ModelUsed:    "Model B",
	},
	{
		Name:         "Project 3",
		Path:         "/path/to/project3",
		Star:         true,
		LastModified: "2025-10-03T09:30:00Z",
		Status:       StatusStarting,
		ModelUsed:    "Model C",
	},
	{
		Name:         "Project 4",
		Path:         "/very/long/abracadabra/path/to/project4",
		Star:         false,
		LastModified: "2026-07-04T22:15:00Z",
		Status:       StatusStopping,
		ModelUsed:    "Model D",
	},
}

func (a *Projects) GetProjects() []Project {
	return projects
}

func (a *Projects) ModifyProject(projectName string, attribute string, value any) int {
	// get pointer to Project with name projectName
	var project *Project
	for i := range projects {
		if projects[i].Name == projectName {
			project = &projects[i]
			break
		}
	}

	// project not found
	if project == nil {
		fmt.Printf("Project `%s` not found\n", projectName)
		return 404
	}

	// set attribute of project to value
	var valid bool = false
	switch attribute {
	case "name":
		if s, ok := value.(string); ok {
			project.Name = s
			valid = true
		}
	case "path":
		if s, ok := value.(string); ok {
			project.Path = s
			valid = true
		}
	case "star":
		if b, ok := value.(bool); ok {
			project.Star = b
			valid = true
		}
	case "status":
		if s, ok := value.(Status); ok {
			project.Status = s
			valid = true
		}
	case "modelUsed":
		if s, ok := value.(string); ok {
			project.ModelUsed = s
			valid = true
		}
	default:
		fmt.Printf("Invalid attribute `%s` for project `%s`\n", attribute, projectName)
		return 400
	}

	if !valid {
		fmt.Printf("Invalid value `%v` for attribute `%s` for project `%s`\n", value, attribute, projectName)
		return 400
	}

	if attribute != "star" {
		project.LastModified = time.Now().Format(time.RFC3339)
	}

	fmt.Printf("Modified project `%s`: set `%s` --> `%v`\n", projectName, attribute, value)
	return 200
}

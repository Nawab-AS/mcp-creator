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
	Status       Status `json:"status,omitempty"`
	ModelUsed    string `json:"modelUsed"`
}

var projects = []Project{
	{
		Name:         "Mac Project",
		Path:         "/Users/User/Documents/project1/",
		Star:         true,
		LastModified: "2026-02-01T03:12:00Z",
		ModelUsed:    "slow",
	},
	{
		Name:         "Windows Project",
		Path:         "C:\\Users\\User\\Documents\\project2\\",
		Star:         false,
		LastModified: "2026-04-02T14:45:00Z",
		ModelUsed:    "balanced",
	},
	{
		Name:         "Linux Project",
		Path:         "/home/user/Documents/project3/",
		Star:         true,
		LastModified: "2025-10-03T09:30:00Z",
		ModelUsed:    "accurate",
	},
	{
		Name:         "Project 4",
		Path:         "/very/long/abracadabra/path/to/project4/",
		Star:         false,
		LastModified: "2026-07-04T22:15:00Z",
		ModelUsed:    "fast",
	},
}

// expose to entire package `backend`
func getProjects() []Project { return projects }

func (a *Projects) GetProjects() []Project { return projects }

func (a *Projects) ModifyProject(projectName string, attribute string, value any) Response {
	// get pointer to Project with name projectName
	var project *Project
	for i := range projects {
		if projects[i].Name == projectName {
			project = &projects[i]
			break
		}
	}

	if project == nil {
		return Response{404, fmt.Sprintf("Project `%s` not found", projectName)}
	}
	var oldProjects []Project = getProjects()
	candidate := *project

	// set attribute of project to value
	var valid bool = false
	switch attribute {
	case "name":
		if s, ok := value.(string); ok {
			if s == "" {
				return Response{400, "name: Name cannot be empty"}
			}
			for _, p := range oldProjects {
				if p.Name == s && p.Name != projectName {
					return Response{409, fmt.Sprintf("name: Project `%s` already exists", s)}
				}
			}
			candidate.Name = s
			valid = true
		}
	case "path":
		if s, ok := value.(string); ok {
			if s == "" {
				return Response{400, "path: Path cannot be empty"}
			}
			for _, p := range oldProjects {
				if p.Path == s && p.Name != projectName {
					return Response{409, "path: Project with this path already exists"}
				}
			}
			candidate.Path = s
			valid = true
		}
	case "star":
		if b, ok := value.(bool); ok {
			candidate.Star = b
			valid = true
		}
	case "status":
		if s, ok := value.(Status); ok {
			candidate.Status = s
			valid = true
		}
	case "modelUsed":
		if s, ok := value.(string); ok {
			candidate.ModelUsed = s
			valid = true
		}
	}

	if !valid {
		return Response{400, fmt.Sprintf("Invalid value `%s` for attribute `%s` for project `%s`", value, attribute, projectName)}
	}

	if attribute != "star" {
		candidate.LastModified = time.Now().Format(time.RFC3339)
	}

	exists, err := FsExists(candidate.Path, true)
	if err != nil || !exists {
		return Response{400, "path: Folder does not exist"}
	}
	*project = candidate

	return Response{200, fmt.Sprintf("Modified project `%s`: set `%s` --> `%v`\n", projectName, attribute, value)}
}

func (a *Projects) CreateProject(name string, path string, model string) Response {
	if name == "" {
		return Response{400, "name: Name is required"}
	}
	if path == "" {
		return Response{400, "path: Path is required"}
	}
	if model == "" {
		return Response{400, "model: Model is required"}
	}

	// check if project with same name already exists
	for _, p := range projects {
		if p.Name == name {
			return Response{409, fmt.Sprintf("name: Project `%s` already exists", name)}
		}
		if p.Path == path {
			return Response{409, "path: Project with this path already exists"}
		}
	}

	var models = GetModels()
	var modelExists bool = false
	for _, m := range models {
		if m.Name == model {
			modelExists = true
			break
		}
	}
	if !modelExists {
		return Response{400, "model: Model does not exist"}
	}
	exists, err := FsExists(path, true)
	if err != nil || !exists {
		return Response{400, "path: Folder does not exist"}
	}

	projects = append(projects, Project{
		Name:         name,
		Path:         path,
		Star:         false,
		LastModified: time.Now().Format(time.RFC3339),
		Status:       StatusUnknown,
		ModelUsed:    model,
	})

	return Response{201, "Project created successfully"}
}

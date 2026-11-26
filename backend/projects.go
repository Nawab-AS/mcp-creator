package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"mcp-creator/backend/mcpserver"
	"mcp-creator/utils"
	"os"
	"regexp"
	"time"
)

// Projects struct
type Projects struct{ ctx context.Context }

var global_ctx context.Context

func (a *Projects) Startup(ctx context.Context) {
	a.ctx = ctx
	global_ctx = ctx

	// load (/save default) models save file
	modelsFilePath := Utils.Files.GetFileStoragePath("projects.json")
	if exists, err := Utils.Files.Exists(modelsFilePath); err != nil && !exists {
		// File does not exist, create it (empty)
		if err := Utils.Files.WriteStructToFile(projects, "projects.json"); err != nil {
			fmt.Println("Error writing default projects to projects.json:", err)
			return
		}
	} else if err == nil {
		// File exists, load it
		file, err := os.Open(modelsFilePath)
		if err != nil {
			fmt.Println("Error opening projects.json:", err)
			return
		}
		defer file.Close()

		var loadedProjects []Project
		if err := json.NewDecoder(file).Decode(&loadedProjects); err != nil {
			fmt.Println("Error decoding projects.json:", err)
			return
		}
		projects = loadedProjects
	}
}

func editStatus(projectName string, status Status) error {
	var project *Project
	for i := range projects {
		if projects[i].Name == projectName {
			project = &projects[i]
			break
		}
	}

	if project == nil {
		return fmt.Errorf("project `%s` not found", projectName)
	}

	project.Status = status
	Events.Emit("project-status", projectName, status)

	return nil
}

var servers = map[string]*mcpserver.Server{} // name --> server
func (a *Backend) InitProjects() {
	for _, project := range projects {
		// create a server for each project
		go func(proj Project) {
			_ = editStatus(proj.Name, StatusStarting)
			server, err := mcpserver.NewServer(proj.Name, proj.Model)
			if err != nil {
				fmt.Printf("Error initializing server for project %s: %v\n", proj.Name, err)
				_ = editStatus(proj.Name, StatusOffline)
				return
			}
			servers[proj.Name] = server

			if err := server.IndexDir(proj.Path, false); err != nil {
				fmt.Printf("Error indexing project %s: %v\n", proj.Name, err)
				_ = server.Close()
				delete(servers, proj.Name)
				_ = editStatus(proj.Name, StatusOffline)
				return
			}
			if err := server.StartServer(proj.Port); err != nil {
				fmt.Printf("Error starting server for project %s: %v\n", proj.Name, err)
				_ = server.Close()
				delete(servers, proj.Name)
				_ = editStatus(proj.Name, StatusOffline)
				return
			}
			_ = editStatus(proj.Name, StatusOnline)
		}(project)
	}
}

func (a *Projects) ReindexProject(projectName string) error {
	server, exists := servers[projectName]
	if !exists {
		return fmt.Errorf("server for project %s not found", projectName)
	}

	var project *Project
	for i := range projects {
		if projects[i].Name == projectName {
			project = &projects[i]
			break
		}
	}
	if project == nil {
		return fmt.Errorf("project %s not found", projectName)
	}
	project.LastModified = time.Now().Format(time.RFC3339)
	server.Paused = true
	project.Status = StatusOffline
	defer func() {
		server.Paused = false
		project.Status = StatusOnline
	}()
	return server.IndexDir(project.Path, true)
}

// StopIndexing cancels an active index operation and restores the database to its pre-index state.
func (a *Projects) StopIndexing(projectName string) error {
	server, exists := servers[projectName]
	if !exists {
		return fmt.Errorf("server for project %s not found", projectName)
	}

	if err := editStatus(projectName, StatusStopping); err != nil {
		return err
	}
	server.Paused = true
	defer func() {
		server.Paused = false
		if err := editStatus(projectName, StatusOnline); err != nil {
			fmt.Printf("Error restoring status for project %s: %v\n", projectName, err)
		}
	}()

	return server.StopIndexing()
}

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
	Status       Status `json:"-"`
	Model        string `json:"model"`
	Port         int    `json:"port"`
}

var projects = []Project{}
var projectNamePattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

func (a *Projects) ModifyProject(projectName string, attribute string, value any) utils.Response {
	// get pointer to Project with name projectName
	var project *Project
	for i := range projects {
		if projects[i].Name == projectName {
			project = &projects[i]
			break
		}
	}

	if project == nil {
		return Utils.NewResponse(true, fmt.Sprintf("Project `%s` not found", projectName))
	}
	var oldProjects []Project = projects
	candidate := *project

	// set attribute of project to value
	var valid bool = false
	switch attribute {
	case "name":
		if s, ok := value.(string); ok {
			if s == "" {
				return Utils.NewResponse(true, "name: Name cannot be empty")
			}
			if !projectNamePattern.MatchString(s) {
				return Utils.NewResponse(true, "name: Name can only contain letters, numbers, and dashes")
			}
			for _, p := range oldProjects {
				if p.Name == s && p.Name != projectName {
					return Utils.NewResponse(true, fmt.Sprintf("name: Project `%s` already exists", s))
				}
			}
			candidate.Name = s
			valid = true
		}
	case "path":
		if s, ok := value.(string); ok {
			if s == "" {
				return Utils.NewResponse(true, "path: Path cannot be empty")
			}
			for _, p := range oldProjects {
				if p.Path == s && p.Name != projectName {
					return Utils.NewResponse(true, "path: Project with this path already exists")
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
			candidate.Model = s
			valid = true
		}
	case "port":
		if port, ok := value.(int); ok {
			if _, err := Utils.Network.IsPortAvailable(port); err != nil {
				return Utils.NewResponse(true, fmt.Sprintf("port: %s", err.Error()))
			}
			candidate.Port = port
			valid = true
		}
	}

	if !valid {
		return Utils.NewResponse(true, fmt.Sprintf("Invalid value for attribute `%s`", attribute))
	}

	if attribute != "star" {
		candidate.LastModified = time.Now().Format(time.RFC3339)
	}

	exists, err := Utils.Files.Exists(candidate.Path)
	if err != nil || !exists {
		return Utils.NewResponse(true, fmt.Sprintf("path: Folder `%s` is not accessible", candidate.Path))
	}
	*project = candidate

	if err := Utils.Files.WriteStructToFile(projects, "projects.json"); err != nil {
		return Utils.NewResponse(true, fmt.Sprintf("Error writing projects to projects.json: %s", err.Error()))
	}

	return Utils.NewResponse(false, fmt.Sprintf("Modified project `%s`: set `%s` --> `%v`\n", projectName, attribute, value))
}

func (a *Projects) CreateProject(name string, path string, model string, port int) utils.Response {
	if name == "" {
		return Utils.NewResponse(true, "name: Name is required")
	}
	if !projectNamePattern.MatchString(name) {
		return Utils.NewResponse(true, "name: Name can only contain letters, numbers, and dashes")
	}
	if path == "" {
		return Utils.NewResponse(true, "path: Path is required")
	}
	if model == "" {
		return Utils.NewResponse(true, "model: Model is required")
	}
	if port == -1 {
		port = Utils.Network.GetAvailablePort()
	}

	// check if project with same name already exists
	for _, p := range projects {
		if p.Name == name {
			return Utils.NewResponse(true, fmt.Sprintf("name: Project `%s` already exists", name))
		}
		if p.Path == path {
			return Utils.NewResponse(true, fmt.Sprintf("path: Project `%s` already exists with this path", p.Name))
		}
	}
	if _, err := Utils.Network.IsPortAvailable(port); err != nil {
		return Utils.NewResponse(true, fmt.Sprintf("port: %s", err.Error()))
	}

	// TODO: make into a function
	var modelExists bool = false
	for _, m := range models {
		if m.Name == model {
			modelExists = true
			break
		}
	}
	if !modelExists {
		return Utils.NewResponse(true, "model: Model does not exist")
	}
	exists, err := Utils.Files.Exists(path)
	if err != nil || !exists {
		return Utils.NewResponse(true, "path: Folder does not exist")
	}

	project := Project{
		Name:         name,
		Path:         path,
		Star:         false,
		LastModified: time.Now().Format(time.RFC3339),
		Status:       StatusStarting,
		Model:        model,
		Port:         port,
	}

	server, err := mcpserver.NewServer(name, model)
	if err != nil {
		fmt.Printf("Error initializing server for project %s: %v\n", name, err)
		return Utils.NewResponse(true, fmt.Sprintf("Error initializing server for project %s: %v", name, err))
	}
	if err := server.StartServer(port); err != nil {
		fmt.Printf("Error starting server for project %s: %v\n", name, err)
		return Utils.NewResponse(true, fmt.Sprintf("Error starting server for project %s: %v", name, err))
	}

	projects = append(projects, project)
	if err := Utils.Files.WriteStructToFile(projects, "projects.json"); err != nil {
		_ = server.Delete()
		projects = projects[:len(projects)-1]
		return Utils.NewResponse(true, fmt.Sprintf("Error writing projects to projects.json: %s", err.Error()))
	}
	servers[name] = server
	go func() {
		if err := a.ReindexProject(name); err != nil {
			fmt.Printf("Error indexing project %s: %v\n", name, err)
		}
	}()
	fmt.Printf("Created Project `%s` at path `%s` with model `%s` and port `%d`\n", name, path, model, port)

	return Utils.NewResponse(false, "Project created successfully")
}

func (a *Projects) DeleteProject(projectName string) utils.Response {
	var project *Project
	for i := range projects {
		if projects[i].Name == projectName {
			project = &projects[i]
			break
		}
	}

	if project == nil {
		return Utils.NewResponse(true, fmt.Sprintf("Project `%s` not found", projectName))
	}

	// Remove the project from the slice
	for i, p := range projects {
		if p.Name == projectName {
			projects = append(projects[:i], projects[i+1:]...)
			break
		}
	}

	if server, exists := servers[projectName]; exists {
		if err := server.Delete(); err != nil {
			return Utils.NewResponse(true, fmt.Sprintf("Error deleting project database: %s", err.Error()))
		}
		delete(servers, projectName)
	}

	if err := Utils.Files.WriteStructToFile(projects, "projects.json"); err != nil {
		return Utils.NewResponse(true, fmt.Sprintf("Error writing projects to projects.json: %s", err.Error()))
	}

	fmt.Printf("Deleted Project `%s`\n", projectName)
	return Utils.NewResponse(false, fmt.Sprintf("Deleted project `%s` successfully", projectName))
}

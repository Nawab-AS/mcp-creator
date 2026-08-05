package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"mcp-creator/backend/mcpserver"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Projects struct
type Projects struct{ ctx context.Context }

var global_ctx context.Context

func (a *Projects) Startup(ctx context.Context) {
	a.ctx = ctx
	global_ctx = ctx

	// load (/save default) models save file
	configDir, err := GetConfigPath()
	if err != nil {
		fmt.Println("Error getting config path:", err)
		return
	}

	modelsFilePath := filepath.Join(configDir, "projects.json")
	if _, err := os.Stat(modelsFilePath); os.IsNotExist(err) {
		// File does not exist, create it with default projects (empty)
		if err := writeSaveFile_projects(); err != nil {
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
	runtime.EventsEmit(global_ctx, "project-status", projectName, status)
	return nil
}

var servers = map[string]*mcpserver.Server{} // name --> server
func (a *Common) InitProjects() {
	for _, project := range getProjects() {
		// create a server for each project
		go func(proj Project) {
			editStatus(proj.Name, StatusStarting)
			server, err := mcpserver.NewServer(proj.Name, proj.ModelUsed)
			if err != nil {
				fmt.Printf("Error initializing server for project %s: %v\n", proj.Name, err)
				return
			}
			servers[proj.Name] = server

			projectsDir, projectName, err := ProjectDBPath(proj.Name)
			if err != nil {
				fmt.Printf("Error getting project DB path for project %s: %v\n", proj.Name, err)
				return
			}
			server.IndexDir(filepath.Join(projectsDir, projectName), false)
			if err := server.StartServer(proj.Port); err != nil {
				fmt.Printf("Error starting server for project %s: %v\n", proj.Name, err)
			}
			editStatus(proj.Name, StatusOnline)
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

var FS_MUTEX_projects sync.Mutex

func writeSaveFile_projects() error {
	FS_MUTEX_projects.Lock()
	defer FS_MUTEX_projects.Unlock()
	// Write `projects` to `projects.json` in the config directory
	configDir, err := GetConfigPath()
	if err != nil {
		fmt.Println("Error getting config path:", err)
		return err
	}

	file, err := os.Create(filepath.Join(configDir, "projects.json"))
	if err != nil {
		fmt.Println("Error creating projects.json:", err)
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(projects)
	if err != nil {
		fmt.Println("Error writing to projects.json:", err)
		return err
	}
	return nil
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
	Status       Status `json:"status,omitempty"`
	ModelUsed    string `json:"modelUsed"`
	Port         int    `json:"port"`
}

var projects = []Project{}
var projectNamePattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

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
			if !projectNamePattern.MatchString(s) {
				return Response{400, "name: Name can only contain letters, numbers, and dashes"}
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
	case "port":
		if port, ok := value.(int); ok {
			if _, err := a.PortAvailable(port); err != nil {
				return Response{409, fmt.Sprintf("port: %s", err.Error())}
			}
			candidate.Port = port
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
		return Response{400, fmt.Sprintf("path: Folder `%s` does not exist", candidate.Path)}
	}
	*project = candidate

	if err := writeSaveFile_projects(); err != nil {
		return Response{500, fmt.Sprintf("Error writing projects to projects.json: %s", err.Error())}
	}

	return Response{200, fmt.Sprintf("Modified project `%s`: set `%s` --> `%v`\n", projectName, attribute, value)}
}

func (a *Projects) CreateProject(name string, path string, model string, port int) Response {
	if name == "" {
		return Response{400, "name: Name is required"}
	}
	if !projectNamePattern.MatchString(name) {
		return Response{400, "name: Name can only contain letters, numbers, and dashes"}
	}
	if path == "" {
		return Response{400, "path: Path is required"}
	}
	if model == "" {
		return Response{400, "model: Model is required"}
	}
	if port == -1 {
		port = a.GetAvailablePort()
	}

	// check if project with same name already exists
	for _, p := range projects {
		if p.Name == name {
			return Response{409, fmt.Sprintf("name: Project `%s` already exists", name)}
		}
		if p.Path == path {
			return Response{409, fmt.Sprintf("path: Project `%s` already exists with this path", p.Name)}
		}
	}
	if _, err := a.PortAvailable(port); err != nil {
		return Response{409, fmt.Sprintf("port: %s", err.Error())}
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
		Status:       StatusStarting,
		ModelUsed:    model,
		Port:         port,
	})

	if err := writeSaveFile_projects(); err != nil {
		return Response{500, fmt.Sprintf("Error writing projects to projects.json: %s", err.Error())}
	}

	server, err := mcpserver.NewServer(name, model)
	if err != nil {
		fmt.Printf("Error initializing server for project %s: %v\n", name, err)
		return Response{500, fmt.Sprintf("Error initializing server for project %s: %v", name, err)}
	}
	if err := server.StartServer(port); err != nil {
		fmt.Printf("Error starting server for project %s: %v\n", name, err)
		return Response{500, fmt.Sprintf("Error starting server for project %s: %v", name, err)}
	}
	a.ReindexProject(name)
	fmt.Printf("Created Project `%s` at path `%s` with model `%s` and port `%d`\n", name, path, model, port)

	return Response{201, "Project created successfully"}
}

func (a *Projects) PortAvailable(port int) (bool, error) {
	if port < 1024 || port > 49150 {
		return false, fmt.Errorf("port must be between 1024 and 49,150 (inclusive)")
	}
	// port used by other projects
	for _, p := range projects {
		if p.Port == port {
			return false, fmt.Errorf("port `%d` is already in use by project `%s`", port, p.Name)
		}
	}

	// port used by other processes
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false, fmt.Errorf("port `%d` is already in use by another process", port)
	}
	_ = ln.Close()
	return true, nil
}

func (a *Projects) GetAvailablePort() int {
	for i := 0; i < 100; i++ {
		port := rand.IntN(49150-1024) + 1024
		if _, err := a.PortAvailable(port); err == nil {
			return port
		}
	}
	return 0 // If you actually get this, Good luck! You'll need it
}

func (a *Projects) DeleteProject(projectName string) Response {
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

	// Remove the project from the slice
	for i, p := range projects {
		if p.Name == projectName {
			projects = append(projects[:i], projects[i+1:]...)
			break
		}
	}

	if server, exists := servers[projectName]; exists {
		server.Delete()
		delete(servers, projectName)
	}

	if err := writeSaveFile_projects(); err != nil {
		return Response{500, fmt.Sprintf("Error writing projects to projects.json: %s", err.Error())}
	}

	fmt.Printf("Deleted Project `%s`\n", projectName)
	return Response{200, fmt.Sprintf("Deleted project `%s` successfully", projectName)}
}

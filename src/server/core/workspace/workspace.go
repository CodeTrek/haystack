package workspace

import (
	"encoding/json"
	"fmt"
	"haystack/conf"
	"haystack/server/core/storage"
	"haystack/utils"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var mutex sync.Mutex
var workspaces map[string]*Workspace
var workspacePaths map[string]*Workspace

type Meta struct {
	ID               string        `json:"id"`
	Path             string        `json:"path"`
	UseGlobalFilters bool          `json:"use_global_filters"`
	Filters          *conf.Filters `json:"filters" optional:"true"`

	CreatedAt    time.Time `json:"created_time"`
	LastAccessed time.Time `json:"last_accessed_time"`
	LastFullSync time.Time `json:"last_full_sync_time"`
}

type Workspace struct {
	Meta Meta

	Indexing *time.Time `json:"-"`
	Deleted  bool       `json:"-"`

	Mutex sync.Mutex
}

func GetAll() []string {
	mutex.Lock()
	defer mutex.Unlock()

	result := []string{}
	for _, workspace := range workspaces {
		result = append(result, workspace.Meta.Path)
	}

	return result
}

func GetByPath(path string) (*Workspace, error) {
	mutex.Lock()
	defer mutex.Unlock()

	path = utils.NormalizePath(path)
	if workspace, ok := workspacePaths[path]; ok {
		return workspace, nil
	}

	return nil, fmt.Errorf("workspace not found")
}

func Get(workspaceId string) (*Workspace, error) {
	mutex.Lock()
	defer mutex.Unlock()

	workspace, ok := workspaces[workspaceId]
	if !ok || workspace.Deleted {
		return nil, fmt.Errorf("workspace not found")
	}

	return workspace, nil
}

func (w *Workspace) Save() error {
	json, err := w.serialize()
	if err != nil {
		return err
	}

	return storage.SaveWorkspace(w.Meta.ID, string(json))
}

func (w *Workspace) UpdateLastFullSync() {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	w.Meta.LastFullSync = time.Now()
}

func (w *Workspace) GetFilters() conf.Filters {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()
	if w.Meta.Filters == nil || w.Meta.UseGlobalFilters {
		return conf.Get().Server.Filters
	}

	return *w.Meta.Filters
}

func (w *Workspace) Delete() error {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	w.Deleted = true
	delete(workspaces, w.Meta.ID)
	delete(workspacePaths, w.Meta.Path)

	// TODO: Delete index
	return nil
}

func GetOrCreate(workspacePath string) (*Workspace, error) {
	workspacePath = utils.NormalizePath(workspacePath)

	mutex.Lock()
	defer mutex.Unlock()

	workspace := workspacePaths[workspacePath]
	if workspace == nil {
		// Validate the workspace path
		// 1. It must be absolute
		// 2. It must be a directory
		if !filepath.IsAbs(workspacePath) {
			return nil, fmt.Errorf("workspace path must be absolute")
		}

		info, err := os.Stat(workspacePath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat workspace: %v", err)
		}

		if !info.IsDir() {
			return nil, fmt.Errorf("workspace path must be a directory")
		}

		var id string
		// Try 10 times to generate a unique workspace id
		for i := 0; i < 10; i++ {
			id, err = storage.GetIncreasedWorkspaceID()
			if err != nil {
				return nil, err
			}

			if _, ok := workspaces[id]; !ok {
				break
			}
		}

		if _, ok := workspaces[id]; ok {
			return nil, fmt.Errorf("failed to generate unique workspace id")
		}

		workspace = &Workspace{
			Meta: Meta{
				ID:               id,
				Path:             workspacePath,
				UseGlobalFilters: true,
				CreatedAt:        time.Now(),
				LastAccessed:     time.Now(),
			},
		}

		if err := workspace.Save(); err != nil {
			return nil, err
		}

		workspaces[id] = workspace
		workspacePaths[workspacePath] = workspace

		log.Printf("New workspace created: %v, path: %v", id, workspacePath)
	}

	return workspace, nil
}

func (w *Workspace) serialize() ([]byte, error) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	return json.Marshal(w.Meta)
}

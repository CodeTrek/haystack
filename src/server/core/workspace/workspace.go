package workspace

import (
	"encoding/json"
	"fmt"
	"haystack/conf"
	"haystack/server/core/storage"
	"haystack/shared/types"
	"haystack/utils"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

var mutex sync.Mutex
var workspaces map[string]*Workspace
var workspacePaths map[string]*Workspace

type IndexingStatus struct {
	StartedAt    *time.Time
	TotalFiles   int
	IndexedFiles int
}

type Workspace struct {
	ID               string        `json:"id"`
	Path             string        `json:"path"`
	UseGlobalFilters bool          `json:"use_global_filters"`
	Filters          *conf.Filters `json:"filters" optional:"true"`
	TotalFiles       int           `json:"total_files"`

	CreatedAt    time.Time `json:"created_time"`
	LastAccessed time.Time `json:"last_accessed_time"`
	LastFullSync time.Time `json:"last_full_sync_time"`

	IndexingStatus *IndexingStatus `json:"-"`
	Deleted        bool            `json:"-"`

	Mutex sync.Mutex `json:"-"`
}

func GetAllPaths() []string {
	mutex.Lock()
	defer mutex.Unlock()

	result := []string{}
	for _, workspace := range workspaces {
		result = append(result, workspace.Path)
	}

	return result
}

func GetAll() []types.Workspace {
	mutex.Lock()
	defer mutex.Unlock()

	result := []types.Workspace{}
	for _, workspace := range workspaces {
		workspace.Mutex.Lock()

		totalFiles := workspace.TotalFiles
		if totalFiles == 0 && workspace.IndexingStatus != nil {
			totalFiles = workspace.IndexingStatus.TotalFiles
		}

		result = append(result, types.Workspace{
			ID:           workspace.ID,
			Path:         workspace.Path,
			CreatedAt:    workspace.CreatedAt,
			TotalFiles:   totalFiles,
			LastAccessed: workspace.LastAccessed,
			LastFullSync: workspace.LastFullSync,
			Indexing:     workspace.IndexingStatus != nil,
		})

		workspace.Mutex.Unlock()
	}

	sort.Slice(result, func(i, j int) bool {
		ri, _ := strconv.Atoi(result[i].ID)
		rj, _ := strconv.Atoi(result[j].ID)
		return ri < rj
	})

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
	json, err := w.Serialize()
	if err != nil {
		return err
	}

	return storage.SaveWorkspace(w.ID, string(json))
}

func (w *Workspace) UpdateLastFullSync() {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.IndexingStatus != nil {
		w.TotalFiles = w.IndexingStatus.TotalFiles
		w.IndexingStatus = nil
	}

	w.LastFullSync = time.Now()
}

func (w *Workspace) GetFilters() conf.Filters {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()
	if w.Filters == nil || w.UseGlobalFilters {
		return conf.Get().Server.Filters
	}

	return *w.Filters
}

func (w *Workspace) Delete() error {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	w.Deleted = true
	delete(workspaces, w.ID)
	delete(workspacePaths, w.Path)

	// TODO: Delete index
	return nil
}

func Create(workspacePath string) (*Workspace, error) {
	mutex.Lock()
	defer mutex.Unlock()

	workspace := workspacePaths[workspacePath]
	if workspace != nil {
		return nil, fmt.Errorf("workspace already exists")
	}

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
	for range 10 {
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
		ID:               id,
		Path:             workspacePath,
		UseGlobalFilters: true,
		CreatedAt:        time.Now(),
		LastAccessed:     time.Now(),
	}

	if err := workspace.Save(); err != nil {
		return nil, err
	}

	workspaces[id] = workspace
	workspacePaths[workspacePath] = workspace

	log.Printf("New workspace created: %v, path: %v", id, workspacePath)

	return workspace, nil
}

func (w *Workspace) Serialize() ([]byte, error) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	return json.Marshal(w)
}

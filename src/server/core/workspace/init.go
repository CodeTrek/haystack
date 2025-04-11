package workspace

import (
	"context"
	"encoding/json"
	"fmt"
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

var (
	workspaces     map[string]*Workspace
	workspacePaths map[string]*Workspace

	mutex sync.Mutex
)

func Init(shutdown context.Context, wg *sync.WaitGroup) error {
	mutex.Lock()
	defer mutex.Unlock()

	workspaces = make(map[string]*Workspace)
	workspacePaths = make(map[string]*Workspace)
	allWorkspaces, err := storage.GetAllWorkspaces()
	if err != nil {
		return err
	}

	for _, workspace := range allWorkspaces {
		space := Workspace{
			ID:               workspace[0],
			UseGlobalFilters: true,
		}

		if err := json.Unmarshal([]byte(workspace[1]), &space); err == nil {
			space.Path = utils.NormalizePath(space.Path)
			workspaces[space.ID] = &space
			workspacePaths[space.Path] = workspaces[space.ID]
			log.Printf("Found workspace: %v, path: %v", space.ID, space.Path)
		} else {
			log.Printf("Error unmarshalling workspace: %v", err)
			// TODO: Delete the malformed workspace
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-shutdown.Done()
		log.Println("Workspace shutdown.")
	}()
	return nil
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
		workspace.mutex.Lock()

		totalFiles := workspace.TotalFiles
		if totalFiles == 0 && workspace.indexingStatus != nil {
			totalFiles = workspace.indexingStatus.TotalFiles
		}

		result = append(result, types.Workspace{
			ID:           workspace.ID,
			Path:         workspace.Path,
			CreatedAt:    workspace.CreatedAt,
			TotalFiles:   totalFiles,
			LastAccessed: workspace.LastAccessed,
			LastFullSync: workspace.LastFullSync,
			Indexing:     workspace.indexingStatus != nil,
		})

		workspace.mutex.Unlock()
	}

	sort.Slice(result, func(i, j int) bool {
		ri, _ := strconv.Atoi(result[i].ID)
		rj, _ := strconv.Atoi(result[j].ID)
		return ri < rj
	})

	return result
}

func GetByPath(path string) (*Workspace, error) {
	path = utils.NormalizePath(path)

	mutex.Lock()
	defer mutex.Unlock()
	if workspace, ok := workspacePaths[path]; ok {
		return workspace, nil
	}

	return nil, fmt.Errorf("workspace not found")
}

func Get(workspaceId string) (*Workspace, error) {
	mutex.Lock()
	defer mutex.Unlock()

	workspace, ok := workspaces[workspaceId]
	if !ok || workspace.deleted {
		return nil, fmt.Errorf("workspace not found")
	}

	return workspace, nil
}

func Create(workspacePath string) (*Workspace, error) {
	workspacePath = utils.NormalizePath(workspacePath)

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

	workspaces[workspace.ID] = workspace
	workspacePaths[workspace.Path] = workspace

	log.Printf("New workspace created: %v, path: %v", id, workspacePath)

	return workspace, nil
}

func Delete(workspaceId string) error {
	mutex.Lock()
	defer mutex.Unlock()

	workspace, ok := workspaces[workspaceId]
	if !ok {
		return fmt.Errorf("workspace not found")
	}

	workspace.SetDeleted()
	delete(workspaces, workspaceId)
	delete(workspacePaths, workspace.Path)

	return storage.DeleteWorkspace(workspaceId)
}

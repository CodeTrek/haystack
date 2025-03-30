package workspace

import (
	"encoding/json"
	"fmt"
	"log"
	"search-indexer/running"
	"search-indexer/server/core/storage"
	"sync"
	"time"
)

func Init(wg *sync.WaitGroup) error {
	mutex.Lock()
	defer mutex.Unlock()

	workspaces = make(map[string]*Workspace)
	allWorkspaces, err := storage.GetAllWorkspaces()
	if err != nil {
		return err
	}

	for _, workspace := range allWorkspaces {
		space := Meta{
			ID:               workspace[0],
			UseGlobalFilters: true,
		}

		if err := json.Unmarshal([]byte(workspace[1]), &space); err == nil {
			workspaces[space.ID] = &Workspace{Meta: space}
			log.Printf("Found workspace: %v, path: %v", space.ID, space.Path)
		} else {
			log.Printf("Error unmarshalling workspace: %v", err)
			workspaces[workspace[0]] = &Workspace{Meta: Meta{ID: workspace[0]}}
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		running.WaitingForShutdown()
		log.Println("Workspace shutdown.")
	}()
	return nil
}

func GetOrCreate(path string) (*Workspace, error) {
	mutex.Lock()
	defer mutex.Unlock()

	var workspace *Workspace
	for _, s := range workspaces {
		if s.Meta.Path == path {
			workspace = s
			break
		}
	}

	if workspace == nil {
		id, err := storage.GetIncreasedWorkspaceID()
		if err != nil {
			return nil, err
		}

		if _, ok := workspaces[id]; ok {
			return nil, fmt.Errorf("workspace id %s already exists", id)
		}

		workspace = &Workspace{
			Meta: Meta{
				ID:               id,
				Path:             path,
				UseGlobalFilters: true,
				CreatedAt:        time.Now(),
				LastAccessed:     time.Now(),
			},
		}

		if err := workspace.Save(); err != nil {
			return nil, err
		}

		workspaces[id] = workspace
	}

	return workspace, nil
}

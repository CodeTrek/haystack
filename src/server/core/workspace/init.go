package workspace

import (
	"encoding/json"
	"haystack/server/core/storage"
	"haystack/shared/running"
	"haystack/utils"
	"log"
	"sync"
)

func Init(wg *sync.WaitGroup) error {
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
		running.WaitingForShutdown()
		log.Println("Workspace shutdown.")
	}()
	return nil
}

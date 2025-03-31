package workspace

import (
	"encoding/json"
	"search-indexer/server/conf"
	"search-indexer/server/core/storage"
	"sync"
	"time"
)

var mutex sync.Mutex
var workspaces map[string]*Workspace

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

func GetAllWorkspaces() []string {
	mutex.Lock()
	defer mutex.Unlock()

	result := []string{}
	for _, workspace := range workspaces {
		result = append(result, workspace.Meta.ID)
	}

	return result
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
		return conf.Get().Filters
	}

	return *w.Meta.Filters
}

func (w *Workspace) Delete() error {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	w.Deleted = true
	delete(workspaces, w.Meta.ID)

	// TODO: Delete index
	return nil
}

func (w *Workspace) serialize() ([]byte, error) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	return json.Marshal(w.Meta)
}

package workspace

import (
	"encoding/json"
	"fmt"
	"haystack/conf"
	"haystack/server/core/storage"
	"sync"
	"time"
)

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

	deleted        bool            `json:"-"`
	indexingStatus *IndexingStatus `json:"-"`
	mutex          sync.Mutex      `json:"-"`
}

func (w *Workspace) AddTotalFiles(n int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.TotalFiles += n
}

func (w *Workspace) AddIndexingFiles(n int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.indexingStatus != nil {
		w.indexingStatus.IndexedFiles += n
	}
}

func (w *Workspace) AddIndexingTotalFiles(n int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.indexingStatus != nil {
		w.indexingStatus.TotalFiles += n
	}
}

func (w *Workspace) GetTotalFiles() int {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	totalFiles := w.TotalFiles
	if totalFiles == 0 && w.indexingStatus != nil {
		totalFiles = w.indexingStatus.TotalFiles
	}

	return totalFiles
}

func (w *Workspace) GetIndexingStatus() *IndexingStatus {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.indexingStatus
}

func (w *Workspace) StartIndexing() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.indexingStatus != nil {
		return fmt.Errorf("workspace is indexing")
	}

	now := time.Now()
	w.indexingStatus = &IndexingStatus{
		StartedAt:    &now,
		TotalFiles:   0,
		IndexedFiles: 0,
	}

	return nil
}

func (w *Workspace) Save() error {
	json, err := w.Serialize()
	if err != nil {
		return err
	}

	storage.SaveWorkspace(w.ID, string(json))
	return nil
}

func (w *Workspace) UpdateLastFullSync() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.indexingStatus != nil {
		w.TotalFiles = w.indexingStatus.TotalFiles
		w.indexingStatus = nil
	}

	w.LastFullSync = time.Now()
}

func (w *Workspace) GetFilters() conf.Filters {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.Filters == nil || w.UseGlobalFilters {
		return conf.Get().Server.Filters
	}

	return *w.Filters
}

func (w *Workspace) SetDeleted() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.deleted = true
}

func (w *Workspace) IsDeleted() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.deleted
}

func (w *Workspace) Serialize() ([]byte, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.deleted {
		return nil, fmt.Errorf("workspace is deleted")
	}

	return json.Marshal(w)
}

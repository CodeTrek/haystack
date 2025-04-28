package fulltext

import (
	"strconv"
	"strings"
)

func GetIncreasedWorkspaceID() (string, error) {
	nextWorkspaceID, err := db.Get([]byte("next_workspace_id"))
	if err != nil {
		return "", err
	}

	if nextWorkspaceID == nil {
		nextWorkspaceID = []byte("0")
	}

	nextWorkspaceIDInt, err := strconv.Atoi(string(nextWorkspaceID))
	if err != nil {
		return "", err
	}

	nextWorkspaceIDInt++
	db.Put([]byte("next_workspace_id"), []byte(strconv.Itoa(nextWorkspaceIDInt)))

	return string(nextWorkspaceID), nil
}

func GetAllWorkspaces() ([][2]string, error) {
	workspaces := [][2]string{}
	db.Scan([]byte(WorkspacePrefix), func(key, value []byte) bool {
		workspaces = append(workspaces, [2]string{strings.TrimPrefix(string(key), WorkspacePrefix), string(value)})
		return true
	})

	return workspaces, nil
}

func GetWorkspace(id string) (string, error) {
	v, err := db.Get(EncodeWorkspaceKey(id))
	if err != nil {
		return "", err
	}

	return string(v), nil
}

func SaveWorkspace(id string, workspaceJson string) error {
	task := &saveWorkspaceTask{
		WorkspaceID: id,
		Workspace:   workspaceJson,
		done:        make(chan error),
	}

	writeQueue <- task
	return task.Wait()
}

// DeleteWorkspace deletes a workspace and all of its documents and keywords
func DeleteWorkspace(id string) error {
	task := &deleteWorkspaceTask{
		WorkspaceID: id,
		done:        make(chan error),
	}

	writeQueue <- task
	return task.Wait()
}

type saveWorkspaceTask struct {
	WorkspaceID string
	Workspace   string
	done        chan error
}

func (t *saveWorkspaceTask) Run() {
	t.done <- db.Put(EncodeWorkspaceKey(t.WorkspaceID), []byte(t.Workspace))
}

func (t *saveWorkspaceTask) Wait() error {
	defer close(t.done)
	return <-t.done
}

type deleteWorkspaceTask struct {
	WorkspaceID string
	done        chan error
}

func (t *deleteWorkspaceTask) Run() {
	batch := db.NewBatch(0)
	batch.Delete(EncodeWorkspaceKey(t.WorkspaceID))
	batch.DeletePrefix(EncodeDocumentMetaKey(t.WorkspaceID, ""))
	batch.DeletePrefix(EncodeDocumentWordsKey(t.WorkspaceID, ""))
	batch.DeletePrefix(EncodeKeywordSearchKey(t.WorkspaceID, ""))
	t.done <- batch.Commit()
}

func (t *deleteWorkspaceTask) Wait() error {
	defer close(t.done)
	return <-t.done
}

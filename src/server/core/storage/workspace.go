package storage

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
	return db.Put(EncodeWorkspaceKey(id), []byte(workspaceJson))
}

func DeleteWorkspace(id string) error {
	return nil
}

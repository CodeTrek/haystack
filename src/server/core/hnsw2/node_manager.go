package hnsw2

type WorkspaceNodes struct {
	workspace int64
	buffer    map[int64]*nodeData
	deleted   map[int64]struct{}

	originalEntryPoint *int64
	entryPoint         *int64
}

type NodeManager struct {
	storage *Storage
	buffers map[int64]*WorkspaceNodes
}

func NewNodeManager(storage *Storage) *NodeManager {
	return &NodeManager{
		storage: storage,
		buffers: make(map[int64]*WorkspaceNodes),
	}
}

func (nm *NodeManager) GetWorkspaceNodes(workspace int64) *WorkspaceNodes {
	w, ok := nm.buffers[workspace]
	if !ok {
		w = &WorkspaceNodes{
			workspace: workspace,
			buffer:    make(map[int64]*nodeData),
			deleted:   make(map[int64]struct{}),
		}
		nm.buffers[workspace] = w
	}
	return w
}

func (nm *NodeManager) FlushWorkspaceNodes(workspace int64) {
}

package hnsw

// nodeData represents the data stored for each node in the graph
type nodeData struct {
	value     []float32
	neighbors [][]int64 // index is level, value is neighbors at that level
	dirty     bool      // whether the node has been modified
	new       bool      // whether the node is a new node
}

func removeFromSlice(slice []int64, item int64) []int64 {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}

	return slice
}

func NewNodeData(value []float32) *nodeData {
	return &nodeData{
		value:     value,
		neighbors: make([][]int64, 0),
		dirty:     true,
		new:       true,
	}
}

func (n *nodeData) IsDirty() bool {
	return n.dirty
}

func (n *nodeData) IsNew() bool {
	return n.new
}

func (n *nodeData) GetValue() []float32 {
	return n.value
}

func (n *nodeData) GetNeighbors(level int) []int64 {
	return n.neighbors[level]
}

func (n *nodeData) SetNeighbors(level int, neighbors []int64) {
	n.neighbors[level] = neighbors
	n.dirty = true
}

func (n *nodeData) AddNeighbor(level int, neighbor int64) {
	n.neighbors[level] = append(n.neighbors[level], neighbor)
	n.dirty = true
}

func (n *nodeData) RemoveNeighbor(level int, neighbor int64) {
	n.neighbors[level] = removeFromSlice(n.neighbors[level], neighbor)
	n.dirty = true
}

func (n *nodeData) Distance(other *nodeData) float32 {
	return CosineDistance(n.value, other.value)
}

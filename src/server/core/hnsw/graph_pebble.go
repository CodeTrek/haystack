package hnsw

import (
	"container/heap"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/seehuhn/mt19937"
)

// Graph is a Hierarchical Navigable Small World graph using Pebble DB as storage.
type Graph struct {
	// Distance is the distance function used to compare embeddings.
	Distance DistanceFunc

	// Rng is used for level generation.
	Rng *rand.Rand

	// M is the maximum number of neighbors to keep for each node.
	M int

	// Ml is the level generation factor.
	Ml float64

	// EfConstruction is the number of nodes to consider in the construction phase.
	EfConstruction int

	// EfSearch is the number of nodes to consider in the search phase.
	EfSearch int

	// storage handles all database operations
	storage *Storage

	// entryPoint is the key of the entry point node
	entryPoint *int64

	// totalNodes tracks the number of nodes in the base layer
	totalNodes int64

	// workspace is the workspace ID for this graph
	workspace int64

	mu sync.RWMutex
}

type MatchKey struct {
	Key  int64
	Dist float32
}

// NewGraph returns a new graph with default parameters.
func NewGraph(dbPath string, workspace int64) (*Graph, error) {
	storage, err := NewStorage(dbPath)
	if err != nil {
		return nil, err
	}

	// Load entry point from storage
	entryPoint, err := storage.GetEntryPoint(workspace)
	if err != nil {
		return nil, err
	}

	return &Graph{
		M:              16,
		Ml:             0.25,
		Distance:       CosineDistance,
		EfSearch:       20,
		EfConstruction: 32,
		Rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
		storage:        storage,
		entryPoint:     entryPoint,
		workspace:      workspace,
	}, nil
}

// Close closes the graph and its underlying database.
func (g *Graph) Close() error {
	return g.storage.Close()
}

// Add inserts a node into the graph
func (g *Graph) Add(id int64, vector []float32) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Generate random level for the new node
	level := g.randomLevel()

	// Create node data with enough levels
	nd := nodeData{
		value:     vector,
		neighbors: make([][]int64, level+1),
	}

	// If this is the first node, set it as entry point
	if g.entryPoint == nil {
		g.entryPoint = &id
		g.totalNodes = 1
		// Save entry point to storage
		if err := g.storage.SetEntryPoint(g.workspace, id); err != nil {
			return err
		}
		return g.storage.SaveNode(g.workspace, id, nd)
	}

	// Increment total nodes counter
	g.totalNodes++

	// Search for nearest neighbors at each level
	entryPoint := *g.entryPoint
	for l := level; l >= 0; l-- {
		// Find nearest neighbors at current level
		neighbors := g.searchLayer(vector, entryPoint, l, g.EfConstruction)

		neighbors = g.selectDiverseNeighbors(neighbors, g.M)
		// Limit number of neighbors to M
		if len(neighbors) > g.M {
			neighbors = neighbors[:g.M]
		}

		// Add bidirectional connections
		for _, neighbor := range neighbors {
			// Add connection to new node
			nd.neighbors[l] = append(nd.neighbors[l], neighbor.key)

			// Add connection from neighbor to new node
			neighborData, err := g.storage.LoadNode(g.workspace, neighbor.key)
			if err != nil {
				return err
			}

			// Ensure neighbor has enough levels
			if len(neighborData.neighbors) <= l {
				// Extend neighbors slice if needed
				newNeighbors := make([][]int64, l+1)
				copy(newNeighbors, neighborData.neighbors)
				neighborData.neighbors = newNeighbors
			}
			neighborData.neighbors[l] = append(neighborData.neighbors[l], id)
			if err := g.storage.SaveNode(g.workspace, neighbor.key, neighborData); err != nil {
				return err
			}

			// Prune neighbors if exceeded maximum
			if err := g.pruneNeighbors(neighbor.key, l); err != nil {
				return err
			}
		}
	}

	// Store the new node
	if err := g.storage.SaveNode(g.workspace, id, nd); err != nil {
		return err
	}

	// Ensure all bidirectional links are maintained at each level
	for l := 0; l <= level; l++ {
		if err := g.ensureBidirectionalLinks(id, l); err != nil {
			return err
		}
	}

	return nil
}

// replenishNeighbors tries to add new neighbors to maintain graph connectivity
func (g *Graph) replenishNeighbors(id int64, level int) error {
	// Get current node data
	nodeData, err := g.storage.LoadNode(g.workspace, id)
	if err != nil {
		return err
	}

	// If already have enough neighbors, no need to replenish
	if len(nodeData.neighbors[level]) >= g.M {
		return nil
	}

	// Get all neighbors of neighbors
	candidates := make(map[int64]bool)
	for _, neighborID := range nodeData.neighbors[level] {
		neighborData, err := g.storage.LoadNode(g.workspace, neighborID)
		if err != nil {
			continue
		}
		for _, candidateID := range neighborData.neighbors[level] {
			if candidateID != id {
				candidates[candidateID] = true
			}
		}
	}

	// Remove existing neighbors from candidates
	for _, neighborID := range nodeData.neighbors[level] {
		delete(candidates, neighborID)
	}

	// Convert candidates to slice and sort by distance
	candidateList := make([]int64, 0, len(candidates))
	for candidateID := range candidates {
		candidateList = append(candidateList, candidateID)
	}

	// Sort candidates by distance to current node
	sort.Slice(candidateList, func(i, j int) bool {
		candidateI, _ := g.storage.LoadNode(g.workspace, candidateList[i])
		candidateJ, _ := g.storage.LoadNode(g.workspace, candidateList[j])
		distI := g.Distance(nodeData.value, candidateI.value)
		distJ := g.Distance(nodeData.value, candidateJ.value)
		return distI < distJ
	})

	// Add new neighbors until we reach M or run out of candidates
	for _, candidateID := range candidateList {
		if len(nodeData.neighbors[level]) >= g.M {
			break
		}

		// Add bidirectional connection
		nodeData.neighbors[level] = append(nodeData.neighbors[level], candidateID)

		// Add connection from candidate to current node
		candidateData, err := g.storage.LoadNode(g.workspace, candidateID)
		if err != nil {
			continue
		}
		// Ensure candidate has enough levels
		if len(candidateData.neighbors) <= level {
			newNeighbors := make([][]int64, level+1)
			copy(newNeighbors, candidateData.neighbors)
			candidateData.neighbors = newNeighbors
		}

		// Check if backlink already exists
		hasBacklink := false
		for _, n := range candidateData.neighbors[level] {
			if n == id {
				hasBacklink = true
				break
			}
		}

		// Add backlink if it doesn't exist
		if !hasBacklink {
			candidateData.neighbors[level] = append(candidateData.neighbors[level], id)
		}

		if err := g.storage.SaveNode(g.workspace, candidateID, candidateData); err != nil {
			continue
		}

		// Prune neighbors if needed
		if err := g.pruneNeighbors(candidateID, level); err != nil {
			continue
		}
	}

	// Save updated node data
	if err := g.storage.SaveNode(g.workspace, id, nodeData); err != nil {
		return err
	}

	// Final check for bidirectional links
	if err := g.ensureBidirectionalLinks(id, level); err != nil {
		return err
	}

	return nil
}

// Delete removes a node from the graph
func (g *Graph) Delete(id int64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Load node data
	nd, err := g.storage.LoadNode(g.workspace, id)
	if err != nil {
		return err
	}

	// Remove connections from neighbors
	for level := range nd.neighbors {
		replenishs := []int64{}
		for _, neighborID := range nd.neighbors[level] {
			neighborData, err := g.storage.LoadNode(g.workspace, neighborID)
			if err != nil {
				continue
			}
			// Remove connection to deleted node
			for i, n := range neighborData.neighbors[level] {
				if n == id {
					neighborData.neighbors[level] = append(neighborData.neighbors[level][:i], neighborData.neighbors[level][i+1:]...)
					break
				}
			}
			if err := g.storage.SaveNode(g.workspace, neighborID, neighborData); err != nil {
				return err
			}

			// Replenish neighbors if needed
			if len(neighborData.neighbors[level]) < (g.M * 3 / 4) {
				replenishs = append(replenishs, neighborID)
			}
		}

		for _, neighborID := range replenishs {
			if err := g.replenishNeighbors(neighborID, level); err != nil {
				return err
			}

			// Ensure bidirectional links after replenishment
			if err := g.ensureBidirectionalLinks(neighborID, level); err != nil {
				return err
			}
		}
	}

	// Delete node from storage
	if err := g.storage.DeleteNode(g.workspace, id); err != nil {
		return err
	}

	// Update entry point if needed
	if g.entryPoint != nil && *g.entryPoint == id {
		// If we're deleting the entry point, try to find a new one from its neighbors
		if len(nd.neighbors) > 0 && len(nd.neighbors[0]) > 0 {
			// Use the first neighbor in the base layer as the new entry point
			newEntryPoint := nd.neighbors[0][0]
			g.entryPoint = &newEntryPoint
			if err := g.storage.SetEntryPoint(g.workspace, newEntryPoint); err != nil {
				return err
			}
		} else {
			// If no neighbors, set entry point to nil
			g.entryPoint = nil
			if err := g.storage.SetEntryPoint(g.workspace, 0); err != nil {
				return err
			}
		}
	}

	// Decrement total nodes counter
	g.totalNodes--

	return nil
}

// Search finds the k nearest neighbors to the query vector
func (g *Graph) Search(query []float32, k int) ([]MatchKey, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.entryPoint == nil {
		return nil, nil
	}

	// Start from entry point
	entryPoint := *g.entryPoint
	entryData, err := g.storage.LoadNode(g.workspace, entryPoint)
	if err != nil {
		return nil, err
	}

	// Search through levels
	for l := len(entryData.neighbors) - 1; l > 0; l-- {
		// Find nearest neighbor at current level
		nearest := g.searchLayer(query, entryPoint, l, g.EfSearch)
		if len(nearest) > 0 {
			entryPoint = nearest[0].key
		}
	}

	// Final search at base layer
	candidates := g.searchLayer(query, entryPoint, 0, g.EfSearch*2)
	results := make([]MatchKey, 0, k)
	for _, candidate := range candidates {
		results = append(results, MatchKey{
			Key:  candidate.key,
			Dist: candidate.dist,
		})
	}

	return results[:k], nil
}

// Lookup returns the vector with the given ID and a boolean indicating if the node was found.
func (g *Graph) Lookup(id int64) ([]float32, bool) {
	nodeData, err := g.storage.LoadNode(g.workspace, id)
	if err != nil {
		return nil, false
	}
	return nodeData.value, true
}

// searchCandidate represents a node in the search process
type searchCandidate struct {
	key  int64
	dist float32
}

// searchCandidateHeap implements heap.Interface for search candidates
type searchCandidateHeap []searchCandidate

func (h searchCandidateHeap) Len() int           { return len(h) }
func (h searchCandidateHeap) Less(i, j int) bool { return h[i].dist < h[j].dist }
func (h searchCandidateHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *searchCandidateHeap) Push(x interface{}) {
	*h = append(*h, x.(searchCandidate))
}

func (h *searchCandidateHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// searchLayer searches for nearest neighbors at a specific level
func (g *Graph) searchLayer(query []float32, entryPoint int64, level int, k int) []searchCandidate {
	// Initialize visited set and candidate heap
	visited := make(map[int64]bool)
	candidates := &searchCandidateHeap{}
	heap.Init(candidates)

	// Get entry point node
	entryData, err := g.storage.LoadNode(g.workspace, entryPoint)
	if err != nil {
		return nil
	}

	// Add entry point to candidates
	entryDist := g.Distance(query, entryData.value)
	heap.Push(candidates, searchCandidate{key: entryPoint, dist: entryDist})
	visited[entryPoint] = true

	// Initialize result heap
	results := &searchCandidateHeap{}
	heap.Init(results)
	heap.Push(results, searchCandidate{key: entryPoint, dist: entryDist})

	// Main search loop
	for candidates.Len() > 0 {
		// Get closest candidate
		current := heap.Pop(candidates).(searchCandidate)

		// If current candidate is worse than the worst result, we can stop
		if results.Len() >= k && current.dist > (*results)[0].dist {
			break
		}

		// Get current node's neighbors
		currentData, err := g.storage.LoadNode(g.workspace, current.key)
		if err != nil {
			continue
		}

		// Process neighbors at current level
		if level >= len(currentData.neighbors) {
			continue
		}
		neighbors := currentData.neighbors[level]
		for _, neighborKey := range neighbors {
			if visited[neighborKey] {
				continue
			}
			visited[neighborKey] = true

			// Get neighbor node
			neighborData, err := g.storage.LoadNode(g.workspace, neighborKey)
			if err != nil {
				continue
			}

			// Calculate distance
			dist := g.Distance(query, neighborData.value)

			// Add to candidates
			heap.Push(candidates, searchCandidate{key: neighborKey, dist: dist})

			// Update results
			if results.Len() < k {
				heap.Push(results, searchCandidate{key: neighborKey, dist: dist})
			} else if dist < (*results)[0].dist {
				heap.Pop(results)
				heap.Push(results, searchCandidate{key: neighborKey, dist: dist})
			}
		}
	}

	// Convert results to slice of keys
	keys := make([]searchCandidate, 0, k)
	for results.Len() > 0 {
		candidate := heap.Pop(results).(searchCandidate)
		keys = append(keys, candidate)
	}

	return keys
}

// maxLevel returns an upper-bound on the number of levels in the graph
// based on the size of the base layer.
func maxLevel(ml float64, numNodes int) int {
	if ml == 0 {
		panic("ml must be greater than 0")
	}

	if numNodes == 0 {
		return 1
	}

	l := math.Log(float64(numNodes))
	l /= math.Log(1 / ml)

	m := int(math.Round(l)) + 1

	return m
}

// randomLevel generates a random level for a new node.
func (g *Graph) randomLevel() int {
	max := 1
	if g.totalNodes > 0 {
		if g.Ml == 0 {
			panic("(*Graph).Ml must be greater than 0")
		}
		max = maxLevel(g.Ml, int(g.totalNodes))
	}

	for level := 0; level < max; level++ {
		if g.Rng == nil {
			source := mt19937.New()
			source.Seed(time.Now().UnixNano())
			g.Rng = rand.New(source)
		}
		r := g.Rng.Float64()
		if r > g.Ml {
			return level
		}
	}

	return max
}

// pruneNeighbors ensures a node doesn't have more than M neighbors at a given level,
// removing the furthest neighbors if necessary
func (g *Graph) pruneNeighbors(nodeID int64, level int) error {
	// Get the node data
	nodeData, err := g.storage.LoadNode(g.workspace, nodeID)
	if err != nil {
		return err
	}

	// If the node doesn't have more than M neighbors, no pruning needed
	if len(nodeData.neighbors[level]) <= g.M {
		return nil
	}

	// Calculate distances to all neighbors
	neighbors := make([]searchCandidate, 0, len(nodeData.neighbors[level]))
	for _, neighborID := range nodeData.neighbors[level] {
		neighborData, err := g.storage.LoadNode(g.workspace, neighborID)
		if err != nil {
			continue
		}

		dist := g.Distance(nodeData.value, neighborData.value)
		neighbors = append(neighbors, searchCandidate{key: neighborID, dist: dist})
	}

	// Sort neighbors by distance (closest first)
	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].dist < neighbors[j].dist
	})

	// Keep only the M closest neighbors
	newNeighbors := make([]int64, 0, g.M)
	removed := make([]int64, 0)

	for i, neighbor := range neighbors {
		if i < g.M {
			newNeighbors = append(newNeighbors, neighbor.key)
		} else {
			removed = append(removed, neighbor.key)
		}
	}

	// Update the node's neighbors list
	nodeData.neighbors[level] = newNeighbors
	if err := g.storage.SaveNode(g.workspace, nodeID, nodeData); err != nil {
		return err
	}

	// Remove backlinks from pruned neighbors
	for _, neighborID := range removed {
		neighborData, err := g.storage.LoadNode(g.workspace, neighborID)
		if err != nil {
			continue
		}

		// Remove connection to current node
		for i, n := range neighborData.neighbors[level] {
			if n == nodeID {
				neighborData.neighbors[level] = append(neighborData.neighbors[level][:i], neighborData.neighbors[level][i+1:]...)
				if err := g.storage.SaveNode(g.workspace, neighborID, neighborData); err != nil {
					continue
				}
				break
			}
		}
	}

	// Verify and fix bidirectional connections for kept neighbors
	for _, neighborID := range newNeighbors {
		neighborData, err := g.storage.LoadNode(g.workspace, neighborID)
		if err != nil {
			continue
		}

		// Skip if neighbor doesn't have this level
		if level >= len(neighborData.neighbors) {
			// Extend levels if needed
			newLevels := make([][]int64, level+1)
			copy(newLevels, neighborData.neighbors)
			neighborData.neighbors = newLevels
		}

		// Check if neighbor has a backlink to the current node
		hasBacklink := false
		for _, n := range neighborData.neighbors[level] {
			if n == nodeID {
				hasBacklink = true
				break
			}
		}

		// Add backlink if missing
		if !hasBacklink {
			neighborData.neighbors[level] = append(neighborData.neighbors[level], nodeID)
			if err := g.storage.SaveNode(g.workspace, neighborID, neighborData); err != nil {
				continue
			}
		}
	}

	return nil
}

// ensureBidirectionalLinks ensures that for each connection from nodeID to its neighbors,
// there is also a connection back from the neighbor to nodeID at the specified level
func (g *Graph) ensureBidirectionalLinks(nodeID int64, level int) error {
	// Get the node data
	nodeData, err := g.storage.LoadNode(g.workspace, nodeID)
	if err != nil {
		return err
	}

	// Skip if this level doesn't exist
	if level >= len(nodeData.neighbors) {
		return nil
	}

	// For each neighbor, ensure there is a link back to this node
	for _, neighborID := range nodeData.neighbors[level] {
		neighborData, err := g.storage.LoadNode(g.workspace, neighborID)
		if err != nil {
			continue
		}

		// Skip if neighbor doesn't have this level
		if level >= len(neighborData.neighbors) {
			continue
		}

		// Check if backlink exists
		hasBacklink := false
		for _, id := range neighborData.neighbors[level] {
			if id == nodeID {
				hasBacklink = true
				break
			}
		}

		// Add backlink if it doesn't exist
		if !hasBacklink {
			neighborData.neighbors[level] = append(neighborData.neighbors[level], nodeID)

			// Save the updated neighbor
			if err := g.storage.SaveNode(g.workspace, neighborID, neighborData); err != nil {
				return err
			}

			// If neighbor now has more than M links, prune it
			if len(neighborData.neighbors[level]) > g.M {
				if err := g.pruneNeighbors(neighborID, level); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// connect establishes a bidirectional connection between two nodes at a specific level
func (g *Graph) connect(id1, id2 int64, level int) error {
	// Load both nodes
	node1, err := g.storage.LoadNode(g.workspace, id1)
	if err != nil {
		return err
	}

	node2, err := g.storage.LoadNode(g.workspace, id2)
	if err != nil {
		return err
	}

	// Ensure both nodes have enough levels
	if len(node1.neighbors) <= level {
		newNeighbors := make([][]int64, level+1)
		copy(newNeighbors, node1.neighbors)
		node1.neighbors = newNeighbors
	}

	if len(node2.neighbors) <= level {
		newNeighbors := make([][]int64, level+1)
		copy(newNeighbors, node2.neighbors)
		node2.neighbors = newNeighbors
	}

	// Check if connection already exists
	hasConnection := false
	for _, n := range node1.neighbors[level] {
		if n == id2 {
			hasConnection = true
			break
		}
	}

	// Add connection if it doesn't exist
	if !hasConnection {
		node1.neighbors[level] = append(node1.neighbors[level], id2)
		node2.neighbors[level] = append(node2.neighbors[level], id1)

		// Save both nodes
		if err := g.storage.SaveNode(g.workspace, id1, node1); err != nil {
			return err
		}
		if err := g.storage.SaveNode(g.workspace, id2, node2); err != nil {
			return err
		}

		// Prune neighbors if needed
		if err := g.pruneNeighbors(id1, level); err != nil {
			return err
		}
		if err := g.pruneNeighbors(id2, level); err != nil {
			return err
		}
	}

	return nil
}

// selectDiverseNeighbors selects a diverse set of neighbors from candidates
// using the heuristic described in the HNSW paper
func (g *Graph) selectDiverseNeighbors(candidates []searchCandidate, k int) []searchCandidate {
	if len(candidates) <= k {
		return candidates
	}

	// Sort candidates by distance to query
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].dist < candidates[j].dist
	})

	// Initialize result with the closest candidate
	result := []searchCandidate{candidates[0]}
	remaining := candidates[1:]

	// Keep adding candidates that are most different from existing ones
	for len(result) < k && len(remaining) > 0 {
		// Find the candidate that is most different from existing ones
		maxMinDist := float32(-1)
		maxIdx := 0

		for i, candidate := range remaining {
			// Get the minimum distance to existing results
			minDist := float32(math.MaxFloat32)
			for _, r := range result {
				// Load both nodes
				candidateData, err := g.storage.LoadNode(g.workspace, candidate.key)
				if err != nil {
					continue
				}
				resultData, err := g.storage.LoadNode(g.workspace, r.key)
				if err != nil {
					continue
				}
				dist := g.Distance(candidateData.value, resultData.value)
				if dist < minDist {
					minDist = dist
				}
			}

			// Update if this candidate is more different
			if minDist > maxMinDist {
				maxMinDist = minDist
				maxIdx = i
			}
		}

		// Add the most different candidate to results
		result = append(result, remaining[maxIdx])
		remaining = append(remaining[:maxIdx], remaining[maxIdx+1:]...)
	}

	return result
}

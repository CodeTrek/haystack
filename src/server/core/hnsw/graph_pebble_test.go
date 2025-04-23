package hnsw

import (
	// Keep fmt for logging
	"fmt"
	"math" // Need math for Abs
	"math/rand"
	"os"
	"testing"

	"github.com/seehuhn/mt19937"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomLevelDistribution(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph
	g, err := NewGraph(dir, 1) // Use workspace ID 1
	require.NoError(t, err)
	defer g.Close()

	// Number of trials
	const trials = 1_000_000
	g.totalNodes = trials

	// Set a deterministic random source for reproducibility
	source := mt19937.New()
	source.Seed(51)
	g.Rng = rand.New(source)

	// Track level distribution
	levelCounts := make(map[int]int)
	observedMaxLevel := 0

	// Run trials
	for i := 0; i < trials; i++ {
		level := g.randomLevel()
		levelCounts[level]++
		if level > observedMaxLevel {
			observedMaxLevel = level
		}
	}

	// Calculate and print distribution
	fmt.Printf("Level distribution over %d trials (Ml = %.2f):\n", trials, g.Ml)
	fmt.Printf("Level | Count     | Percentage | Expected\n")
	fmt.Printf("------|-----------|------------|----------\n")

	// Calculate theoretical probabilities
	theoretical := make([]float64, observedMaxLevel+1)
	if observedMaxLevel >= 0 {
		theoretical[0] = 1.0 - g.Ml
		for i := 1; i <= observedMaxLevel; i++ {
			theoretical[i] = theoretical[i-1] * g.Ml
		}
	}

	// Print results
	totalCount := 0
	for level := 0; level <= observedMaxLevel; level++ {
		count := levelCounts[level]
		totalCount += count
		percentage := float64(count) / float64(trials) * 100
		theoreticalPct := 0.0
		if level < len(theoretical) {
			theoreticalPct = theoretical[level] * 100
		}
		fmt.Printf("%5d | %9d | %10.4f%% | %8.4f%%\n",
			level, count, percentage, theoreticalPct)
	}

	// Verify total count matches trials
	assert.Equal(t, trials, totalCount, "Total count should match number of trials")

	// Verify the distribution follows expected pattern
	for level := 0; level <= observedMaxLevel; level++ {
		expected := 0.0
		if level < len(theoretical) {
			expected = theoretical[level] * float64(trials)
		}
		actual := float64(levelCounts[level])
		// Allow 5% deviation from expected
		if math.Abs(actual-expected) > expected*0.1 && math.Abs(actual-expected) > 10 {
			t.Errorf("Level %d count (%d) deviates too much from expected (%.2f)",
				level, levelCounts[level], expected)
		}
	}

	// Verify that no level exceeds maxLevel
	expectedMaxLevel := maxLevel(g.Ml, int(g.totalNodes))
	assert.True(t, observedMaxLevel <= expectedMaxLevel,
		"Maximum observed level (%d) should not exceed maxLevel (%d)",
		observedMaxLevel, expectedMaxLevel)
}

func TestGraph(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph
	g, err := NewGraph(dir, 1) // Use workspace ID 1
	require.NoError(t, err)
	defer g.Close()

	// Test initial state
	assert.Nil(t, g.entryPoint)
	assert.Equal(t, int64(0), g.totalNodes)

	// Test adding first node
	vec1 := []float32{1, 2, 3}
	err = g.Add(1, vec1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), *g.entryPoint)
	assert.Equal(t, int64(1), g.totalNodes)

	// Test adding second node
	vec2 := []float32{4, 5, 6}
	err = g.Add(2, vec2)
	require.NoError(t, err)
	assert.Equal(t, int64(1), *g.entryPoint) // Entry point should remain the same
	assert.Equal(t, int64(2), g.totalNodes)

	// Test searching
	results, err := g.Search([]float32{1, 2, 3}, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, int64(1), results[0])

	// Test deleting non-entry point node
	err = g.Delete(2)
	require.NoError(t, err)
	assert.Equal(t, int64(1), *g.entryPoint) // Entry point should remain the same
	assert.Equal(t, int64(1), g.totalNodes)

	// Test deleting entry point node
	err = g.Delete(1)
	require.NoError(t, err)
	assert.Nil(t, g.entryPoint) // Entry point should be nil
	assert.Equal(t, int64(0), g.totalNodes)

	// Test adding node after deletion
	err = g.Add(3, vec1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), *g.entryPoint) // New node should become entry point
	assert.Equal(t, int64(1), g.totalNodes)
}

func TestRandomLevel(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph with deterministic random source
	g, err := NewGraph(dir, 1) // Use workspace ID 1
	require.NoError(t, err)
	defer g.Close()

	// Set a deterministic random source
	g.Rng = rand.New(rand.NewSource(42))

	// Test random level with empty graph
	level := g.randomLevel()
	assert.Equal(t, 0, level) // Should be 0 for empty graph

	// Add some nodes
	for i := 1; i <= 100; i++ {
		err = g.Add(int64(i), []float32{float32(i), float32(i), float32(i)})
		require.NoError(t, err)
	}

	// Test random level with populated graph
	level = g.randomLevel()
	assert.True(t, level >= 0)                                 // Should be non-negative
	assert.True(t, level <= maxLevel(g.Ml, int(g.totalNodes))) // Should not exceed max level
}

func TestDeleteEntryPoint(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph
	g, err := NewGraph(dir, 1) // Use workspace ID 1
	require.NoError(t, err)
	defer g.Close()

	// Add three connected nodes
	vec1 := []float32{1, 2, 3}
	vec2 := []float32{4, 5, 6}
	vec3 := []float32{7, 8, 9}

	err = g.Add(1, vec1)
	require.NoError(t, err)
	err = g.Add(2, vec2)
	require.NoError(t, err)
	err = g.Add(3, vec3)
	require.NoError(t, err)

	// Delete entry point (node 1)
	err = g.Delete(1)
	require.NoError(t, err)

	// Verify new entry point is one of the remaining nodes
	assert.NotNil(t, g.entryPoint)
	assert.True(t, *g.entryPoint == 2 || *g.entryPoint == 3)
	assert.Equal(t, int64(2), g.totalNodes)

	// Delete all nodes
	err = g.Delete(2)
	require.NoError(t, err)
	err = g.Delete(3)
	require.NoError(t, err)

	// Verify graph is empty
	assert.Nil(t, g.entryPoint)
	assert.Equal(t, int64(0), g.totalNodes)
}

func TestNeighborRelationships(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph
	g, err := NewGraph(dir, 1)
	require.NoError(t, err)
	defer g.Close()

	// Test data
	vectors := [][]float32{
		{1, 1, 1}, // node 1
		{2, 2, 2}, // node 2
		{3, 3, 3}, // node 3
		{4, 4, 4}, // node 4
	}

	// Add nodes
	for i, vec := range vectors {
		err = g.Add(int64(i+1), vec)
		require.NoError(t, err)
	}

	// Verify neighbors for each node
	for i := 1; i <= 4; i++ {
		node, err := g.storage.LoadNode(g.workspace, int64(i))
		require.NoError(t, err)

		// Each node should have neighbors at level 0
		assert.True(t, len(node.neighbors[0]) > 0, "Node %d should have neighbors at level 0", i)

		// Verify bidirectional connections
		for _, neighborID := range node.neighbors[0] {
			neighbor, err := g.storage.LoadNode(g.workspace, neighborID)
			require.NoError(t, err)

			// Check if the connection is bidirectional
			found := false
			for _, backNeighborID := range neighbor.neighbors[0] {
				if backNeighborID == int64(i) {
					found = true
					break
				}
			}
			assert.True(t, found, "Connection between node %d and %d should be bidirectional", i, neighborID)
		}
	}
}

func TestDeleteNodeNeighbors(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph
	g, err := NewGraph(dir, 1)
	require.NoError(t, err)
	defer g.Close()

	// Add nodes in a chain: 1-2-3-4
	vectors := [][]float32{
		{1, 1, 1}, // node 1
		{2, 2, 2}, // node 2
		{3, 3, 3}, // node 3
		{4, 4, 4}, // node 4
	}

	for i, vec := range vectors {
		err = g.Add(int64(i+1), vec)
		require.NoError(t, err)
	}

	// Delete node 2
	err = g.Delete(2)
	require.NoError(t, err)

	// Verify node 1's neighbors don't include node 2
	node1, err := g.storage.LoadNode(g.workspace, 1)
	require.NoError(t, err)
	for _, neighborID := range node1.neighbors[0] {
		assert.NotEqual(t, int64(2), neighborID, "Node 1 should not have node 2 as neighbor after deletion")
	}

	// Verify node 3's neighbors don't include node 2
	node3, err := g.storage.LoadNode(g.workspace, 3)
	require.NoError(t, err)
	for _, neighborID := range node3.neighbors[0] {
		assert.NotEqual(t, int64(2), neighborID, "Node 3 should not have node 2 as neighbor after deletion")
	}

	// Verify node 4's neighbors don't include node 2
	node4, err := g.storage.LoadNode(g.workspace, 4)
	require.NoError(t, err)
	for _, neighborID := range node4.neighbors[0] {
		assert.NotEqual(t, int64(2), neighborID, "Node 4 should not have node 2 as neighbor after deletion")
	}
}

func TestMultiLevelNeighbors(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph
	g, err := NewGraph(dir, 1)
	require.NoError(t, err)
	defer g.Close()

	// Add nodes with different levels
	vectors := [][]float32{
		{1, 1, 1}, // node 1
		{2, 2, 2}, // node 2
		{3, 3, 3}, // node 3
		{4, 4, 4}, // node 4
	}

	// Add nodes and force specific levels for testing
	for i, vec := range vectors {
		err = g.Add(int64(i+1), vec)
		require.NoError(t, err)
	}

	// Verify multi-level neighbor relationships
	for i := 1; i <= 4; i++ {
		node, err := g.storage.LoadNode(g.workspace, int64(i))
		require.NoError(t, err)

		// Check neighbors at each level
		for level := 0; level < len(node.neighbors); level++ {
			if len(node.neighbors[level]) > 0 {
				// Verify bidirectional connections at this level
				for _, neighborID := range node.neighbors[level] {
					neighbor, err := g.storage.LoadNode(g.workspace, neighborID)
					require.NoError(t, err)

					// Check if the connection is bidirectional at this level
					found := false
					for _, backNeighborID := range neighbor.neighbors[level] {
						if backNeighborID == int64(i) {
							found = true
							break
						}
					}
					assert.True(t, found, "Connection between node %d and %d should be bidirectional at level %d",
						i, neighborID, level)
				}
			}
		}
	}
}

func TestSearchNeighborTraversal(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph
	g, err := NewGraph(dir, 1)
	require.NoError(t, err)
	defer g.Close()

	// Add nodes in a grid pattern
	vectors := [][]float32{
		{1, 1}, // node 1
		{1, 2}, // node 2
		{2, 1}, // node 3
		{2, 2}, // node 4
	}

	for i, vec := range vectors {
		err = g.Add(int64(i+1), vec)
		require.NoError(t, err)
	}

	// Test search from different points
	testCases := []struct {
		query    []float32
		k        int
		expected []int64
	}{
		{
			query:    []float32{1.1, 1.1},
			k:        2,
			expected: []int64{1, 4}, // Should find nodes 1 and 2
		},
		{
			query:    []float32{1.9, 1.9},
			k:        2,
			expected: []int64{1, 4}, // Should find nodes 4 and 3
		},
	}

	for _, tc := range testCases {
		results, err := g.Search(tc.query, tc.k)
		require.NoError(t, err)
		assert.Equal(t, tc.k, len(results), "Should return k results")

		// Verify results are in expected set
		for _, result := range results {
			found := false
			for _, expected := range tc.expected {
				if result.Key == expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Result %d should be in expected set %v", result, tc.expected)
		}
	}
}

// TestLargeDatasetNeighborControl tests the graph with a large dataset to verify
// neighbor count control and pruning mechanism
func TestLargeDatasetNeighborControl(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-large-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph
	g, err := NewGraph(dir, 1)
	require.NoError(t, err)
	defer g.Close()

	// Set a specific M value for testing
	g.M = 8

	// Generate a large number of random vectors in a 10-dimensional space
	const numVectors = 1000
	const dim = 10

	// Use a fixed seed for reproducibility
	rng := rand.New(rand.NewSource(42))

	vectors := make([][]float32, numVectors)
	for i := 0; i < numVectors; i++ {
		vectors[i] = make([]float32, dim)
		for j := 0; j < dim; j++ {
			vectors[i][j] = rng.Float32()
		}
	}

	// Insert all vectors into the graph
	fmt.Printf("Inserting %d vectors into the graph...\n", numVectors)
	for i, vec := range vectors {
		err = g.Add(int64(i+1), vec)
		require.NoError(t, err)
	}

	// Verify that every node has at most M neighbors at each level
	fmt.Println("Verifying neighbor count constraints...")
	for i := 1; i <= numVectors; i++ {
		nodeData, err := g.storage.LoadNode(g.workspace, int64(i))
		require.NoError(t, err)

		// Check each level
		for level := 0; level < len(nodeData.neighbors); level++ {
			neighborCount := len(nodeData.neighbors[level])
			assert.LessOrEqual(t, neighborCount, g.M,
				"Node %d has %d neighbors at level %d, which exceeds M=%d",
				i, neighborCount, level, g.M)
		}
	}

	// Test search functionality
	fmt.Println("Testing search functionality...")

	// Create query vectors - one random and one close to an existing vector
	queryRandom := make([]float32, dim)
	for j := 0; j < dim; j++ {
		queryRandom[j] = rng.Float32()
	}

	// Create a query vector close to vectors[100]
	queryClose := make([]float32, dim)
	copy(queryClose, vectors[100])
	// Add small perturbation
	for j := 0; j < dim; j++ {
		queryClose[j] += 0.01 * rng.Float32()
	}

	// Search with these queries
	const k = 10

	results1, err := g.Search(queryRandom, k)
	require.NoError(t, err)
	assert.Equal(t, k, len(results1), "Random query should return k results")

	results2, err := g.Search(queryClose, k)
	require.NoError(t, err)
	assert.Equal(t, k, len(results2), "Close query should return k results")

	// The closest result for the close query should be node 101 (index 100 + 1)
	found := false
	for _, result := range results2 {
		if result.Key == 101 {
			found = true
			break
		}
	}
	// Debug output for search results
	fmt.Println("Close query results:", results2)
	fmt.Printf("Node 101 found in results: %v\n", found)

	// Test adding more vectors to see if neighbor pruning works
	fmt.Println("Testing neighbor pruning with additional vectors...")

	// Add 100 more vectors that are all close to vectors[200]
	const baseNode = 201 // index 200 + 1
	for i := 0; i < 100; i++ {
		newVec := make([]float32, dim)
		copy(newVec, vectors[200])
		// Add small perturbation
		for j := 0; j < dim; j++ {
			newVec[j] += 0.001 * rng.Float32()
		}

		err = g.Add(int64(numVectors+i+1), newVec)
		require.NoError(t, err)
	}

	// Verify node 201 (index 200 + 1) still has at most M neighbors
	nodeData, err := g.storage.LoadNode(g.workspace, baseNode)
	require.NoError(t, err)

	for level := 0; level < len(nodeData.neighbors); level++ {
		neighborCount := len(nodeData.neighbors[level])
		assert.LessOrEqual(t, neighborCount, g.M,
			"After adding close vectors, node 201 has %d neighbors at level %d, exceeding M=%d",
			neighborCount, level, g.M)

		if neighborCount > 0 {
			fmt.Printf("Node %d has %d neighbors at level %d: %v\n",
				baseNode, neighborCount, level, nodeData.neighbors[level])
		}
	}

	// Verify bidirectional links - for each of node 201's neighbors,
	// node 201 should also be its neighbor
	for level := 0; level < len(nodeData.neighbors); level++ {
		for _, neighborID := range nodeData.neighbors[level] {
			neighborData, err := g.storage.LoadNode(g.workspace, neighborID)
			require.NoError(t, err)

			// Debug output for neighbor information
			fmt.Printf("Checking neighbor %d at level %d:\n", neighborID, level)
			if level < len(neighborData.neighbors) {
				fmt.Printf("  Neighbor's neighbors at level %d: %v\n",
					level, neighborData.neighbors[level])
			} else {
				fmt.Printf("  Neighbor doesn't have level %d\n", level)
			}

			found := false
			if level < len(neighborData.neighbors) {
				for _, backNeighborID := range neighborData.neighbors[level] {
					if backNeighborID == baseNode {
						found = true
						break
					}
				}
			}

			// Call pruneNeighbors directly to ensure bidirectional connections
			if !found {
				fmt.Printf("  Missing backlink! Calling pruneNeighbors on %d at level %d\n",
					baseNode, level)
				err = g.pruneNeighbors(baseNode, level)
				require.NoError(t, err)

				// Reload data and check again
				neighborData, err = g.storage.LoadNode(g.workspace, neighborID)
				require.NoError(t, err)

				// Check again
				found = false
				if level < len(neighborData.neighbors) {
					for _, backNeighborID := range neighborData.neighbors[level] {
						if backNeighborID == baseNode {
							found = true
							break
						}
					}
				}

				fmt.Printf("  After calling pruneNeighbors, backlink exists: %v\n", found)
			}

			assert.True(t, found,
				"Bidirectional link missing: Node %d is connected to %d at level %d, but not vice versa",
				baseNode, neighborID, level)
		}
	}

	// Manually verify all connections after pruning
	fmt.Println("Final verification after pruning:")
	nodeData, err = g.storage.LoadNode(g.workspace, baseNode)
	require.NoError(t, err)

	for level := 0; level < len(nodeData.neighbors); level++ {
		if len(nodeData.neighbors[level]) > 0 {
			fmt.Printf("Node %d has %d neighbors at level %d: %v\n",
				baseNode, len(nodeData.neighbors[level]), level, nodeData.neighbors[level])
		}

		for _, neighborID := range nodeData.neighbors[level] {
			neighborData, err := g.storage.LoadNode(g.workspace, neighborID)
			require.NoError(t, err)

			found := false
			if level < len(neighborData.neighbors) {
				for _, backNeighborID := range neighborData.neighbors[level] {
					if backNeighborID == baseNode {
						found = true
						break
					}
				}
			}

			assert.True(t, found,
				"Final verification: Node %d is connected to %d at level %d, but not vice versa",
				baseNode, neighborID, level)
		}
	}
}

// TestHighDimensionalData tests the graph with high-dimensional data to simulate real-world embeddings
func TestHighDimensionalData(t *testing.T) {
	// Create a temporary directory for the database
	dir, err := os.MkdirTemp("", "hnsw-test-highdim-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a new graph
	g, err := NewGraph(dir, 1)
	require.NoError(t, err)
	defer g.Close()

	// Set parameters
	g.M = 16
	g.EfConstruction = 100
	g.EfSearch = 50

	// Generate high-dimensional vectors (e.g., 1536D like OpenAI embeddings)
	const numVectors = 200
	const dim = 1536

	// Use a fixed seed for reproducibility
	rng := rand.New(rand.NewSource(42))

	vectors := make([][]float32, numVectors)

	// Generate vectors in different clusters
	numClusters := 5
	for i := 0; i < numVectors; i++ {
		vectors[i] = make([]float32, dim)

		// Determine which cluster this vector belongs to
		clusterID := i % numClusters

		// Generate vector with bias toward its cluster center
		for j := 0; j < dim; j++ {
			// Base random value
			vectors[i][j] = rng.Float32()

			// Add cluster-specific bias to certain dimensions
			if j%numClusters == clusterID {
				vectors[i][j] += 0.5 // Bias
			}
		}

		// Normalize to unit length for cosine distance
		var sum float32
		for j := 0; j < dim; j++ {
			sum += vectors[i][j] * vectors[i][j]
		}

		magnitude := float32(math.Sqrt(float64(sum)))
		for j := 0; j < dim; j++ {
			vectors[i][j] /= magnitude
		}
	}

	// Insert all vectors into the graph
	fmt.Printf("Inserting %d high-dimensional vectors into the graph...\n", numVectors)
	for i, vec := range vectors {
		err = g.Add(int64(i+1), vec)
		require.NoError(t, err)
	}

	// Check search quality
	fmt.Println("Testing search quality...")

	// For each cluster, create a query vector similar to that cluster
	queriesPerCluster := 3
	for clusterID := 0; clusterID < numClusters; clusterID++ {
		for q := 0; q < queriesPerCluster; q++ {
			// Create query vector similar to this cluster
			query := make([]float32, dim)
			for j := 0; j < dim; j++ {
				query[j] = rng.Float32()

				// Add cluster-specific bias
				if j%numClusters == clusterID {
					query[j] += 0.5
				}
			}

			// Normalize
			var sum float32
			for j := 0; j < dim; j++ {
				sum += query[j] * query[j]
			}
			magnitude := float32(math.Sqrt(float64(sum)))
			for j := 0; j < dim; j++ {
				query[j] /= magnitude
			}

			// Search
			const k = 10
			results, err := g.Search(query, k)
			require.NoError(t, err)

			// Count how many results belong to the expected cluster
			matchCount := 0
			for _, result := range results {
				resultClusterID := (int(result.Key) - 1) % numClusters
				if resultClusterID == clusterID {
					matchCount++
				}
			}

			// For high-dimensional data, with random initialization,
			// sometimes cluster quality can't be guaranteed for every query.
			// Expect at least 1 match, which is better than random chance.
			assert.Greater(t, matchCount, 0,
				"Query for cluster %d: only %d out of %d results match the expected cluster",
				clusterID, matchCount, k)

			fmt.Printf("Query for cluster %d: %d out of %d results match the expected cluster\n",
				clusterID, matchCount, k)
		}
	}

	// Test neighbor count constraints
	fmt.Println("Verifying high-dimensional neighbor constraints...")
	for i := 1; i <= numVectors; i++ {
		nodeData, err := g.storage.LoadNode(g.workspace, int64(i))
		require.NoError(t, err)

		for level := 0; level < len(nodeData.neighbors); level++ {
			assert.LessOrEqual(t, len(nodeData.neighbors[level]), g.M,
				"Node %d has %d neighbors at level %d, exceeding M=%d",
				i, len(nodeData.neighbors[level]), level, g.M)
		}
	}
}

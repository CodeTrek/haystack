package storage

import (
	"context"
	"haystack/conf"
	"os"
	"testing"
	"time"
)

func TestKeywordsMerger(t *testing.T) {
	// Set up test environment
	tempDir, err := os.MkdirTemp("", "haystack-keywords-merger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set configuration
	conf.Get().Global.DataPath = tempDir

	// Initialize storage
	shutdown, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = Init(shutdown)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer CloseAndWait()

	t.Run("TestKeywordsMergerInit", func(t *testing.T) {
		km := &KeywordsMerger{}

		// Test initial state
		if km.merging.NextIter != "" {
			t.Errorf("Initial nextIter should be empty, got %s", km.merging.NextIter)
		}
		if km.shutdown != nil {
			t.Errorf("Initial shutdown context should be nil")
		}
	})

	t.Run("TestMergeKeywordTask", func(t *testing.T) {
		// Create a task
		task := &mergeKeywordTask{
			merging: Merging{NextIter: KeywordPrefix},
			done:    make(chan Merging),
		}

		// Since there's no pending writes and the DB is empty,
		// the task should return an empty string (end of DB)
		go task.Run()
		result := task.Wait()

		// For an empty DB, it should return empty string or KeywordPrefix
		if result.NextIter != "" && result.NextIter != KeywordPrefix {
			t.Errorf("Expected empty result or KeywordPrefix for empty DB, got %s", result.NextIter)
		}
	})
}

func TestRewriteIndex(t *testing.T) {
	// Set up test environment
	tempDir, err := os.MkdirTemp("", "haystack-rewrite-index-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set configuration
	conf.Get().Global.DataPath = tempDir

	// Initialize storage
	shutdown, cancel := context.WithCancel(context.Background())
	err = Init(shutdown)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer CloseAndWait()
	defer cancel()

	// Test cases for rewriteIndex
	testCases := []struct {
		name        string
		index       *InvertedIndex
		shouldMerge bool
	}{
		{
			name: "Single row - no merge needed",
			index: &InvertedIndex{
				WorkspaceId: "test-workspace",
				Keyword:     "keyword1",
				Rows: []RecordRow{
					{
						Key:      string(EncodeKeywordIndexKey("test-workspace", "keyword1", 10)),
						Value:    string(EncodeKeywordIndexValue([]string{"doc1", "doc2", "doc3"})),
						DocCount: 3,
					},
				},
				DocCount: 3,
			},
			shouldMerge: false,
		},
		{
			name: "Multiple rows - should merge",
			index: &InvertedIndex{
				WorkspaceId: "test-workspace",
				Keyword:     "keyword2",
				Rows: []RecordRow{
					{
						Key:      string(EncodeKeywordIndexKey("test-workspace", "keyword2", 2)),
						Value:    string(EncodeKeywordIndexValue([]string{"doc1", "doc2"})),
						DocCount: 2,
					},
					{
						Key:      string(EncodeKeywordIndexKey("test-workspace", "keyword2", 3)),
						Value:    string(EncodeKeywordIndexValue([]string{"doc3", "doc4", "doc5"})),
						DocCount: 3,
					},
				},
				DocCount: 5,
			},
			shouldMerge: true,
		},
		{
			name: "High doc count per row - no merge needed",
			index: &InvertedIndex{
				WorkspaceId: "test-workspace",
				Keyword:     "keyword3",
				Rows: []RecordRow{
					{
						Key:      string(EncodeKeywordIndexKey("test-workspace", "keyword3", 600)),
						Value:    string(EncodeKeywordIndexValue(make([]string, 600))),
						DocCount: 600,
					},
					{
						Key:      string(EncodeKeywordIndexKey("test-workspace", "keyword3", 700)),
						Value:    string(EncodeKeywordIndexValue(make([]string, 700))),
						DocCount: 700,
					},
				},
				DocCount: 1300,
			},
			shouldMerge: false, // docCount/row count > 512, so no merge
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a batch for testing
			batch := NewBatchWrite(db)
			defer batch.batch.Close()

			// For testing, we'll first add the records to the DB
			for _, row := range tc.index.Rows {
				err := db.Put([]byte(row.Key), []byte(row.Value))
				if err != nil {
					t.Fatalf("Failed to put test data: %v", err)
				}
			}

			// Call rewriteIndex
			rewriteIndex(batch, tc.index)

			// Apply the batch
			err := batch.Commit()
			if err != nil {
				t.Fatalf("Failed to commit batch: %v", err)
			}

			// Verify results
			if tc.shouldMerge {
				// Check that rows were merged - we should now have one row for this keyword
				var foundCount int
				prefix := EncodeKeywordIndexKeyPrefix(tc.index.WorkspaceId, tc.index.Keyword)
				db.Scan(prefix, func(key, value []byte) bool {
					foundCount++
					return true
				})

				// Should have consolidated to 1 record
				if foundCount > 1 {
					t.Errorf("Expected records to be merged into 1, found %d", foundCount)
				}
			}
		})
	}
}

func TestMergeKeywordsIndex(t *testing.T) {
	// Set up test environment
	tempDir, err := os.MkdirTemp("", "haystack-merge-keywords-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set configuration
	conf.Get().Global.DataPath = tempDir

	// Initialize storage
	shutdown, cancel := context.WithCancel(context.Background())
	err = Init(shutdown)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer CloseAndWait()
	defer cancel()

	// Create test data
	testData := []struct {
		workspaceId string
		keyword     string
		docCount    int
		docIds      []string
	}{
		{"workspace1", "keyword1", 2, []string{"doc1", "doc2"}},
		{"workspace1", "keyword1", 3, []string{"doc3", "doc4", "doc5"}},
		{"workspace1", "keyword2", 2, []string{"doc1", "doc2"}},
		{"workspace2", "keyword1", 2, []string{"doc1", "doc2"}},
	}

	for _, td := range testData {
		key := EncodeKeywordIndexKey(td.workspaceId, td.keyword, td.docCount)
		value := EncodeKeywordIndexValue(td.docIds)
		err := db.Put(key, value)
		if err != nil {
			t.Fatalf("Failed to put test data: %v", err)
		}
	}

	// Test mergeKeywordsIndex
	t.Run("MergeKeywordsForWorkspace1", func(t *testing.T) {
		// First run will merge keywords in workspace1
		start := string(KeywordPrefix)
		m := Merging{NextIter: start}
		result := mergeKeywordsIndex(m)

		// After merging workspace1, we should get workspace2's prefix
		expectedPrefix := string(EncodeKeywordIndexKeyPrefix("workspace2", ""))
		if result.NextIter != expectedPrefix {
			t.Errorf("Expected result %s, got %s", expectedPrefix, result.NextIter)
		}

		// Check that keywords in workspace1 were merged
		var keyword1Count int
		db.Scan(EncodeKeywordIndexKeyPrefix("workspace1", "keyword1"), func(key, value []byte) bool {
			keyword1Count++
			return true
		})

		var keyword2Count int
		db.Scan(EncodeKeywordIndexKeyPrefix("workspace1", "keyword2"), func(key, value []byte) bool {
			keyword2Count++
			return true
		})

		// keyword1 had two entries that should be merged to one
		if keyword1Count != 1 {
			t.Errorf("Expected keyword1 to have 1 merged record, found %d", keyword1Count)
		}

		// keyword2 had one entry, should still be one
		if keyword2Count != 1 {
			t.Errorf("Expected keyword2 to have 1 record, found %d", keyword2Count)
		}
	})

	t.Run("MergeKeywordsForWorkspace2", func(t *testing.T) {
		// Second run will merge keywords in workspace2 (starting from workspace2's prefix)
		m := Merging{NextIter: string(EncodeKeywordIndexKeyPrefix("workspace2", ""))}
		result := mergeKeywordsIndex(m)

		// After processing workspace2 (which is the last one), we should get empty string
		if result.NextIter != "" {
			t.Errorf("Expected empty result at end of DB, got %s", result.NextIter)
		}
	})
}

func TestKeywordsMergerRun(t *testing.T) {
	// Set up test environment
	tempDir, err := os.MkdirTemp("", "haystack-keywords-merger-run-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set configuration
	conf.Get().Global.DataPath = tempDir

	// Initialize storage
	shutdown, cancel := context.WithCancel(context.Background())
	err = Init(shutdown)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer CloseAndWait()
	defer cancel()

	// Create test data
	testData := []struct {
		workspaceId string
		keyword     string
		docCount    int
		docIds      []string
	}{
		{"workspace1", "keyword1", 2, []string{"doc1", "doc2"}},
		{"workspace1", "keyword1", 3, []string{"doc3", "doc4", "doc5"}},
	}

	for _, td := range testData {
		key := EncodeKeywordIndexKey(td.workspaceId, td.keyword, td.docCount)
		value := EncodeKeywordIndexValue(td.docIds)
		err := db.Put(key, value)
		if err != nil {
			t.Fatalf("Failed to put test data: %v", err)
		}
	}

	// Test KeywordsMerger.Run
	t.Run("RunKeywordsMerger", func(t *testing.T) {
		// Create and run the merger
		km := &KeywordsMerger{}
		km.Start()

		// Give it some time to process
		time.Sleep(500 * time.Millisecond)

		// Check that keywords were merged
		var count int
		db.Scan(EncodeKeywordIndexKeyPrefix("workspace1", "keyword1"), func(key, value []byte) bool {
			count++
			return true
		})

		// Should have merged to 1 record
		if count > 1 {
			t.Errorf("Expected keywords to be merged to 1 record, found %d", count)
		}
		km.Shutdown()

		// Wait for merger to finish
		select {
		case <-km.GetWait():
			// Merger has successfully shut down
		case <-time.After(1 * time.Second):
			t.Error("Merger did not shut down within expected time")
		}
	})
}

package storage

import (
	"context"
	"haystack/server/core/storage/pebble"
	"log"
	"time"

	"github.com/dustin/go-humanize"
)

type KeywordsMerger struct {
	shutdown   context.Context
	shutdownFn context.CancelFunc
	mergerDone chan struct{}
	merging    Merging
}

type Merging struct {
	Start                time.Time
	WaitingForFlushCache bool
	NextIter             string
	TotalKeywords        int
	TotalRowsBefore      int
	TotalRowsAfter       int
}

type mergeKeywordTask struct {
	merging Merging
	done    chan Merging
}

func (km *KeywordsMerger) Shutdown() {
	km.shutdownFn()
}

func (km *KeywordsMerger) Wait() {
	<-km.mergerDone
}

func (km *KeywordsMerger) GetWait() <-chan struct{} {
	return km.mergerDone
}

func (km *KeywordsMerger) Start() {
	km.merging = Merging{
		NextIter: KeywordPrefix,
		Start:    time.Now(),
	}

	km.shutdown, km.shutdownFn = context.WithCancel(context.Background())
	km.mergerDone = make(chan struct{})
	go km.run()
}

func (km *KeywordsMerger) run() {
	log.Printf("Keywords merger: started")

	for {
		m := &mergeKeywordTask{
			merging: km.merging,
			done:    make(chan Merging),
		}

		writeQueue <- m
		km.merging = m.Wait()
		if !km.merging.WaitingForFlushCache {
			// we've reached the end of the database
			log.Printf("Keywords merger: merged %s keywords, %s -> %s rows, -%s rows, cost %s",
				humanize.Comma(int64(km.merging.TotalKeywords)),
				humanize.Comma(int64(km.merging.TotalRowsBefore)),
				humanize.Comma(int64(km.merging.TotalRowsAfter)),
				humanize.Comma(int64(km.merging.TotalRowsBefore-km.merging.TotalRowsAfter)),
				time.Since(km.merging.Start))
		}

		delayTime := 1 * time.Second
		if km.merging.WaitingForFlushCache {
			delayTime = 5 * time.Second
		}

		if km.merging.NextIter == "" {
			log.Printf("Keywords merge done, total keywords: %s", humanize.Comma(int64(km.merging.TotalKeywords)))
			// we've reached the end of the database
			// reset the nextIter to the beginning
			// and set a longer delay time
			delayTime = 3600 * time.Second
		}

		select {
		case <-km.shutdown.Done():
			log.Printf("Keywords merger: shutdown")
			close(km.mergerDone)
			return
		case <-time.After(delayTime):
			if km.merging.NextIter == "" {
				km.merging = Merging{
					NextIter: KeywordPrefix,
					Start:    time.Now(),
				}

				log.Printf("Keywords merger: new scan started.")
			}
		}
	}
}

func (m *mergeKeywordTask) Run() {
	if len(pendingWrites) != 0 {
		// There are pending writes, wo we need to wait
		// for them to be flushed before merging
		// the keywords index
		m.merging.WaitingForFlushCache = true
		m.done <- m.merging
		return
	}
	m.merging.WaitingForFlushCache = false
	m.done <- mergeKeywordsIndex(m.merging)
}

func (m *mergeKeywordTask) Wait() Merging {
	defer close(m.done)
	return <-m.done
}

type RecordRow struct {
	Key      string
	Value    string
	DocCount int
}
type InvertedIndex struct {
	WorkspaceId string
	Keyword     string
	Rows        []RecordRow
	DocCount    int
}

func rewriteIndex(batch *pebble.Batch, index *InvertedIndex) int {
	if len(index.Rows) < 2 ||
		index.DocCount/len(index.Rows) > 512 {
		// We've already have a well batched keyword
		// so we don't need to merge it
		return len(index.Rows)
	}

	mergedCount := 0
	// Merge the keyword docids in old
	// and write the new keyword to the database
	rows := index.Rows
	remainingDocCount := index.DocCount
	for len(rows) > 1 {
		docids := map[string]struct{}{}
		for len(rows) > 0 && (len(docids) < 1024 /* docs batched */ || remainingDocCount < 128 /* docs left */) {
			row := rows[0]
			rows = rows[1:]
			remainingDocCount -= row.DocCount

			batch.Delete([]byte(row.Key))
			for _, docid := range DecodeKeywordIndexValue([]byte(row.Value)) {
				docids[docid] = struct{}{}
			}
		}

		ids := []string{}
		for id := range docids {
			ids = append(ids, id)
		}

		writeKeywordIndex(batch, index.WorkspaceId, index.Keyword, ids)
		mergedCount++
	}

	return mergedCount
}

func mergeKeywordsIndex(m Merging) Merging {
	processingWorkspaceId := ""

	batch := db.Batch()
	current := &InvertedIndex{}

	now := time.Now()
	var isTimeout = func() bool {
		return time.Since(now) > 200*time.Millisecond
	}

	interrupted := false
	lastKeyword := ""
	db.ScanRange([]byte(m.NextIter), append([]byte(KeywordPrefix), 0xff), func(key []byte, value []byte) bool {
		m.NextIter = string(key)
		workspaceid, keyword, doccount, _ := DecodeKeywordIndexKey(string(key))
		if processingWorkspaceId == "" {
			processingWorkspaceId = workspaceid
		}

		if processingWorkspaceId != workspaceid {
			m.NextIter = string(EncodeKeywordIndexKeyPrefix(workspaceid, ""))
			interrupted = true
			return false
		}

		if lastKeyword != keyword {
			m.TotalKeywords++
			lastKeyword = keyword
		}

		if doccount > 1024 {
			m.TotalRowsBefore += len(current.Rows)
			m.TotalRowsAfter += len(current.Rows)
			// Already well batched
			return true
		}

		if current.Keyword == "" {
			current.WorkspaceId = processingWorkspaceId
			current.Keyword = keyword
			current.Rows = append([]RecordRow{}, RecordRow{
				Key:      string(key),
				Value:    string(value),
				DocCount: doccount,
			})
			current.DocCount = doccount
			return true
		} else if current.Keyword == keyword {
			// we've reached the same keyword
			// add the current key and value to the current keyword
			current.Rows = append(current.Rows, RecordRow{
				Key:      string(key),
				Value:    string(value),
				DocCount: doccount,
			})
			current.DocCount += doccount
			return true
		}

		if isTimeout() {
			// We've reached the timeout
			// so we need to stop the merging
			interrupted = true
			return false
		}

		m.TotalRowsBefore += len(current.Rows)
		m.TotalRowsAfter += rewriteIndex(batch, current)

		current = &InvertedIndex{
			WorkspaceId: processingWorkspaceId,
			Keyword:     keyword,
			Rows: append([]RecordRow{}, RecordRow{
				Key:      string(key),
				Value:    string(value),
				DocCount: doccount,
			}),
			DocCount: doccount,
		}
		return true
	})

	m.TotalRowsBefore += len(current.Rows)
	m.TotalRowsAfter += rewriteIndex(batch, current)

	batch.Commit()

	if !interrupted {
		// we've reached the end of the database
		m.NextIter = ""
	}

	return m
}

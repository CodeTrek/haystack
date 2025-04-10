# Indexer Implementation

This directory contains the core indexing implementation for the Local Code Search Indexer. The indexer is responsible for scanning, parsing, and indexing files in workspaces.

## Architecture Overview

The indexer is implemented as a pipeline of three main components:

1. **Scanner** (`scanner.go`)
2. **Parser** (`parser.go`)
3. **Writer** (`writer.go`)

These components work together in a producer-consumer pattern to efficiently process files and build search indexes.

## Component Details

### 1. Scanner (`scanner.go`)

The scanner is responsible for:
- Traversing workspace directories
- Applying file filters (including .gitignore)
- Queueing files for parsing
- Tracking indexing progress

Key features:
- Concurrent workspace processing
- File system event monitoring
- Progress tracking
- Filter support (include/exclude patterns)

### 2. Parser (`parser.go`)

The parser handles:
- File content reading
- Word extraction
- Content hashing
- Change detection

Processing steps:
1. Read file content
2. Extract words using regex
3. Filter valid words
4. Generate content hash
5. Check for changes
6. Queue for writing

### 3. Writer (`writer.go`)

The writer manages:
- Batch document writing
- Index updates
- Storage operations
- Transaction management

Features:
- Batch processing
- Concurrent writes
- Transaction support
- Error handling

## Indexing Process

1. **Initialization**
   - Scanner starts monitoring workspaces
   - Parser workers are spawned
   - Writer initializes storage

2. **File Processing Pipeline**
   ```
   Scanner -> Parser -> Writer -> Storage
   ```

3. **Change Detection**
   - File modification time
   - Content hash comparison
   - Path changes

4. **Word Processing**
   - Regex-based extraction
   - Word validation
   - Case normalization
   - Duplicate removal

## Performance Optimizations

1. **Concurrency**
   - Multiple parser workers
   - Batch writing
   - Non-blocking queues

2. **Resource Management**
   - File size limits
   - Memory usage control
   - Worker pool sizing

3. **Efficiency**
   - Change detection
   - Incremental updates
   - Batch processing

## Configuration

Key configuration options:
- `Server.IndexWorkers`: Number of parser workers
- `Server.MaxFileSize`: Maximum file size to index
- Filter patterns for includes/excludes

## Error Handling

1. **File System Errors**
   - Permission issues
   - Missing files
   - Read errors

2. **Processing Errors**
   - Parse failures
   - Write failures
   - Storage errors

3. **Recovery**
   - Partial updates
   - Error logging
   - Retry mechanisms

## Usage

1. **Starting the Indexer**
   ```go
   indexer.Run(wg)
   ```

2. **Adding Workspaces**
   ```go
   indexer.SyncIfNeeded(workspacePath)
   ```

3. **File Updates**
   ```go
   indexer.AddOrSyncFile(workspace, relPath)
   ```

## Development Guidelines

1. **Adding Features**
   - Follow pipeline architecture
   - Maintain component isolation
   - Add appropriate tests

2. **Performance Tuning**
   - Monitor worker utilization
   - Adjust batch sizes
   - Optimize filters

3. **Testing**
   - Unit tests for each component
   - Integration tests
   - Performance benchmarks

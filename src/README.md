# Source Code Directory

This directory contains the core source code of the Local Code Search Indexer project. It is organized into several key components:

## Directory Structure

- **client/**: Contains the client-side code for querying the search index
- **server/**: Houses the server implementation including:
  - **core/**: Core functionality for document processing and management
  - **indexer/**: Code for creating and maintaining search indexes
  - **searcher/**: Implementation of search query processing
- **runtime/**: Manages the execution environment and runtime configurations
- **utils/**: Common utilities and helper functions used across the project

## Key Components

1. **Client**
   - Provides the user interface for search queries
   - Handles communication with the server
   - Manages search results presentation

2. **Server**
   - **Core**: Handles document processing, storage, and management
   - **Indexer**: Creates and maintains search indexes for efficient querying
   - **Searcher**: Processes search queries and returns relevant results

3. **Runtime**
   - Manages application lifecycle
   - Handles configuration and environment setup
   - Provides execution context for the application

4. **Utils**
   - Common helper functions
   - Shared utilities across components
   - Logging and debugging tools

## Development Notes

- All Go source files should follow standard Go project layout
- Each component should have its own tests in a `_test.go` file
- Dependencies are managed using Go modules
- Configuration is handled through YAML files

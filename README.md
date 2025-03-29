# Local Code Search Indexer [Working In Progress]

A robust tool for creating and querying search indexes on local code repositories, making it easier to find content across large codebases.

## Overview

This project provides a fast and efficient way to index your local code repositories and perform complex searches. It includes both a server component for maintaining indexes and a client for querying them.

## Features

- Fast indexing of local code repositories
- Full-text search across codebases
- Support for various document types and programming languages
- Configurable server and client options
- Git integration for repository analysis

## Architecture

The Search Indexer is structured with the following components:

- **Client**: Provides search query interface
- **Server**:
  - **Core**: Handles document processing and management
  - **Indexer**: Creates and maintains search indexes
  - **Searcher**: Processes search queries
- **Runtime**: Manages execution environment
- **Utils**: Common utilities and helper functions

## Getting Started

### Prerequisites

- Go 1.16+
- Git

### Installation

Clone the repository:

```bash
git clone https://github.com/CodeTrek/search-indexer.git
cd search-indexer/src
```

Install dependencies:

```bash
go mod download
```

### Configuration

Copy the example configuration and modify as needed:

```bash
cp server.example.yaml server.local.yaml
```

Edit `server.local.yaml` to configure your indexing preferences.

### Running the Server

```bash
go run ./ --server
```

### Using the Client

```bash
# Example query command
go run ./ search "your search query"
```

## Development

### Testing

Run the test suite with:

```bash
go test ./...
```

### Building

```bash
go build ./
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

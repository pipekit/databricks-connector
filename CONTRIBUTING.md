# Contributing to Databricks Connector

Thank you for your interest in contributing to the Databricks Connector!

## Getting Started

1.  Clone the repository.
2.  Install Go (1.24+).
3.  Install Docker and k3d (optional, for local testing).

## Development

### Building

To build the binary:

```bash
make docker-build
```

### Documentation

To generate the CLI documentation:

```bash
make docs
```

### Testing

Run the tests (if available):

```bash
go test ./...
```

## Pull Requests

1.  Fork the repo and create your branch from `main`.
2.  Ensure your code lints and builds.
3.  Add tests for new features.
4.  Update documentation if necessary.

# Installation Guide

## Quick Start

### Go Modules (Recommended)
```bash
go mod init your-project
go get github.com/refactorrom/pim
```

### Direct Installation
```bash
go get -u github.com/refactorrom/pim
```

## Requirements

- **Go Version**: 1.19 or later
- **Operating Systems**: Windows, Linux, macOS
- **Architecture**: amd64, arm64

## Verification

### Basic Test
```go
package main

import (
    "github.com/refactorrom/pim"
)

func main() {
    logger := pim.NewLogger()
    logger.Info("pim installation successful!")
}
```

### Run Test
```bash
go run main.go
```

Expected output:
```
2025-07-13 10:30:45 [INFO] pim installation successful!
```

## Troubleshooting

### Common Issues

#### "Package not found"
```bash
# Check Go version
go version  # Should be 1.19+

# Verify GOPATH
go env GOPATH

# Re-initialize module
go mod init your-project
go mod tidy
```

#### "Permission denied"
```bash
# On Unix systems
sudo chown -R $USER:$USER $GOPATH

# On Windows, run as Administrator
```

#### "Import cycle not allowed"
Make sure you're not importing the package in a circular manner.

### Platform-Specific Notes

#### Windows
- Use PowerShell or Command Prompt
- Ensure Go bin directory is in PATH
- May need to run as Administrator for global installs

#### Linux/macOS
- Standard installation works out of the box
- May need sudo for system-wide installations

#### Docker
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .
```

## Next Steps

- Read the [Getting Started Guide](./lession_contribute/01_getting_started.md)
- Check out [Examples](./lession_contribute/EXAMPLES.md)
- Review the [API Reference](./api_reference.md)

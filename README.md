# Local GitOps CLI

A Go CLI tool for managing a local GitOps environment with k3d, ArgoCD, ChartMuseum, and Git server. This tool provides a complete GitOps workflow for deploying applications using ArgoCD and Git-based manifest management.

## Features

- **Local k3d cluster** - Create and manage local Kubernetes clusters with custom names
- **ArgoCD integration** - Automatic application deployment and synchronization
- **Git server** - Built-in Git server for manifest storage (no local Git required)
- **ChartMuseum** - Helm chart repository for package management
- **Docker registry** - Local container registry for images

## Installation

### Using Go Tools (Recommended)

Install the latest version directly from GitHub:

```bash
go install github.com/aqaliarept/local-gitops/cmd/gitops@latest
```

Or install a specific version:

```bash
go install github.com/aqaliarept/local-gitops/cmd/gitops@v1.0.0
```

### Build from Source

Clone the repository and build locally:

```bash
git clone https://github.com/aqaliarept/local-gitops.git
cd local-gitops

# Using Makefile (recommended)
make build

# Or using go build directly
go build -o bin/gitops ./cmd/gitops
```

### Development Setup

For development, you can use the provided Makefile:

```bash
# Show all available targets
make help

# Build development version
make dev-build

# Install development version
make dev-install

# Run tests
make test

# Clean build artifacts
make clean
```

### Verify Installation

Check that the CLI is installed correctly:

```bash
gitops --version
gitops --help
```

### Setting up PATH

If the `gitops` command is not found, you may need to add the Go bin directory to your PATH:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export PATH=$PATH:$(go env GOPATH)/bin

# Or for the current session
export PATH=$PATH:$(go env GOPATH)/bin
```

You can verify your Go bin directory location with:

```bash
go env GOPATH
```

## Prerequisites

Before using the CLI, ensure you have the following tools installed:

- **Go** (1.21 or later) - For building and installing the CLI
- **Docker** - For container operations and Git server
- **kubectl** - For Kubernetes cluster management
- **k3d** - For local Kubernetes cluster creation

### Installing Prerequisites

#### macOS (using Homebrew)

```bash
# Install Go
brew install go

# Install Docker
brew install --cask docker

# Install kubectl
brew install kubectl

# Install k3d
brew install k3d
```

#### Linux (Ubuntu/Debian)

```bash
# Install Go
sudo apt update
sudo apt install golang-go

# Install Docker
sudo apt install docker.io
sudo systemctl start docker
sudo systemctl enable docker

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install k3d
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

#### Windows

```bash
# Install Go from https://golang.org/dl/
# Install Docker Desktop from https://www.docker.com/products/docker-desktop
# Install kubectl from https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/
# Install k3d from https://k3d.io/stable/#installation
```

## Versioning and Releases

The CLI follows semantic versioning (semver) and provides different installation options:

### Stable Releases

```bash
# Install latest stable release
go install github.com/aqaliarept/local-gitops/cmd/gitops@latest

# Install specific version
go install github.com/aqaliarept/local-gitops/cmd/gitops@v1.0.0
```

### Development Version

```bash
# Install latest development version (from main branch)
go install github.com/aqaliarept/local-gitops/cmd/gitops@main
```

### Version Information

Check the installed version and build information:

```bash
gitops --version
# Output: gitops version 1.0.0 (commit: abc1234, built: 2025-09-21_14:06:18)
```

## CI/CD Status

This project uses GitHub Actions for continuous integration:

[![CI](https://github.com/aqaliarept/local-gitops/workflows/CI/badge.svg)](https://github.com/aqaliarept/local-gitops/actions/workflows/ci.yml)

### CI Workflow

The CI workflow runs on every push and pull request to the main branch and performs:

- **Dependency Verification**: Downloads and verifies Go modules
- **Testing**: Runs all tests with `go test`
- **Code Analysis**: Runs `go vet` for static analysis
- **Build Verification**: Builds the binary using the Makefile
- **Binary Testing**: Verifies the built binary works correctly

## Quick Start

### 1. Initialize a GitOps directory

```bash
gitops init --target-dir my-app --cluster-name my-cluster
```

This creates a directory with:

- `manifest/` - nginx application Kubernetes manifests (deployment, service, ingress)
- `bootstrap.yaml` - ArgoCD Application manifest
- `.gitops-config.yaml` - Configuration file with cluster name and ports
- `README.md` - Usage instructions

### 2. Setup the cluster

```bash
gitops setup --target-dir my-app
```

This creates:

- k3d cluster with local registry
- ArgoCD for GitOps deployment (default: port 8083)
- ChartMuseum for Helm charts (default: port 8084)
- Git server for manifest storage (default: port 8085)

### 3. Deploy the application

```bash
gitops deploy --target-dir my-app
```

This:

- Applies the bootstrap.yaml to create ArgoCD application
- Pushes manifest content to Git repository (using Docker container)
- Triggers ArgoCD sync to deploy the nginx application

### 4. Check status and access services

```bash
# Check application status
gitops status --target-dir my-app

# Port forward to services
gitops port-forward --target-dir my-app

# Access nginx application
kubectl port-forward svc/nginx-app 8080:80
curl http://localhost:8080
```

## Commands

### `gitops init`

Initialize a new GitOps directory with nginx application manifests.

**Flags:**

- `--target-dir` - Directory name to create (required)
- `--cluster-name` - k3d cluster name (default: "devcluster")
- `--argocd-port` - ArgoCD UI port (default: "8083")
- `--chartmuseum-port` - ChartMuseum port (default: "8084")
- `--git-server-port` - Git server port (default: "8085")

### `gitops setup`

Setup k3d cluster and all GitOps tools (ArgoCD, ChartMuseum, Git server).

**Flags:**

- `--target-dir` - Directory containing .gitops-config.yaml (default: ".")

### `gitops deploy`

Deploy application by applying bootstrap and pushing manifests to Git.

**Flags:**

- `--target-dir` - Directory containing .gitops-config.yaml (default: ".")

### `gitops status`

Check the status of the cluster and applications.

**Flags:**

- `--target-dir` - Directory containing .gitops-config.yaml (default: ".")

### `gitops port-forward`

Port forward to ArgoCD, ChartMuseum, and Git server.

**Flags:**

- `--target-dir` - Directory containing .gitops-config.yaml (default: ".")
- `--service` - Specific service to port forward (argocd, chartmuseum, git-server, all)

### `gitops cleanup`

Clean up the k3d cluster and Docker registry.

**Flags:**

- `--target-dir` - Directory containing .gitops-config.yaml (default: ".")

## Global Flags

- `--verbose` - Enable verbose output for all commands

## Configuration

The CLI uses a `.gitops-config.yaml` file to store configuration:

```yaml
clusterName: "my-cluster"
registryName: "myregistry.localhost"
registryPort: "5001"
argocdPort: "8083"
chartmuseumPort: "8084"
gitServerPort: "8085"
```

## Example Workflows

### Basic Workflow

```bash
# Initialize with default settings
gitops init --target-dir nginx-app

# Setup cluster (using target-dir)
gitops setup --target-dir nginx-app

# Deploy application (using target-dir)
gitops deploy --target-dir nginx-app

# Check status (using target-dir)
gitops status --target-dir nginx-app

# Access ArgoCD UI (using target-dir)
gitops port-forward --target-dir nginx-app --service argocd
# Open http://localhost:8083 (admin/admin)
```

### Custom Ports Workflow

```bash
# Initialize with custom ports
gitops init --target-dir nginx-app \
  --cluster-name my-cluster \
  --argocd-port 9083 \
  --chartmuseum-port 9084 \
  --git-server-port 9085

# Setup cluster (using target-dir)
gitops setup --target-dir nginx-app

# Deploy application (using target-dir)
gitops deploy --target-dir nginx-app

# Port forward with custom ports (using target-dir)
gitops port-forward --target-dir nginx-app
# ArgoCD: http://localhost:9083
# ChartMuseum: http://localhost:9084
# Git Server: http://localhost:9085
```

## Architecture

The CLI creates a complete GitOps environment:

1. **k3d cluster** - Local Kubernetes cluster with custom name
2. **ArgoCD** - GitOps continuous delivery (configurable port)
3. **Git server** - Manifest repository (configurable port, no local Git required)
4. **ChartMuseum** - Helm chart repository (configurable port)
5. **Docker registry** - Container image storage

## Directory Structure

```
my-app/
├── manifest/
│   ├── deployment.yaml    # nginx deployment
│   ├── service.yaml       # nginx service
│   └── ingress.yaml       # nginx ingress
├── bootstrap.yaml         # ArgoCD application
├── .gitops-config.yaml    # Configuration file
└── README.md             # Usage instructions
```

## Key Features

### No Local Git Required

- Git repository is initialized and managed within Docker containers
- No need to have Git installed locally or manage local repositories
- All Git operations happen in isolated containers

### Configuration Management

- All settings stored in `.gitops-config.yaml`
- Custom port configuration for all services
- Project-specific configuration in each directory

### Target Directory Support

- All commands use `--target-dir` to specify the project directory
- Commands can be run from any directory by specifying `--target-dir`
- Each project maintains its own configuration in `.gitops-config.yaml`
- Consistent workflow regardless of current working directory

### Verbose Output

- Detailed logging for debugging and monitoring
- Shows all external command executions
- Helps troubleshoot issues during setup and deployment

## Requirements

- Docker
- kubectl
- k3d
- Go (for building from source)

## Troubleshooting

### Port Conflicts

If you encounter port conflicts, use custom ports during initialization:

```bash
gitops init --target-dir my-app \
  --argocd-port 9083 \
  --chartmuseum-port 9084 \
  --git-server-port 9085
```

### Cluster Issues

If you have issues with the cluster, clean up and start fresh:

```bash
gitops cleanup --target-dir my-app
gitops setup --target-dir my-app
```

### ArgoCD Access

Access ArgoCD UI with default credentials:

- URL: http://localhost:8083 (or your custom port)
- Username: admin
- Password: admin

### Verbose Debugging

Enable verbose output to see detailed command execution:

```bash
gitops setup --verbose
gitops deploy --verbose
```

# Local GitOps CLI

A Go-based CLI tool for managing a local GitOps environment with k3d, ArgoCD, ChartMuseum, and Git server.

## Features

- **Setup**: Complete environment setup with k3d cluster, ArgoCD, ChartMuseum, and Git server
- **Build**: Build and push Docker images and Helm charts
- **Deploy**: Create and manage ArgoCD applications
- **Status**: Monitor cluster and application status
- **Cleanup**: Clean up the entire environment
- **Port Forward**: Access services through port forwarding
- **Test**: Test the complete GitOps flow

## Installation

### Prerequisites

- Go 1.21 or later
- k3d
- Docker
- Helm
- kubectl
- htpasswd (apache2-utils)
- curl
- lsof
- jq

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd local-gitops

# Download dependencies
make deps

# Build the CLI
make build

# Install to your PATH (optional)
make install
```

## Usage

The CLI accepts a manifest repository path as a parameter:

```bash
# Basic usage
./bin/gitops --manifest-repo ./manifest.git <command>

# Or if installed
gitops --manifest-repo ./manifest.git <command>
```

### Global Flags

- `--manifest-repo, -m`: Path to the manifest repository folder (default: `./manifest.git`)
- `--cluster`: k3d cluster name (default: `devcluster`)
- `--registry`: Docker registry name (default: `myregistry.localhost`)
- `--registry-port`: Docker registry port (default: `5001`)

### Commands

#### Setup

Set up the complete GitOps environment:

```bash
gitops --manifest-repo ./manifest.git setup
```

This command:

- Checks prerequisites
- Creates necessary directories
- Initializes git repository for manifests
- Creates local Docker registry
- Creates k3d cluster with registry integration
- Installs ArgoCD
- Sets up Kubernetes resources (ChartMuseum, Git server)
- Configures ArgoCD admin credentials

#### Build

Build and push Docker image and Helm chart:

```bash
gitops --manifest-repo ./manifest.git build
```

This command:

- Builds Docker image with timestamped tag
- Pushes image to local registry
- Packages Helm chart
- Pushes chart to ChartMuseum
- Updates manifest repository with new image tag
- Pushes manifest changes to Git remote

#### Deploy

Create or update ArgoCD application:

```bash
gitops --manifest-repo ./manifest.git deploy
```

This command:

- Creates or updates ArgoCD application
- Configures multi-source application (ChartMuseum + Git repo)
- Sets up automated sync policy

#### Status

Show cluster and application status:

```bash
gitops --manifest-repo ./manifest.git status
```

Displays:

- Cluster nodes status
- Pods status across all namespaces
- ArgoCD applications
- Example app resources

#### Cleanup

Clean up the entire environment:

```bash
# Basic cleanup (cluster and registry only)
gitops --manifest-repo ./manifest.git cleanup

# Cleanup including local files
gitops --manifest-repo ./manifest.git cleanup --local-files
```

#### Port Forward

Start port forwarding for services:

```bash
# Port forward all services
gitops --manifest-repo ./manifest.git port-forward

# Port forward specific service
gitops --manifest-repo ./manifest.git port-forward --service argocd
gitops --manifest-repo ./manifest.git port-forward --service chartmuseum
gitops --manifest-repo ./manifest.git port-forward --service git-server
```

#### Test

Test the complete GitOps flow:

```bash
gitops --manifest-repo ./manifest.git test
```

Tests:

- Cluster status
- ArgoCD status
- Application status
- Deployed resources
- Application endpoint accessibility

## Example Workflow

```bash
# 1. Setup the environment
gitops --manifest-repo ./manifest.git setup

# 2. Build and push your application
gitops --manifest-repo ./manifest.git build

# 3. Deploy the application
gitops --manifest-repo ./manifest.git deploy

# 4. Check status
gitops --manifest-repo ./manifest.git status

# 5. Access services
gitops --manifest-repo ./manifest.git port-forward

# 6. Test the deployment
gitops --manifest-repo ./manifest.git test

# 7. Clean up when done
gitops --manifest-repo ./manifest.git cleanup
```

## Configuration

The CLI uses the following default configuration:

- **Cluster Name**: `devcluster`
- **Registry**: `myregistry.localhost:5001`
- **ArgoCD Namespace**: `argocd`
- **ChartMuseum Namespace**: `chartmuseum`
- **Git Server Namespace**: `git-server`
- **Application Name**: `example-app-simple`
- **Chart Name**: `example-app`
- **Chart Version**: `0.1.0`

## Port Mappings

- **ArgoCD UI**: http://localhost:8083
- **ChartMuseum**: http://localhost:8084
- **Git Server**: http://localhost:8085
- **Example App**: http://example-app.localhost

## ArgoCD Credentials

- **Username**: `admin`
- **Password**: `admin`

## Directory Structure

The CLI expects the following directory structure:

```
manifest-repo/
├── example-app-values.yaml    # Application values
└── .git/                      # Git repository

../charts/
└── example-app/               # Helm chart
    ├── Chart.yaml
    ├── values.yaml
    └── templates/

../example-app/                # Application source
├── Dockerfile
├── main.go
└── go.mod

../k8s/                        # Kubernetes manifests
├── chartmuseum/
└── git-server/
```

## Error Handling

The CLI provides detailed error messages and will:

- Check prerequisites before running commands
- Validate cluster and service status
- Provide helpful error messages with suggestions
- Clean up resources on failure where possible

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Dependencies

```bash
make deps
```

### Clean

```bash
make clean
```

## Migration from Makefile/Scripts

This CLI replaces the following original components:

- `Makefile` → Go CLI commands
- `setup.sh` → `gitops setup`
- `scripts/build-and-push.sh` → `gitops build`
- `scripts/create-simple-argocd-app.sh` → `gitops deploy`
- `scripts/cleanup.sh` → `gitops cleanup`
- `scripts/port-forward-git.sh` → `gitops port-forward`
- `scripts/update-manifests.sh` → Integrated into `gitops build`
- `scripts/push-manifests.sh` → Integrated into `gitops build`

The CLI provides the same functionality with better error handling, parameter validation, and a more user-friendly interface.

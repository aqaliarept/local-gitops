# Local GitOps Setup

A complete GitOps workflow using k3d, ArgoCD, ChartMuseum, and a local Git server.

## Architecture

- **k3d**: Local Kubernetes cluster
- **ArgoCD**: GitOps operator for continuous deployment
- **ChartMuseum**: Helm chart repository
- **Git Server**: HTTP-based Git repository for application values
- **Docker Registry**: Local registry for container images

## Quick Start

1. **Setup Environment**:

   ```bash
   ./setup.sh
   ```

2. **Build & Deploy**:

   ```bash
   make build    # Build app, push to registry, package & push chart
   make deploy   # Create/update ArgoCD application
   ```

3. **Check Status**:
   ```bash
   make status   # View application status
   make logs     # View application logs
   ```

## Project Structure

```
├── example-app/           # Go HTTP server application
│   ├── main.go           # Application source code
│   ├── go.mod            # Go module definition
│   ├── go.sum            # Go module checksums
│   └── Dockerfile        # Container build definition
├── charts/               # Helm charts
│   └── example-app/      # Application Helm chart
├── k8s/                  # Kubernetes manifests
│   ├── chartmuseum/      # ChartMuseum deployment
│   └── git-server/       # Git server deployment
├── manifest.git/         # Git repository for values
│   └── example-app-values.yaml
├── scripts/              # Automation scripts
│   ├── build-and-push.sh # Build & push workflow
│   ├── create-simple-argocd-app.sh # ArgoCD app creation
│   ├── setup-k8s-resources.sh # K8s resources setup
│   └── update-manifests.sh # Update values with new tags
├── Makefile              # Build automation
└── setup.sh              # Environment setup
```

## Workflow

1. **Development**: Modify `example-app/main.go`
2. **Build**: `make build` - Builds Docker image, pushes to registry, packages Helm chart, pushes to ChartMuseum, updates values in Git
3. **Deploy**: `make deploy` - Creates/updates ArgoCD application
4. **Sync**: ArgoCD automatically syncs and deploys to cluster

## Access Points

- **ArgoCD UI**: http://localhost:8083 (admin/admin)
- **ChartMuseum**: http://localhost:8084
- **Git Server**: http://localhost:8085
- **Application**: http://example-app.localhost (via ingress)

## Commands

- `make build` - Build and push everything
- `make deploy` - Deploy to ArgoCD
- `make status` - Check application status
- `make logs` - View application logs
- `make clean` - Clean up resources
- `make restart` - Restart the environment

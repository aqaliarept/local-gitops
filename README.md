# Local GitOps Environment

A complete local Kubernetes development environment using k3d, ArgoCD, local Docker registry, and ChartMuseum for a fully self-contained GitOps workflow.

## 🚀 Features

- **k3d Cluster**: Lightweight Kubernetes cluster running in Docker
- **Local Docker Registry**: Push and pull container images without external dependencies
- **ArgoCD**: GitOps continuous delivery with local manifest folder support
- **ChartMuseum**: Local Helm chart repository for chart management
- **Example Applications**: Ready-to-use Node.js app with Kubernetes manifests and Helm charts

## 📋 Prerequisites

Before running the setup, ensure you have the following tools installed:

- [k3d](https://k3d.io/) - Kubernetes in Docker
- [Docker](https://www.docker.com/) - Container runtime
- [kubectl](https://kubernetes.io/docs/tasks/tools/) - Kubernetes CLI
- [Helm](https://helm.sh/) - Package manager for Kubernetes
- [curl](https://curl.se/) - For API calls

### Automatic Installation

The easiest way to install all prerequisites is using the provided script:

```bash
./install-prerequisites.sh
```

This script automatically detects your operating system and installs:

- k3d (Kubernetes in Docker)
- kubectl (Kubernetes CLI)
- Helm (Package manager)
- Docker (Container runtime)
- curl (HTTP client)

### Manual Installation

If you prefer to install manually:

```bash
# Install k3d
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Install kubectl (macOS)
brew install kubectl

# Install Helm (macOS)
brew install helm

# Verify installations
k3d version
kubectl version --client
helm version
```

## 🛠️ Quick Start

1. **Install Prerequisites**

   ```bash
   git clone <your-repo>
   cd local-gitops
   ./install-prerequisites.sh
   ```

2. **Setup Environment**

   ```bash
   ./setup.sh
   ```

3. **Access Services**

   - ArgoCD UI: http://localhost:8080 (admin / password from setup output)
   - ChartMuseum: http://localhost:8081
   - Local Registry: myregistry.localhost:5001

4. **Build and Deploy Example App**

   ```bash
   # Build and push the example app
   ./scripts/build-and-push.sh example-app latest ./example-app/Dockerfile

   # Deploy using ArgoCD
   ./scripts/deploy-app.sh example-app default
   ```

## 📁 Project Structure

```
local-gitops/
├── setup.sh                 # Main setup script
├── scripts/                 # Helper scripts
│   ├── build-and-push.sh    # Build and push Docker images
│   ├── deploy-app.sh        # Deploy applications with ArgoCD
│   ├── push-chart.sh        # Package and push Helm charts
│   └── cleanup.sh           # Clean up environment
├── manifests/               # Kubernetes manifests (GitOps source)
│   └── example-app.yaml     # Example application manifests
├── charts/                  # Helm charts
│   └── example-app/         # Example application chart
├── example-app/             # Example Node.js application
│   ├── Dockerfile
│   ├── package.json
│   └── server.js
└── packages/                # Generated Helm packages
```

## 🔧 Usage Guide

### Building and Pushing Images

Build and push Docker images to the local registry:

```bash
# Build and push an image
./scripts/build-and-push.sh <image-name> <tag> [dockerfile-path]

# Example
./scripts/build-and-push.sh my-app v1.0.0 ./my-app/Dockerfile
```

### Deploying Applications

Deploy applications using ArgoCD with local manifests:

```bash
# Deploy an application
./scripts/deploy-app.sh <app-name> [namespace]

# Example
./scripts/deploy-app.sh my-app production
```

### Working with Helm Charts

Package and push Helm charts to ChartMuseum:

```bash
# Package and push a chart
./scripts/push-chart.sh <chart-directory> [version]

# Example
./scripts/push-chart.sh ./charts/my-app v1.0.0
```

### GitOps Workflow

1. **Develop locally**: Create or modify Kubernetes manifests in `manifests/`
2. **Commit changes**: Git commit triggers ArgoCD sync
3. **Automatic deployment**: ArgoCD applies changes to the cluster

```bash
# Example workflow
cd manifests/
# Edit your YAML files
git add .
git commit -m "Update application configuration"
# ArgoCD automatically syncs the changes
```

## 🎯 Example Applications

### Node.js Example App

The included example application demonstrates:

- **Health checks**: `/health` and `/ready` endpoints
- **Environment variables**: Configurable via Kubernetes
- **Resource limits**: CPU and memory constraints
- **Ingress**: External access via nginx ingress

**Endpoints:**

- `GET /` - Application info
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /info` - Detailed application information

### Deploying the Example

```bash
# 1. Build and push the image
./scripts/build-and-push.sh example-app latest ./example-app/Dockerfile

# 2. Deploy using manifests
./scripts/deploy-app.sh example-app default

# 3. Or deploy using Helm chart
./scripts/push-chart.sh ./charts/example-app v1.0.0
```

## 🔍 Monitoring and Debugging

### Check Cluster Status

```bash
# Check cluster nodes
kubectl get nodes

# Check all pods
kubectl get pods --all-namespaces

# Check ArgoCD applications
kubectl get applications -n argocd
```

### Access Logs

```bash
# ArgoCD logs
kubectl logs -n argocd deployment/argocd-server

# Application logs
kubectl logs -n default deployment/example-app

# ChartMuseum logs
kubectl logs -n chartmuseum deployment/chartmuseum
```

### ArgoCD CLI

```bash
# Get ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Login to ArgoCD
argocd login localhost:8080

# List applications
argocd app list

# Sync application
argocd app sync example-app
```

## 🧹 Cleanup

Remove the entire environment:

```bash
./scripts/cleanup.sh
```

This will:

- Delete the k3d cluster
- Remove the local registry
- Optionally clean up local files

## 🔧 Configuration

### Customizing the Setup

Edit `setup.sh` to modify:

- Cluster name
- Registry configuration
- Port mappings
- Resource limits

### Adding New Applications

1. **Using Manifests**:

   - Add YAML files to `manifests/`
   - Commit changes to trigger ArgoCD sync

2. **Using Helm Charts**:
   - Create charts in `charts/`
   - Package and push to ChartMuseum
   - Create ArgoCD Application pointing to the chart

### Registry Configuration

The local registry is configured to:

- Listen on `myregistry.localhost:5000`
- Trust all images (no authentication required)
- Persist images in Docker volumes

## 🐛 Troubleshooting

### Common Issues

1. **Registry not accessible**:

   ```bash
   # Check if registry is running
   docker ps | grep registry

   # Restart registry
   k3d registry delete myregistry.localhost
   k3d registry create myregistry.localhost --port 5000
   ```

2. **ArgoCD not syncing**:

   ```bash
   # Check ArgoCD repo-server logs
   kubectl logs -n argocd deployment/argocd-repo-server

   # Force sync
   argocd app sync <app-name>
   ```

3. **Images not pulling**:

   ```bash
   # Verify image exists in registry
   curl http://myregistry.localhost:5000/v2/_catalog

   # Check cluster registry configuration
   kubectl get nodes -o yaml | grep -A 10 registries
   ```

### Reset Everything

```bash
# Complete reset
./scripts/cleanup.sh
./setup.sh
```

## 📚 Additional Resources

- [k3d Documentation](https://k3d.io/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [ChartMuseum Documentation](https://chartmuseum.com/)
- [Helm Documentation](https://helm.sh/docs/)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test the setup
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Happy GitOps! 🎉**

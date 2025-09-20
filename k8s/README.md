# Kubernetes Manifests

This directory contains all Kubernetes manifests organized by component.

## Structure

```
k8s/
├── argocd/           # ArgoCD configuration
│   └── ssh-config.yaml
├── chartmuseum/      # ChartMuseum Helm repository
│   └── chartmuseum.yaml
├── git-server/       # Git server for manifests
│   └── git-server.yaml
└── README.md         # This file
```

## Components

### ArgoCD (`argocd/`)
- SSH configuration for Git server access
- Known hosts configuration
- Command parameters for SSH agent settings

### ChartMuseum (`chartmuseum/`)
- Helm chart repository server
- Provides HTTP API for chart storage and retrieval
- Used by ArgoCD to fetch Helm charts

### Git Server (`git-server/`)
- SSH-based Git server using gitasservice-k8s
- Stores Kubernetes manifests as Git repository
- Provides SSH access for ArgoCD to fetch values files

## Usage

All resources are deployed automatically by the `setup-k8s-resources.sh` script, which:
1. Creates necessary namespaces
2. Deploys ChartMuseum
3. Deploys Git server
4. Configures SSH keys and authentication
5. Sets up Git repository with manifests
6. Configures ArgoCD to access both services

## Manual Deployment

To deploy individual components:

```bash
# Deploy ChartMuseum
kubectl apply -f k8s/chartmuseum/chartmuseum.yaml

# Deploy Git server
kubectl apply -f k8s/git-server/git-server.yaml

# Configure ArgoCD SSH access
kubectl apply -f k8s/argocd/ssh-config.yaml
```

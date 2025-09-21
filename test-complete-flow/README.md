# GitOps Directory

This directory contains a complete GitOps setup with ArgoCD ApplicationSet.

## Structure

- manifest/ - Contains all application manifests and charts
  - applications/ - ArgoCD Application manifests
  - charts/ - Helm charts for nginx application
  - nginx-app-values.yaml - Application values
- bootstrap.yaml - ArgoCD Application to bootstrap the GitOps workflow

## Quick Start

1. Setup the GitOps environment:
   gitops setup

2. Push manifest content to Git repository:
   gitops build

3. Apply bootstrap.yaml to create ArgoCD application:
   kubectl apply -f bootstrap.yaml

4. Check status:
   gitops status

## Nginx Application

The nginx application uses the official nginx:1.25-alpine image and serves the default nginx welcome page:
- / - Default nginx welcome page
- Health checks are performed on the root path

## Access

- ArgoCD UI: http://localhost:8083 (admin/admin)
- Nginx App: http://nginx-app.localhost
- ChartMuseum: http://localhost:8084

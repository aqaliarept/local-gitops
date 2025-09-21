# GitOps Directory

This directory contains a simple nginx application deployed using GitOps principles.

## Structure

- manifest/ - Contains nginx application Kubernetes manifests
  - deployment.yaml - nginx deployment
  - service.yaml - nginx service
  - ingress.yaml - nginx ingress
- bootstrap.yaml - ArgoCD Application manifest

## Usage

1. Setup the cluster:
   gitops setup

2. Deploy the application:
   gitops deploy

The nginx application will be available at http://nginx.localhost

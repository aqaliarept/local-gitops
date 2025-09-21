#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

set -e

echo -e "${YELLOW}🚀 Setting up Kubernetes resources...${NC}"

# Create namespaces
echo -e "${YELLOW}📦 Creating namespaces...${NC}"
kubectl create namespace chartmuseum --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace git-server --dry-run=client -o yaml | kubectl apply -f -

# Deploy ChartMuseum
echo -e "${YELLOW}📊 Deploying ChartMuseum...${NC}"
kubectl apply -f k8s/chartmuseum/chartmuseum.yaml

# Wait for ChartMuseum to be ready
echo -e "${YELLOW}⏳ Waiting for ChartMuseum to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/chartmuseum -n chartmuseum

# Deploy Git server
echo -e "${YELLOW}📦 Deploying Git server...${NC}"
kubectl apply -f k8s/git-server/git-server.yaml

# Wait for Git server to be ready
echo -e "${YELLOW}⏳ Waiting for Git server to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/git-server -n git-server

# HTTP Git server doesn't need SSH keys
echo -e "${YELLOW}📡 Using HTTP-based Git server (no SSH keys needed)...${NC}"

# Create Git repository
echo -e "${YELLOW}📁 Creating Git repository...${NC}"
kubectl exec -n git-server deployment/git-server -- mkdir -p /git

# Copy manifests to Git server
echo -e "${YELLOW}📋 Copying manifests to Git server...${NC}"
kubectl cp manifest.git/. git-server/$(kubectl get pods -n git-server -l app=git-server -o jsonpath='{.items[0].metadata.name}'):/tmp/manifests

# Initialize Git repository
echo -e "${YELLOW}🔧 Initializing Git repository...${NC}"
kubectl exec -n git-server deployment/git-server -- sh -c "
cd /tmp/manifests
echo 'Initializing bare repository...'
git init --bare /git/manifests.git
echo 'Adding remote origin...'
git remote add origin /git/manifests.git
echo 'Pushing to master branch...'
git push origin master
echo 'Git repository initialized successfully'
"

# HTTP Git server doesn't need SSH configuration
echo -e "${YELLOW}📡 HTTP Git server configured - no SSH setup needed...${NC}"

echo -e "${GREEN}✅ Kubernetes resources setup completed successfully!${NC}"
echo ""
echo "Deployed resources:"
echo "  ChartMuseum: chartmuseum.chartmuseum.svc.cluster.local:8080"
echo "  Git Server: git-server.git-server.svc.cluster.local:80"
echo "  Repository: http://git-server.git-server.svc.cluster.local/manifests.git"
echo ""
echo "To test HTTP Git access:"
echo "  kubectl run git-test --image=alpine/git:latest --rm -it --restart=Never --command -- sh -c 'apk add --no-cache git && git clone http://git-server.git-server.svc.cluster.local/manifests.git test-repo'"

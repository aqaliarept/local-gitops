#!/bin/bash

# Local Kubernetes (k3d) with ArgoCD, Local Registry & Chart Repo Setup
# This script sets up a complete local GitOps environment

set -e

# Cleanup function for failed setup
cleanup_on_failure() {
    echo -e "${RED}❌ Setup failed. Cleaning up...${NC}"
    # Kill any background processes
    jobs -p | xargs -r kill 2>/dev/null || true
    # Clean up port forwards
    pkill -f "kubectl port-forward" 2>/dev/null || true
    echo -e "${YELLOW}💡 You can run this script again to retry the setup${NC}"
    exit 1
}

# Set trap for cleanup on failure
trap cleanup_on_failure ERR

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REGISTRY_NAME="myregistry.localhost"
REGISTRY_PORT="5001"
CLUSTER_NAME="devcluster"
MANIFESTS_DIR="./manifests"
CHARTS_DIR="./charts"

echo -e "${BLUE}🚀 Setting up Local GitOps Environment${NC}"

# Check prerequisites
echo -e "${YELLOW}📋 Checking prerequisites...${NC}"
command -v k3d >/dev/null 2>&1 || { echo -e "${RED}❌ k3d is required but not installed. Please install k3d first.${NC}"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo -e "${RED}❌ Docker is required but not installed.${NC}"; exit 1; }
command -v helm >/dev/null 2>&1 || { echo -e "${RED}❌ Helm is required but not installed.${NC}"; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo -e "${RED}❌ kubectl is required but not installed.${NC}"; exit 1; }
command -v htpasswd >/dev/null 2>&1 || { echo -e "${RED}❌ htpasswd is required but not installed. Please install apache2-utils.${NC}"; exit 1; }
command -v curl >/dev/null 2>&1 || { echo -e "${RED}❌ curl is required but not installed.${NC}"; exit 1; }
command -v lsof >/dev/null 2>&1 || { echo -e "${RED}❌ lsof is required but not installed.${NC}"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo -e "${RED}❌ jq is required but not installed.${NC}"; exit 1; }

echo -e "${GREEN}✅ All prerequisites found${NC}"

# Check for port conflicts
echo -e "${YELLOW}🔍 Checking for port conflicts...${NC}"
CONFLICT_PORTS=""
for port in 8083 8084 8085 5001; do
    if lsof -i :$port >/dev/null 2>&1; then
        CONFLICT_PORTS="$CONFLICT_PORTS $port"
    fi
done

if [ -n "$CONFLICT_PORTS" ]; then
    echo -e "${YELLOW}⚠️  Warning: The following ports are already in use:$CONFLICT_PORTS${NC}"
    echo -e "${YELLOW}   This may cause issues with the setup. Consider stopping services using these ports.${NC}"
    echo -e "${YELLOW}   Continuing anyway...${NC}"
else
    echo -e "${GREEN}✅ No port conflicts detected${NC}"
fi

# Create directories
echo -e "${YELLOW}📁 Creating directories...${NC}"
mkdir -p "$MANIFESTS_DIR"
mkdir -p "$CHARTS_DIR"
mkdir -p "./packages"

# Initialize git repository for manifests
if [ ! -d "$MANIFESTS_DIR/.git" ]; then
    echo -e "${YELLOW}🔧 Initializing git repository for manifests...${NC}"
    cd "$MANIFESTS_DIR"
    git init
    git config user.name "Local GitOps"
    git config user.email "local@gitops.dev"
    cd ..
fi

# Create local registry
echo -e "${YELLOW}🐳 Creating local Docker registry...${NC}"
if k3d registry list | grep -q "$REGISTRY_NAME"; then
    echo -e "${BLUE}ℹ️  Registry $REGISTRY_NAME already exists${NC}"
else
    k3d registry create "$REGISTRY_NAME" --port "$REGISTRY_PORT"
    echo -e "${GREEN}✅ Registry created at $REGISTRY_NAME:$REGISTRY_PORT${NC}"
fi

# Create k3d cluster with registry and volume mounts
echo -e "${YELLOW}☸️  Creating k3d cluster...${NC}"
if k3d cluster list | grep -q "$CLUSTER_NAME"; then
    echo -e "${BLUE}ℹ️  Cluster $CLUSTER_NAME already exists, deleting...${NC}"
    k3d cluster delete "$CLUSTER_NAME"
fi

k3d cluster create "$CLUSTER_NAME" \
    --registry-use "k3d-$REGISTRY_NAME:$REGISTRY_PORT" \
    -v "$(pwd)/$MANIFESTS_DIR:/data/manifests@server:0" \
    -p "8083:80@loadbalancer" \
    -p "8084:8080@loadbalancer" \
    -p "8085:8080@loadbalancer" \
    --wait

echo -e "${GREEN}✅ Cluster created with registry integration${NC}"

# Wait for cluster to be ready
echo -e "${YELLOW}⏳ Waiting for cluster to be ready...${NC}"
kubectl wait --for=condition=Ready nodes --all --timeout=300s

# Set kubeconfig
export KUBECONFIG=$(k3d kubeconfig write "$CLUSTER_NAME")
echo -e "${GREEN}✅ Kubeconfig configured${NC}"

# Install ArgoCD
echo -e "${YELLOW}🔄 Installing ArgoCD...${NC}"
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for ArgoCD to be ready
echo -e "${YELLOW}⏳ Waiting for ArgoCD to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/argocd-server -n argocd

# Configure ArgoCD admin password after installation
echo -e "${YELLOW}🔑 Setting up ArgoCD admin credentials...${NC}"
# Generate bcrypt hash for 'admin' password
ADMIN_HASH=$(echo -n 'admin' | htpasswd -niB admin | cut -d: -f2)
ADMIN_HASH_B64=$(echo -n "$ADMIN_HASH" | base64 | tr -d '\n')
ADMIN_MTIME_B64=$(echo -n "$(date +%Y-%m-%dT%H:%M:%S)" | base64 | tr -d '\n')

# Update the existing secret with admin password using a more reliable method
kubectl get secret argocd-secret -n argocd -o json | \
jq --arg password "$ADMIN_HASH_B64" --arg mtime "$ADMIN_MTIME_B64" \
'.data["admin.password"] = $password | .data["admin.passwordMtime"] = $mtime' | \
kubectl apply -f -

# Restart ArgoCD server to pick up the new password
echo -e "${YELLOW}🔄 Restarting ArgoCD server to apply new credentials...${NC}"
kubectl rollout restart deployment/argocd-server -n argocd
kubectl rollout status deployment/argocd-server -n argocd


# Install Kubernetes resources (ChartMuseum, Git server, etc.)
echo -e "${YELLOW}📦 Installing Kubernetes resources...${NC}"
./scripts/setup-k8s-resources.sh

# ArgoCD admin credentials are now set to admin/admin
echo -e "${YELLOW}🔑 ArgoCD admin credentials configured...${NC}"
ARGOCD_PASSWORD="admin"

# Verify ArgoCD login works
echo -e "${YELLOW}🔍 Verifying ArgoCD login...${NC}"
# Use a random port to avoid conflicts
VERIFY_PORT=$((8082 + RANDOM % 1000))
kubectl port-forward -n argocd svc/argocd-server $VERIFY_PORT:443 > /dev/null 2>&1 &
PORT_FORWARD_PID=$!
sleep 5

# Test login
LOGIN_TEST=$(curl -k -s -o /dev/null -w "%{http_code}" https://localhost:$VERIFY_PORT/api/v1/session -d '{"username":"admin","password":"admin"}' -H "Content-Type: application/json" 2>/dev/null || echo "000")

if [ "$LOGIN_TEST" = "200" ]; then
    echo -e "${GREEN}✅ ArgoCD login verified successfully${NC}"
else
    echo -e "${YELLOW}⚠️  ArgoCD login test failed (HTTP $LOGIN_TEST), but setup continues${NC}"
fi

# Clean up port forward
kill $PORT_FORWARD_PID 2>/dev/null || true

# Create initial commit in manifests directory
echo -e "${YELLOW}📝 Creating initial commit...${NC}"
cd "$MANIFESTS_DIR"
if [ ! -f "README.md" ]; then
    echo "# Local GitOps Manifests" > README.md
    echo "This directory contains Kubernetes manifests for local GitOps deployment." >> README.md
    git add README.md
    git commit -m "Initial commit"
fi
cd ..

echo -e "${GREEN}🎉 Setup completed successfully!${NC}"
echo ""
echo -e "${BLUE}📋 Access Information:${NC}"
echo -e "  ArgoCD UI: http://localhost:8083"
echo -e "  ArgoCD Username: admin"
echo -e "  ArgoCD Password: admin"
echo -e "  ChartMuseum: http://localhost:8084"
echo -e "  Local Registry: $REGISTRY_NAME:$REGISTRY_PORT"
echo ""
echo -e "${BLUE}📁 Directory Structure:${NC}"
echo -e "  Manifests: $(pwd)/$MANIFESTS_DIR"
echo -e "  Charts: $(pwd)/$CHARTS_DIR"
echo -e "  Packages: $(pwd)/packages"
echo ""
echo -e "${YELLOW}🔧 Next Steps:${NC}"
echo -e "  1. Add your Kubernetes manifests to $MANIFESTS_DIR"
echo -e "  2. Create Helm charts in $CHARTS_DIR"
echo -e "  3. Use ./scripts/build-and-push.sh to build and push images"
echo -e "  4. Use ./scripts/deploy-app.sh to deploy applications"
echo ""
echo -e "${GREEN}✅ Local GitOps environment is ready!${NC}"
echo ""
echo -e "${BLUE}🚀 Quick Start Commands:${NC}"
echo -e "  # Access ArgoCD UI:"
echo -e "  open http://localhost:8083"
echo -e ""
echo -e "  # Access ChartMuseum:"
echo -e "  open http://localhost:8084"
echo -e ""
echo -e "  # Build and push an image:"
echo -e "  ./scripts/build-and-push.sh"
echo -e ""
echo -e "  # Deploy an application:"
echo -e "  ./scripts/deploy-app.sh"
echo ""
echo -e "${GREEN}🎉 Happy GitOps-ing!${NC}"

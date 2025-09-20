#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

set -e

echo -e "${YELLOW}🚀 Setting up Git server...${NC}"

# Generate SSH key pair for ArgoCD
echo -e "${YELLOW}🔑 Generating SSH key pair for ArgoCD...${NC}"
ssh-keygen -t rsa -b 4096 -f /tmp/argocd_key -N "" -C "argocd@local" >/dev/null 2>&1

# Deploy Git server
echo -e "${YELLOW}📦 Deploying Git server...${NC}"
kubectl apply -f manifests/git-server.yaml

# Wait for Git server to be ready
echo -e "${YELLOW}⏳ Waiting for Git server to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/git-server -n git-server

# Update Git server with the generated public key
echo -e "${YELLOW}🔧 Configuring Git server with SSH key...${NC}"
kubectl create configmap git-server-keys \
  --from-file=id_rsa.pub=/tmp/argocd_key.pub \
  -n git-server \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart Git server to pick up the new SSH key
kubectl rollout restart deployment/git-server -n git-server
kubectl rollout status deployment/git-server -n git-server

# Create Git repository
echo -e "${YELLOW}📁 Creating Git repository...${NC}"
kubectl exec -n git-server deployment/git-server -- mkdir -p /git-server/repos

# Copy manifests to Git server
echo -e "${YELLOW}📋 Copying manifests to Git server...${NC}"
kubectl cp manifests/. git-server/$(kubectl get pods -n git-server -l app=git-server -o jsonpath='{.items[0].metadata.name}'):/tmp/manifests

# Initialize Git repository
echo -e "${YELLOW}🔧 Initializing Git repository...${NC}"
kubectl exec -n git-server deployment/git-server -- sh -c "
cd /tmp/manifests
git init --bare /git-server/repos/manifests.git
git remote add origin /git-server/repos/manifests.git
git push origin master
"

# Configure ArgoCD with SSH credentials
echo -e "${YELLOW}🔐 Configuring ArgoCD with SSH credentials...${NC}"

# Create SSH secret for ArgoCD
kubectl create secret generic git-server-ssh-key \
  --from-file=sshPrivateKey=/tmp/argocd_key \
  --from-literal=type=ssh \
  --from-literal=url=ssh://git@git-server.git-server.svc.cluster.local:/git-server/repos/manifests.git \
  -n argocd \
  --dry-run=client -o yaml | kubectl apply -f -

# Add Git server to ArgoCD known hosts
kubectl get configmap argocd-ssh-known-hosts-cm -n argocd -o jsonpath='{.data.ssh_known_hosts}' > /tmp/current-hosts
kubectl exec -n git-server deployment/git-server -- ssh-keyscan -p 22 localhost >> /tmp/current-hosts
kubectl create configmap argocd-ssh-known-hosts-cm \
  --from-file=ssh_known_hosts=/tmp/current-hosts \
  -n argocd \
  --dry-run=client -o yaml | kubectl apply -f -

# Configure ArgoCD to disable SSH agent
kubectl patch configmap argocd-cmd-params-cm -n argocd --type merge -p '{
  "data": {
    "server.insecure": "true",
    "server.disable.auth": "true",
    "repo.server.ssh.agent": "false"
  }
}'

# Restart ArgoCD to pick up new configuration
echo -e "${YELLOW}🔄 Restarting ArgoCD to apply new configuration...${NC}"
kubectl rollout restart deployment/argocd-repo-server -n argocd
kubectl rollout status deployment/argocd-repo-server -n argocd

# Clean up temporary files
rm -f /tmp/argocd_key /tmp/argocd_key.pub /tmp/current-hosts

echo -e "${GREEN}✅ Git server setup completed successfully!${NC}"
echo ""
echo "Git server details:"
echo "  Namespace: git-server"
echo "  Service: git-server.git-server.svc.cluster.local:22"
echo "  Repository: ssh://git@git-server.git-server.svc.cluster.local:/git-server/repos/manifests.git"
echo ""
echo "To test SSH connection:"
echo "  kubectl run ssh-test --image=alpine/git:latest --rm -it --restart=Never --command -- sh -c 'apk add --no-cache openssh-client && ssh -o StrictHostKeyChecking=no git@git-server.git-server.svc.cluster.local'"

#!/bin/bash

# Port forward Git server for local access
# Usage: ./scripts/port-forward-git.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
GIT_SERVER_NAMESPACE="git-server"
GIT_SERVER_SERVICE="git-server"
LOCAL_PORT="8085"
REMOTE_PORT="80"

echo -e "${YELLOW}🌐 Starting Git server port forward...${NC}"
echo -e "${GREEN}📡 Git Server: http://localhost:$LOCAL_PORT${NC}"

# Check if Git server is running
if ! kubectl get svc "$GIT_SERVER_SERVICE" -n "$GIT_SERVER_NAMESPACE" >/dev/null 2>&1; then
    echo -e "${RED}❌ Git server service not found. Please run 'make setup' first.${NC}"
    exit 1
fi

# Start port forwarding
kubectl port-forward -n "$GIT_SERVER_NAMESPACE" "svc/$GIT_SERVER_SERVICE" "$LOCAL_PORT:$REMOTE_PORT"


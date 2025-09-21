#!/bin/bash

# Cleanup local GitOps environment
# Usage: ./scripts/cleanup.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_NAME="devcluster"
REGISTRY_NAME="myregistry.localhost"

echo -e "${BLUE}🧹 Cleaning up Local GitOps Environment${NC}"

# Confirm cleanup
read -p "Are you sure you want to delete the cluster and registry? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}❌ Cleanup cancelled${NC}"
    exit 0
fi

# Delete k3d cluster
echo -e "${YELLOW}🗑️  Deleting k3d cluster...${NC}"
if k3d cluster list | grep -q "$CLUSTER_NAME"; then
    k3d cluster delete "$CLUSTER_NAME"
    echo -e "${GREEN}✅ Cluster deleted${NC}"
else
    echo -e "${BLUE}ℹ️  Cluster $CLUSTER_NAME not found${NC}"
fi

# Delete registry
echo -e "${YELLOW}🗑️  Deleting registry...${NC}"
if k3d registry list | grep -q "$REGISTRY_NAME"; then
    k3d registry delete "$REGISTRY_NAME"
    echo -e "${GREEN}✅ Registry deleted${NC}"
else
    echo -e "${BLUE}ℹ️  Registry $REGISTRY_NAME not found${NC}"
fi

# Clean up local files (optional)
read -p "Do you want to clean up local manifests and charts? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}🗑️  Cleaning up local files...${NC}"
    
    # Remove packages directory
    if [ -d "./packages" ]; then
        rm -rf "./packages"
        echo -e "${GREEN}✅ Packages directory removed${NC}"
    fi
    
    # Reset manifest.git repository
    if [ -d "./manifest.git/.git" ]; then
        cd "./manifest.git"
        git clean -fd
        git reset --hard HEAD
        cd ..
        echo -e "${GREEN}✅ Manifest.git repository reset${NC}"
    fi
    
    # Remove charts directory contents
    if [ -d "./charts" ]; then
        find "./charts" -mindepth 1 -delete
        echo -e "${GREEN}✅ Charts directory cleaned${NC}"
    fi
fi

echo -e "${GREEN}🎉 Cleanup completed successfully!${NC}"
echo -e "${BLUE}💡 To start fresh, run: ./setup.sh${NC}"

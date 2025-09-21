#!/bin/bash

# Push manifest changes to Git remote
# Usage: ./scripts/push-manifests.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
MANIFEST_DIR="./manifest.git"
GIT_REMOTE="origin"
GIT_BRANCH="master"

echo -e "${YELLOW}📤 Pushing manifest changes to Git remote...${NC}"

# Check if manifest directory exists
if [ ! -d "$MANIFEST_DIR" ]; then
    echo -e "${RED}❌ Manifest directory not found: $MANIFEST_DIR${NC}"
    exit 1
fi

# Check if it's a git repository
if [ ! -d "$MANIFEST_DIR/.git" ]; then
    echo -e "${RED}❌ Not a git repository: $MANIFEST_DIR${NC}"
    exit 1
fi

# Change to manifest directory
cd "$MANIFEST_DIR"

# Check if there are any changes
if git diff --quiet && git diff --cached --quiet; then
    echo -e "${GREEN}✅ No changes to commit${NC}"
    exit 0
fi

# Add all changes
echo -e "${YELLOW}📝 Adding changes to git...${NC}"
git add .

# Commit changes
echo -e "${YELLOW}💾 Committing changes...${NC}"
git commit -m "Update application values - $(date '+%Y-%m-%d %H:%M:%S')" || {
    echo -e "${GREEN}✅ No changes to commit${NC}"
    exit 0
}

# Push to remote
echo -e "${YELLOW}🚀 Pushing to remote repository...${NC}"
git push "$GIT_REMOTE" "$GIT_BRANCH"

echo -e "${GREEN}✅ Manifest changes pushed successfully${NC}"
echo -e "${GREEN}📋 Repository: $(git remote get-url origin)${NC}"
echo -e "${GREEN}📋 Branch: $GIT_BRANCH${NC}"
echo -e "${GREEN}📋 Commit: $(git rev-parse --short HEAD)${NC}"


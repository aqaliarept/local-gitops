#!/bin/bash

# Update manifest repository with latest image tag
# Usage: ./scripts/update-manifests.sh

set -e

# Configuration
VALUES_FILE="./manifest.git/example-app-values.yaml"
TAG_FILE=".image-tag"

# Check if tag file exists
if [ ! -f "$TAG_FILE" ]; then
    echo "Tag file not found. Run 'make build' first."
    exit 1
fi

TAG=$(cat "$TAG_FILE")
echo "Updating values manifest with tag: $TAG"

# Update the image tag in the values file
sed -i.bak "s/tag: \".*\"/tag: \"$TAG\"/" "$VALUES_FILE"

# Restore backup if it exists
if [ -f "$VALUES_FILE.bak" ]; then
    rm "$VALUES_FILE.bak"
fi

echo "Values manifest updated: $VALUES_FILE"
echo "Using image: k3d-myregistry.localhost:5001/example-app:$TAG"

#!/bin/bash

# Build and push example-app Docker image and Helm chart
# Usage: ./scripts/build-and-push.sh

set -e

# Configuration
REGISTRY_NAME="myregistry.localhost"
REGISTRY_PORT="5001"
REGISTRY_URL="localhost:$REGISTRY_PORT"
IMAGE_NAME="example-app"
DOCKERFILE_PATH="./example-app/Dockerfile"
BUILD_CONTEXT="./example-app"
CHART_DIR="./charts/example-app"
PACKAGES_DIR="./packages"
CHARTMUSEUM_URL="http://localhost:8084"
CHART_VERSION="0.1.0"
CLUSTER_NAME="devcluster"
TAG_FILE=".image-tag"

# Generate timestamp
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
echo "$TIMESTAMP" > "$TAG_FILE"
echo "Generated timestamp: $TIMESTAMP"

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed."; exit 1; }
command -v helm >/dev/null 2>&1 || { echo "Helm is required but not installed."; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "curl is required but not installed."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo "kubectl is required but not installed."; exit 1; }

# Check if registry is running
if ! docker ps | grep -q "k3d-$REGISTRY_NAME"; then
    echo "Local registry is not running. Please run ./setup.sh first."
    exit 1
fi

# Check if k3d cluster is running
if ! k3d cluster list | grep -q "$CLUSTER_NAME"; then
    echo "Cluster $CLUSTER_NAME is not running. Please run ./setup.sh first."
    exit 1
fi

# Set kubeconfig
export KUBECONFIG=$(k3d kubeconfig write "$CLUSTER_NAME")

# Check if ChartMuseum is running
if ! kubectl get deployment chartmuseum -n chartmuseum >/dev/null 2>&1; then
    echo "ChartMuseum is not running. Please run ./setup.sh first."
    exit 1
fi

echo "Building and pushing Docker image with tag: $TIMESTAMP"

# Build the image
docker build -t "$IMAGE_NAME:$TIMESTAMP" -f "$DOCKERFILE_PATH" "$BUILD_CONTEXT"

# Tag for local registry
FULL_IMAGE_NAME="$REGISTRY_URL/$IMAGE_NAME:$TIMESTAMP"
docker tag "$IMAGE_NAME:$TIMESTAMP" "$FULL_IMAGE_NAME"

# Push to local registry
docker push "$FULL_IMAGE_NAME"

echo "Docker image built and pushed successfully: $FULL_IMAGE_NAME"

echo "Building and pushing Helm chart..."

# Get chart name from Chart.yaml
CHART_NAME=$(grep '^name:' "$CHART_DIR/Chart.yaml" | awk '{print $2}')

# Update chart values with timestamped tag
echo "Updating chart values with tag: $TIMESTAMP"
sed -i.bak "s/tag: \".*\"/tag: \"$TIMESTAMP\"/" "$CHART_DIR/values.yaml"

# Create packages directory if it doesn't exist
mkdir -p "$PACKAGES_DIR"

# Package the chart
helm package "$CHART_DIR" --version "$CHART_VERSION" --destination "$PACKAGES_DIR"

PACKAGE_FILE="$PACKAGES_DIR/$CHART_NAME-$CHART_VERSION.tgz"

if [ ! -f "$PACKAGE_FILE" ]; then
    echo "Failed to create package file"
    exit 1
fi

# Push to ChartMuseum using port-forward
echo "Pushing chart to ChartMuseum..."
kubectl port-forward -n chartmuseum svc/chartmuseum 8084:8080 > /dev/null 2>&1 &
PORT_FORWARD_PID=$!
sleep 2

# Push the chart
curl --data-binary "@$PACKAGE_FILE" "$CHARTMUSEUM_URL/api/charts"

# Clean up port forward
kill $PORT_FORWARD_PID 2>/dev/null || true

# Restore original values.yaml
if [ -f "$CHART_DIR/values.yaml.bak" ]; then
    mv "$CHART_DIR/values.yaml.bak" "$CHART_DIR/values.yaml"
fi

echo "Helm chart built and pushed successfully: $CHART_NAME version $CHART_VERSION"
echo "Using image: $FULL_IMAGE_NAME"
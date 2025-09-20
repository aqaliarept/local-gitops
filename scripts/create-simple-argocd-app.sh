#!/bin/bash

# Create simple ArgoCD application targeting Git repository directly
# Usage: ./scripts/create-simple-argocd-app.sh

set -e

# Configuration
APP_NAME="example-app-simple"
NAMESPACE="argocd"
CHARTMUSEUM_URL="http://chartmuseum.chartmuseum.svc.cluster.local:8080"
CHART_NAME="example-app"
CHART_VERSION="0.1.0"
VALUES_REPO_URL="http://git-server.git-server.svc.cluster.local/manifest"
VALUES_PATH="example-app-values.yaml"
CLUSTER_NAME="devcluster"

# Check prerequisites
command -v kubectl >/dev/null 2>&1 || { echo "kubectl is required but not installed."; exit 1; }

# Check if k3d cluster is running
if ! k3d cluster list | grep -q "$CLUSTER_NAME"; then
    echo "Cluster $CLUSTER_NAME is not running. Please run ./setup.sh first."
    exit 1
fi

# Set kubeconfig
export KUBECONFIG=$(k3d kubeconfig write "$CLUSTER_NAME")

# Check if ArgoCD is running
if ! kubectl get deployment argocd-server -n "$NAMESPACE" >/dev/null 2>&1; then
    echo "ArgoCD is not running. Please run ./setup.sh first."
    exit 1
fi

# Check if application already exists
if kubectl get application "$APP_NAME" -n "$NAMESPACE" >/dev/null 2>&1; then
    echo "ArgoCD application '$APP_NAME' already exists."
    echo "Updating application..."
    
    # Update existing application
    kubectl patch application "$APP_NAME" -n "$NAMESPACE" --type merge -p '{
        "spec": {
            "source": null,
            "sources": [
                {
                    "repoURL": "'"$CHARTMUSEUM_URL"'",
                    "chart": "'"$CHART_NAME"'",
                    "targetRevision": "'"$CHART_VERSION"'",
                    "helm": {
                        "valueFiles": ["$values/'"$VALUES_PATH"'"]
                    }
                },
                {
                    "repoURL": "'"$VALUES_REPO_URL"'",
                    "targetRevision": "HEAD",
                    "ref": "values"
                }
            ],
            "destination": {
                "server": "https://kubernetes.default.svc",
                "namespace": "default"
            },
            "syncPolicy": {
                "automated": {
                    "prune": true,
                    "selfHeal": true
                }
            }
        }
    }'
    
    echo "ArgoCD application updated successfully."
else
    echo "Creating ArgoCD application '$APP_NAME'..."
    
    # Create new application
    kubectl apply -f - <<EOF
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: $APP_NAME
  namespace: $NAMESPACE
spec:
  project: default
  sources:
  - repoURL: $CHARTMUSEUM_URL
    chart: $CHART_NAME
    targetRevision: $CHART_VERSION
    helm:
      valueFiles:
      - \$values/$VALUES_PATH
  - repoURL: $VALUES_REPO_URL
    targetRevision: HEAD
    ref: values
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
EOF
    
    echo "ArgoCD application created successfully."
fi

echo ""
echo "Application details:"
echo "  Name: $APP_NAME"
echo "  Chart: $CHARTMUSEUM_URL/$CHART_NAME:$CHART_VERSION"
echo "  Values: $VALUES_REPO_URL/$VALUES_PATH"
echo "  Namespace: default"
echo ""
echo "To sync the application:"
echo "  kubectl patch application $APP_NAME -n $NAMESPACE --type merge -p '{\"operation\":{\"sync\":{}}}'"
echo ""
echo "To view application status:"
echo "  kubectl get application $APP_NAME -n $NAMESPACE"
echo ""
echo "To view application details:"
echo "  kubectl describe application $APP_NAME -n $NAMESPACE"

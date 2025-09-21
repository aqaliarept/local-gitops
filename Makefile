# Local GitOps Environment - Makefile

.PHONY: help setup build deploy clean status logs sync port-forward test full-flow dev-flow

# Configuration
TAG_FILE := .image-tag
CLUSTER_NAME := devcluster

# Default target
help: ## Show this help message
	@echo "Local GitOps Environment - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# === SETUP TARGETS ===
setup: ## Run the complete environment setup
	@echo "🚀 Setting up Local GitOps Environment..."
	./setup.sh

setup-k8s: ## Setup only Kubernetes resources (ChartMuseum, Git server)
	@echo "📦 Setting up Kubernetes resources..."
	./scripts/setup-k8s-resources.sh

# === BUILD TARGETS ===
build: ## Build and push Docker image and Helm chart
	@echo "🐳 Building and pushing Docker image and Helm chart..."
	@./scripts/build-and-push.sh
	@echo "📝 Updating manifest repository..."
	@./scripts/update-manifests.sh
	@echo "🌐 Starting Git server port forward..."
	@./scripts/port-forward-git.sh &
	@echo "⏳ Waiting for Git server to be accessible..."
	@sleep 3
	@echo "📤 Pushing manifest changes to Git remote..."
	@./scripts/push-manifests.sh
	@echo "🛑 Stopping Git server port forward..."
	@pkill -f "kubectl port-forward.*git-server" || true

build-image: ## Build and push only Docker image
	@echo "🐳 Building and pushing Docker image..."
	@./scripts/build-and-push.sh

build-chart: ## Build and push only Helm chart
	@echo "📦 Building and pushing Helm chart..."
	@./scripts/build-and-push.sh

push-manifests: ## Push manifest changes to Git remote
	@echo "📤 Pushing manifest changes to Git remote..."
	@./scripts/push-manifests.sh

# === DEPLOY TARGETS ===
deploy: ## Create/update ArgoCD application
	@echo "🚀 Creating/updating ArgoCD application..."
	@./scripts/create-simple-argocd-app.sh

sync: ## Manually sync ArgoCD application
	@echo "🔄 Syncing ArgoCD application..."
	@kubectl patch application example-app-simple -n argocd --type merge -p '{"operation":{"sync":{}}}'

# === MONITORING TARGETS ===
status: ## Show cluster and application status
	@echo "📊 Cluster Status:"
	@kubectl get nodes
	@echo ""
	@echo "📊 Pods Status:"
	@kubectl get pods --all-namespaces
	@echo ""
	@echo "📊 ArgoCD Applications:"
	@kubectl get applications -n argocd 2>/dev/null || echo "No ArgoCD applications found"
	@echo ""
	@echo "📊 Example App Status:"
	@kubectl get pods,svc,ingress -l app=example-app 2>/dev/null || echo "Example app not deployed"

logs: ## Show logs for all components
	@echo "📋 ArgoCD Server Logs:"
	@kubectl logs -n argocd deployment/argocd-server --tail=20
	@echo ""
	@echo "📋 ArgoCD Repo Server Logs:"
	@kubectl logs -n argocd deployment/argocd-repo-server --tail=20
	@echo ""
	@echo "📋 ChartMuseum Logs:"
	@kubectl logs -n chartmuseum deployment/chartmuseum --tail=20
	@echo ""
	@echo "📋 Git Server Logs:"
	@kubectl logs -n git-server deployment/git-server --tail=20
	@echo ""
	@echo "📋 Example App Logs:"
	@kubectl logs -n default deployment/example-app-simple --tail=20 2>/dev/null || echo "Example app not deployed"

# === ACCESS TARGETS ===
port-forward: ## Start port forwarding for all services
	@echo "🌐 Starting port forwarding..."
	@echo "ArgoCD UI: http://localhost:8083 (admin/admin)"
	@echo "ChartMuseum: http://localhost:8084"
	@echo "Git Server: http://localhost:8085"
	@echo "Example App: http://example-app.localhost"
	@echo ""
	@echo "Press Ctrl+C to stop port forwarding"
	@kubectl port-forward -n argocd svc/argocd-server 8083:443 &
	@kubectl port-forward -n chartmuseum svc/chartmuseum 8084:8080 &
	@kubectl port-forward -n git-server svc/git-server 8085:80 &
	@wait

argocd-ui: ## Port forward only ArgoCD UI
	@echo "🌐 Starting ArgoCD UI port forward..."
	@echo "ArgoCD UI: http://localhost:8083 (admin/admin)"
	@kubectl port-forward -n argocd svc/argocd-server 8083:443

chartmuseum-ui: ## Port forward only ChartMuseum
	@echo "🌐 Starting ChartMuseum port forward..."
	@echo "ChartMuseum: http://localhost:8084"
	@kubectl port-forward -n chartmuseum svc/chartmuseum 8084:8080

git-server-ui: ## Port forward only Git Server
	@echo "🌐 Starting Git Server port forward..."
	@echo "Git Server: http://localhost:8085"
	@kubectl port-forward -n git-server svc/git-server 8085:80

port-forward-git: ## Port forward Git server for build process
	@echo "🌐 Starting Git server port forward..."
	@./scripts/port-forward-git.sh

# === TESTING TARGETS ===
test: ## Test the complete GitOps flow
	@echo "🧪 Testing GitOps flow..."
	@echo "1. Checking cluster status..."
	@kubectl get nodes
	@echo ""
	@echo "2. Checking ArgoCD status..."
	@kubectl get pods -n argocd
	@echo ""
	@echo "3. Checking application status..."
	@kubectl get application example-app-simple -n argocd
	@echo ""
	@echo "4. Checking deployed resources..."
	@kubectl get pods,svc,ingress -l app=example-app
	@echo ""
	@echo "5. Testing application endpoint..."
	@curl -s http://example-app.localhost/health || echo "Application not accessible"

test-git: ## Test Git server connectivity
	@echo "🧪 Testing Git server..."
	@kubectl run git-test --image=alpine/git:latest --rm -it --restart=Never --command -- sh -c 'apk add --no-cache curl && git clone http://git-server.git-server.svc.cluster.local/manifest.git test-repo && ls -la test-repo/'

test-registry: ## Test Docker registry connectivity
	@echo "🧪 Testing Docker registry..."
	@curl -s http://localhost:5001/v2/_catalog

# === WORKFLOW TARGETS ===
full-flow: setup build deploy ## Complete GitOps flow: setup + build + deploy
	@echo "🎉 Full GitOps flow completed!"
	@echo "Run 'make status' to check the deployment"
	@echo "Run 'make port-forward' to access the services"
	@echo "Run 'make test' to verify everything is working"

dev-flow: build deploy sync ## Development flow: build + deploy + sync
	@echo "🔄 Development flow completed!"
	@echo "Run 'make status' to check the deployment"

# === CLEANUP TARGETS ===
clean: ## Clean up the entire environment
	@echo "🧹 Cleaning up environment..."
	./scripts/cleanup.sh
	@rm -f $(TAG_FILE)

clean-cluster: ## Delete only the k3d cluster
	@echo "🗑️ Deleting k3d cluster..."
	@kubectl config unset current-context 2>/dev/null || true
	@k3d cluster delete $(CLUSTER_NAME) 2>/dev/null || echo "Cluster not found"

clean-app: ## Delete only the ArgoCD application
	@echo "🗑️ Deleting ArgoCD application..."
	@kubectl delete application example-app-simple -n argocd 2>/dev/null || echo "Application not found"

restart: clean setup ## Restart the entire environment

# === UTILITY TARGETS ===
kubeconfig: ## Set kubeconfig for the cluster
	@echo "🔧 Setting kubeconfig..."
	@export KUBECONFIG=$$(k3d kubeconfig write $(CLUSTER_NAME))
	@echo "Kubeconfig set to: $$KUBECONFIG"

check-prereqs: ## Check if all prerequisites are installed
	@echo "🔍 Checking prerequisites..."
	@command -v k3d >/dev/null 2>&1 && echo "✅ k3d" || echo "❌ k3d"
	@command -v docker >/dev/null 2>&1 && echo "✅ docker" || echo "❌ docker"
	@command -v helm >/dev/null 2>&1 && echo "✅ helm" || echo "❌ helm"
	@command -v kubectl >/dev/null 2>&1 && echo "✅ kubectl" || echo "❌ kubectl"
	@command -v htpasswd >/dev/null 2>&1 && echo "✅ htpasswd" || echo "❌ htpasswd"
	@command -v curl >/dev/null 2>&1 && echo "✅ curl" || echo "❌ curl"
	@command -v jq >/dev/null 2>&1 && echo "✅ jq" || echo "❌ jq"

get-password: ## Get ArgoCD admin password
	@echo "🔑 ArgoCD admin password: admin"
	@echo "ArgoCD UI: http://localhost:8083 (admin/admin)"
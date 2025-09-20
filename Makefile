# Local GitOps Environment - Makefile

.PHONY: help setup build deploy clean status logs

# Configuration
TAG_FILE := .image-tag

# Default target
help: ## Show this help message
	@echo "Local GitOps Environment - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Run the complete setup
	@echo "🚀 Setting up Local GitOps Environment..."
	./setup.sh

build: ## Build and push Docker image and Helm chart
	@echo "🐳 Building and pushing Docker image and Helm chart..."
	@./scripts/build-and-push.sh
	@echo "📝 Updating manifest repository..."
	@./scripts/update-manifests.sh

deploy: ## Create ArgoCD app and deploy to local git
	@echo "🚀 Creating ArgoCD application..."
	@./scripts/create-argocd-app.sh

status: ## Show cluster and application status
	@echo "📊 Cluster Status:"
	@kubectl get nodes
	@echo ""
	@echo "📊 Pods Status:"
	@kubectl get pods --all-namespaces
	@echo ""
	@echo "📊 ArgoCD Applications:"
	@kubectl get applications -n argocd 2>/dev/null || echo "No ArgoCD applications found"

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
	@echo "📋 Example App Logs:"
	@kubectl logs -n default deployment/example-app --tail=20 2>/dev/null || echo "Example app not deployed"

clean: ## Clean up the entire environment
	@echo "🧹 Cleaning up environment..."
	./scripts/cleanup.sh
	@rm -f $(TAG_FILE)

restart: clean setup ## Restart the entire environment


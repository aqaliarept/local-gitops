package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the complete GitOps environment",
	Long:  "Sets up k3d cluster, ArgoCD, ChartMuseum, Git server, and all necessary resources",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Setting up Local GitOps Environment...")

	// Check prerequisites
	if err := checkPrerequisites(); err != nil {
		return fmt.Errorf("prerequisites check failed: %w", err)
	}

	// Create directories
	if err := createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Initialize git repository for manifests
	if err := initManifestRepo(); err != nil {
		return fmt.Errorf("failed to initialize manifest repository: %w", err)
	}

	// Create local registry
	if err := createRegistry(); err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	// Create k3d cluster
	if err := createCluster(); err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	// Install ArgoCD
	if err := installArgoCD(); err != nil {
		return fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	// Setup Kubernetes resources
	if err := setupK8sResources(); err != nil {
		return fmt.Errorf("failed to setup Kubernetes resources: %w", err)
	}

	fmt.Println("🎉 Setup completed successfully!")
	printAccessInfo()

	return nil
}

func checkPrerequisites() error {
	fmt.Println("📋 Checking prerequisites...")

	required := []string{"k3d", "docker", "helm", "kubectl", "htpasswd", "curl", "lsof", "jq"}

	for _, cmd := range required {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("required command not found: %s", cmd)
		}
	}

	fmt.Println("✅ All prerequisites found")
	return nil
}

func createDirectories() error {
	fmt.Println("📁 Creating directories...")

	dirs := []string{
		filepath.Join(manifestRepoPath, "..", "charts"),
		filepath.Join(manifestRepoPath, "..", "packages"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func initManifestRepo() error {
	fmt.Println("🔧 Initializing git repository for manifests...")

	// Check if already a git repo
	if _, err := os.Stat(filepath.Join(manifestRepoPath, ".git")); err == nil {
		fmt.Println("ℹ️  Manifest repository already initialized")
		return nil
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = manifestRepoPath
	if err := runCommandInteractive(cmd, "git init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user
	configCmd := exec.Command("git", "config", "user.name", "Local GitOps")
	configCmd.Dir = manifestRepoPath
	if err := runCommandInteractive(configCmd, "git config user.name"); err != nil {
		return fmt.Errorf("failed to configure git user: %w", err)
	}

	configCmd = exec.Command("git", "config", "user.email", "local@gitops.dev")
	configCmd.Dir = manifestRepoPath
	if err := runCommandInteractive(configCmd, "git config user.email"); err != nil {
		return fmt.Errorf("failed to configure git email: %w", err)
	}

	return nil
}

func createRegistry() error {
	fmt.Println("🐳 Creating local Docker registry...")

	// Check if registry already exists
	cmd := exec.Command("k3d", "registry", "list")
	output, err := runCommand(cmd, "k3d registry list")
	if err != nil {
		return fmt.Errorf("failed to list registries: %w", err)
	}

	if strings.Contains(string(output), registryName) {
		fmt.Println("ℹ️  Registry already exists")
		return nil
	}

	// Create registry
	cmd = exec.Command("k3d", "registry", "create", registryName, "--port", registryPort)
	if err := runCommandInteractive(cmd, "k3d registry create"); err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	fmt.Printf("✅ Registry created at %s:%s\n", registryName, registryPort)
	return nil
}

func createCluster() error {
	fmt.Println("☸️  Creating k3d cluster...")

	// Check if cluster already exists
	cmd := exec.Command("k3d", "cluster", "list")
	output, err := runCommand(cmd, "k3d cluster list")
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	if strings.Contains(string(output), clusterName) {
		fmt.Println("ℹ️  Cluster already exists, deleting...")
		cmd = exec.Command("k3d", "cluster", "delete", clusterName)
		if err := runCommandInteractive(cmd, "k3d cluster delete"); err != nil {
			return fmt.Errorf("failed to delete existing cluster: %w", err)
		}
	}

	// Create cluster with registry and port mappings
	// Convert manifest repo path to absolute path for k3d
	absManifestPath, err := filepath.Abs(manifestRepoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for manifest repository: %w", err)
	}

	cmd = exec.Command("k3d", "cluster", "create", clusterName,
		"--registry-use", fmt.Sprintf("k3d-%s:%s", registryName, registryPort),
		"-v", fmt.Sprintf("%s:/data/manifests@server:0", absManifestPath),
		"-p", "8083:80@loadbalancer",
		"-p", "8084:8080@loadbalancer",
		"-p", "8085:8080@loadbalancer",
		"--wait")

	if err := runCommandInteractive(cmd, "k3d cluster create"); err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	fmt.Println("✅ Cluster created with registry integration")

	// Wait for cluster to be ready
	fmt.Println("⏳ Waiting for cluster to be ready...")
	cmd = exec.Command("kubectl", "wait", "--for=condition=Ready", "nodes", "--all", "--timeout=300s")
	if err := runCommandInteractive(cmd, "kubectl wait for cluster readiness"); err != nil {
		return fmt.Errorf("failed to wait for cluster readiness: %w", err)
	}

	// Set kubeconfig
	cmd = exec.Command("k3d", "kubeconfig", "write", clusterName)
	output, err = runCommand(cmd, "k3d kubeconfig write")
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	os.Setenv("KUBECONFIG", strings.TrimSpace(string(output)))
	fmt.Println("✅ Kubeconfig configured")

	return nil
}

func installArgoCD() error {
	fmt.Println("🔄 Installing ArgoCD...")

	// Create namespace
	cmd := exec.Command("kubectl", "create", "namespace", "argocd", "--dry-run=client", "-o", "yaml")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	cmd = exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(string(output))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply namespace: %w", err)
	}

	// Install ArgoCD
	cmd = exec.Command("kubectl", "apply", "-n", "argocd", "-f", "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	// Wait for ArgoCD to be ready
	fmt.Println("⏳ Waiting for ArgoCD to be ready...")
	cmd = exec.Command("kubectl", "wait", "--for=condition=available", "--timeout=300s", "deployment/argocd-server", "-n", "argocd")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to wait for ArgoCD: %w", err)
	}

	// Configure admin password
	if err := configureArgoCDPassword(); err != nil {
		return fmt.Errorf("failed to configure ArgoCD password: %w", err)
	}

	return nil
}

func configureArgoCDPassword() error {
	fmt.Println("🔑 Setting up ArgoCD admin credentials...")

	// Generate bcrypt hash for 'admin' password
	cmd := exec.Command("htpasswd", "-niB", "admin")
	cmd.Stdin = strings.NewReader("admin")
	output, err := runCommand(cmd, "htpasswd generate password hash")
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	hash := strings.TrimSpace(strings.Split(string(output), ":")[1])

	// Update secret using kubectl patch with proper base64 encoding
	hashB64 := base64.StdEncoding.EncodeToString([]byte(hash))
	mtimeB64 := base64.StdEncoding.EncodeToString([]byte("2024-01-01T00:00:00"))

	patch := fmt.Sprintf(`{"data":{"admin.password":"%s","admin.passwordMtime":"%s"}}`,
		hashB64, mtimeB64)

	cmd = exec.Command("kubectl", "patch", "secret", "argocd-secret", "-n", "argocd", "--type", "merge", "-p", patch)
	if err := runCommandInteractive(cmd, "kubectl patch argocd secret"); err != nil {
		return fmt.Errorf("failed to update ArgoCD secret: %w", err)
	}

	// Restart ArgoCD server
	cmd = exec.Command("kubectl", "rollout", "restart", "deployment/argocd-server", "-n", "argocd")
	if err := runCommandInteractive(cmd, "kubectl rollout restart argocd-server"); err != nil {
		return fmt.Errorf("failed to restart ArgoCD server: %w", err)
	}

	cmd = exec.Command("kubectl", "rollout", "status", "deployment/argocd-server", "-n", "argocd")
	if err := runCommandInteractive(cmd, "kubectl rollout status argocd-server"); err != nil {
		return fmt.Errorf("failed to wait for ArgoCD server restart: %w", err)
	}

	fmt.Println("✅ ArgoCD admin credentials configured (admin/admin)")
	return nil
}

func setupK8sResources() error {
	fmt.Println("📦 Installing Kubernetes resources...")

	// Create namespaces
	namespaces := []string{"chartmuseum", "git-server"}
	for _, ns := range namespaces {
		cmd := exec.Command("kubectl", "create", "namespace", ns, "--dry-run=client", "-o", "yaml")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to create namespace %s: %w", ns, err)
		}

		cmd = exec.Command("kubectl", "apply", "-f", "-")
		cmd.Stdin = strings.NewReader(string(output))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to apply namespace %s: %w", ns, err)
		}
	}

	// Deploy ChartMuseum
	fmt.Println("📊 Deploying ChartMuseum...")
	chartmuseumPath := filepath.Join(manifestRepoPath, "..", "k8s", "chartmuseum", "chartmuseum.yaml")
	cmd := exec.Command("kubectl", "apply", "-f", chartmuseumPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy ChartMuseum: %w", err)
	}

	// Wait for ChartMuseum
	fmt.Println("⏳ Waiting for ChartMuseum to be ready...")
	cmd = exec.Command("kubectl", "wait", "--for=condition=available", "--timeout=300s", "deployment/chartmuseum", "-n", "chartmuseum")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to wait for ChartMuseum: %w", err)
	}

	// Deploy Git server
	fmt.Println("📦 Deploying Git server...")
	gitServerPath := filepath.Join(manifestRepoPath, "..", "k8s", "git-server", "git-server.yaml")
	cmd = exec.Command("kubectl", "apply", "-f", gitServerPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy Git server: %w", err)
	}

	// Wait for Git server
	fmt.Println("⏳ Waiting for Git server to be ready...")
	cmd = exec.Command("kubectl", "wait", "--for=condition=available", "--timeout=300s", "deployment/git-server", "-n", "git-server")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to wait for Git server: %w", err)
	}

	// Setup Git repository
	if err := setupGitRepository(); err != nil {
		return fmt.Errorf("failed to setup Git repository: %w", err)
	}

	fmt.Println("✅ Kubernetes resources setup completed successfully!")
	return nil
}

func setupGitRepository() error {
	fmt.Println("📁 Creating Git repository...")

	// Create Git repository directory
	cmd := exec.Command("kubectl", "exec", "-n", "git-server", "deployment/git-server", "--", "mkdir", "-p", "/git")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create git directory: %w", err)
	}

	// Initialize bare Git repository
	fmt.Println("🔧 Initializing bare Git repository...")
	initScript := `
echo 'Initializing bare repository...'
git init --bare /git/manifest.git
echo 'Git repository initialized successfully'
`

	cmd = exec.Command("kubectl", "exec", "-n", "git-server", "deployment/git-server", "--", "sh", "-c", initScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	fmt.Println("✅ Git repository setup completed")
	return nil
}

func printAccessInfo() {
	fmt.Println("")
	fmt.Println("📋 Access Information:")
	fmt.Println("  ArgoCD UI: http://localhost:8083")
	fmt.Println("  ArgoCD Username: admin")
	fmt.Println("  ArgoCD Password: admin")
	fmt.Println("  ChartMuseum: http://localhost:8084")
	fmt.Printf("  Local Registry: %s:%s\n", registryName, registryPort)
	fmt.Println("")
	fmt.Printf("📁 Manifest Repository: %s\n", manifestRepoPath)
	fmt.Println("")
	fmt.Println("✅ Local GitOps environment is ready!")
}

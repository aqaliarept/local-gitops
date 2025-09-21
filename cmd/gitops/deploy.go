package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Create/update ArgoCD application",
	Long:  "Creates or updates an ArgoCD application to deploy the example app using ChartMuseum and Git repository",
	RunE:  runDeploy,
}

var (
	appName            = "nginx-app"
	argocdNamespace    = "argocd"
	deployChartName    = "nginx-app"
	deployChartVersion = "0.1.0"
	valuesPath         = "nginx-app-values.yaml"
)

func runDeploy(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Deploying application...")

	// Check prerequisites
	if err := checkDeployPrerequisites(); err != nil {
		return fmt.Errorf("prerequisites check failed: %w", err)
	}

	// Apply bootstrap.yaml to create ArgoCD application
	if err := applyBootstrap(); err != nil {
		return fmt.Errorf("failed to apply bootstrap: %w", err)
	}

	// Push manifest content to Git repository
	if err := pushManifestContent(); err != nil {
		return fmt.Errorf("failed to push manifest content: %w", err)
	}

	fmt.Println("✅ Deployment completed successfully!")
	return nil
}

func checkDeployPrerequisites() error {
	// Check if kubectl is available
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl is required but not installed")
	}

	// Check if k3d cluster is running
	cmd := exec.Command("k3d", "cluster", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	if !strings.Contains(string(output), clusterName) {
		return fmt.Errorf("cluster %s is not running. Please run 'gitops setup' first", clusterName)
	}

	// Set kubeconfig
	cmd = exec.Command("k3d", "kubeconfig", "write", clusterName)
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	os.Setenv("KUBECONFIG", strings.TrimSpace(string(output)))

	// Check if ArgoCD is running
	cmd = exec.Command("kubectl", "get", "deployment", "argocd-server", "-n", argocdNamespace)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ArgoCD is not running. Please run 'gitops setup' first")
	}

	return nil
}

func applyBootstrap() error {
	fmt.Println("📋 Applying bootstrap.yaml...")

	// Check if bootstrap.yaml exists
	bootstrapPath := filepath.Join(manifestRepoPath, "bootstrap.yaml")
	if _, err := os.Stat(bootstrapPath); err != nil {
		return fmt.Errorf("bootstrap.yaml not found: %w", err)
	}

	// Apply bootstrap.yaml
	cmd := exec.Command("kubectl", "apply", "-f", bootstrapPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply bootstrap.yaml: %w", err)
	}

	fmt.Println("✅ Bootstrap.yaml applied successfully")
	return nil
}

func pushManifestContent() error {
	fmt.Println("📤 Pushing manifest content to Git repository...")

	// Start port forward for git server
	fmt.Println("🌐 Starting Git server port forward...")
	cmd := exec.Command("kubectl", "port-forward", "-n", "git-server", "svc/git-server", "8085:80")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start port forward: %w", err)
	}
	defer cmd.Process.Kill()

	// Wait for port forward to establish
	time.Sleep(2 * time.Second)

	// Use Docker container to copy manifest folder content, init git repo, and push
	fmt.Println("🐳 Using Docker container to copy manifest folder, init, and push...")
	pushScript := `
set -e  # Exit on any error
echo "=== DOCKER CONTAINER DEBUG ==="
echo "Source directory contents:"
ls -la /source/
echo "Manifest directory contents:"
ls -la /source/manifest/
echo "Creating workspace in container /tmp..."
mkdir -p /tmp/workspace
echo "Copying manifest folder content to container /tmp/workspace..."
cp -r /source/manifest/* /tmp/workspace/
cd /tmp/workspace
echo "Files in workspace after copy:"
ls -la
echo "Initializing git repository..."
git init
git config --global user.email 'gitops@example.com'
git config --global user.name 'GitOps CLI'
echo "Adding remote origin..."
git remote add origin http://host.docker.internal:8085/manifest.git
echo "Adding and committing files..."
git add .
echo "Git status after add:"
git status
git commit -m "Initial commit: nginx application manifests"
echo "Pushing to remote repository..."
git push -u origin master --force
echo "=== END DOCKER CONTAINER DEBUG ==="
`

	cmd = exec.Command("docker", "run", "--rm", "--entrypoint=",
		"-v", fmt.Sprintf("%s:/source", manifestRepoPath),
		"alpine/git:latest", "/bin/sh", "-c", pushScript)

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push manifest content: %w", err)
	}

	fmt.Println("✅ Manifest content pushed successfully")
	return nil
}

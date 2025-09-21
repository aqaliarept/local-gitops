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
	appName         = "nginx-app"
	argocdNamespace = "argocd"
	deployTargetDir string
)

func init() {
	deployCmd.Flags().StringVar(&deployTargetDir, "target-dir", ".", "Target directory containing .gitops-config.yaml")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Deploying application...")

	// Read configuration
	config, err := readConfig(deployTargetDir)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if verbose {
		fmt.Printf("📋 Using cluster name: %s\n", config.ClusterName)
	}

	// Check prerequisites
	if err := checkDeployPrerequisites(config.ClusterName); err != nil {
		return fmt.Errorf("prerequisites check failed: %w", err)
	}

	// Apply bootstrap.yaml to create ArgoCD application
	if err := applyBootstrap(deployTargetDir); err != nil {
		return fmt.Errorf("failed to apply bootstrap: %w", err)
	}

	// Push manifest content to Git repository
	if err := pushManifestContent(config, deployTargetDir); err != nil {
		return fmt.Errorf("failed to push manifest content: %w", err)
	}

	// Sync ArgoCD application
	if err := syncArgoCDApplication(); err != nil {
		return fmt.Errorf("failed to sync ArgoCD application: %w", err)
	}

	fmt.Println("✅ Deployment completed successfully!")
	return nil
}

func checkDeployPrerequisites(clusterName string) error {
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

func applyBootstrap(targetDir string) error {
	fmt.Println("📋 Applying bootstrap.yaml...")

	// Check if bootstrap.yaml exists in target directory
	bootstrapPath := filepath.Join(targetDir, "bootstrap.yaml")
	if _, err := os.Stat(bootstrapPath); err != nil {
		return fmt.Errorf("bootstrap.yaml not found in %s: %w", targetDir, err)
	}

	// Apply bootstrap.yaml
	cmd := exec.Command("kubectl", "apply", "-f", bootstrapPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply bootstrap.yaml: %w", err)
	}

	fmt.Println("✅ Bootstrap.yaml applied successfully")
	return nil
}

func pushManifestContent(config *Config, targetDir string) error {
	fmt.Println("📤 Pushing manifest content to Git repository...")

	// Start port forward for git server
	fmt.Println("🌐 Starting Git server port forward...")
	portForwardCmd := exec.Command("kubectl", "port-forward", "-n", "git-server", "svc/git-server", fmt.Sprintf("%s:80", config.GitServerPort))
	portForwardCmd.Stdout = nil
	portForwardCmd.Stderr = nil
	if err := portForwardCmd.Start(); err != nil {
		return fmt.Errorf("failed to start port forward: %w", err)
	}
	defer func() {
		if portForwardCmd.Process != nil {
			portForwardCmd.Process.Kill()
		}
	}()

	// Wait for port forward to establish
	fmt.Println("⏳ Waiting for port forward to establish...")
	time.Sleep(3 * time.Second)

	// Test if port forward is working
	fmt.Println("🔍 Testing port forward connection...")
	testCmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", fmt.Sprintf("http://localhost:%s/manifest.git/info/refs?service=git-upload-pack", config.GitServerPort))
	output, err := testCmd.Output()
	if err != nil || strings.TrimSpace(string(output)) != "200" {
		return fmt.Errorf("port forward test failed: %w", err)
	}
	fmt.Println("✅ Port forward is working")

	// Use Docker container to copy manifest folder content, init git repo, and push
	fmt.Println("🐳 Using Docker container to copy manifest folder, init, and push...")
	pushScript := fmt.Sprintf(`
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
git remote add origin http://host.docker.internal:%s/manifest.git
echo "Adding and committing files..."
git add .
echo "Git status after add:"
git status
git commit -m "Initial commit: nginx application manifests"
echo "Pushing to remote repository..."
git push -u origin master --force
echo "=== END DOCKER CONTAINER DEBUG ==="
`, config.GitServerPort)

	dockerCmd := exec.Command("docker", "run", "--rm", "--entrypoint=",
		"-v", fmt.Sprintf("%s:/source", targetDir),
		"alpine/git:latest", "/bin/sh", "-c", pushScript)

	if verbose {
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
	}

	if err := dockerCmd.Run(); err != nil {
		return fmt.Errorf("failed to push manifest content: %w", err)
	}

	fmt.Println("✅ Manifest content pushed successfully")
	return nil
}

func syncArgoCDApplication() error {
	fmt.Println("🔄 Syncing ArgoCD application...")

	// Wait a moment for the git push to be available
	time.Sleep(2 * time.Second)

	// Trigger ArgoCD application sync
	cmd := exec.Command("kubectl", "patch", "application", appName, "-n", argocdNamespace, "--type", "merge", "--patch", `{"operation":{"sync":{}}}`)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to trigger ArgoCD sync: %w", err)
	}

	// Wait for sync to complete
	fmt.Println("⏳ Waiting for ArgoCD sync to complete...")
	time.Sleep(5 * time.Second)

	// Check sync status
	cmd = exec.Command("kubectl", "get", "application", appName, "-n", argocdNamespace, "-o", "jsonpath={.status.sync.status}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check sync status: %w", err)
	}

	syncStatus := strings.TrimSpace(string(output))
	if verbose {
		fmt.Printf("📋 ArgoCD sync status: %s\n", syncStatus)
	}

	if syncStatus == "Synced" {
		fmt.Println("✅ ArgoCD application synced successfully")
	} else {
		fmt.Printf("⚠️  ArgoCD application sync status: %s (may still be in progress)\n", syncStatus)
	}

	return nil
}

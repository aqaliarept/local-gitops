package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build and push Docker image and Helm chart",
	Long:  "Builds Docker image, pushes to registry, packages Helm chart, pushes to ChartMuseum, and updates manifest repository",
	RunE:  runBuild,
}

var (
	imageName         = "nginx-app"
	dockerfilePath    = "./manifest/nginx-app/Dockerfile"
	buildContext      = "./manifest/nginx-app"
	chartDir          = "./manifest/charts/nginx-app"
	packagesDir       = "./packages"
	buildChartVersion = "0.1.0"
	tagFile           = ".image-tag"
)

func runBuild(cmd *cobra.Command, args []string) error {
	fmt.Println("🐳 Building and pushing Docker image and Helm chart...")

	// Check prerequisites
	if err := checkBuildPrerequisites(); err != nil {
		return fmt.Errorf("prerequisites check failed: %w", err)
	}

	// Generate timestamp
	timestamp := time.Now().Format("20060102-150405")
	if err := ioutil.WriteFile(tagFile, []byte(timestamp), 0644); err != nil {
		return fmt.Errorf("failed to write tag file: %w", err)
	}
	fmt.Printf("Generated timestamp: %s\n", timestamp)

	// Build and push Docker image (skip for nginx)
	if err := buildAndPushImage(timestamp); err != nil {
		fmt.Printf("⚠️  Skipping Docker build (likely using nginx): %v\n", err)
	}

	// Build and push Helm chart
	if err := buildAndPushChart(timestamp); err != nil {
		return fmt.Errorf("failed to build and push chart: %w", err)
	}

	// Update manifest repository
	if err := updateManifests(timestamp); err != nil {
		return fmt.Errorf("failed to update manifests: %w", err)
	}

	// Push manifest changes
	if err := pushManifests(); err != nil {
		return fmt.Errorf("failed to push manifests: %w", err)
	}

	fmt.Println("✅ Build and push completed successfully!")
	return nil
}

func checkBuildPrerequisites() error {
	required := []string{"docker", "helm", "curl", "kubectl"}

	for _, cmd := range required {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("required command not found: %s", cmd)
		}
	}

	// Check if registry is running
	cmd := exec.Command("docker", "ps")
	output, err := runCommand(cmd, "docker ps")
	if err != nil {
		return fmt.Errorf("failed to check docker ps: %w", err)
	}

	if !strings.Contains(string(output), fmt.Sprintf("k3d-%s", registryName)) {
		return fmt.Errorf("local registry is not running. Please run 'gitops setup' first")
	}

	// Check if k3d cluster is running
	cmd = exec.Command("k3d", "cluster", "list")
	output, err = runCommand(cmd, "k3d cluster list")
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	if !strings.Contains(string(output), clusterName) {
		return fmt.Errorf("cluster %s is not running. Please run 'gitops setup' first", clusterName)
	}

	// Set kubeconfig
	cmd = exec.Command("k3d", "kubeconfig", "write", clusterName)
	output, err = runCommand(cmd, "k3d kubeconfig write")
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	os.Setenv("KUBECONFIG", strings.TrimSpace(string(output)))

	// Check if ChartMuseum is running
	cmd = exec.Command("kubectl", "get", "deployment", "chartmuseum", "-n", "chartmuseum")
	if err := runCommandInteractive(cmd, "kubectl get chartmuseum deployment"); err != nil {
		return fmt.Errorf("ChartMuseum is not running. Please run 'gitops setup' first")
	}

	return nil
}

func buildAndPushImage(timestamp string) error {
	fmt.Printf("Building and pushing Docker image with tag: %s\n", timestamp)

	registryURL := fmt.Sprintf("localhost:%s", registryPort)
	fullImageName := fmt.Sprintf("%s/%s:%s", registryURL, imageName, timestamp)

	// Build the image
	cmd := exec.Command("docker", "build", "-t", fmt.Sprintf("%s:%s", imageName, timestamp), "-f", dockerfilePath, buildContext)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build docker image: %w", err)
	}

	// Tag for local registry
	cmd = exec.Command("docker", "tag", fmt.Sprintf("%s:%s", imageName, timestamp), fullImageName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	// Push to local registry
	cmd = exec.Command("docker", "push", fullImageName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	fmt.Printf("Docker image built and pushed successfully: %s\n", fullImageName)
	return nil
}

func buildAndPushChart(timestamp string) error {
	fmt.Println("Building and pushing Helm chart...")

	// Get chart name from Chart.yaml
	chartYamlPath := filepath.Join(chartDir, "Chart.yaml")
	content, err := ioutil.ReadFile(chartYamlPath)
	if err != nil {
		return fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	chartName := ""
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "name:") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				chartName = parts[1]
				break
			}
		}
	}

	if chartName == "" {
		return fmt.Errorf("could not find chart name in Chart.yaml")
	}

	// Update chart values with timestamped tag
	fmt.Printf("Updating chart values with tag: %s\n", timestamp)
	valuesPath := filepath.Join(chartDir, "values.yaml")
	content, err = ioutil.ReadFile(valuesPath)
	if err != nil {
		return fmt.Errorf("failed to read values.yaml: %w", err)
	}

	// Replace tag in values.yaml
	updatedContent := strings.ReplaceAll(string(content),
		`tag: ".*"`,
		fmt.Sprintf(`tag: "%s"`, timestamp))

	// Create backup
	backupPath := valuesPath + ".bak"
	if err := ioutil.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Write updated content
	if err := ioutil.WriteFile(valuesPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated values.yaml: %w", err)
	}

	// Create packages directory if it doesn't exist
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	// Package the chart
	cmd := exec.Command("helm", "package", chartDir, "--version", buildChartVersion, "--destination", packagesDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to package chart: %w", err)
	}

	packageFile := filepath.Join(packagesDir, fmt.Sprintf("%s-%s.tgz", chartName, buildChartVersion))
	if _, err := os.Stat(packageFile); err != nil {
		return fmt.Errorf("package file not created: %s", packageFile)
	}

	// Push to ChartMuseum using port-forward
	fmt.Println("Pushing chart to ChartMuseum...")
	cmd = exec.Command("kubectl", "port-forward", "-n", "chartmuseum", "svc/chartmuseum", "8084:8080")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start port forward: %w", err)
	}

	// Wait a bit for port forward to establish
	time.Sleep(2 * time.Second)

	// Push the chart
	cmd = exec.Command("curl", "--data-binary", "@"+packageFile, "http://localhost:8084/api/charts")
	if err := cmd.Run(); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to push chart to ChartMuseum: %w", err)
	}

	// Clean up port forward
	cmd.Process.Kill()

	// Restore original values.yaml
	if err := os.Rename(backupPath, valuesPath); err != nil {
		return fmt.Errorf("failed to restore values.yaml: %w", err)
	}

	fmt.Printf("Helm chart built and pushed successfully: %s version %s\n", chartName, buildChartVersion)
	return nil
}

func updateManifests(timestamp string) error {
	fmt.Println("📝 Updating manifest repository...")

	valuesFile := filepath.Join(manifestRepoPath, "example-app-values.yaml")
	if _, err := os.Stat(valuesFile); err != nil {
		return fmt.Errorf("values file not found: %s", valuesFile)
	}

	// Read current values
	content, err := ioutil.ReadFile(valuesFile)
	if err != nil {
		return fmt.Errorf("failed to read values file: %w", err)
	}

	// Update the image tag
	updatedContent := strings.ReplaceAll(string(content),
		`tag: ".*"`,
		fmt.Sprintf(`tag: "%s"`, timestamp))

	// Write updated content
	if err := ioutil.WriteFile(valuesFile, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated values file: %w", err)
	}

	fmt.Printf("Values manifest updated: %s\n", valuesFile)
	fmt.Printf("Using image: k3d-%s:%s/%s:%s\n", registryName, registryPort, imageName, timestamp)
	return nil
}

func pushManifests() error {
	fmt.Println("📤 Pushing manifest changes to Git remote...")

	// Check if manifest directory exists
	if _, err := os.Stat(manifestRepoPath); err != nil {
		return fmt.Errorf("manifest directory not found: %s", manifestRepoPath)
	}

	// Check if it's a git repository
	if _, err := os.Stat(filepath.Join(manifestRepoPath, ".git")); err != nil {
		return fmt.Errorf("not a git repository: %s", manifestRepoPath)
	}

	// Change to manifest directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(manifestRepoPath); err != nil {
		return fmt.Errorf("failed to change to manifest directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// Check if there are any changes
	cmd := exec.Command("git", "diff", "--quiet")
	if err := cmd.Run(); err == nil {
		cmd = exec.Command("git", "diff", "--cached", "--quiet")
		if err := cmd.Run(); err == nil {
			fmt.Println("✅ No changes to commit")
			return nil
		}
	}

	// Add all changes
	fmt.Println("📝 Adding changes to git...")
	cmd = exec.Command("git", "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit changes
	fmt.Println("💾 Committing changes...")
	commitMsg := fmt.Sprintf("Update application values - %s", time.Now().Format("2006-01-02 15:04:05"))
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if err := cmd.Run(); err != nil {
		// Check if there's nothing to commit
		if strings.Contains(err.Error(), "nothing to commit") {
			fmt.Println("✅ No changes to commit")
			return nil
		}
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Start port forward for git server
	fmt.Println("🌐 Starting Git server port forward...")
	cmd = exec.Command("kubectl", "port-forward", "-n", "git-server", "svc/git-server", "8085:80")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start git port forward: %w", err)
	}

	// Wait for port forward to establish
	time.Sleep(3 * time.Second)

	// Push to remote
	fmt.Println("🚀 Pushing to remote repository...")
	cmd = exec.Command("git", "push", "origin", "master")
	if err := cmd.Run(); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to push to remote: %w", err)
	}

	// Clean up port forward
	cmd.Process.Kill()

	fmt.Println("✅ Manifest changes pushed successfully")

	// Get repository info
	cmd = exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err == nil {
		fmt.Printf("📋 Repository: %s\n", strings.TrimSpace(string(output)))
	}

	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err = cmd.Output()
	if err == nil {
		fmt.Printf("📋 Commit: %s\n", strings.TrimSpace(string(output)))
	}

	return nil
}

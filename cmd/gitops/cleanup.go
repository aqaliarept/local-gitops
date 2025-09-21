package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up local GitOps environment",
	Long:  "Removes k3d cluster, registry, and all associated resources",
	RunE:  runCleanup,
}

var cleanupTargetDir string

func init() {
	cleanupCmd.Flags().StringVar(&cleanupTargetDir, "target-dir", ".", "Target directory containing .gitops-config.yaml")
}

func runCleanup(cmd *cobra.Command, args []string) error {
	fmt.Println("🧹 Cleaning up Local GitOps Environment...")

	// Read config from target directory
	targetClusterName, err := readConfigClusterName()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if verbose {
		fmt.Printf("📋 Using cluster name: %s\n", targetClusterName)
	}

	// Set kubeconfig
	if err := setKubeconfig(targetClusterName); err != nil {
		return fmt.Errorf("failed to set kubeconfig: %w", err)
	}

	// Delete k3d cluster
	if err := deleteCluster(targetClusterName); err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	// Delete registry
	if err := deleteRegistry(); err != nil {
		return fmt.Errorf("failed to delete registry: %w", err)
	}

	fmt.Println("✅ Cleanup completed successfully!")
	return nil
}

func readConfigClusterName() (string, error) {
	configPath := filepath.Join(cleanupTargetDir, ".gitops-config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		// If no config file, use default cluster name
		return "devcluster", nil
	}

	// Read config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	// Simple parsing - look for cluster_name line
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cluster_name:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	// If not found, use default
	return "devcluster", nil
}

func deleteCluster(clusterName string) error {
	fmt.Printf("🗑️  Deleting k3d cluster: %s\n", clusterName)

	// Check if cluster exists
	cmd := exec.Command("k3d", "cluster", "list")
	output, err := runCommand(cmd, "k3d cluster list")
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("📋 Available clusters: %s\n", strings.TrimSpace(string(output)))
	}

	if !strings.Contains(string(output), clusterName) {
		fmt.Printf("ℹ️  Cluster %s not found\n", clusterName)
		return nil
	}

	// Delete cluster
	cmd = exec.Command("k3d", "cluster", "delete", clusterName)
	if _, err := runCommand(cmd, "k3d cluster delete"); err != nil {
		return err
	}

	fmt.Printf("✅ Cluster %s deleted\n", clusterName)
	return nil
}

func deleteRegistry() error {
	fmt.Printf("🗑️  Deleting Docker registry: %s\n", registryName)

	// Check if registry exists
	cmd := exec.Command("k3d", "registry", "list")
	output, err := runCommand(cmd, "k3d registry list")
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("📋 Available registries: %s\n", strings.TrimSpace(string(output)))
	}

	if !strings.Contains(string(output), registryName) {
		fmt.Printf("ℹ️  Registry %s not found\n", registryName)
		return nil
	}

	// Delete registry
	cmd = exec.Command("k3d", "registry", "delete", registryName)
	if _, err := runCommand(cmd, "k3d registry delete"); err != nil {
		return err
	}

	fmt.Printf("✅ Registry %s deleted\n", registryName)
	return nil
}

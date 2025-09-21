package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up the entire environment",
	Long:  "Deletes the k3d cluster, registry, and optionally cleans up local files",
	RunE:  runCleanup,
}

var (
	cleanupLocalFiles bool
)

func init() {
	cleanupCmd.Flags().BoolVar(&cleanupLocalFiles, "local-files", false, "Also clean up local manifests and charts")
}

func runCleanup(cmd *cobra.Command, args []string) error {
	fmt.Println("🧹 Cleaning up Local GitOps Environment")

	// Confirm cleanup
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Are you sure you want to delete the cluster and registry? (y/N): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("❌ Cleanup cancelled")
		return nil
	}

	// Delete k3d cluster
	if err := deleteCluster(); err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	// Delete registry
	if err := deleteRegistry(); err != nil {
		return fmt.Errorf("failed to delete registry: %w", err)
	}

	// Clean up local files if requested
	if cleanupLocalFiles {
		if err := cleanupLocalFilesFunc(); err != nil {
			return fmt.Errorf("failed to cleanup local files: %w", err)
		}
	}

	// Remove tag file
	if err := os.Remove(tagFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove tag file: %v\n", err)
	}

	fmt.Println("🎉 Cleanup completed successfully!")
	fmt.Println("💡 To start fresh, run: gitops setup")
	return nil
}

func deleteCluster() error {
	fmt.Println("🗑️  Deleting k3d cluster...")

	cmd := exec.Command("k3d", "cluster", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	if strings.Contains(string(output), clusterName) {
		cmd = exec.Command("k3d", "cluster", "delete", clusterName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to delete cluster: %w", err)
		}
		fmt.Println("✅ Cluster deleted")
	} else {
		fmt.Printf("ℹ️  Cluster %s not found\n", clusterName)
	}

	return nil
}

func deleteRegistry() error {
	fmt.Println("🗑️  Deleting registry...")

	cmd := exec.Command("k3d", "registry", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list registries: %w", err)
	}

	if strings.Contains(string(output), registryName) {
		cmd = exec.Command("k3d", "registry", "delete", registryName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to delete registry: %w", err)
		}
		fmt.Println("✅ Registry deleted")
	} else {
		fmt.Printf("ℹ️  Registry %s not found\n", registryName)
	}

	return nil
}

func cleanupLocalFilesFunc() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to clean up local manifests and charts? (y/N): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("ℹ️  Skipping local files cleanup")
		return nil
	}

	fmt.Println("🗑️  Cleaning up local files...")

	// Remove packages directory
	packagesPath := filepath.Join(manifestRepoPath, "..", "packages")
	if _, err := os.Stat(packagesPath); err == nil {
		if err := os.RemoveAll(packagesPath); err != nil {
			return fmt.Errorf("failed to remove packages directory: %w", err)
		}
		fmt.Println("✅ Packages directory removed")
	}

	// Reset manifest.git repository
	gitPath := filepath.Join(manifestRepoPath, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		// Change to manifest directory
		originalDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		if err := os.Chdir(manifestRepoPath); err != nil {
			return fmt.Errorf("failed to change to manifest directory: %w", err)
		}

		// Clean and reset git repository
		cmd := exec.Command("git", "clean", "-fd")
		if err := cmd.Run(); err != nil {
			os.Chdir(originalDir)
			return fmt.Errorf("failed to clean git repository: %w", err)
		}

		cmd = exec.Command("git", "reset", "--hard", "HEAD")
		if err := cmd.Run(); err != nil {
			os.Chdir(originalDir)
			return fmt.Errorf("failed to reset git repository: %w", err)
		}

		os.Chdir(originalDir)
		fmt.Println("✅ Manifest.git repository reset")
	}

	// Remove charts directory contents
	chartsPath := filepath.Join(manifestRepoPath, "..", "charts")
	if _, err := os.Stat(chartsPath); err == nil {
		// Remove all contents but keep the directory
		entries, err := os.ReadDir(chartsPath)
		if err != nil {
			return fmt.Errorf("failed to read charts directory: %w", err)
		}

		for _, entry := range entries {
			entryPath := filepath.Join(chartsPath, entry.Name())
			if err := os.RemoveAll(entryPath); err != nil {
				return fmt.Errorf("failed to remove chart entry: %w", err)
			}
		}
		fmt.Println("✅ Charts directory cleaned")
	}

	return nil
}

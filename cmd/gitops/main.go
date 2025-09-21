package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	manifestRepoPath string
	clusterName      = "devcluster"
	registryName     = "myregistry.localhost"
	registryPort     = "5001"
	verbose          bool
	initDir          string
	rootCmd          = &cobra.Command{
		Use:   "gitops",
		Short: "Local GitOps Environment CLI",
		Long:  "A CLI tool for managing a local GitOps environment with k3d, ArgoCD, ChartMuseum, and Git server",
	}
)

func main() {

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&manifestRepoPath, "manifest-repo", "m", "./manifest.git", "Path to the manifest repository folder")
	rootCmd.PersistentFlags().StringVar(&clusterName, "cluster", "devcluster", "k3d cluster name")
	rootCmd.PersistentFlags().StringVar(&registryName, "registry", "myregistry.localhost", "Docker registry name")
	rootCmd.PersistentFlags().StringVar(&registryPort, "registry-port", "5001", "Docker registry port")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&initDir, "init-dir", "", "Initialize a new GitOps directory with ApplicationSet (fails if directory exists)")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cleanupCmd)
	rootCmd.AddCommand(portForwardCmd)
	rootCmd.AddCommand(testCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

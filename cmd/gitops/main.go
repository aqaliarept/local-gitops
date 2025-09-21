package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version      = "dev"
	commit       = "unknown"
	buildTime    = "unknown"
	clusterName  = "devcluster"
	registryName = "myregistry.localhost"
	registryPort = "5001"
	verbose      bool
	targetDir    string
	rootCmd      = &cobra.Command{
		Use:     "gitops",
		Short:   "Local GitOps Environment CLI",
		Long:    "A CLI tool for managing a local GitOps environment with k3d, ArgoCD, ChartMuseum, and Git server",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildTime),
	}
)

func main() {

	// Global flags
	rootCmd.PersistentFlags().StringVar(&clusterName, "cluster", "devcluster", "k3d cluster name")
	rootCmd.PersistentFlags().StringVar(&registryName, "registry", "myregistry.localhost", "Docker registry name")
	rootCmd.PersistentFlags().StringVar(&registryPort, "registry-port", "5001", "Docker registry port")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&targetDir, "target-dir", "", "Initialize a new GitOps directory (fails if directory exists)")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cleanupCmd)
	rootCmd.AddCommand(portForwardCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

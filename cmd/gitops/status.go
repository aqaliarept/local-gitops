package main

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	statusTargetDir string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cluster and application status",
	Long:  "Displays the status of the cluster, pods, ArgoCD applications, and example app",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().StringVar(&statusTargetDir, "target-dir", ".", "Target directory containing .gitops-config.yaml")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("📊 Cluster Status:")

	// Read configuration
	config, err := readConfig(statusTargetDir)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if verbose {
		fmt.Printf("📋 Using cluster name: %s\n", config.ClusterName)
	}

	// Set kubeconfig
	if err := setKubeconfig(config.ClusterName); err != nil {
		return fmt.Errorf("failed to set kubeconfig: %w", err)
	}

	// Show cluster status
	if err := showClusterStatus(); err != nil {
		return fmt.Errorf("failed to show cluster status: %w", err)
	}

	// Show pods status
	if err := showPodsStatus(); err != nil {
		return fmt.Errorf("failed to show pods status: %w", err)
	}

	// Show ArgoCD applications
	if err := showArgoCDApplications(); err != nil {
		return fmt.Errorf("failed to show ArgoCD applications: %w", err)
	}

	// Show example app status
	if err := showExampleAppStatus(); err != nil {
		return fmt.Errorf("failed to show example app status: %w", err)
	}

	return nil
}

func showClusterStatus() error {
	cmd := exec.Command("kubectl", "get", "nodes")
	output, err := runCommand(cmd, "kubectl get nodes")
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func showPodsStatus() error {
	fmt.Println("")
	fmt.Println("📊 Pods Status:")
	cmd := exec.Command("kubectl", "get", "pods", "--all-namespaces")
	output, err := runCommand(cmd, "kubectl get pods --all-namespaces")
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func showArgoCDApplications() error {
	fmt.Println("")
	fmt.Println("📊 ArgoCD Applications:")
	cmd := exec.Command("kubectl", "get", "applications", "-n", "argocd")
	output, err := runCommand(cmd, "kubectl get applications -n argocd")
	if err != nil {
		fmt.Println("No ArgoCD applications found")
		return nil
	}
	fmt.Println(string(output))
	return nil
}

func showExampleAppStatus() error {
	fmt.Println("")
	fmt.Println("📊 Example App Status:")
	cmd := exec.Command("kubectl", "get", "pods,svc,ingress", "-l", "app=example-app")
	output, err := runCommand(cmd, "kubectl get pods,svc,ingress -l app=example-app")
	if err != nil {
		fmt.Println("Example app not deployed")
		return nil
	}
	fmt.Println(string(output))
	return nil
}

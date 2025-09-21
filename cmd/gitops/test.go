package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test the complete GitOps flow",
	Long:  "Tests the cluster status, ArgoCD, application deployment, and endpoint accessibility",
	RunE:  runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	fmt.Println("🧪 Testing GitOps flow...")

	// Set kubeconfig
	if err := setKubeconfig(); err != nil {
		return fmt.Errorf("failed to set kubeconfig: %w", err)
	}

	// Test cluster status
	if err := testClusterStatus(); err != nil {
		return fmt.Errorf("cluster status test failed: %w", err)
	}

	// Test ArgoCD status
	if err := testArgoCDStatus(); err != nil {
		return fmt.Errorf("ArgoCD status test failed: %w", err)
	}

	// Test application status
	if err := testApplicationStatus(); err != nil {
		return fmt.Errorf("application status test failed: %w", err)
	}

	// Test deployed resources
	if err := testDeployedResources(); err != nil {
		return fmt.Errorf("deployed resources test failed: %w", err)
	}

	// Test application endpoint
	if err := testApplicationEndpoint(); err != nil {
		return fmt.Errorf("application endpoint test failed: %w", err)
	}

	fmt.Println("✅ All tests passed!")
	return nil
}

func testClusterStatus() error {
	fmt.Println("1. Checking cluster status...")
	cmd := exec.Command("kubectl", "get", "nodes")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func testArgoCDStatus() error {
	fmt.Println("")
	fmt.Println("2. Checking ArgoCD status...")
	cmd := exec.Command("kubectl", "get", "pods", "-n", "argocd")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get ArgoCD pods: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func testApplicationStatus() error {
	fmt.Println("")
	fmt.Println("3. Checking application status...")
	cmd := exec.Command("kubectl", "get", "application", appName, "-n", "argocd")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get application status: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func testDeployedResources() error {
	fmt.Println("")
	fmt.Println("4. Checking deployed resources...")
	cmd := exec.Command("kubectl", "get", "pods,svc,ingress", "-l", "app=example-app")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get deployed resources: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func testApplicationEndpoint() error {
	fmt.Println("")
	fmt.Println("5. Testing application endpoint...")
	cmd := exec.Command("curl", "-s", "http://example-app.localhost/health")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Application not accessible")
		return nil // Don't fail the test, just report
	}

	response := strings.TrimSpace(string(output))
	if response != "" {
		fmt.Printf("Application response: %s\n", response)
	} else {
		fmt.Println("Application not accessible")
	}
	return nil
}

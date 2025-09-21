package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var portForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Start port forwarding for all services",
	Long:  "Starts port forwarding for ArgoCD UI, ChartMuseum, Git server, and example app",
	RunE:  runPortForward,
}

var (
	portForwardService string
)

func init() {
	portForwardCmd.Flags().StringVarP(&portForwardService, "service", "s", "", "Port forward specific service (argocd, chartmuseum, git-server)")
}

func runPortForward(cmd *cobra.Command, args []string) error {
	// Set kubeconfig
	if err := setKubeconfig(); err != nil {
		return fmt.Errorf("failed to set kubeconfig: %w", err)
	}

	if portForwardService != "" {
		return portForwardSpecificService(portForwardService)
	}

	return portForwardAllServices()
}

func portForwardAllServices() error {
	fmt.Println("🌐 Starting port forwarding...")
	fmt.Println("ArgoCD UI: http://localhost:8083 (admin/admin)")
	fmt.Println("ChartMuseum: http://localhost:8084")
	fmt.Println("Git Server: http://localhost:8085")
	fmt.Println("Example App: http://example-app.localhost")
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop port forwarding")

	// Start port forwards
	processes := make([]*exec.Cmd, 0, 3)

	// ArgoCD
	argocdCmd := exec.Command("kubectl", "port-forward", "-n", "argocd", "svc/argocd-server", "8083:443")
	argocdCmd.Stdout = nil
	argocdCmd.Stderr = nil
	if err := argocdCmd.Start(); err != nil {
		return fmt.Errorf("failed to start ArgoCD port forward: %w", err)
	}
	processes = append(processes, argocdCmd)

	// ChartMuseum
	chartmuseumCmd := exec.Command("kubectl", "port-forward", "-n", "chartmuseum", "svc/chartmuseum", "8084:8080")
	chartmuseumCmd.Stdout = nil
	chartmuseumCmd.Stderr = nil
	if err := chartmuseumCmd.Start(); err != nil {
		killProcesses(processes)
		return fmt.Errorf("failed to start ChartMuseum port forward: %w", err)
	}
	processes = append(processes, chartmuseumCmd)

	// Git Server
	gitServerCmd := exec.Command("kubectl", "port-forward", "-n", "git-server", "svc/git-server", "8085:80")
	gitServerCmd.Stdout = nil
	gitServerCmd.Stderr = nil
	if err := gitServerCmd.Start(); err != nil {
		killProcesses(processes)
		return fmt.Errorf("failed to start Git server port forward: %w", err)
	}
	processes = append(processes, gitServerCmd)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Stopping port forwarding...")
	killProcesses(processes)
	fmt.Println("✅ Port forwarding stopped")

	return nil
}

func portForwardSpecificService(service string) error {
	switch strings.ToLower(service) {
	case "argocd":
		return portForwardArgoCD()
	case "chartmuseum":
		return portForwardChartMuseum()
	case "git-server":
		return portForwardGitServer()
	default:
		return fmt.Errorf("unknown service: %s. Available services: argocd, chartmuseum, git-server", service)
	}
}

func portForwardArgoCD() error {
	fmt.Println("🌐 Starting ArgoCD UI port forward...")
	fmt.Println("ArgoCD UI: http://localhost:8083 (admin/admin)")

	cmd := exec.Command("kubectl", "port-forward", "-n", "argocd", "svc/argocd-server", "8083:443")
	return cmd.Run()
}

func portForwardChartMuseum() error {
	fmt.Println("🌐 Starting ChartMuseum port forward...")
	fmt.Println("ChartMuseum: http://localhost:8084")

	cmd := exec.Command("kubectl", "port-forward", "-n", "chartmuseum", "svc/chartmuseum", "8084:8080")
	return cmd.Run()
}

func portForwardGitServer() error {
	fmt.Println("🌐 Starting Git Server port forward...")
	fmt.Println("Git Server: http://localhost:8085")

	cmd := exec.Command("kubectl", "port-forward", "-n", "git-server", "svc/git-server", "8085:80")
	return cmd.Run()
}

func killProcesses(processes []*exec.Cmd) {
	for _, proc := range processes {
		if proc.Process != nil {
			proc.Process.Kill()
		}
	}
}

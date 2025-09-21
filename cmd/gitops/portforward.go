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
	portForwardService   string
	portForwardTargetDir string
)

func init() {
	portForwardCmd.Flags().StringVarP(&portForwardService, "service", "s", "", "Port forward specific service (argocd, chartmuseum, git-server)")
	portForwardCmd.Flags().StringVar(&portForwardTargetDir, "target-dir", ".", "Target directory containing .gitops-config.yaml")
}

func runPortForward(cmd *cobra.Command, args []string) error {
	// Read configuration
	config, err := readConfig(portForwardTargetDir)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Set kubeconfig
	if err := setKubeconfig(config.ClusterName); err != nil {
		return fmt.Errorf("failed to set kubeconfig: %w", err)
	}

	if portForwardService != "" {
		return portForwardSpecificService(portForwardService, config)
	}

	return portForwardAllServices(config)
}

func portForwardAllServices(config *Config) error {
	// Start all services by calling individual service functions
	services := []string{"argocd", "chartmuseum", "git-server"}

	for _, service := range services {
		go func(svc string) {
			switch svc {
			case "argocd":
				portForwardArgoCD(config)
			case "chartmuseum":
				portForwardChartMuseum(config)
			case "git-server":
				portForwardGitServer(config)
			}
		}(service)
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Stopping port forwarding...")
	fmt.Println("✅ Port forwarding stopped")

	return nil
}

// startPortForward is a helper function to start a port forward command
func startPortForward(namespace, service, localPort, remotePort string) (*exec.Cmd, error) {
	cmd := exec.Command("kubectl", "port-forward", "-n", namespace, service, fmt.Sprintf("%s:%s", localPort, remotePort))
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

// printPortForwardInfo prints the port forward information for a service
func printPortForwardInfo(serviceName, localPort, credentials string) {
	if credentials != "" {
		fmt.Printf("🌐 Starting %s port forward...\n", serviceName)
		fmt.Printf("%s: http://localhost:%s (%s)\n", serviceName, localPort, credentials)
	} else {
		fmt.Printf("🌐 Starting %s port forward...\n", serviceName)
		fmt.Printf("%s: http://localhost:%s\n", serviceName, localPort)
	}
}

func portForwardSpecificService(service string, config *Config) error {
	switch strings.ToLower(service) {
	case "argocd":
		return portForwardArgoCD(config)
	case "chartmuseum":
		return portForwardChartMuseum(config)
	case "git-server":
		return portForwardGitServer(config)
	default:
		return fmt.Errorf("unknown service: %s. Available services: argocd, chartmuseum, git-server", service)
	}
}

func portForwardArgoCD(config *Config) error {
	printPortForwardInfo("ArgoCD UI", config.ArgoCDPort, "credentials: admin/admin")

	cmd, err := startPortForward("argocd", "svc/argocd-server", config.ArgoCDPort, "443")
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func portForwardChartMuseum(config *Config) error {
	printPortForwardInfo("ChartMuseum", config.ChartMuseumPort, "")

	cmd, err := startPortForward("chartmuseum", "svc/chartmuseum", config.ChartMuseumPort, "8080")
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func portForwardGitServer(config *Config) error {
	printPortForwardInfo("Git Server", config.GitServerPort, "")

	cmd, err := startPortForward("git-server", "svc/git-server", config.GitServerPort, "80")
	if err != nil {
		return err
	}
	return cmd.Wait()
}

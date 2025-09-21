package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// setKubeconfig sets the KUBECONFIG environment variable for the specified cluster
func setKubeconfig(clusterName string) error {
	cmd := exec.Command("k3d", "kubeconfig", "write", clusterName)
	output, err := runCommand(cmd, "k3d kubeconfig write")
	if err != nil {
		return err
	}

	os.Setenv("KUBECONFIG", strings.TrimSpace(string(output)))
	return nil
}

// Config holds the GitOps configuration
type Config struct {
	ClusterName     string
	RegistryName    string
	RegistryPort    string
	ArgoCDPort      string
	ChartMuseumPort string
	GitServerPort   string
}

// readConfig reads the GitOps configuration from the specified directory
func readConfig(configDir string) (*Config, error) {
	configPath := filepath.Join(configDir, ".gitops-config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		// If no config file, use default values
		return &Config{
			ClusterName:     "devcluster",
			RegistryName:    "myregistry.localhost",
			RegistryPort:    "5001",
			ArgoCDPort:      "8083",
			ChartMuseumPort: "8084",
			GitServerPort:   "8085",
		}, nil
	}

	// Read config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse config file
	config := &Config{
		ClusterName:     "devcluster",
		RegistryName:    "myregistry.localhost",
		RegistryPort:    "5001",
		ArgoCDPort:      "8083",
		ChartMuseumPort: "8084",
		GitServerPort:   "8085",
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "cluster_name":
				config.ClusterName = value
			case "registry_name":
				config.RegistryName = value
			case "registry_port":
				config.RegistryPort = value
			case "argocd_port":
				config.ArgoCDPort = value
			case "chartmuseum_port":
				config.ChartMuseumPort = value
			case "git_server_port":
				config.GitServerPort = value
			}
		}
	}

	return config, nil
}

// runCommand executes a command and returns its output with enhanced error handling
func runCommand(cmd *exec.Cmd, description string) ([]byte, error) {
	if verbose {
		fmt.Printf("🔧 Running: %s %s\n", cmd.Path, strings.Join(cmd.Args[1:], " "))
	}

	output, err := cmd.Output()
	if err != nil {
		// If the command failed, try to get stderr as well
		var stderr []byte
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr = exitError.Stderr
		}

		// Build detailed error message
		errorMsg := fmt.Sprintf("command failed: %s", description)
		if len(output) > 0 {
			errorMsg += fmt.Sprintf("\nstdout: %s", string(output))
		}
		if len(stderr) > 0 {
			errorMsg += fmt.Sprintf("\nstderr: %s", string(stderr))
		}
		errorMsg += fmt.Sprintf("\nexit code: %v", err)

		return output, errors.New(errorMsg)
	}

	if verbose && len(output) > 0 {
		fmt.Printf("📤 Output: %s\n", string(output))
	}

	return output, nil
}

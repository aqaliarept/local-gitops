package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// setKubeconfig sets the KUBECONFIG environment variable for the specified cluster
func setKubeconfig() error {
	cmd := exec.Command("k3d", "kubeconfig", "write", clusterName)
	output, err := runCommand(cmd, "k3d kubeconfig write")
	if err != nil {
		return err
	}

	os.Setenv("KUBECONFIG", strings.TrimSpace(string(output)))
	return nil
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

		return output, fmt.Errorf(errorMsg)
	}

	if verbose && len(output) > 0 {
		fmt.Printf("📤 Output: %s\n", string(output))
	}

	return output, nil
}

// runCommandWithInput executes a command with input and returns its output with enhanced error handling
func runCommandWithInput(cmd *exec.Cmd, input string, description string) ([]byte, error) {
	if verbose {
		fmt.Printf("🔧 Running: %s %s\n", cmd.Path, strings.Join(cmd.Args[1:], " "))
	}

	cmd.Stdin = strings.NewReader(input)
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

		return output, fmt.Errorf(errorMsg)
	}

	if verbose && len(output) > 0 {
		fmt.Printf("📤 Output: %s\n", string(output))
	}

	return output, nil
}

// runCommandInteractive executes a command that might need user interaction
func runCommandInteractive(cmd *exec.Cmd, description string) error {
	if verbose {
		fmt.Printf("🔧 Running: %s %s\n", cmd.Path, strings.Join(cmd.Args[1:], " "))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		errorMsg := fmt.Sprintf("command failed: %s\nexit code: %v", description, err)
		return fmt.Errorf(errorMsg)
	}

	return nil
}

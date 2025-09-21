package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup local GitOps environment",
	Long:  "Creates k3d cluster, installs ArgoCD, ChartMuseum, and Git server",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Setting up Local GitOps Environment...")

	// Read configuration
	config, err := readConfig(".")
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if verbose {
		fmt.Printf("📋 Using cluster name: %s\n", config.ClusterName)
	}

	// Check prerequisites
	if err := checkPrerequisites(); err != nil {
		return fmt.Errorf("prerequisites check failed: %w", err)
	}

	// Create local registry
	if err := createRegistry(); err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	// Create k3d cluster
	if err := createCluster(config.ClusterName); err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	// Install ArgoCD
	if err := installArgoCD(); err != nil {
		return fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	// Install ChartMuseum
	if err := installChartMuseum(); err != nil {
		return fmt.Errorf("failed to install ChartMuseum: %w", err)
	}

	// Install Git server
	if err := setupGitServer(); err != nil {
		return fmt.Errorf("failed to install Git server: %w", err)
	}

	// Setup Git repository
	if err := setupGitRepository(); err != nil {
		return fmt.Errorf("failed to setup Git repository: %w", err)
	}

	// Print status
	printStatus(config)

	fmt.Println("✅ Local GitOps Environment setup completed!")
	return nil
}

func checkPrerequisites() error {
	fmt.Println("🔍 Checking prerequisites...")

	// Check if k3d is installed
	if _, err := runCommand(exec.Command("k3d", "version"), "k3d version"); err != nil {
		return fmt.Errorf("k3d is not installed or not in PATH")
	}

	// Check if kubectl is installed
	if _, err := runCommand(exec.Command("kubectl", "version", "--client"), "kubectl version"); err != nil {
		return fmt.Errorf("kubectl is not installed or not in PATH")
	}

	// Check if docker is running
	if _, err := runCommand(exec.Command("docker", "version"), "docker version"); err != nil {
		return errors.New("docker is not running or not in PATH")
	}

	fmt.Println("✅ Prerequisites check passed")
	return nil
}

func createRegistry() error {
	fmt.Println("🐳 Creating local Docker registry...")

	// Check if registry already exists
	cmd := exec.Command("k3d", "registry", "list")
	output, err := runCommand(cmd, "k3d registry list")
	if err != nil {
		return err
	}

	if strings.Contains(string(output), registryName) {
		fmt.Printf("ℹ️  Registry %s already exists\n", registryName)
		return nil
	}

	// Create registry
	cmd = exec.Command("k3d", "registry", "create", registryName, "--port", registryPort)
	if _, err := runCommand(cmd, "k3d registry create"); err != nil {
		return err
	}

	fmt.Printf("✅ Registry created at %s:%s\n", registryName, registryPort)
	return nil
}

func createCluster(clusterName string) error {
	fmt.Println("🏗️  Creating k3d cluster...")

	// Check if cluster already exists
	cmd := exec.Command("k3d", "cluster", "list")
	output, err := runCommand(cmd, "k3d cluster list")
	if err != nil {
		return err
	}

	if strings.Contains(string(output), clusterName) {
		fmt.Printf("ℹ️  Cluster %s already exists\n", clusterName)
		return nil
	}

	// Create cluster with registry
	cmd = exec.Command("k3d", "cluster", "create", clusterName,
		"--registry-use", fmt.Sprintf("k3d-%s:%s", registryName, registryPort),
		"--port", "8080:80@loadbalancer",
		"--port", "8443:443@loadbalancer")
	if _, err := runCommand(cmd, "k3d cluster create"); err != nil {
		return err
	}

	// Set kubeconfig
	if err := setKubeconfig(clusterName); err != nil {
		return err
	}

	fmt.Printf("✅ Cluster %s created successfully\n", clusterName)
	return nil
}

func installArgoCD() error {
	fmt.Println("🚀 Installing ArgoCD...")

	// Create argocd namespace
	cmd := exec.Command("kubectl", "create", "namespace", "argocd")
	runCommand(cmd, "kubectl create namespace argocd") // Ignore error if namespace exists

	// Install ArgoCD
	cmd = exec.Command("kubectl", "apply", "-n", "argocd", "-f", "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml")
	if _, err := runCommand(cmd, "kubectl apply ArgoCD"); err != nil {
		return err
	}

	// Wait for ArgoCD to be ready
	fmt.Println("⏳ Waiting for ArgoCD to be ready...")
	cmd = exec.Command("kubectl", "wait", "--for=condition=available", "--timeout=300s", "deployment/argocd-server", "-n", "argocd")
	if _, err := runCommand(cmd, "kubectl wait ArgoCD"); err != nil {
		return err
	}

	// Configure ArgoCD password
	if err := configureArgoCDPassword(); err != nil {
		return fmt.Errorf("failed to configure ArgoCD password: %w", err)
	}

	fmt.Println("✅ ArgoCD installed successfully")
	return nil
}

func configureArgoCDPassword() error {
	fmt.Println("🔐 Configuring ArgoCD password...")

	// Hardcoded bcrypt hash for password "admin"
	// Generated with: htpasswd -nbB admin admin
	adminPasswordHash := "$2y$05$JCbgj53Ew6p9TAaNUAk9cu15EB4yrjPp6yI3ucUl5MdmuiYM54.O2"

	// Update password to 'admin' using the hardcoded hash
	patchData := fmt.Sprintf(`{"stringData":{"admin.password":"%s"}}`, adminPasswordHash)
	cmd := exec.Command("kubectl", "-n", "argocd", "patch", "secret", "argocd-secret", "-p", patchData)
	if _, err := runCommand(cmd, "kubectl patch ArgoCD secret"); err != nil {
		return err
	}

	// Verify password was set correctly
	fmt.Println("🔍 Verifying ArgoCD password configuration...")
	cmd = exec.Command("kubectl", "-n", "argocd", "get", "secret", "argocd-secret", "-o", "jsonpath={.data.admin\\.password}")
	output, err := runCommand(cmd, "kubectl get ArgoCD secret")
	if err != nil {
		return fmt.Errorf("failed to verify password: %w", err)
	}

	// Decode base64 and compare with expected hash
	expectedHash := adminPasswordHash
	actualHashBase64 := strings.TrimSpace(string(output))

	// Decode the base64 value from the secret
	actualHashBytes, err := base64.StdEncoding.DecodeString(actualHashBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64 hash: %w", err)
	}
	actualHash := string(actualHashBytes)

	if actualHash != expectedHash {
		return fmt.Errorf("password verification failed: expected %s, got %s", expectedHash, actualHash)
	}

	fmt.Println("✅ ArgoCD password configured and verified (admin/admin)")
	return nil
}

func installChartMuseum() error {
	fmt.Println("📦 Installing ChartMuseum...")

	// Create chartmuseum namespace
	cmd := exec.Command("kubectl", "create", "namespace", "chartmuseum")
	runCommand(cmd, "kubectl create namespace chartmuseum") // Ignore error if namespace exists

	// ChartMuseum deployment
	chartmuseumYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: chartmuseum
  namespace: chartmuseum
spec:
  replicas: 1
  selector:
    matchLabels:
      app: chartmuseum
  template:
    metadata:
      labels:
        app: chartmuseum
    spec:
      containers:
        - name: chartmuseum
          image: chartmuseum/chartmuseum:latest
          ports:
            - containerPort: 8080
          env:
            - name: PORT
              value: "8080"
            - name: STORAGE
              value: "local"
            - name: STORAGE_LOCAL_ROOTDIR
              value: "/charts"
          volumeMounts:
            - name: chart-storage
              mountPath: /charts
      volumes:
        - name: chart-storage
          emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: chartmuseum
  namespace: chartmuseum
spec:
  ports:
    - port: 8080
      targetPort: 8080
  selector:
    app: chartmuseum`

	cmd = exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(chartmuseumYAML)
	if _, err := runCommand(cmd, "kubectl apply ChartMuseum"); err != nil {
		return err
	}

	// Wait for ChartMuseum to be ready
	fmt.Println("⏳ Waiting for ChartMuseum to be ready...")
	cmd = exec.Command("kubectl", "wait", "--for=condition=available", "--timeout=300s", "deployment/chartmuseum", "-n", "chartmuseum")
	if _, err := runCommand(cmd, "kubectl wait ChartMuseum"); err != nil {
		return err
	}

	fmt.Println("✅ ChartMuseum installed successfully")
	return nil
}

func setupGitServer() error {
	fmt.Println("📁 Installing Git server...")

	// Git server deployment
	gitServerYAML := `apiVersion: v1
kind: Namespace
metadata:
  name: git-server
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: git-server-pv
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: /data/git-server
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - k3d-local-gitops-server-0
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: git-server-pvc
  namespace: git-server
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: git-server
  namespace: git-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: git-server
  template:
    metadata:
      labels:
        app: git-server
    spec:
      containers:
        - name: git-server
          image: moikot/basic-git-server
          ports:
            - containerPort: 8080
          volumeMounts:
            - mountPath: /repos
              name: repo-storage
      volumes:
        - name: repo-storage
          persistentVolumeClaim:
            claimName: git-server-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: git-server
  namespace: git-server
spec:
  ports:
    - port: 80
      targetPort: 8080
      name: http
  selector:
    app: git-server`

	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(gitServerYAML)
	if _, err := runCommand(cmd, "kubectl apply Git server"); err != nil {
		return err
	}

	// Wait for Git server to be ready
	fmt.Println("⏳ Waiting for Git server to be ready...")
	cmd = exec.Command("kubectl", "wait", "--for=condition=available", "--timeout=300s", "deployment/git-server", "-n", "git-server")
	if _, err := runCommand(cmd, "kubectl wait Git server"); err != nil {
		return err
	}

	fmt.Println("✅ Git server installed successfully")
	return nil
}

func setupGitRepository() error {
	fmt.Println("📁 Git repository setup...")

	// The moikot/basic-git-server automatically provides git server functionality
	// No manual repository initialization needed

	fmt.Println("✅ Git repository setup completed")
	return nil
}

func printStatus(config *Config) error {
	fmt.Println("")
	fmt.Println("📊 Setup Status:")
	fmt.Println("==================")
	fmt.Printf("  Cluster: %s\n", config.ClusterName)
	fmt.Printf("  Local Registry: %s:%s\n", config.RegistryName, config.RegistryPort)
	fmt.Printf("  ArgoCD: http://localhost:%s (admin/admin)\n", config.ArgoCDPort)
	fmt.Printf("  ChartMuseum: http://localhost:%s\n", config.ChartMuseumPort)
	fmt.Printf("  Git Server: http://localhost:%s\n", config.GitServerPort)
	fmt.Println("")
	fmt.Println("🌐 To access services, run:")
	fmt.Println("  gitops port-forward")
	return nil
}

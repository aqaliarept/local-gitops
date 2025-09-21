package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new GitOps directory with nginx application",
	Long:  "Creates a new directory with ArgoCD Application and nginx application manifests",
	RunE:  runInit,
}

var (
	initClusterName     string
	initArgoCDPort      string
	initChartMuseumPort string
	initGitServerPort   string
)

func init() {
	initCmd.Flags().StringVar(&initClusterName, "cluster", "devcluster", "k3d cluster name for this project")
	initCmd.Flags().StringVar(&initArgoCDPort, "argocd-port", "8083", "ArgoCD server port")
	initCmd.Flags().StringVar(&initChartMuseumPort, "chartmuseum-port", "8084", "ChartMuseum server port")
	initCmd.Flags().StringVar(&initGitServerPort, "git-server-port", "8085", "Git server port")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	if targetDir == "" {
		return fmt.Errorf("init directory is required")
	}

	// Check if directory already exists
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("directory %s already exists", targetDir)
	}

	// Create directory structure
	if err := createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Create bootstrap.yaml
	if err := createBootstrapYAML(); err != nil {
		return fmt.Errorf("failed to create bootstrap.yaml: %w", err)
	}

	// Create example application manifests
	if err := createExampleManifests(); err != nil {
		return fmt.Errorf("failed to create example manifests: %w", err)
	}

	// Create README
	if err := createREADME(); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	// Create config file
	if err := createConfig(); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	fmt.Printf("✅ GitOps directory initialized successfully: %s\n", targetDir)
	fmt.Println("")
	fmt.Println("📋 Next steps:")
	fmt.Printf("  1. cd %s\n", targetDir)
	fmt.Println("  2. gitops setup")
	fmt.Println("  3. gitops deploy")
	fmt.Println("")
	fmt.Println("📁 Directory structure created:")
	fmt.Println("  ├── manifest/")
	fmt.Println("  │   ├── deployment.yaml")
	fmt.Println("  │   ├── service.yaml")
	fmt.Println("  │   └── ingress.yaml")
	fmt.Println("  ├── bootstrap.yaml")
	fmt.Println("  ├── .gitops-config.yaml")
	fmt.Println("  └── README.md")

	return nil
}

func createDirectoryStructure() error {
	dirs := []string{
		filepath.Join(targetDir, "manifest"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func createBootstrapYAML() error {
	bootstrapYAML := `apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: nginx-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: http://git-server.git-server.svc.cluster.local/manifest.git
    targetRevision: HEAD
    path: .
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
`

	bootstrapPath := filepath.Join(targetDir, "bootstrap.yaml")
	return os.WriteFile(bootstrapPath, []byte(bootstrapYAML), 0644)
}

func createExampleManifests() error {
	// Create nginx deployment
	deploymentYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-app
  labels:
    app: nginx-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx-app
  template:
    metadata:
      labels:
        app: nginx-app
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
`

	deploymentPath := filepath.Join(targetDir, "manifest", "deployment.yaml")
	if err := os.WriteFile(deploymentPath, []byte(deploymentYAML), 0644); err != nil {
		return err
	}

	// Create nginx service
	serviceYAML := `apiVersion: v1
kind: Service
metadata:
  name: nginx-app
  labels:
    app: nginx-app
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
  selector:
    app: nginx-app
`

	servicePath := filepath.Join(targetDir, "manifest", "service.yaml")
	if err := os.WriteFile(servicePath, []byte(serviceYAML), 0644); err != nil {
		return err
	}

	// Create nginx ingress
	ingressYAML := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-app
  labels:
    app: nginx-app
spec:
  ingressClassName: nginx
  rules:
  - host: nginx.localhost
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nginx-app
            port:
              number: 80
`

	ingressPath := filepath.Join(targetDir, "manifest", "ingress.yaml")
	if err := os.WriteFile(ingressPath, []byte(ingressYAML), 0644); err != nil {
		return err
	}

	return nil
}

func createREADME() error {
	readmeContent := `# GitOps Directory

This directory contains a simple nginx application deployed using GitOps principles.

## Structure

- manifest/ - Contains nginx application Kubernetes manifests
  - deployment.yaml - nginx deployment
  - service.yaml - nginx service
  - ingress.yaml - nginx ingress
- bootstrap.yaml - ArgoCD Application manifest

## Usage

1. Setup the cluster:
   gitops setup

2. Deploy the application:
   gitops deploy

The nginx application will be available at http://nginx.localhost
`

	readmePath := filepath.Join(targetDir, "README.md")
	return os.WriteFile(readmePath, []byte(readmeContent), 0644)
}

func createConfig() error {
	configContent := fmt.Sprintf(`# GitOps Configuration
cluster_name: %s
registry_name: myregistry.localhost
registry_port: 5001
argocd_port: %s
chartmuseum_port: %s
git_server_port: %s
`, initClusterName, initArgoCDPort, initChartMuseumPort, initGitServerPort)

	configPath := filepath.Join(targetDir, ".gitops-config.yaml")
	return os.WriteFile(configPath, []byte(configContent), 0644)
}

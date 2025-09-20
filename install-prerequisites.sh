#!/bin/bash

# Install Prerequisites for Local GitOps Environment
# This script installs all required tools for the local GitOps setup

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Installing Prerequisites for Local GitOps Environment${NC}"

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     MACHINE=Linux;;
    Darwin*)    MACHINE=Mac;;
    CYGWIN*)    MACHINE=Cygwin;;
    MINGW*)     MACHINE=MinGw;;
    *)          MACHINE="UNKNOWN:${OS}"
esac

echo -e "${YELLOW}📋 Detected OS: $MACHINE${NC}"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install on macOS
install_macos() {
    echo -e "${YELLOW}🍎 Installing on macOS...${NC}"
    
    # Check if Homebrew is installed
    if ! command_exists brew; then
        echo -e "${YELLOW}📦 Installing Homebrew...${NC}"
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    else
        echo -e "${GREEN}✅ Homebrew already installed${NC}"
    fi
    
    # Install k3d
    if ! command_exists k3d; then
        echo -e "${YELLOW}📦 Installing k3d...${NC}"
        curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
    else
        echo -e "${GREEN}✅ k3d already installed${NC}"
    fi
    
    # Install kubectl
    if ! command_exists kubectl; then
        echo -e "${YELLOW}📦 Installing kubectl...${NC}"
        brew install kubectl
    else
        echo -e "${GREEN}✅ kubectl already installed${NC}"
    fi
    
    # Install Helm
    if ! command_exists helm; then
        echo -e "${YELLOW}📦 Installing Helm...${NC}"
        brew install helm
    else
        echo -e "${GREEN}✅ Helm already installed${NC}"
    fi
    
    # Install Docker (if not installed)
    if ! command_exists docker; then
        echo -e "${YELLOW}📦 Installing Docker Desktop...${NC}"
        echo -e "${RED}⚠️  Docker Desktop needs to be installed manually from: https://www.docker.com/products/docker-desktop${NC}"
        echo -e "${YELLOW}Please install Docker Desktop and restart this script.${NC}"
        exit 1
    else
        echo -e "${GREEN}✅ Docker already installed${NC}"
    fi
}

# Function to install on Linux
install_linux() {
    echo -e "${YELLOW}🐧 Installing on Linux...${NC}"
    
    # Detect Linux distribution
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
    else
        echo -e "${RED}❌ Cannot detect Linux distribution${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}📋 Detected distribution: $DISTRO${NC}"
    
    case $DISTRO in
        ubuntu|debian)
            # Update package list
            echo -e "${YELLOW}📦 Updating package list...${NC}"
            sudo apt-get update
            
            # Install curl if not present
            if ! command_exists curl; then
                echo -e "${YELLOW}📦 Installing curl...${NC}"
                sudo apt-get install -y curl
            fi
            
            # Install k3d
            if ! command_exists k3d; then
                echo -e "${YELLOW}📦 Installing k3d...${NC}"
                curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
            else
                echo -e "${GREEN}✅ k3d already installed${NC}"
            fi
            
            # Install kubectl
            if ! command_exists kubectl; then
                echo -e "${YELLOW}📦 Installing kubectl...${NC}"
                curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
                chmod +x kubectl
                sudo mv kubectl /usr/local/bin/
            else
                echo -e "${GREEN}✅ kubectl already installed${NC}"
            fi
            
            # Install Helm
            if ! command_exists helm; then
                echo -e "${YELLOW}📦 Installing Helm...${NC}"
                curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
            else
                echo -e "${GREEN}✅ Helm already installed${NC}"
            fi
            
            # Install Docker
            if ! command_exists docker; then
                echo -e "${YELLOW}📦 Installing Docker...${NC}"
                curl -fsSL https://get.docker.com -o get-docker.sh
                sudo sh get-docker.sh
                sudo usermod -aG docker $USER
                echo -e "${YELLOW}⚠️  Please log out and log back in for Docker group changes to take effect${NC}"
            else
                echo -e "${GREEN}✅ Docker already installed${NC}"
            fi
            ;;
        centos|rhel|fedora)
            # Install curl if not present
            if ! command_exists curl; then
                echo -e "${YELLOW}📦 Installing curl...${NC}"
                sudo yum install -y curl
            fi
            
            # Install k3d
            if ! command_exists k3d; then
                echo -e "${YELLOW}📦 Installing k3d...${NC}"
                curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
            else
                echo -e "${GREEN}✅ k3d already installed${NC}"
            fi
            
            # Install kubectl
            if ! command_exists kubectl; then
                echo -e "${YELLOW}📦 Installing kubectl...${NC}"
                curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
                chmod +x kubectl
                sudo mv kubectl /usr/local/bin/
            else
                echo -e "${GREEN}✅ kubectl already installed${NC}"
            fi
            
            # Install Helm
            if ! command_exists helm; then
                echo -e "${YELLOW}📦 Installing Helm...${NC}"
                curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
            else
                echo -e "${GREEN}✅ Helm already installed${NC}"
            fi
            
            # Install Docker
            if ! command_exists docker; then
                echo -e "${YELLOW}📦 Installing Docker...${NC}"
                curl -fsSL https://get.docker.com -o get-docker.sh
                sudo sh get-docker.sh
                sudo usermod -aG docker $USER
                echo -e "${YELLOW}⚠️  Please log out and log back in for Docker group changes to take effect${NC}"
            else
                echo -e "${GREEN}✅ Docker already installed${NC}"
            fi
            ;;
        *)
            echo -e "${RED}❌ Unsupported Linux distribution: $DISTRO${NC}"
            echo -e "${YELLOW}Please install the following tools manually:${NC}"
            echo -e "  - k3d: https://k3d.io/"
            echo -e "  - kubectl: https://kubernetes.io/docs/tasks/tools/"
            echo -e "  - Helm: https://helm.sh/"
            echo -e "  - Docker: https://www.docker.com/"
            exit 1
            ;;
    esac
}

# Main installation logic
case $MACHINE in
    Mac)
        install_macos
        ;;
    Linux)
        install_linux
        ;;
    *)
        echo -e "${RED}❌ Unsupported operating system: $MACHINE${NC}"
        echo -e "${YELLOW}Please install the following tools manually:${NC}"
        echo -e "  - k3d: https://k3d.io/"
        echo -e "  - kubectl: https://kubernetes.io/docs/tasks/tools/"
        echo -e "  - Helm: https://helm.sh/"
        echo -e "  - Docker: https://www.docker.com/"
        exit 1
        ;;
esac

# Verify installations
echo -e "${YELLOW}🔍 Verifying installations...${NC}"

# Check k3d
if command_exists k3d; then
    K3D_VERSION=$(k3d version | head -n1 | cut -d' ' -f3)
    echo -e "${GREEN}✅ k3d installed: $K3D_VERSION${NC}"
else
    echo -e "${RED}❌ k3d installation failed${NC}"
    exit 1
fi

# Check kubectl
if command_exists kubectl; then
    KUBECTL_VERSION=$(kubectl version --client --short 2>/dev/null | cut -d' ' -f3)
    echo -e "${GREEN}✅ kubectl installed: $KUBECTL_VERSION${NC}"
else
    echo -e "${RED}❌ kubectl installation failed${NC}"
    exit 1
fi

# Check Helm
if command_exists helm; then
    HELM_VERSION=$(helm version --short)
    echo -e "${GREEN}✅ Helm installed: $HELM_VERSION${NC}"
else
    echo -e "${RED}❌ Helm installation failed${NC}"
    exit 1
fi

# Check Docker
if command_exists docker; then
    DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | cut -d',' -f1)
    echo -e "${GREEN}✅ Docker installed: $DOCKER_VERSION${NC}"
    
    # Check if Docker daemon is running
    if docker info >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Docker daemon is running${NC}"
    else
        echo -e "${YELLOW}⚠️  Docker daemon is not running. Please start Docker Desktop or Docker service.${NC}"
    fi
else
    echo -e "${RED}❌ Docker installation failed${NC}"
    exit 1
fi

# Check curl
if command_exists curl; then
    CURL_VERSION=$(curl --version | head -n1 | cut -d' ' -f2)
    echo -e "${GREEN}✅ curl installed: $CURL_VERSION${NC}"
else
    echo -e "${RED}❌ curl installation failed${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 All prerequisites installed successfully!${NC}"
echo ""
echo -e "${BLUE}📋 Installed Tools:${NC}"
echo -e "  k3d: $K3D_VERSION"
echo -e "  kubectl: $KUBECTL_VERSION"
echo -e "  Helm: $HELM_VERSION"
echo -e "  Docker: $DOCKER_VERSION"
echo -e "  curl: $CURL_VERSION"
echo ""
echo -e "${YELLOW}🔧 Next Steps:${NC}"
echo -e "  1. Make sure Docker is running"
echo -e "  2. Run: ./setup.sh"
echo -e "  3. Follow the setup instructions"
echo ""
echo -e "${GREEN}✅ Ready to set up your Local GitOps Environment!${NC}"

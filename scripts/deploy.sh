#!/bin/bash

# Logmojo Auto-Deploy Script
# Usage: curl -fsSL https://raw.githubusercontent.com/saiarlen/logmojo/main/scripts/deploy.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
REPO_URL="https://github.com/saiarlen/logmojo"
INSTALL_DIR="/opt/logmojo"
SERVICE_USER="logmojo"
SERVICE_NAME="logmojo"

# Print banner
print_banner() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                                                              ║"
    echo "║                    🚀 LOGMOJO INSTALLER                      ║"
    echo "║                                                              ║"
    echo "║            High-Performance Log Management System            ║"
    echo "║                                                              ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Print step
print_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# Print success
print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Print warning
print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Print error
print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Detect OS and architecture
detect_system() {
    print_step "Detecting system architecture..."
    
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    case $OS in
        linux)
            BINARY_NAME="logmojo-linux-$ARCH"
            ;;
        darwin)
            BINARY_NAME="logmojo-darwin-$ARCH"
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    print_success "Detected: $OS $ARCH"
}

# Check system requirements
check_requirements() {
    print_step "Checking system requirements..."
    
    # Check for required commands
    local missing_deps=()
    
    for cmd in curl wget grep systemctl; do
        if ! command -v $cmd &> /dev/null; then
            missing_deps+=($cmd)
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        print_error "Please install them and run the script again"
        exit 1
    fi
    
    # Check available disk space (minimum 100MB)
    available_space=$(df / | awk 'NR==2 {print $4}')
    if [ $available_space -lt 102400 ]; then
        print_warning "Low disk space detected. At least 100MB recommended."
    fi
    
    print_success "System requirements met"
}

# Get latest release version
get_latest_version() {
    print_step "Fetching latest release version..."
    
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/saiarlen/logmojo/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$LATEST_VERSION" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi
    
    print_success "Latest version: $LATEST_VERSION"
}

# Download and install binary
install_binary() {
    print_step "Downloading Logmojo binary..."
    
    DOWNLOAD_URL="https://github.com/saiarlen/logmojo/releases/download/$LATEST_VERSION/$BINARY_NAME"
    
    # Create installation directory
    mkdir -p $INSTALL_DIR
    
    # Download binary
    if ! curl -L -o "$INSTALL_DIR/logmojo" "$DOWNLOAD_URL"; then
        print_error "Failed to download binary from $DOWNLOAD_URL"
        exit 1
    fi
    
    # Make executable
    chmod +x "$INSTALL_DIR/logmojo"
    
    print_success "Binary installed to $INSTALL_DIR/logmojo"
}

# Download required files
download_files() {
    print_step "Downloading required files..."
    
    cd $INSTALL_DIR
    
    # Download views directory
    print_step "Downloading views..."
    mkdir -p views/layouts
    curl -L -o views/layouts/main.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/layouts/main.jet.html"
    curl -L -o views/dashboard.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/dashboard.jet.html"
    curl -L -o views/logs.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/logs.jet.html"
    curl -L -o views/processes.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/processes.jet.html"
    curl -L -o views/services.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/services.jet.html"
    curl -L -o views/alerts.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/alerts.jet.html"
    curl -L -o views/settings.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/settings.jet.html"
    curl -L -o views/login.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/login.jet.html"
    
    # Download public directory
    print_step "Downloading public assets..."
    mkdir -p public/{css,js,images}
    curl -L -o public/css/style.css "https://raw.githubusercontent.com/saiarlen/logmojo/main/public/css/style.css"
    curl -L -o public/js/main.js "https://raw.githubusercontent.com/saiarlen/logmojo/main/public/js/main.js"
    curl -L -o public/images/logo.png "https://raw.githubusercontent.com/saiarlen/logmojo/main/public/images/logo.png"
    curl -L -o public/images/favicon.png "https://raw.githubusercontent.com/saiarlen/logmojo/main/public/images/favicon.png"
    
    # Download configuration files
    print_step "Downloading configuration files..."
    curl -L -o config.yaml "https://raw.githubusercontent.com/saiarlen/logmojo/main/config.yaml"
    curl -L -o .env.example "https://raw.githubusercontent.com/saiarlen/logmojo/main/.env.example"
    
    # Create .env from example if it doesn't exist
    if [ ! -f .env ]; then
        cp .env.example .env
        # Set the actual version from GitHub release
        sed -i "s/MONITOR_GENERAL_VERSION=\"dev\"/MONITOR_GENERAL_VERSION=\"$LATEST_VERSION\"/g" .env
        print_success "Created .env file with version $LATEST_VERSION"
    else
        # Update existing .env with correct version
        sed -i "s/MONITOR_GENERAL_VERSION=\"[^\"]*\"/MONITOR_GENERAL_VERSION=\"$LATEST_VERSION\"/g" .env
        print_success "Updated .env with version $LATEST_VERSION"
    fi
    
    print_success "Required files downloaded"
}

# Create system user
create_user() {
    print_step "Creating system user..."
    
    if ! id "$SERVICE_USER" &>/dev/null; then
        useradd --system --no-create-home --shell /bin/false $SERVICE_USER
        print_success "Created user: $SERVICE_USER"
    else
        print_success "User $SERVICE_USER already exists"
    fi
    
    # Add user to required groups for comprehensive log access
    # System log groups
    if getent group systemd-journal >/dev/null 2>&1; then
        usermod -a -G systemd-journal $SERVICE_USER 2>/dev/null || print_warning "Failed to add user to systemd-journal group"
    fi
    if getent group adm >/dev/null 2>&1; then
        usermod -a -G adm $SERVICE_USER 2>/dev/null || print_warning "Failed to add user to adm group"
    fi
    
    # Add to all existing user groups for application log access
    for group in $(getent group | grep -E '^[a-zA-Z][a-zA-Z0-9_-]*:[^:]*:[0-9]{4,}:' | cut -d: -f1); do
        if [ "$group" != "$SERVICE_USER" ] && [ "$group" != "root" ] && [ "$group" != "nobody" ]; then
            usermod -a -G "$group" $SERVICE_USER 2>/dev/null || true
        fi
    done
    
    print_success "Added $SERVICE_USER to all available log access groups"
    
}

# Set permissions
set_permissions() {
    print_step "Setting permissions..."
    
    chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
    chmod 755 $INSTALL_DIR
    chmod +x $INSTALL_DIR/logmojo
    
    print_success "Permissions set"
}

# Create systemd service
create_service() {
    print_step "Creating systemd service..."
    
    cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=Logmojo - High-Performance Log Management System
Documentation=https://github.com/saiarlen/logmojo
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/logmojo
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=5
Restart=always
RestartSec=5
StartLimitInterval=0

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
#ProtectHome=yes
ReadWritePaths=$INSTALL_DIR
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE

# Environment
Environment=MONITOR_SERVER_LISTEN_ADDR=0.0.0.0:7005
Environment=MONITOR_DATABASE_PATH=$INSTALL_DIR/monitor.db

[Install]
WantedBy=multi-user.target
EOF

    # Create sudoers file for logmojo user
    cat > /etc/sudoers.d/logmojo << EOF
# Allow logmojo user to manage systemd services and kill processes
$SERVICE_USER ALL=(ALL) NOPASSWD: /bin/systemctl start *, /bin/systemctl stop *, /bin/systemctl restart *, /bin/systemctl enable *, /bin/systemctl disable *
$SERVICE_USER ALL=(ALL) NOPASSWD: /bin/kill -9 *
EOF
    
    chmod 440 /etc/sudoers.d/logmojo

    systemctl daemon-reload
    systemctl enable $SERVICE_NAME
    
    print_success "Systemd service and sudo permissions created"
}

# Configure firewall (if ufw is available)
configure_firewall() {
    if command -v ufw &> /dev/null; then
        print_step "Configuring firewall..."
        ufw allow 7005/tcp comment "Logmojo Web Interface"
        print_success "Firewall configured (port 7005 opened)"
    fi
}

# Start service
start_service() {
    print_step "Starting Logmojo service..."
    
    systemctl start $SERVICE_NAME
    
    # Wait a moment for service to start
    sleep 3
    
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "Logmojo service started successfully"
    else
        print_error "Failed to start Logmojo service"
        print_error "Check logs with: journalctl -u $SERVICE_NAME -f"
        exit 1
    fi
}

# Print completion message
print_completion() {
    echo
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                              ║${NC}"
    echo -e "${GREEN}║                  🎉 INSTALLATION COMPLETE!                   ║${NC}"
    echo -e "${GREEN}║                                                              ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
    echo -e "${CYAN}📍 Access Logmojo:${NC}"
    echo -e "   🌐 Web Interface: ${BLUE}http://$(hostname -I | awk '{print $1}'):7005${NC}"
    echo -e "   🌐 Local Access:  ${BLUE}http://localhost:7005${NC}"
    echo
    echo -e "${CYAN}🔐 Default Credentials:${NC}"
    echo -e "   👤 Username: ${YELLOW}admin${NC}"
    echo -e "   🔑 Password: ${YELLOW}admin${NC}"
    echo -e "   ${RED}⚠️  Change default password immediately!${NC}"
    echo
    echo -e "${CYAN}🛠️  Service Management:${NC}"
    echo -e "   ▶️  Start:   ${GREEN}sudo systemctl start $SERVICE_NAME${NC}"
    echo -e "   ⏹️  Stop:    ${GREEN}sudo systemctl stop $SERVICE_NAME${NC}"
    echo -e "   🔄 Restart: ${GREEN}sudo systemctl restart $SERVICE_NAME${NC}"
    echo -e "   📊 Status:  ${GREEN}sudo systemctl status $SERVICE_NAME${NC}"
    echo -e "   📋 Logs:    ${GREEN}sudo journalctl -u $SERVICE_NAME -f${NC}"
    echo
    echo -e "${CYAN}📁 Installation Directory:${NC} ${BLUE}$INSTALL_DIR${NC}"
    echo -e "${CYAN}⚙️  Configuration File:${NC} ${BLUE}$INSTALL_DIR/config.yaml${NC}"
    echo -e "${CYAN}🔧 Environment File:${NC} ${BLUE}$INSTALL_DIR/.env${NC}"
    echo
    echo -e "${YELLOW}📖 Documentation: https://github.com/saiarlen/logmojo${NC}"
    echo
}

# Main installation function
main() {
    print_banner
    
    check_root
    detect_system
    check_requirements
    get_latest_version
    install_binary
    download_files
    create_user
    set_permissions
    create_service
    configure_firewall
    start_service
    print_completion
}

# Run main function
main "$@"
#!/bin/bash

# Logmojo Update Script
# Usage: curl -fsSL https://raw.githubusercontent.com/saiarlen/logmojo/main/scripts/update.sh | sudo bash

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
SERVICE_NAME="logmojo"

# Print banner
print_banner() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                                                              ║"
    echo "║                    🔄 LOGMOJO UPDATER                        ║"
    echo "║                                                              ║"
    echo "║              Zero-Downtime Binary & Assets Update           ║"
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

# Check if Logmojo is installed
check_installation() {
    print_step "Checking existing installation..."
    
    if [ ! -f "$INSTALL_DIR/logmojo" ]; then
        print_error "Logmojo not found at $INSTALL_DIR"
        print_error "Please run the installation script first:"
        print_error "curl -fsSL https://raw.githubusercontent.com/saiarlen/logmojo/main/scripts/deploy.sh | sudo bash"
        exit 1
    fi
    
    if [ ! -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
        print_error "Logmojo service not found"
        print_error "Please run the installation script first"
        exit 1
    fi
    
    print_success "Existing installation found"
}

# Get current version
get_current_version() {
    print_step "Getting current version..."
    
    # Try to get version from binary
    if [ -f "$INSTALL_DIR/logmojo" ]; then
        # Try --version flag first, then -version, then fallback
        CURRENT_VERSION=$($INSTALL_DIR/logmojo --version 2>/dev/null | head -n1 | awk '{print $2}' || \
                         $INSTALL_DIR/logmojo -version 2>/dev/null | head -n1 | awk '{print $2}' || \
                         echo "unknown")
        # If version doesn't start with 'v', add it
        if [[ ! $CURRENT_VERSION =~ ^v ]] && [[ $CURRENT_VERSION != "unknown" ]]; then
            CURRENT_VERSION="v$CURRENT_VERSION"
        fi
    else
        CURRENT_VERSION="unknown"
    fi
    
    print_success "Current version: $CURRENT_VERSION"
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

# Check if update is needed
check_update_needed() {
    print_step "Checking if update is needed..."
    
    if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
        print_success "Already running the latest version ($CURRENT_VERSION)"
        echo -e "${YELLOW}No update needed. Exiting.${NC}"
        exit 0
    fi
    
    print_success "Update available: $CURRENT_VERSION → $LATEST_VERSION"
}

# Create backup
create_backup() {
    print_step "Creating backup..."
    
    BACKUP_DIR="/tmp/logmojo-backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # Backup current binary
    cp "$INSTALL_DIR/logmojo" "$BACKUP_DIR/logmojo.backup"
    
    # Backup database and config (if they exist)
    [ -f "$INSTALL_DIR/monitor.db" ] && cp "$INSTALL_DIR/monitor.db" "$BACKUP_DIR/"
    [ -f "$INSTALL_DIR/.env" ] && cp "$INSTALL_DIR/.env" "$BACKUP_DIR/"
    [ -f "$INSTALL_DIR/config.yaml" ] && cp "$INSTALL_DIR/config.yaml" "$BACKUP_DIR/"
    
    print_success "Backup created at: $BACKUP_DIR"
}

# Stop service
stop_service() {
    print_step "Stopping Logmojo service..."
    
    if systemctl is-active --quiet $SERVICE_NAME; then
        systemctl stop $SERVICE_NAME
        print_success "Service stopped"
    else
        print_warning "Service was not running"
    fi
}

# Download new binary
download_binary() {
    print_step "Downloading new binary..."
    
    DOWNLOAD_URL="https://github.com/saiarlen/logmojo/releases/download/$LATEST_VERSION/$BINARY_NAME"
    
    # Download to temporary location
    if ! curl -L -o "$INSTALL_DIR/logmojo.new" "$DOWNLOAD_URL"; then
        print_error "Failed to download binary from $DOWNLOAD_URL"
        exit 1
    fi
    
    # Make executable
    chmod +x "$INSTALL_DIR/logmojo.new"
    
    print_success "New binary downloaded"
}

# Update assets
update_assets() {
    print_step "Updating views and assets..."
    
    cd $INSTALL_DIR
    
    # Backup existing views and public directories
    [ -d "views" ] && mv views views.backup
    [ -d "public" ] && mv public public.backup
    
    # Download new views
    mkdir -p views/layouts
    curl -s -L -o views/layouts/main.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/layouts/main.jet.html"
    curl -s -L -o views/dashboard.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/dashboard.jet.html"
    curl -s -L -o views/logs.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/logs.jet.html"
    curl -s -L -o views/processes.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/processes.jet.html"
    curl -s -L -o views/services.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/services.jet.html"
    curl -s -L -o views/alerts.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/alerts.jet.html"
    curl -s -L -o views/settings.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/settings.jet.html"
    curl -s -L -o views/login.jet.html "https://raw.githubusercontent.com/saiarlen/logmojo/main/views/login.jet.html"
    
    # Download new public assets
    mkdir -p public/{css,js,images}
    curl -s -L -o public/css/style.css "https://raw.githubusercontent.com/saiarlen/logmojo/main/public/css/style.css"
    curl -s -L -o public/js/main.js "https://raw.githubusercontent.com/saiarlen/logmojo/main/public/js/main.js"
    curl -s -L -o public/images/logo.png "https://raw.githubusercontent.com/saiarlen/logmojo/main/public/images/logo.png"
    curl -s -L -o public/images/favicon.png "https://raw.githubusercontent.com/saiarlen/logmojo/main/public/images/favicon.png"
    
    # Update config.yaml (preserve existing if user modified)
    if [ ! -f "config.yaml.user" ]; then
        curl -s -L -o config.yaml.new "https://raw.githubusercontent.com/saiarlen/logmojo/main/config.yaml"
        if [ -f "config.yaml" ]; then
            mv config.yaml config.yaml.old
            print_warning "Existing config.yaml backed up as config.yaml.old"
        fi
        mv config.yaml.new config.yaml
    fi
    
    print_success "Assets updated"
}

# Replace binary atomically
replace_binary() {
    print_step "Replacing binary..."
    
    # Atomic replacement
    mv "$INSTALL_DIR/logmojo" "$INSTALL_DIR/logmojo.old"
    mv "$INSTALL_DIR/logmojo.new" "$INSTALL_DIR/logmojo"
    
    print_success "Binary replaced"
}

# Set permissions
set_permissions() {
    print_step "Setting permissions..."
    
    chown -R logmojo:logmojo $INSTALL_DIR
    chmod +x $INSTALL_DIR/logmojo
    
    # Ensure logmojo user still has proper group memberships
    # System log groups
    if getent group systemd-journal >/dev/null 2>&1; then
        usermod -a -G systemd-journal logmojo 2>/dev/null || print_warning "Failed to add user to systemd-journal group"
    fi
    if getent group adm >/dev/null 2>&1; then
        usermod -a -G adm logmojo 2>/dev/null || print_warning "Failed to add user to adm group"
    fi
    
    # Add to all existing user groups for application log access
    for group in $(getent group | grep -E '^[a-zA-Z][a-zA-Z0-9_-]*:[^:]*:[0-9]{4,}:' | cut -d: -f1); do
        if [ "$group" != "logmojo" ] && [ "$group" != "root" ] && [ "$group" != "nobody" ]; then
            usermod -a -G "$group" logmojo 2>/dev/null || true
        fi
    done
    
    print_success "Permissions set"
}

# Start service
start_service() {
    print_step "Starting Logmojo service..."
    
    systemctl start $SERVICE_NAME
    
    # Wait for service to start
    sleep 3
    
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "Service started successfully"
    else
        print_error "Failed to start service"
        print_error "Rolling back..."
        
        # Rollback
        systemctl stop $SERVICE_NAME 2>/dev/null || true
        mv "$INSTALL_DIR/logmojo" "$INSTALL_DIR/logmojo.failed"
        mv "$INSTALL_DIR/logmojo.old" "$INSTALL_DIR/logmojo"
        systemctl start $SERVICE_NAME
        
        print_error "Rollback completed. Check logs with: journalctl -u $SERVICE_NAME -f"
        exit 1
    fi
}

# Cleanup
cleanup() {
    print_step "Cleaning up..."
    
    # Remove old binary
    [ -f "$INSTALL_DIR/logmojo.old" ] && rm -f "$INSTALL_DIR/logmojo.old"
    
    # Remove old assets backups after successful update
    [ -d "$INSTALL_DIR/views.backup" ] && rm -rf "$INSTALL_DIR/views.backup"
    [ -d "$INSTALL_DIR/public.backup" ] && rm -rf "$INSTALL_DIR/public.backup"
    
    print_success "Cleanup completed"
}

# Print completion message
print_completion() {
    echo
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                              ║${NC}"
    echo -e "${GREEN}║                   🎉 UPDATE COMPLETE!                        ║${NC}"
    echo -e "${GREEN}║                                                              ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
    echo -e "${CYAN}📈 Version Update:${NC} ${YELLOW}$CURRENT_VERSION${NC} → ${GREEN}$LATEST_VERSION${NC}"
    echo
    echo -e "${CYAN}🌐 Access Logmojo:${NC} ${BLUE}http://$(hostname -I | awk '{print $1}'):7005${NC}"
    echo
    echo -e "${CYAN}🛠️  Service Status:${NC}"
    echo -e "   📊 Status:  ${GREEN}sudo systemctl status $SERVICE_NAME${NC}"
    echo -e "   📋 Logs:    ${GREEN}sudo journalctl -u $SERVICE_NAME -f${NC}"
    echo
    echo -e "${CYAN}💾 Database & Config:${NC} ${GREEN}Preserved${NC}"
    echo -e "${CYAN}🔄 Backup Location:${NC} ${BLUE}$BACKUP_DIR${NC}"
    echo
    echo -e "${YELLOW}📖 Changelog: https://github.com/saiarlen/logmojo/releases/tag/$LATEST_VERSION${NC}"
    echo
}

# Main update function
main() {
    print_banner
    
    check_root
    check_installation
    get_current_version
    detect_system
    get_latest_version
    check_update_needed
    create_backup
    stop_service
    download_binary
    update_assets
    replace_binary
    set_permissions
    
    # Update version in .env file
    if [ -f "$INSTALL_DIR/.env" ]; then
        sed -i "s/MONITOR_GENERAL_VERSION=\"[^\"]*\"/MONITOR_GENERAL_VERSION=\"$LATEST_VERSION\"/g" "$INSTALL_DIR/.env"
        print_success "Updated version to $LATEST_VERSION in .env"
    fi
    
    start_service
    cleanup
    print_completion
}

# Run main function
main "$@"
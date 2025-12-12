# Logmojo - High-Performance Log Management System

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/saiarlen/logmojo)](https://github.com/saiarlen/logmojo/releases)
[![GitHub Stars](https://img.shields.io/github/stars/saiarlen/logmojo?style=social)](https://github.com/saiarlen/logmojo/stargazers)

A high-performance, centralized log management and server monitoring agent designed for speed and simplicity. Provides instant log searching using `grep`/`zgrep`, system metrics, and advanced service management without heavy database ingestion.

**Perfect for DevOps, SysAdmins, Backend Engineers, SREs, and Platform Teams** who need fast, accurate, and zero-bloat log search with a beautiful real-time UI.

---

## ⚡ Quick Start

### Production Deployment (1 minute)

**Linux/macOS** - Automated installation:

```bash
curl -fsSL https://raw.githubusercontent.com/saiarlen/logmojo/main/scripts/deploy.sh | sudo bash
```

**Windows** - Manual installation:

1. Download [latest release](https://github.com/saiarlen/logmojo/releases/latest)
2. Download extract assets and files (views folder, public folder, .env, config.yaml) from the logmojo-assets.tar.gz
3. Setup .env and config.yaml from examples.env and config.yaml
4. Extract and run `logmojo-windows-amd64.exe`

**Access**: `http://localhost:7005` or `http://your-server-ip:7005`

> [!WARNING] > **Default Credentials**: `admin` / `admin`  
> ⚠️ **Change password immediately** after first login via Settings page! or CLI Commands

### Development Setup (2 minutes)

```bash
# Clone and setup
git clone https://github.com/saiarlen/logmojo.git
cd logmojo
cp .env.example .env
go mod download

# Run with live reload (recommended)
go install github.com/cosmtrek/air@latest
air

# Or standard run
go run main.go
```

**Access**: `http://localhost:7005` (admin/admin)

---

## 📑 Table of Contents

- [Key Features](#-key-features)
- [Why Logmojo?](#-why-logmojo)
- [Architecture](#️-architecture)
- [Installation](#-installation)
  - [System Requirements](#system-requirements)
  - [Production Deployment](#production-deployment)
  - [Development Setup](#development-setup)
- [Getting Started](#️-getting-started)
- [Configuration](#️-configuration)
- [API Reference](#-api-reference)
- [How It Works](#-how-it-works)
- [Performance](#-performance)
- [Troubleshooting](#-troubleshooting)
- [Project Structure](#-project-structure)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🚀 Key Features

### 📊 **Log Management & Search**

- **Centralized Log Management**: Configurable log aggregation from multiple apps and services
- **High-Performance Search**: Direct file-based search using `grep` - no database ingestion lag
- **Multi-Format Archive Support**: Automatically searches `.gz`, `.bz2`, `.xz`, `.lz4` files
- **Real-Time Log Streaming**: Live log tailing with WebSocket connections
- **Advanced Timestamp Parsing**: Supports ISO 8601, syslog, Unix timestamps, and more
- **Smart File Discovery**: Intelligent scanning to find log files and rotated siblings

### 🖥️ **System Monitoring**

- **Real-Time Metrics**: CPU, RAM, Disk, and Network metrics via WebSocket
- **Historical Data**: SQLite storage for metric history and graphing
- **Process Manager**: View and manage running processes with kill functionality
- **Advanced Service Management**: Complete systemd service control and monitoring

### 🚨 **Intelligent Alert System**

- **Multi-Type Alert Rules**: System metrics, log patterns, exception detection, service status
- **Smart Duplicate Prevention**: Hash-based tracking prevents repeated alerts
- **Persistent Alert Memory**: Database storage survives application restarts
- **Email Notifications**: HTML-formatted alerts with severity color coding
- **Real-Time Processing**: Instant alerts for new errors without cooldown delays

### 🔐 **Security & Authentication**

- **JWT Token-Based Auth**: Secure, stateless authentication with 24-hour validity
- **HTTP-Only Cookies**: Prevents XSS attacks
- **Bcrypt Password Hashing**: Industry-standard encryption
- **User Management CLI**: Create, update, delete users via command line

### 🎨 **Modern User Interface**

- **Server-Side Rendering**: Built with Jet templates, TailwindCSS, and DaisyUI
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Theme Consistency**: Dark theme with glassmorphism effects
- **Copy & Export**: Copy log messages and export search results
- **Screenshot Capture**: Built-in screenshot functionality

---

## 🤔 Why Logmojo?

### Comparison with Popular Alternatives

| Feature               | Logmojo                      | ELK Stack              | Grafana Loki | Splunk |
| --------------------- | ---------------------------- | ---------------------- | ------------ | ------ |
| **Setup Time**        | 1 minute                     | 2-4 hours              | 30+ minutes  | Hours  |
| **Memory Usage**      | ~100MB                       | 4GB+                   | 500MB+       | 2GB+   |
| **Database Required** | No (SQLite for metrics only) | Yes (Elasticsearch)    | Yes          | Yes    |
| **Search Speed**      | Instant (grep)               | Fast                   | Medium       | Fast   |
| **Complexity**        | Single Binary                | Complex (3+ services)  | Medium       | High   |
| **Cost**              | Free (MIT)                   | Free                   | Free         | Paid   |
| **Learning Curve**    | Minutes                      | Days                   | Hours        | Days   |
| **Compressed Logs**   | Native support               | Requires preprocessing | Limited      | Yes    |

### Key Advantages

✅ **Zero Database Ingestion** - Logs stay as files, search happens in real-time  
✅ **Single Binary** - No complex setup, no dependencies  
✅ **Instant Search** - Direct grep on files, no indexing delay  
✅ **Lightweight** - Runs on 512MB RAM  
✅ **Production Ready** - Used in production environments

---

## 🏗️ Architecture

Single binary Go application acting as both web server and monitoring daemon.

### **Core Components**

- **Web Server**: Fiber v2 framework serving UI and REST API
- **Log Engine**: Direct `grep`/`zgrep` execution on files (no DB ingestion)
- **Metrics Engine**: gopsutil for host metrics, SQLite for history
- **Alert Engine**: Real-time monitoring with duplicate prevention
- **WebSocket**: Live updates for metrics, logs, and alerts

### **Tech Stack**

- **Backend**: Go 1.24+, Fiber v2, SQLite3
- **Frontend**: Vanilla JS, TailwindCSS, DaisyUI, ECharts
- **Templates**: Jet templating engine
- **Authentication**: JWT with bcrypt
- **Configuration**: Viper (YAML + environment variables)

---

## 📦 Installation

### System Requirements

**Minimum:**

- **OS**: Linux, macOS, Windows (x86_64 or ARM64)
- **RAM**: 512MB
- **Disk**: 100MB free space
- **Dependencies**: `grep`, `zgrep` (for log search)

**Recommended:**

- **RAM**: 1GB+
- **Disk**: 1GB+ (for log storage)
- **Network**: 1Gbps+ for high log volume

### Production Deployment

#### Option 1: Automated Installation (Linux/macOS)

**One-line installation:**

```bash
curl -fsSL https://raw.githubusercontent.com/saiarlen/logmojo/main/scripts/deploy.sh | sudo bash
```

**What this does:**

- ✅ Auto-detects OS/architecture (Linux/macOS, x86_64/ARM64)
- ✅ Downloads latest binary from GitHub releases
- ✅ Downloads required assets (views, public files, configs)
- ✅ Creates system user (`logmojo`)
- ✅ Creates systemd service
- ✅ Configures firewall (if UFW is available)
- ✅ Starts service automatically
- ✅ Sets up proper permissions and security

**Installation location**: `/opt/logmojo`

**After installation:**

```bash
# Check service status
sudo systemctl status logmojo

# View logs
sudo journalctl -u logmojo -f

# Full logs
cd /opt/logmojo/logs/

# Manage service
sudo systemctl start|stop|restart logmojo
```

#### Option 2: Manual Installation (All Platforms)

**Step 1: Download Binary**

```bash
# Download latest release for your platform
# Linux x86_64
wget https://github.com/saiarlen/logmojo/releases/latest/download/logmojo-linux-amd64

# macOS ARM64
wget https://github.com/saiarlen/logmojo/releases/latest/download/logmojo-darwin-arm64

# Windows x86_64
# Download from: https://github.com/saiarlen/logmojo/releases/latest
```

**Step 2: Download Required Files**

```bash
# Clone repository for assets
git clone https://github.com/saiarlen/logmojo.git
cd logmojo

# move binary into logmojo folder
mv ../logmojo-* ../logmojo

# Or download individual files from GitHub
```

**Step 3: Setup Configuration**

```bash
# Copy environment file
cp .env.example .env

# Edit configuration
nano config.yaml  # Add your log paths
nano .env         # Configure server settings
```

**Step 4: Run**

```bash
# Make executable
chmod +x logmojo-linux-amd64

# Run
./logmojo-linux-amd64
```

**Step 5: Access**

- Open browser: `http://localhost:7005`
- Login: `admin` / `admin`
- **Change password immediately!**

#### Option 3: Systemd Service (Manual Setup)

Create `/etc/systemd/system/logmojo.service`:

```ini
[Unit]
Description=Logmojo - High-Performance Log Management System
After=network.target

[Service]
Type=simple
User=logmojo
WorkingDirectory=/opt/logmojo  
ExecStart=/opt/logmojo/logmojo
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable logmojo
sudo systemctl start logmojo
```

### Development Setup

**Prerequisites:**

- Go 1.24+ installed
- Git

**Setup steps:**

```bash
# 1. Clone repository
git clone https://github.com/saiarlen/logmojo.git
cd logmojo

# 2. Install dependencies
go mod download

# 3. Setup environment
cp .env.example .env

# 4. Run with live reload (recommended)
go install github.com/cosmtrek/air@latest
air

# Or run normally
go run main.go
```

**Access**: `http://localhost:7005` (admin/admin)

**Development workflow:**

```bash
# Pull latest changes
git pull origin main

# Update dependencies
go mod tidy

# Build binary
go build -o logmojo .
```
### To test the alerts follow tests/test-alerts.md
---

## ⚙️ Getting Started

### 1. First Login

- Access web interface: `http://your-server:7005`
- Default credentials: `admin` / `admin`
- **⚠️ Change password immediately via Settings**

### 2. Configure Log Sources

Edit `config.yaml` to add your applications:

```yaml
apps:
  - name: "My Application"
    service_name: "my-app"
    logs:
      - name: "Application Log"
        path: "/var/log/myapp/app.log"
      - name: "Error Log"
        path: "/var/log/myapp/error.log"
```

Restart service:

```bash
sudo systemctl restart logmojo
```

### 3. Setup Alerts (Optional)

**Configure email notifications** in `.env`:

```bash
MONITOR_NOTIFIERS_EMAIL_ENABLED=true
MONITOR_NOTIFIERS_EMAIL_SMTP_HOST=smtp.gmail.com
MONITOR_NOTIFIERS_EMAIL_SMTP_PORT=587
MONITOR_NOTIFIERS_EMAIL_USERNAME=your-email@gmail.com
MONITOR_NOTIFIERS_EMAIL_PASSWORD=your-app-password  # Gmail App Password
MONITOR_NOTIFIERS_EMAIL_FROM=alerts@company.com
MONITOR_NOTIFIERS_EMAIL_TO=admin@company.com
```

> **Note**: For Gmail, use [App Passwords](https://support.google.com/accounts/answer/185833), not your regular password.

**Create alert rules** via web interface:

- Navigate to Alerts page
- Click "Create Alert Rule"
- Configure conditions and notifications
- Test alerts with built-in test function

### 4. User Management

**Via Web Interface:**

- Go to Settings → Change Password

**Via Command Line:**

```bash
# Create user
./logmojo --user=create --username=john --password=SecurePass123

# List users
./logmojo --user=list

# Update password
./logmojo --user=update --username=john --password=NewPass456

# Delete user
./logmojo --user=delete --username=john
```

### 5. Service Management

**Linux/macOS (systemd):**

```bash
# Start service
sudo systemctl start logmojo

# Stop service
sudo systemctl stop logmojo

# Restart service
sudo systemctl restart logmojo

# Check status
sudo systemctl status logmojo

# View logs
sudo journalctl -u logmojo -f
```

**Windows:**

- Run `logmojo.exe` directly or use NSSM to create a Windows service

### 6. Updates

#### Zero-Downtime Update (Linux/macOS)

**Automated update** (preserves data and configuration):

```bash
curl -fsSL https://raw.githubusercontent.com/saiarlen/logmojo/main/scripts/update.sh | sudo bash
```

This update script will:
- ✅ Download the latest binary
- ✅ Preserve your existing configuration and data
- ✅ Restart the service with zero downtime
- ✅ Maintain all user accounts and settings

#### Manual Update (Windows/Linux/macOS)

**For Linux/macOS:**

```bash
# 1. Backup current installation
cp .env .env.backup
cp config.yaml config.yaml.backup
cp monitor.db monitor.db.backup

# 2. Stop service
sudo systemctl stop logmojo

# 3. Download and replace binary
wget https://github.com/saiarlen/logmojo/releases/latest/download/logmojo-linux-amd64
chmod +x logmojo-linux-amd64

# 4. Update views/ and public/ folders from latest release
# 5. Restart service
sudo systemctl start logmojo
```

**For Windows:**

1. Backup `.env`, `config.yaml`, and `monitor.db` files
2. Download latest release from [GitHub](https://github.com/saiarlen/logmojo/releases/latest)
3. Stop Logmojo service (`nssm stop logmojo`)
4. Replace binary and update `views/` and `public/` folders
5. Restart service (`nssm start logmojo`)

> **Note**: Database schema updates happen automatically on startup

---

## 🛠️ Configuration

### Firewall Setup

**UFW (Ubuntu/Debian):**

```bash
sudo ufw allow 7005/tcp
```

**Firewalld (CentOS/RHEL):**

```bash
sudo firewall-cmd --permanent --add-port=7005/tcp
sudo firewall-cmd --reload
```

**Windows Firewall:**

```powershell
netsh advfirewall firewall add rule name="Logmojo" dir=in action=allow protocol=TCP localport=7005
```

### Environment Variables

Key configuration options in `.env`:

```bash
# Server Configuration
MONITOR_SERVER_LISTEN_ADDR=0.0.0.0:7005
MONITOR_DATABASE_PATH=./monitor.db

# Security
MONITOR_SECURITY_JWT_SECRET=your-secret-key-change-this
MONITOR_SECURITY_SESSION_TIMEOUT=24h

# System Alerts
MONITOR_ALERTS_CPU_HIGH_ENABLED=true
MONITOR_ALERTS_CPU_HIGH_THRESHOLD=80.0
MONITOR_ALERTS_DISK_LOW_ENABLED=true
MONITOR_ALERTS_DISK_LOW_THRESHOLD_PERCENT_FREE=10.0

# Email Notifications
MONITOR_NOTIFIERS_EMAIL_ENABLED=true
MONITOR_NOTIFIERS_EMAIL_SMTP_HOST=smtp.gmail.com
MONITOR_NOTIFIERS_EMAIL_SMTP_PORT=587
MONITOR_NOTIFIERS_EMAIL_USERNAME=your-email@gmail.com
MONITOR_NOTIFIERS_EMAIL_PASSWORD=your-app-password

# Webhook Notifications (Slack, Discord, etc.)
MONITOR_NOTIFIERS_WEBHOOK_ENABLED=false
MONITOR_NOTIFIERS_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK
```

---

## 📡 API Reference

### **Log Search**

```bash
# Basic search
GET /api/logs/search?query=error&app=MyApp&limit=100

# Search with filters
GET /api/logs/search?query=exception&app=MyApp&log=ErrorLog&level=ERROR

# Live log streaming (WebSocket)
WS /api/ws/logs?app=MyApp&log=ErrorLog
```

### **System Metrics**

```bash
# Current metrics
GET /api/metrics/host

# Historical data
GET /api/metrics/history?type=cpu&range=24h

# Live metrics (WebSocket)
WS /api/ws/metrics
```

### **Alert Management**

```bash
# List alert rules
GET /api/alerts/rules

# Create alert rule
POST /api/alerts/rules
Content-Type: application/json
{
  "name": "High CPU Alert",
  "type": "cpu",
  "threshold": 80.0,
  "enabled": true
}

# Get alert history
GET /api/alerts/history
```

### **Service Management**

```bash
# List services
GET /api/services

# Control services
POST /api/services/start
POST /api/services/stop
POST /api/services/restart
Content-Type: application/json
{
  "service_name": "nginx"
}
```

---

## 🔍 How It Works

### **Log Search Engine**

1. **File Discovery**: Intelligently finds log files and archives
2. **Tool Selection**: Uses appropriate grep variant (grep/zgrep/bzgrep/xzgrep)
3. **Streaming Output**: Uses `StdoutPipe` to stream results without loading entire files
4. **Timestamp Parsing**: Extracts timestamps from multiple log formats
5. **Safety Limits**: 10-second timeout, 2000-line limits, max 15 files per search

### **Performance Features**

- **No Database Ingestion**: Logs remain as files for maximum performance
- **Compressed File Support**: Native support for gzip, bzip2, xz, and lz4
- **Smart Caching**: Intelligent file selection based on modification time
- **Resource Management**: Automatic timeout and memory management

---

## 📈 Performance

### Benchmarks

- **Search 1GB log file**: ~200ms
- **Search 10GB compressed logs**: ~2 seconds
- **Memory usage**: 50-100MB (idle), 200MB (under load)
- **Concurrent users**: 100+
- **WebSocket connections**: 1000+
- **Log ingestion**: N/A (no ingestion, direct file access)

### Performance Optimization

- **File Limits**: Maximum 15 files per search, fallback to 3 most recent
- **Search Timeout**: 10-second timeout for better UX
- **Result Limits**: 2000 lines maximum per search
- **Memory Management**: Streaming output prevents OOM issues

---

## 🐛 Troubleshooting

### Common Issues

**"No Logs Found"**

- Check file permissions: `ls -la /var/log/myapp/`
- Verify paths in `config.yaml`
- Ensure log files exist and are readable

**Search Not Working**

- Ensure `grep`/`zgrep` tools are installed: `which grep zgrep`
- Check `$PATH` environment variable
- Test grep manually: `grep "error" /var/log/myapp/app.log`

**Service Won't Start**

- Check logs: `sudo journalctl -u logmojo -f`
- Verify port 7005 is available: `sudo lsof -i :7005`
- Check permissions: `ls -la /opt/logmojo`

**Port 7005 Already in Use**

- Change port in `.env`: `MONITOR_SERVER_LISTEN_ADDR=0.0.0.0:8080`
- Restart service: `sudo systemctl restart logmojo`

**Email Alerts Not Working**

- Verify SMTP settings in `.env`
- For Gmail, use App Password (not regular password)
- Test with built-in alert test function
- Check logs for SMTP errors

### Database Notes

- `monitor.db` stores only metadata (metrics, alerts, users)
- Logs are **never** stored in database (performance feature)
- Safe to delete `monitor.db` to reset (will lose metrics history)
- Database is automatically created on first run

---

## 📁 Project Structure

```
.
├── main.go                 # Application entry point
├── internal/
│   ├── api/               # HTTP routes & handlers
│   ├── config/            # Configuration management
│   ├── db/                # SQLite operations
│   ├── logs/              # Log search engine
│   ├── metrics/           # System metrics collection
│   ├── alerts/            # Alert management system
│   ├── processes/         # Process management
│   ├── services/          # Service management
│   ├── auth/              # Authentication & JWT
│   ├── ws/                # WebSocket handlers
│   └── version/           # Version information
├── views/                 # HTML templates (Jet)
├── public/                # Static assets (CSS/JS)
├── scripts/               # Deployment scripts
├── config.yaml            # Runtime configuration
└── .env.example           # Environment variables template
```

---

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### **Development Workflow**

```bash
git pull origin main
go mod tidy
air  # Start with live reload
```

### **Areas for Contribution**

- Additional log format parsers
- New alert rule types
- UI/UX improvements
- Performance optimizations
- Documentation improvements
- Docker/Kubernetes support
- Multi-server log aggregation

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Fast HTTP framework
- [TailwindCSS](https://tailwindcss.com/) - CSS framework
- [DaisyUI](https://daisyui.com/) - Component library
- [ECharts](https://echarts.apache.org/) - Charting library
- [gopsutil](https://github.com/shirou/gopsutil) - System metrics

---

## 📞 Support

- 📖 **Documentation**: [GitHub Repository](https://github.com/saiarlen/logmojo)
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/saiarlen/logmojo/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/saiarlen/logmojo/discussions)
- ⭐ **Star us on GitHub** if you find this useful!

---

**Made with ❤️ for the DevOps community**

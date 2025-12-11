# Logmojo (Log Monitor)

A high-performance, centralized log management and server monitoring agent. Designed for speed and simplicity, it provides instant log searching (using `grep`/`zgrep` under the hood), system metrics, and advance service management without heavy database ingestion for logs.

## 🚀 Key Features

### 📊 Log Management & Search
- **Centralized Log Management**: Configurable log aggregation from multiple apps and services
- **High-Performance Search**: Direct file-based search using `grep`, supporting Regex and massive files instantly. No database ingestion lag
- **Multi-Format Archive Support**: Automatically discovers and searches rotated/archived logs (`.gz`, `.bz2`, `.xz`, `.lz4`, `.1`, etc.)
- **Historical Search**: 10-day historical log search with date range selection (24h, 3d, 7d, 10d, 30d)
- **Real-Time Log Streaming**: Live log tailing with WebSocket connections
- **Smart File Discovery**: Intelligent file scanning to find log files and their rotated siblings
- **Performance Optimized**: 10-second timeouts, 2000-line limits, and smart file selection (max 15 files)
- **Advanced Timestamp Parsing**: Supports multiple formats including macOS system logs, ISO 8601, syslog, and Unix timestamps
- **Intelligent Alert System**: Real-time monitoring with duplicate prevention and persistent tracking

### 🖥️ System Monitoring
- **Real-Time Metrics**: CPU, RAM, Disk, and Network metrics using WebSocket
- **Historical Data**: SQLite storage for metric history and graphing
- **Process Manager**: View and manage running processes
- **Advanced Service Management**: Comprehensive systemd service control and monitoring
- **Advanced Alert Management**: Multi-type alert rules with smart duplicate prevention and persistent tracking

### 🔐 Security & Authentication
- **JWT Token-Based Auth**: Secure, stateless authentication with 24-hour validity
- **HTTP-Only Cookies**: Prevents XSS attacks
- **Bcrypt Password Hashing**: Industry-standard encryption
- **User Management CLI**: Create, update, delete users via command line
- **Auto-Logout on Expiry**: Frontend redirects on invalid token

### 🎨 User Interface
- **Modern UI**: Built with server-side Jet templates, TailwindCSS, and DaisyUI
- **Theme Consistency**: DaisyUI theme colors and consistent styling
- **Copy & Export**: Copy log messages and export search results
- **Screenshot Capture**: Built-in screenshot functionality
- **Row Highlighting**: Visual feedback for selected log entries
- **Responsive Design**: Works on desktop and mobile devices

### ⚙️ Settings & Customization
- **Application Branding**: Customize app name and copyright text
- **Logo Management**: Upload custom logo and favicon with toggle between text/image display
- **Password Management**: Secure password change functionality with current password verification
- **System Information**: Display system metrics, version info, and build details
- **Persistent Settings**: Database storage for all customization preferences

### ⚙️ Advanced Service Management
- **Dual View Modes**: Professional table view and compact card view
- **Real-Time Status**: Live service status monitoring with auto-refresh
- **Complete Service Control**: Start, stop, restart, enable, and disable services
- **Resource Monitoring**: CPU, memory, PID, and uptime tracking per service
- **Integrated Log Access**: Direct navigation to service logs with one click
- **Configuration Management**: Quick access to service configuration files
- **Status Indicators**: Color-coded status display (running, inactive, failed)
- **Bulk Operations**: Manage multiple services efficiently
- **Service Discovery**: Automatic detection of common system services

### ⚡ Performance Features
- **No Database Ingestion**: Logs are not stored in database for maximum performance
- **Compressed File Support**: Native support for gzip, bzip2, xz, and lz4 archives
- **Smart Caching**: Intelligent file selection based on modification time
- **Resource Management**: Automatic timeout and memory management
- **Fallback Mechanisms**: Graceful degradation when tools are unavailable

## 🏗 Architecture

The application is a single binary Go agent acting as a web server and monitoring daemon.

### Core Components

1.  **Web Server (`internal/api`)**:
    - Powered by [Fiber](https://gofiber.io/).
    - Serves UI (Jet templates) and REST API.
    - Handles WebSocket connections for live metrics.
2.  **Log Engine (`internal/logs`)**:
    - **No DB Ingestion**: Logs are _not_ stored in the internal SQLite DB.
    - **Search**: Uses `exec.Command` to run optimized `grep` (or `zgrep` for archives) directly on files.
    - **Discovery**: Intelligent file scanning to find log files and their rotated siblings.
3.  **Metrics Engine (`internal/metrics`)**:
    - Collects host metrics (gopsutil).
    - Stores historical metric data in SQLite (`internal/db`) for graphing.
4.  **Config (`config.yaml`)**:
    - Central source of truth for defined Apps, Services, and Log paths.

### Tech Stack

- **Backend**: Go 1.21+
- **Web Framework**: Fiber v2
- **Templates**: Jet
- **Database**: SQLite3 (Metrics/Auth only)
- **Frontend**: Vanilla JS, TailwindCSS, DaisyUI, ECharts
- **Log Search**: `grep`, `zgrep` (System dependencies)

## 📂 Codebase Structure

```text
.
├── main.go   # Main entry point (main.go)
├── internal/
│   ├── api/              # HTTP Routes & Handlers
│   ├── config/           # Config loading (Viper)
│   ├── db/               # SQLite init & migrations (Metrics/Users)
│   ├── logs/             # Log search & discovery logic (KEY COMPONENT)
│   ├── metrics/          # Host metrics collection
│   ├── processes/        # Process listing
│   └── ws/               # WebSocket handlers
├── views/                # HTML Templates (Jet)
│   ├── layouts/          # Base layouts
│   └── *.jet.html        # Page views
├── public/               # Static assets (JS/CSS usually CDN based)
├── config.yaml           # Runtime configuration
└── demo_logs/            # (Optional) Local test logs
```

## ⚙️ Configuration

Logmojo supports multiple configuration methods:

### Environment Variables (.env)

Create a `.env` file in the project root for easy configuration:

```bash
# Copy the example file
cp .env.example .env

# Edit with your settings
vim .env
```

Key environment variables:

```bash
# Server Configuration
MONITOR_SERVER_LISTEN_ADDR=0.0.0.0:7005
MONITOR_SERVER_AUTH_TOKEN=your-secret-token

# Database
MONITOR_DB_PATH=./monitor.db

# Security
MONITOR_JWT_SECRET=your-jwt-secret
MONITOR_SESSION_TIMEOUT=24h

# Alerts
MONITOR_ALERTS_CPU_HIGH_THRESHOLD=80.0
MONITOR_ALERTS_DISK_LOW_THRESHOLD_PERCENT_FREE=10.0

# Email Notifications
MONITOR_NOTIFIERS_EMAIL_ENABLED=true
MONITOR_NOTIFIERS_EMAIL_SMTP_HOST=smtp.gmail.com
MONITOR_NOTIFIERS_EMAIL_USERNAME=your-email@gmail.com
```

### YAML Configuration (`config.yaml`)

The system uses a hierarchical log definition:

```yaml
apps:
  - name: "App Name" # Grouping in Sidebar
    service_name: "service" # Internal ID
    logs:
      - name: "Access Log" # Display Name
        path: "/var/log/nginx/access.log" # File or Directory path
```

### Configuration Priority

1. **Environment Variables** (highest priority)
2. **config.yaml** file
3. **Default values** (lowest priority)

### Service Configuration

Services are organized within apps for better management:

```yaml
apps:
  - name: "System Services"
    logs: [...]
    services:
      - name: "Nginx"
        service_name: "nginx"
        enabled: true
        description: "HTTP and reverse proxy server"
        config_path: "/etc/nginx/nginx.conf"
        log_path: "/var/log/nginx/error.log"
```

### Log Path Configuration

- If `path` is a **Directory**: All `.log`, `.txt`, `.gz`, `.bz2`, `.xz`, `.lz4` files in it are discovered.
- If `path` is a **File**: The file and its rotated siblings (e.g., `app.log.1`, `app.log.gz`) are discovered.
- **Archive Support**: Automatically detects and searches compressed logs with appropriate tools.
- **Date Range Filtering**: Search logs within specific time ranges (24h, 3d, 7d, 10d, 30d).
- **Service Log Integration**: Direct access to service logs via journalctl integration.

### 🚨 Advanced Alert Management System
- **Multi-Type Alert Rules**: System metrics, log patterns, exception detection, and service status monitoring
- **Smart Duplicate Prevention**: Hash-based tracking prevents repeated alerts for same log entries
- **Persistent Alert Memory**: Database storage survives application restarts - no duplicate alerts after restart
- **Time-Based Filtering**: Only alerts on recent errors (last 10 minutes) to avoid historical noise
- **Automatic Cleanup**: Removes processed entries older than 24 hours to maintain performance
- **Real-Time Processing**: Instant alerts for new errors without cooldown delays
- **Multi-Language Exception Detection**: Automatic detection for Java, Python, JavaScript, PHP, Ruby, and Go
- **Email Notifications**: HTML-formatted alerts with severity color coding and custom FROM addresses
- **Configurable Severity Levels**: Low, Medium, High, Critical with color-coded indicators
- **Alert History & Management**: Complete audit trail with resolve/unresolve capabilities
- **Rule Management UI**: Create, edit, enable/disable, and delete alert rules through web interface
- **Performance Optimized**: Hash-based keys and database indexes for efficient processing
- **Real-Time Updates**: WebSocket-based live updates for alert history and rule status
- **Intelligent Cache Management**: Rules cache automatically refreshes when modified
- **Modern UI**: Clean, responsive interface with subtle styling and intuitive controls

## 🆕 New Features (Latest Updates)

### 🔧 Advanced Service Management
- **Professional Service Manager**: Complete systemd service management interface
- **Dual View Modes**: Switch between detailed table view and compact card view
- **Real-Time Monitoring**: Live service status with CPU, memory, and uptime tracking
- **Integrated Log Access**: One-click navigation to service logs with URL parameters
- **Configuration Management**: Quick copy of service configuration file paths
- **Auto-Refresh**: Configurable automatic service status updates
- **Service Discovery**: Automatic detection and management of common services
- **Status Indicators**: Color-coded visual status (running/inactive/failed)
- **Bulk Operations**: Efficient management of multiple services
- **Service-Log Integration**: Seamless connection between services and their logs

### 📝 Event Logging & Audit Trail
- **Comprehensive Audit Logging**: All user actions logged with timestamps and user context
- **File Rotation**: 10MB log files with 10-file retention for efficient storage
- **Detailed Event Tracking**: Process kills, service management, settings changes, file uploads
- **Structured Logging**: Consistent format with event type, user, and detailed information
- **HTTP Request Logging**: Integrated Fiber middleware for complete request tracking

### 📅 Historical Log Search
- **Date Range Selection**: Search logs from last 24 hours up to 30 days
- **Smart File Selection**: Automatically selects relevant files based on modification time
- **Performance Optimized**: Limits file count to maintain fast search times
- **UI Integration**: Dropdown selector in logs page for easy date range selection

### 📋 Enhanced Log Management
- **Copy Log Messages**: One-click copy functionality for individual log entries
- **Export Search Results**: Export filtered logs to text files
- **Screenshot Capture**: Built-in screenshot functionality for documentation
- **Row Highlighting**: Visual feedback for selected log entries
- **Real-Time Streaming**: Live log tailing with WebSocket connections

### 🔧 Multi-Format Archive Support
- **Compression Formats**: Support for `.gz`, `.bz2`, `.xz`, `.lz4` archives
- **Automatic Tool Selection**: Uses appropriate grep variant (zgrep, bzgrep, xzgrep)
- **Fallback Mechanisms**: Graceful degradation when compression tools unavailable
- **Mixed File Handling**: Seamlessly searches both compressed and uncompressed logs

### ⚡ Performance Improvements
- **Optimized Search**: 10-second timeout with 2000-line result limits
- **Smart File Limiting**: Maximum 15 files per search, fallback to 3 most recent
- **Memory Management**: Streaming output prevents out-of-memory issues
- **Efficient Parsing**: Enhanced timestamp parsing for multiple log formats

### 🎨 UI/UX Enhancements
- **Theme Consistency**: Updated styling to match DaisyUI theme colors
- **Better Typography**: Improved readability and visual hierarchy
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Loading States**: Visual feedback during search operations
- **Error Handling**: User-friendly error messages and recovery options

## 🛠 Development & Debugging

### Prerequisites

- Go 1.21+
- `grep` (and `zgrep` for archive support) installed on system.

### Running Locally

Use [Air](https://github.com/cosmtrek/air) for live reloading:

```bash
air
```

Or standard Go run:

```bash
go run ./cmd/logmojo/main.go
# Server listens on :7005
# Default Auth: admin / admin (created on first run)
```

### Debugging Notes

- **"No Logs Found"**: Check permissions. The agent must have read access to the target log files.
- **Search Issues**: The search relies on system `grep`. Ensure `grep` is in `$PATH`.
- **Database**: `monitor.db` (SQLite) stores _only_ metrics history, alerts, and users. Deleting it resets auth and graphs, but not logs.

## 📡 API Endpoints

### Log Search API
```bash
# Basic search
GET /api/logs/search?query=error&app=MyApp&limit=100

# Search with date range
GET /api/logs/search?query=error&dateRange=7d&limit=100

# Live log streaming (WebSocket)
WS /ws/logs/stream?file=/path/to/log.log
```

### Alert Management API
```bash
# List all alert rules
GET /api/alerts/rules

# Create alert rule
POST /api/alerts/rules

# Update alert rule
PUT /api/alerts/rules/{id}

# Delete alert rule
DELETE /api/alerts/rules/{id}

# Toggle rule status
POST /api/alerts/rules/{id}/toggle

# Get alert history
GET /api/alerts/history

# Resolve alert
POST /api/alerts/{id}/resolve

# Test alert
POST /api/alerts/test
```

### Service Management API
```bash
# List all services
GET /api/services

# Service actions
POST /api/services/start
POST /api/services/stop
POST /api/services/restart
POST /api/services/enable
POST /api/services/disable

# Service logs
GET /api/services/{service}/logs?lines=100
```

### System Metrics API
```bash
# Current system metrics
GET /api/metrics

# Historical metrics
GET /api/metrics/history?hours=24

# Live metrics (WebSocket)
WS /ws/metrics
```

### Process Management API
```bash
# List processes
GET /api/processes

# Kill process
POST /api/processes/kill
```

### Settings API
```bash
# Get app settings
GET /api/settings/app

# Update app settings (branding, logo, favicon)
POST /api/settings/app

# Change password
POST /api/settings/password

# Get system information
GET /api/system/info
```

## 🔍 Log Search Logic (`internal/logs/logs.go`)

The search does **not** read files into memory completely for maximum performance.

### Search Process
1. **File Discovery**: Intelligently discovers log files and archives based on configuration
2. **Date Range Filtering**: Filters files by modification time when date range specified
3. **Tool Selection**: Automatically selects appropriate grep tool (grep/zgrep/bzgrep/xzgrep/lz4grep)
4. **Command Construction**: Builds optimized grep command with performance flags
5. **Streaming Output**: Uses `StdoutPipe` to stream results without loading entire files
6. **Timestamp Parsing**: Extracts and parses timestamps from multiple log formats
7. **Result Processing**: Parses lines into structured `LogResult` objects
8. **Safety Limits**: Applies caps (2000 lines, 10s timeout) to prevent resource exhaustion
9. **Sorting**: Sorts results by timestamp (newest first) before returning JSON

### Supported Log Formats
- **macOS System Logs**: `Tue Dec 10 15:30:45 2024`
- **ISO 8601**: `2024-12-10T15:30:45Z`
- **Syslog**: `Dec 10 15:30:45`
- **Unix Timestamps**: `1702215045` (seconds/milliseconds)
- **Custom Formats**: Configurable regex patterns for application-specific formats

### Archive Handling
- **Automatic Detection**: Recognizes compressed files by extension
- **Tool Fallback**: Falls back to compatible tools when preferred tool unavailable
- **Mixed Processing**: Handles both compressed and uncompressed files in single search
- **Performance Optimization**: Limits compressed file processing for speed

## 🔐 Authentication & User Management

### Security Features

✅ **JWT Token-Based Auth** - Secure, stateless authentication  
✅ **HTTP-Only Cookies** - Prevents XSS attacks  
✅ **Token Expiration** - 24-hour validity  
✅ **Bcrypt Password Hashing** - Industry-standard encryption  
✅ **Auto-Logout on Expiry** - Frontend redirects on invalid token  

### User Management Commands

**Commands:**
```bash
# List all users
./logmojo --user=list

# Create new user
./logmojo --user=create --username=john --password=SecurePass123

# Update password
./logmojo --user=update --username=john --password=NewSecurePass456

# Delete user
./logmojo --user=delete --username=john

# Custom database path
./logmojo --user=list --db=/path/to/monitor.db
```

### Production Security Setup

**1. Generate Strong JWT Secret:**
```bash
openssl rand -base64 32
```

**2. Set Environment Variable:**
```bash
# Add to .env file
MONITOR_JWT_SECRET="generated-secret-from-step-1"
```

**3. Change Default Admin Password:**
```bash
./logmojo --user=update --username=admin --password=YourStrongPassword123!
```

**4. Create Additional Users:**
```bash
./logmojo --user=create --username=operator --password=SecurePass123
./logmojo --user=create --username=viewer --password=ViewerPass456
```

### Security Best Practices

1. ✅ Use strong JWT secret (min 32 characters)
2. ✅ Enable HTTPS in production
3. ✅ Change default admin password
4. ✅ Use environment variables for secrets
5. ✅ Regular password rotation
6. ✅ Monitor failed login attempts

## 📧 Email Notifications

Configure email alerts for system events:

```bash
# Enable email notifications
MONITOR_NOTIFIERS_EMAIL_ENABLED=true
MONITOR_NOTIFIERS_EMAIL_SMTP_HOST=smtp.gmail.com
MONITOR_NOTIFIERS_EMAIL_USERNAME=your-email@gmail.com
MONITOR_NOTIFIERS_EMAIL_PASSWORD=your-app-password
```

## 📊 Performance Optimizations

- **File Limits**: Maximum 15 files per search, fallback to 3 most recent
- **Search Timeout**: 10-second timeout for better UX
- **Result Limits**: 2000 lines maximum per search
- **Smart Caching**: Files selected by modification time
- **Compression Support**: Native handling of multiple archive formats
- **Memory Management**: Streaming output to prevent OOM

## 🤝 Contributing

- **Frontend Changes**: Edit `views/*.jet.html`. No build step required (just reload).
- **Backend Changes**: Restart server (or use Air).
- **Testing**: Use demo_logs/ directory for local testing.
- **Security**: Follow authentication best practices for any auth-related changes.

```

```

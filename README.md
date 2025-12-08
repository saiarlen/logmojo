# Logger EMP (Local Monitor)

A high-performance, centralized log management and server monitoring agent. Designed for speed and simplicity, it provides instant log searching (using `grep`/`zgrep` under the hood), system metrics, and basic service management without heavy database ingestion for logs.

## 🚀 Key Features

- **Centralized Log Management**: Configurable log aggregation from multiple apps and services.
- **High-Performance Search**: Direct file-based search using `grep`, supporting Regex and massive files instantly. No database ingestion lag.
- **Archive Support**: Automatically discovers and searches rotated/archived logs (`.gz`, `.1`, etc.).
- **System Monitoring**: Real-time CPU, RAM, Disk, and Network metrics using WebSocket.
- **Process Manager**: View and manage running processes.
- **Modern UI**: Built with server-side Jet templates, TailwindCSS, and DaisyUI.

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
├── cmd/
│   └── monitor-agent/    # Main entry point (main.go)
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

## ⚙️ Configuration (`config.yaml`)

The system uses a hierarchical log definition:

```yaml
apps:
  - name: "App Name" # Grouping in Sidebar
    service_name: "service" # Internal ID
    logs:
      - name: "Access Log" # Display Name
        path: "/var/log/nginx/access.log" # File or Directory path
```

- If `path` is a **Directory**: All `.log`, `.txt`, `.gz` files in it are discovered.
- If `path` is a **File**: The file and its rotated siblings (e.g., `app.log.1`, `app.log.gz`) are discovered.

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
go run ./cmd/monitor-agent/main.go
# Server listens on :9000
# Default Auth: admin / admin (created on first run)
```

### Debugging Notes

- **"No Logs Found"**: Check permissions. The agent must have read access to the target log files.
- **Search Issues**: The search relies on system `grep`. Ensure `grep` is in `$PATH`.
- **Database**: `monitor.db` (SQLite) stores _only_ metrics history, alerts, and users. Deleting it resets auth and graphs, but not logs.

## 🔍 Log Search Logic (`internal/logs/logs.go`)

The search does **not** read files into memory completely.

1.  It constructs a `grep` command with the query.
2.  It streams the output (`StdoutPipe`).
3.  It parses lines into structs (`LogResult`).
4.  It applies a safety cap (e.g., 5000 lines) to prevent OOM on massive result sets.
5.  It sorts results by timestamp (newest first) before returning JSON.

## 🤝 Contributing

- **Frontend Changes**: Edit `views/*.jet.html`. No build step required (just reload).
- **Backend Changes**: Restart server (or use Air).

```

```

# Changelog

All notable changes to Logmojo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive configuration cleanup using Viper consistently
- Version information display in footer and settings
- Proper JWT-based authentication middleware
- Alert system with email and webhook notifications
- Real-time log streaming via WebSocket
- System metrics monitoring (CPU, RAM, Disk, Network)
- Process management with kill functionality
- Service management (systemd integration)
- User management CLI commands
- Compressed log file support (.gz, .bz2, .xz, .lz4)

### Changed
- Unified configuration management through Viper
- Removed direct environment variable access in favor of config structure
- Improved version initialization flow
- Enhanced security with HTTP-only cookies and bcrypt password hashing

### Fixed
- Version not displaying in UI components
- Configuration inconsistencies between modules
- Authentication middleware proper initialization

## [1.0.0] - 2024-01-XX

### Added
- Initial release of Logmojo
- High-performance log search using grep/zgrep
- Single binary deployment
- SQLite database for metadata storage
- Web-based user interface with TailwindCSS
- Real-time metrics collection
- Alert rule management
- Multi-format timestamp parsing
- Zero-database log ingestion (direct file search)

### Security
- JWT token-based authentication
- Bcrypt password hashing
- HTTP-only cookie implementation
- Session timeout management

---

## Release Notes

### Version Numbering
- **Major**: Breaking changes or significant new features
- **Minor**: New features, backward compatible
- **Patch**: Bug fixes, backward compatible

### Categories
- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security improvements
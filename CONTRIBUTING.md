# Contributing to Logmojo

First off, thank you for considering contributing to Logmojo! 🎉

It's people like you that make Logmojo such a great tool for the DevOps community.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Commit Messages](#commit-messages)

## 📜 Code of Conduct

This project and everyone participating in it is governed by respect and professionalism. By participating, you are expected to uphold this standard. Please report unacceptable behavior to the project maintainers.

## 🤝 How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (code snippets, log outputs, screenshots)
- **Describe the behavior you observed** and what you expected
- **Include your environment details** (OS, Go version, Logmojo version)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful**
- **List any similar features** in other tools if applicable

### Pull Requests

We actively welcome your pull requests! Here are some areas where contributions are especially welcome:

- 🔍 Additional log format parsers
- 🚨 New alert rule types
- 🎨 UI/UX improvements
- ⚡ Performance optimizations
- 📚 Documentation improvements
- 🧪 Test coverage improvements
- 🐛 Bug fixes

## 🛠️ Development Setup

### Prerequisites

- **Go 1.24+** installed
- **Git** for version control
- **Air** (optional, for live reload): `go install github.com/cosmtrek/air@latest`

### Setup Steps

1. **Fork the repository** on GitHub

2. **Clone your fork**:

   ```bash
   git clone https://github.com/YOUR_USERNAME/logmojo.git
   cd logmojo
   ```

3. **Add upstream remote**:

   ```bash
   git remote add upstream https://github.com/saiarlen/logmojo.git
   ```

4. **Install dependencies**:

   ```bash
   go mod download
   ```

5. **Copy environment file**:

   ```bash
   cp .env.example .env
   ```

6. **Run the application**:

   ```bash
   # With live reload
   air

   # Or standard run
   go run main.go
   ```

7. **Access the application**: `http://localhost:7005` (admin/admin)

### Project Structure

```
.
├── main.go                 # Application entry point
├── internal/
│   ├── api/               # HTTP routes & handlers
│   ├── logs/              # Log search engine
│   ├── metrics/           # System metrics collection
│   ├── alerts/            # Alert management
│   ├── auth/              # Authentication & JWT
│   └── ...
├── views/                 # Jet HTML templates
├── public/                # Static assets (CSS/JS)
└── tests/                 # Test files
```

## 🔄 Pull Request Process

1. **Create a new branch** for your feature/fix:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards

3. **Test your changes** thoroughly:

   ```bash
   go test ./...
   ```

4. **Commit your changes** with clear commit messages

5. **Push to your fork**:

   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request** against the `main` branch

7. **Ensure your PR**:
   - Has a clear title and description
   - References any related issues
   - Passes all CI checks
   - Includes tests for new functionality
   - Updates documentation if needed

### PR Review Process

- Maintainers will review your PR within 2-3 business days
- Address any requested changes
- Once approved, a maintainer will merge your PR

## 💻 Coding Standards

### Go Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Run `go vet` to catch common mistakes
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

### Example:

```go
// SearchLogs searches log files using grep and returns matching results.
// It supports compressed files (.gz, .bz2, .xz, .lz4) and applies
// configurable limits to prevent resource exhaustion.
func SearchLogs(query string, filters LogFilters) ([]LogResult, error) {
    // Implementation
}
```

### Frontend Code

- Use vanilla JavaScript (no frameworks unless necessary)
- Follow consistent naming conventions
- Keep functions small and focused
- Add comments for complex UI logic

## 📝 Commit Messages

Write clear, concise commit messages following this format:

```
<type>: <subject>

<body (optional)>

<footer (optional)>
```

### Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples:

```
feat: add support for JSON log format parsing

Implements a new parser for JSON-formatted logs with automatic
field detection and timestamp extraction.

Closes #123
```

```
fix: resolve memory leak in WebSocket connections

Properly close WebSocket connections when clients disconnect
to prevent goroutine leaks.
```

## 🧪 Testing

- Write tests for new features
- Ensure existing tests pass: `go test ./...`
- Aim for meaningful test coverage
- Include both unit tests and integration tests where appropriate

## 📚 Documentation

- Update README.md if you change functionality
- Add inline code comments for complex logic
- Update API documentation if you modify endpoints
- Include examples in your documentation

## 🎯 Areas for Contribution

We especially welcome contributions in these areas:

### High Priority

- Docker and Kubernetes deployment support
- Multi-server log aggregation
- Additional authentication providers (LDAP, OAuth)
- Plugin system for custom log parsers

### Medium Priority

- Enhanced alert rule conditions
- Log export formats (CSV, JSON)
- Dashboard customization
- Mobile app support

### Always Welcome

- Bug fixes
- Performance improvements
- Documentation improvements
- Test coverage
- UI/UX enhancements

## ❓ Questions?

- 💬 Open a [GitHub Discussion](https://github.com/saiarlen/logmojo/discussions)
- 🐛 Report bugs via [GitHub Issues](https://github.com/saiarlen/logmojo/issues)
- 📧 Contact maintainers through GitHub

## 🙏 Thank You!

Your contributions make Logmojo better for everyone. We appreciate your time and effort! ⭐

---

**Happy Coding!** 🚀

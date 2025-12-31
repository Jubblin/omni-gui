# Talos Omni GUI

A cross-platform desktop GUI application for managing and monitoring Sidero Omni resources, providing an intuitive interface to explore clusters, machines, machine sets, and related infrastructure.

## Overview

Talos Omni GUI is a Go-based desktop application built with Fyne that provides a graphical interface to the Sidero Omni management platform. It allows you to browse, inspect, and interact with Omni resources through an organized tree hierarchy with detailed resource views.

### Key Features

- **Tree-Based Navigation**: Hierarchical view of all Omni resources (clusters, machines, machine sets, etc.)
- **Resource Details**: Detailed JSON view of any selected resource
- **Interactive Actions**: Quick access to resource-specific actions (status, metrics, configurations, etc.)
- **Cross-Platform**: Runs on Windows, macOS, and Linux
- **Multi-Language Support**: Internationalization support for multiple languages
- **Real-Time Updates**: Refresh tree data to see latest resource state
- **Settings Management**: Configure Omni endpoint and authentication through a settings window
- **Structured Logging**: All logs output as valid JSON to both console and log file (`omni-api.log` in execution directory)

## Requirements

### Runtime Requirements

- **Go**: Version 1.25.5 or later
- **Sidero Omni**: Access to a running Omni instance
- **Network Access**: Ability to connect to the Omni endpoint

### Authentication

The application requires authentication to connect to Omni. You can use one of the following methods:

- **OIDC Authentication**: Using OpenID Connect (OIDC) with Client Credentials flow
- **Service Account Authentication**: Using a service account key

## Installation

### From Source

1. Clone the repository:

  ```bash
  git clone <repository-url>
  cd omni-api
  ```

2. Install dependencies:

  ```bash
  go mod download
  ```

3. Build the binary:

  ```bash
  make build-gui
  ```

Or build manually:

  ```bash
  go build -o omni-gui .
  ```

### Using Make

The project includes a Makefile with common commands:

```bash
make build-gui    # Build the GUI application
make run-gui      # Run the application
make test         # Run tests
make tidy         # Tidy Go modules
make clean        # Clean build artifacts
make version      # Get current version
make version-patch  # Increment patch version (0.0.1 → 0.0.2)
make version-minor  # Increment minor version (0.0.1 → 0.1.0)
make version-major  # Increment major version (0.0.1 → 1.0.0)
```

## Configuration

The application is configured using environment variables, the Settings window, or command-line flags. The Settings window provides a user-friendly interface for all configuration options and is the recommended method for most users.

### Settings Window

Access the Settings window via **☰ Menu → Settings**. The Settings window provides:

- **Application Settings Tab**:
  - Language selection (applies immediately)
  - Empty Label Filtering (applies immediately, no restart required)
  - Ignore Environment Variables (applies immediately, no restart required)
  - gRPC Debug Level (applies immediately, no restart required)

- **Authentication Settings Tab**:
  - Omni Endpoint configuration
  - Authentication method selection (Service Account or OIDC)
  - Authentication credentials

Settings are saved to `~/.omni-api/settings.json` and persist across application restarts. All application settings (Language, Empty Label Filtering, Ignore Environment Variables, gRPC Debug Level) apply immediately without restart. Only authentication and endpoint changes require a restart.

### Command-Line Flags

Command-line flags are available for compatibility but take precedence over settings file values:

- `--show-empty-labels`: Show nodes with empty labels in the tree (disable filtering)
- `--ignore-env`: Ignore environment variables and use only settings file
- `--grpc-debug`: gRPC debug level (0=disabled, 1=query dumps, 2=query+response, 3=full debugging with UI dumps)

**Note**: Command-line flag values are not persisted to the settings file. Use the Settings window to make permanent changes.

### Required Environment Variables

- **`OMNI_ENDPOINT`**: The Omni API endpoint URL (e.g., `https://omni.example.com` or `http://localhost:8080`)

### Authentication Environment Variables

Choose one of the following authentication methods:

**OIDC Authentication:**

- `OMNI_OIDC_ISSUER_URL`: OIDC provider issuer URL (required)
- `OMNI_OIDC_CLIENT_ID`: OIDC client ID (required)
- `OMNI_OIDC_CLIENT_SECRET`: OIDC client secret (required)

**Service Account Authentication:**

- `OMNI_SERVICE_ACCOUNT` or `OMNI_SERVICE_ACCOUNT_KEY`: Service account key (base64 encoded)

### Example Configuration

**OIDC Authentication:**
```bash
export OMNI_ENDPOINT="https://omni.example.com"
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
```

**Service Account Authentication:**
```bash
export OMNI_ENDPOINT="https://omni.example.com"
export OMNI_SERVICE_ACCOUNT_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Usage

### Starting the Application

1. **Configure Settings** (recommended method):
   - Run the application: `./omni-gui` or `make run-gui`
   - Open Settings via **☰ Menu → Settings**
   - Configure Omni endpoint and authentication credentials
   - Save settings (most apply immediately, authentication changes require restart)

2. **Or use environment variables** (alternative method):

**Using OIDC:**
```bash
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
./omni-gui
```

**Using Service Account:**
```bash
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_SERVICE_ACCOUNT_KEY="your-service-account-key"
./omni-gui
```

3. **Or use command-line flags** (for compatibility):
```bash
./omni-gui --show-empty-labels --ignore-env
```

**Note**: Settings configured through the Settings window are persisted to `~/.omni-api/settings.json` and persist across restarts. Environment variables and command-line flags take precedence but are not persisted.

### Using the Application

1. **Tree Navigation**: Use the left panel to browse resources organized by type (Clusters, Machines, Machine Sets)
2. **Resource Details**: Click on any resource to see its details in the right panel
3. **Actions**: Use action buttons to access resource-specific operations (status, metrics, configurations, etc.)
4. **Refresh**: Use the refresh button (circular arrow icon) in the top-left corner or the menu (☰) to refresh the tree data
5. **Settings**: Use the settings button (cog icon) in the top-left corner or access through the menu to configure endpoint and authentication

For detailed information about the tree hierarchy and navigation, see [docs/FYNE_GUI_TREE_HIERARCHY.md](docs/FYNE_GUI_TREE_HIERARCHY.md).

## Project Structure

``` shell
omni-api/
├── main.go                    # Application entry point
├── types.go                   # Type definitions
├── constants.go               # Constants
├── tree.go                    # Tree management
├── resource_labels.go         # Resource label formatting
├── resource_enrichment.go     # Resource enrichment
├── ui_components.go           # UI component creation
├── ui_layout.go               # Layout assembly
├── ui_settings.go             # Settings window
├── ui_detail.go               # Detail pane management
├── Makefile                   # Build and development commands
├── go.mod                     # Go module dependencies
├── internal/
│   ├── client/                # Omni client wrapper
│   ├── i18n/                  # Internationalization
│   └── resource/              # Resource conversion utilities
└── docs/                      # Documentation
    └── FYNE_GUI_TREE_HIERARCHY.md
```

## Development

### Testing

Run tests:

```bash
make test
```

Or:

```bash
go test ./...
```

### Building

Build the GUI application:

```bash
make build-gui
```

## Versioning

The project uses [Semantic Versioning](https://semver.org/) (SemVer).

### Manual Version Management

You can manually manage versions using the Makefile:

```bash
# Get current version
make version

# Increment patch version (0.0.1 → 0.0.2)
make version-patch

# Increment minor version (0.0.1 → 0.1.0)
make version-minor

# Increment major version (0.0.1 → 1.0.0)
make version-major
```

Or use the version script directly:

```bash
./scripts/version.sh get          # Get current version
./scripts/version.sh patch        # Increment patch
./scripts/version.sh minor        # Increment minor
./scripts/version.sh major        # Increment major
./scripts/version.sh set 1.2.3   # Set specific version
```

### Version File

The current version is stored in the `VERSION` file in the repository root.

## Documentation

Additional documentation is available in the `docs/` directory:

- **[docs/FYNE_GUI_TREE_HIERARCHY.md](docs/FYNE_GUI_TREE_HIERARCHY.md)** - Fyne GUI tree structure hierarchy and navigation guide
- **[docs/OIDC_INTEGRATION.md](docs/OIDC_INTEGRATION.md)** - OIDC authentication integration guide
- **[REFACTORING_SUMMARY.md](REFACTORING_SUMMARY.md)** - Code refactoring summary and structure

## License

MIT - See LICENSE file for details.

## Contributing

Contributions are welcome! Please ensure:

1. Code follows Go best practices
2. Tests are included for new features
3. Code is properly formatted (`go fmt`)
4. All tests pass (`make test`)

## Support

For issues and questions:

- Review application logs for error messages (logs are written to `omni-api.log` in the execution directory)
- All logs are output as valid JSON for easy parsing and analysis
- Verify environment variable configuration
- Check the Settings window for configuration options

## Logging

The application logs all output as structured JSON to both the console (stderr) and a log file:

- **Log File Location**: `omni-api.log` in the execution directory (where you run the application)
- **Log Format**: Valid JSON with structured fields (time, level, msg, source, etc.)
- **Startup Banner**: Each execution begins with a startup banner containing version, timestamp, and configuration
- **Log Levels**: Info, Debug, Warn, Error

Example log entry:
```json
{"time":"2024-01-15T10:30:45Z","level":"INFO","msg":"Application Startup Banner","application":"Omni API GUI Application","version":"1.0.0","started_at":"2024-01-15T10:30:45Z","ignore_env":false,"show_empty_labels":false,"grpc_debug_level":0}
```

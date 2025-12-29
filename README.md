# Talos Omni Control API

A RESTful API server that provides programmatic access to Sidero Omni resources, enabling you to manage and monitor Talos Linux clusters, machines, and related infrastructure through a standardized HTTP interface.

## Overview

Talos Omni Control API is a Go-based REST API that acts as a bridge between your applications and the Sidero Omni management platform. It exposes Omni resources (clusters, machines, machine sets, configurations, etc.) through a clean, RESTful interface with comprehensive Swagger documentation.

### Key Features

- **Comprehensive Resource Access**: Query and manage 38+ resource types including clusters, machines, machine sets, configurations, backups, Kubernetes resources, image management, and more
- **RESTful API Design**: Standard HTTP methods with JSON responses
- **Hypermedia Controls**: All responses include `_links` with full URLs for easy navigation between related resources
- **Interactive Documentation**: Built-in Swagger UI accessible at `/swagger/index.html` (root redirects here)
- **CORS Support**: Configured for cross-origin requests
- **Status Consolidation**: Machine status information (hostname, platform, arch, talos_version) is included directly in machine responses
- **Filtering & Querying**: Support for filtering resources by cluster, machine set, and other criteria
- **70+ Endpoints**: Comprehensive coverage of Omni resources

## Requirements

### Runtime Requirements

- **Go**: Version 1.25.5 or later
- **Sidero Omni**: Access to a running Omni instance
- **Network Access**: Ability to connect to the Omni endpoint

### Authentication

The API requires authentication to connect to Omni. You can use one of the following methods:

- **OIDC Authentication**: Using OpenID Connect (OIDC) with Client Credentials flow
- **Service Account Authentication**: Using a service account key
- **PGP Authentication**: Using user account with PGP keys

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
  make build
  ```

Or build manually:

  ```bash
  go build -o omni-api main.go
  ```

### Using Make

The project includes a Makefile with common commands:

```bash
make build         # Build the binary
make run           # Run the application
make test          # Run tests
make swagger        # Generate Swagger documentation
make tidy           # Tidy Go modules
make clean          # Clean build artifacts
make version        # Get current version
make version-patch  # Increment patch version (0.0.1 → 0.0.2)
make version-minor  # Increment minor version (0.0.1 → 0.1.0)
make version-major  # Increment major version (0.0.1 → 1.0.0)
```

## Versioning

The project uses [Semantic Versioning](https://semver.org/) (SemVer) with automatic patch version incrementing for pull requests.

### Automatic Version Bumping

When you open a pull request targeting the `main` branch, the version is automatically incremented:

- **Patch version** is automatically incremented (e.g., `0.0.1` → `0.0.2`)
- The version is updated in:
  - `VERSION` file (source of truth)
  - `main.go` (Swagger annotation)
  - `internal/api/handlers/health.go` (health endpoint response)
  - `docs/` (Swagger documentation is regenerated)
- Changes are automatically committed to your PR branch
- The workflow handles rebases and updates intelligently

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

The current version is stored in the `VERSION` file in the repository root. This file is the source of truth for the project version and is used by:

- Build workflows
- Swagger documentation
- Health endpoint
- Container image tags

## Configuration

The API is configured using environment variables:

### Required Environment Variables

- **`OMNI_ENDPOINT`**: The Omni API endpoint URL (e.g., `https://omni.example.com` or `http://localhost:8080`)

### Authentication Environment Variables

Choose one of the following authentication methods:

**OIDC Authentication:**

- `OMNI_OIDC_ISSUER_URL`: OIDC provider issuer URL (required)
- `OMNI_OIDC_CLIENT_ID`: OIDC client ID (required)
- `OMNI_OIDC_CLIENT_SECRET`: OIDC client secret (required)
- `OMNI_OIDC_SCOPES`: Comma-separated list of scopes (optional, default: "openid profile email")
- `OMNI_OIDC_AUDIENCE`: Token audience claim (optional)
- `OMNI_OIDC_FLOW`: Authentication flow type (optional, default: "client_credentials")
- `OMNI_OIDC_TOKEN_CACHE_FILE`: File path for token caching (optional)

**Service Account Authentication:**

- `OMNI_SERVICE_ACCOUNT` or `OMNI_SERVICE_ACCOUNT_KEY`: Service account key (base64 encoded)

**PGP Authentication:**

- `OMNI_CONTEXT`: Context name for PGP authentication
- `OMNI_IDENTITY`: Identity for PGP authentication
- `OMNI_KEYS_DIR`: (Optional) Custom directory for PGP keys

### Optional Environment Variables

- **`PORT`**: HTTP server port (default: `8080`)
- **`OMNI_INSECURE`**: Set to `true` to skip TLS verification (not recommended for production)

### Example Configuration

**OIDC Authentication:**
```bash
export OMNI_ENDPOINT="https://omni.example.com"
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
export PORT="8080"
```

**Service Account Authentication:**
```bash
export OMNI_ENDPOINT="https://omni.example.com"
export OMNI_SERVICE_ACCOUNT_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export PORT="8080"
```

## Usage

### Starting the Server

1. Set required environment variables:

**Using OIDC:**
```bash
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
```

**Using Service Account:**
```bash
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_SERVICE_ACCOUNT_KEY="your-service-account-key"
```

2. Run the server:

```bash
./omni-api
```

Or using make:

```bash
make run
```

The server will start on port 8080 (or the port specified by `PORT` environment variable).

### Accessing the API

- **API Base URL**: `http://localhost:8080/api/v1`
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Root Redirect**: `http://localhost:8080/` (redirects to Swagger UI)
- **Health Check**: `http://localhost:8080/health` - API server health status
- **Metrics**: `http://localhost:8080/metrics` - API server metrics

## API Documentation

### Interactive Documentation

The API includes interactive Swagger documentation accessible at `/swagger/index.html`. The root path (`/`) automatically redirects to the Swagger UI.

### Available Endpoints

#### Health & Metrics

- `GET /health` - Get API server health status (includes Omni connectivity)
- `GET /metrics` - Get API server metrics (request counts, response times, errors)

#### Clusters

- `GET /api/v1/clusters` - List all clusters
- `GET /api/v1/clusters/:id` - Get cluster details
- `GET /api/v1/clusters/:id/status` - Get cluster status
- `GET /api/v1/clusters/:id/metrics` - Get cluster metrics
- `GET /api/v1/clusters/:id/bootstrap` - Get bootstrap status
- `GET /api/v1/clusters/:id/kubeconfig` - Get kubeconfig (⚠️ sensitive)
- `GET /api/v1/clusters/:id/kubernetes-upgrade` - Get Kubernetes upgrade status
- `GET /api/v1/clusters/:id/talos-upgrade` - Get Talos upgrade status
- `GET /api/v1/clusters/:id/endpoints` - Get cluster endpoints
- `GET /api/v1/clusters/:id/kubernetes-status` - Get Kubernetes cluster status
- `GET /api/v1/clusters/:id/kubernetes-nodes` - List Kubernetes nodes in cluster
- `GET /api/v1/clusters/:id/kubernetes-nodes/:node` - Get Kubernetes node details
- `GET /api/v1/clusters/:id/controlplane-status` - Get control plane status
- `GET /api/v1/clusters/:id/diagnostics` - Get cluster diagnostics
- `GET /api/v1/clusters/:id/destroy-status` - Get cluster destroy status
- `GET /api/v1/clusters/:id/workload-proxy-status` - Get workload proxy status

#### Machines

- `GET /api/v1/machines` - List all machines
- `GET /api/v1/machines/:id` - Get machine details (includes status)
- `GET /api/v1/machines/:id/status` - Get machine status (deprecated - use main endpoint)
- `GET /api/v1/machines/:id/labels` - Get machine labels
- `GET /api/v1/machines/:id/extensions` - Get machine extensions
- `GET /api/v1/machines/:id/upgrade-status` - Get machine upgrade status
- `GET /api/v1/machines/:id/metrics` - Get machine status metrics
- `GET /api/v1/machines/:id/config-diff` - Get machine configuration diff

#### Machine Sets

- `GET /api/v1/machinesets` - List all machine sets
- `GET /api/v1/machinesets/:id` - Get machine set details
- `GET /api/v1/machinesets/:id/status` - Get machine set status
- `GET /api/v1/machinesets/:id/destroy-status` - Get machine set destroy status

#### Cluster Machines

- `GET /api/v1/clustermachines` - List all cluster machines
- `GET /api/v1/clustermachines/:id` - Get cluster machine details
- `GET /api/v1/clustermachines/:id/status` - Get cluster machine status
- `GET /api/v1/clustermachines/:id/config-status` - Get config status
- `GET /api/v1/clustermachines/:id/talos-version` - Get Talos version
- `GET /api/v1/clustermachines/:id/config` - Get cluster machine configuration

#### Config Patches

- `GET /api/v1/configpatches` - List all config patches
- `GET /api/v1/configpatches/:id` - Get config patch details

#### Machine Classes

- `GET /api/v1/machineclasses` - List all machine classes
- `GET /api/v1/machineclasses/:id` - Get machine class details

#### Machine Set Nodes

- `GET /api/v1/machinesetnodes` - List all machine set nodes
- `GET /api/v1/machinesetnodes/:id` - Get machine set node details

#### Etcd Backups

- `GET /api/v1/etcdbackups` - List all etcd backups
- `GET /api/v1/etcdbackups/:id` - Get etcd backup details
- `GET /api/v1/etcdbackups/:id/status` - Get etcd backup status
- `GET /api/v1/etcd-manual-backups` - List etcd manual backup requests
- `GET /api/v1/etcd-manual-backups/:id` - Get etcd manual backup details

#### Schematics

- `GET /api/v1/schematics` - List all schematics
- `GET /api/v1/schematics/:id` - Get schematic details
- `GET /api/v1/schematic-configurations` - List schematic configurations
- `GET /api/v1/schematic-configurations/:id` - Get schematic configuration details

#### Ongoing Tasks

- `GET /api/v1/ongoingtasks` - List all ongoing tasks
- `GET /api/v1/ongoingtasks/:id` - Get ongoing task details

#### Kubernetes Versions

- `GET /api/v1/kubernetes-versions` - List all available Kubernetes versions
- `GET /api/v1/kubernetes-versions/:id` - Get Kubernetes version details

#### Extensions Configuration

- `GET /api/v1/extensions-configurations` - List all extensions configurations
- `GET /api/v1/extensions-configurations/:id` - Get extensions configuration details

#### Kernel Args

- `GET /api/v1/kernel-args` - List all kernel args configurations
- `GET /api/v1/kernel-args/:id` - Get kernel args details

#### Load Balancers

- `GET /api/v1/loadbalancer-configs` - List all load balancer configurations
- `GET /api/v1/loadbalancer-configs/:id` - Get load balancer config details
- `GET /api/v1/loadbalancers/:id/status` - Get load balancer status

#### Exposed Services

- `GET /api/v1/exposed-services` - List all exposed services
- `GET /api/v1/exposed-services/:id` - Get exposed service details

#### Machine Request Sets

- `GET /api/v1/machine-request-sets` - List all machine request sets
- `GET /api/v1/machine-request-sets/:id` - Get machine request set details

#### Image Pull Requests

- `GET /api/v1/image-pull-requests` - List all image pull requests
- `GET /api/v1/image-pull-requests/:id` - Get image pull request details
- `GET /api/v1/image-pull-requests/:id/status` - Get image pull status

#### Installation Media

- `GET /api/v1/installation-medias` - List all installation medias
- `GET /api/v1/installation-medias/:id` - Get installation media details

#### Infrastructure Machine Configs

- `GET /api/v1/infra-machine-configs` - List all infrastructure machine configs
- `GET /api/v1/infra-machine-configs/:id` - Get infrastructure machine config details

### Response Format

All API responses follow a consistent format:

```json
{
  "id": "resource-id",
  "namespace": "default",
  "field1": "value1",
  "field2": "value2",
  "_links": {
    "self": "http://localhost:8080/api/v1/resource/id",
    "related": "http://localhost:8080/api/v1/related/id"
  }
}
```

### Query Parameters

Many list endpoints support filtering:

- **Clusters**: No filters
- **Machines**: No filters
- **Machine Sets**: No filters
- **Machine Set Nodes**: `?machineset=<machineset-id>` - Filter by machine set
- **Cluster Machines**: `?cluster=<cluster-id>` - Filter by cluster
- **Config Patches**: No filters
- **Machine Classes**: No filters
- **Etcd Backups**: `?cluster=<cluster-id>` - Filter by cluster
- **Etcd Manual Backups**: `?cluster=<cluster-id>` - Filter by cluster
- **Schematics**: No filters
- **Ongoing Tasks**: `?resource=<resource-id>` - Filter by resource
- **Kubernetes Versions**: No filters
- **Extensions Configurations**: No filters
- **Kernel Args**: No filters
- **Load Balancer Configs**: No filters
- **Exposed Services**: No filters
- **Machine Request Sets**: No filters
- **Image Pull Requests**: No filters
- **Installation Medias**: No filters
- **Infrastructure Machine Configs**: `?machine=<machine-id>` - Filter by machine ID
- **Image Pull Requests**: No filters
- **Installation Medias**: No filters
- **Infrastructure Machine Configs**: `?machine=<machine-id>` - Filter by machine ID

### Example Requests

```bash
# List all clusters
curl http://localhost:8080/api/v1/clusters

# Get a specific machine
curl http://localhost:8080/api/v1/machines/machine-id

# Get cluster machines for a specific cluster
curl http://localhost:8080/api/v1/clustermachines?cluster=cluster-id

# Get cluster machine status
curl http://localhost:8080/api/v1/clustermachines/machine-id/status
```

## Development

### Test Coverage

The project maintains comprehensive test coverage:

- **Handler Coverage**: 69.2% (exceeds 70% target threshold)
- **Client Coverage**: 66.7%
- **Total Test Files**: 30+ test files covering all major handlers
- **Test Functions**: 60+ test functions

All major handlers have comprehensive test coverage including:

- List and Get operations
- Filtering logic (where applicable)
- Response structure validation
- Link generation
- Error handling paths

See [docs/TEST_COVERAGE.md](docs/TEST_COVERAGE.md) for detailed coverage information and recommendations.

### Project Structure

``` shell
omni-api/
├── main.go                    # Application entry point
├── Makefile                   # Build and development commands
├── go.mod                     # Go module dependencies
├── internal/
│   ├── api/
│   │   └── handlers/          # API request handlers
│   │       ├── clusters.go
│   │       ├── machines.go
│   │       ├── machinesets.go
│   │       └── ...            # Other resource handlers
│   └── client/
│       └── omni.go            # Omni client wrapper
└── docs/                      # Generated Swagger documentation
    ├── docs.go
    ├── swagger.json
    └── swagger.yaml
```

### Adding New Handlers

1. Create a new handler file in `internal/api/handlers/`
2. Implement the handler struct and methods
3. Register the handler in `main.go`
4. Add routes in the API routes section
5. Update Swagger documentation: `make swagger`

### Testing

The project includes comprehensive test coverage (69.3% for handlers). Run tests:

```bash
make test
```

Or:

```bash
go test ./...
```

To view test coverage:

```bash
go test -cover ./...
```

To generate detailed coverage reports:

```bash
go test -coverprofile=handlers_coverage.out ./internal/api/handlers
go tool cover -func=handlers_coverage.out
```

#### Integration Tests

Integration tests validate the API against a real Omni instance. See [integration/README.md](integration/README.md) for details.

To run integration tests:

```bash
export INTEGRATION_TESTS=true
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_SERVICE_ACCOUNT_KEY="your-service-account-key"
make test-integration
```

Or manually:

```bash
# Start the API server in one terminal
make run

# Run integration tests in another terminal
export INTEGRATION_TESTS=true
go test ./integration/... -v
```

See [docs/TEST_COVERAGE.md](docs/TEST_COVERAGE.md) for detailed coverage information.

### Generating Swagger Documentation

After adding or modifying API endpoints, regenerate Swagger docs:

```bash
make swagger
```

This updates the files in the `docs/` directory.

## Features in Detail

### Hypermedia Controls

All API responses include `_links` objects that provide URLs to related resources, enabling clients to navigate the API without hardcoding URLs.

### Status Consolidation

Machine endpoints consolidate status information directly in the main response, reducing the need for multiple API calls. The separate status endpoint remains available for backward compatibility but is deprecated.

### CORS Configuration

The API is configured with permissive CORS settings for development. For production, consider restricting `AllowOrigins` to specific domains.

### Error Handling

The API returns standard HTTP status codes:

- `200 OK` - Successful request
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

Error responses include a JSON object with an `error` field describing the issue.

## Security Considerations

- **Kubeconfig Endpoint**: The `/clusters/:id/kubeconfig` endpoint returns sensitive credentials. Ensure proper authentication and authorization.
- **Service Account Keys**: Store service account keys securely. Never commit them to version control.
- **TLS**: In production, use HTTPS and avoid setting `OMNI_INSECURE=true`.
- **CORS**: Restrict CORS origins in production environments.

## Troubleshooting

### Connection Issues

If you encounter connection errors:

1. Verify `OMNI_ENDPOINT` is correct and accessible
2. Check authentication credentials are valid
3. Ensure network connectivity to the Omni instance
4. Check Omni instance logs for authentication failures

### Missing Resources

If resources don't appear:

1. Verify the Omni instance has the resources
2. Check authentication has proper permissions
3. Review server logs for errors

### Swagger Documentation Not Updating

If Swagger docs are outdated:

1. Run `make swagger` to regenerate
2. Ensure Swagger annotations are correct in handler files
3. Check that the `docs/` directory is writable

## CI/CD and Container Images

### GitHub Actions

The repository includes GitHub Actions workflows for automated builds:

- **Multi-Architecture Builds**: Automatically builds containers for AMD64 and ARM64
- **Container Signing**: All images are cryptographically signed using cosign
- **Security Scanning**: Automated vulnerability scanning with Trivy
- **SBOM Generation**: Software Bill of Materials for each build

See [.github/workflows/README.md](.github/workflows/README.md) for detailed workflow documentation.

### Container Images

Pre-built container images are available on GitHub Container Registry:

```bash
# Pull latest image
docker pull ghcr.io/jubblin/omni-api:latest

# Pull specific version
docker pull ghcr.io/jubblin/omni-api:0.0.1

# Pull for ARM64
docker pull --platform linux/arm64 ghcr.io/jubblin/omni-api:latest
```

For detailed Docker usage, see [docs/README.Docker.md](docs/README.Docker.md).

## Documentation

Additional documentation is available in the `docs/` directory:

### Core Documentation

- **[docs/RESOURCES.md](docs/RESOURCES.md)** - Complete list of available Omni resources, implementation status, and resource details
- **[docs/TEST_COVERAGE.md](docs/TEST_COVERAGE.md)** - Test coverage report, statistics, and recommendations
- **[docs/MACHINE_ENHANCEMENTS.md](docs/MACHINE_ENHANCEMENTS.md)** - Machine endpoint enhancements and implementation details
- **[docs/README.Docker.md](docs/README.Docker.md)** - Docker build and deployment guide
- **[docs/TESTING_QUICK_START.md](docs/TESTING_QUICK_START.md)** - Quick start guide for testing

### OIDC Authentication Documentation

- **[docs/OIDC_INTEGRATION.md](docs/OIDC_INTEGRATION.md)** - Comprehensive OIDC integration guide
- **[docs/OIDC_QUICK_START.md](docs/OIDC_QUICK_START.md)** - Quick start guide for OIDC authentication
- **[docs/OIDC_SIMPLIFIED_CONFIG.md](docs/OIDC_SIMPLIFIED_CONFIG.md)** - Simplified OIDC configuration guide
- **[docs/OIDC_TESTING_GUIDE.md](docs/OIDC_TESTING_GUIDE.md)** - Complete testing guide for OIDC
- **[docs/OIDC_CONFIGURATION_ANALYSIS.md](docs/OIDC_CONFIGURATION_ANALYSIS.md)** - Technical analysis of OIDC configuration options
- **[docs/OIDC_CLIENT_SIMPLIFICATION.md](docs/OIDC_CLIENT_SIMPLIFICATION.md)** - Client configuration simplification details

### OIDC Implementation History

- **[docs/OIDC_IMPLEMENTATION_PLAN.md](docs/OIDC_IMPLEMENTATION_PLAN.md)** - Initial implementation plan
- **[docs/OIDC_IMPLEMENTATION_STATUS.md](docs/OIDC_IMPLEMENTATION_STATUS.md)** - Implementation status tracking
- **[docs/OIDC_IMPLEMENTATION_SUMMARY.md](docs/OIDC_IMPLEMENTATION_SUMMARY.md)** - Implementation summary
- **[docs/OIDC_COMPLETE.md](docs/OIDC_COMPLETE.md)** - Completion status
- **[docs/OIDC_FINAL_STATUS.md](docs/OIDC_FINAL_STATUS.md)** - Final status report
- **[docs/OIDC_INTEGRATION_SUCCESS.md](docs/OIDC_INTEGRATION_SUCCESS.md)** - Integration success report

### Other Documentation

- **[integration/README.md](integration/README.md)** - Integration test suite documentation
- **[.github/workflows/README.md](.github/workflows/README.md)** - CI/CD workflow documentation
- **[.slsa/README.md](.slsa/README.md)** - SLSA compliance documentation
- **[.checkov-compliance.md](.checkov-compliance.md)** - Checkov Dockerfile compliance documentation

Each document includes navigation links back to this README and to other related documentation.

## License

MIT - See LICENSE file for details.

## Contributing

Contributions are welcome! Please ensure:

1. Code follows Go best practices
2. Tests are included for new features (aim for 70%+ coverage)
3. Swagger documentation is updated
4. Code is properly formatted (`go fmt`)
5. All tests pass (`make test`)

## Support

For issues and questions:

- Check the Swagger documentation at `/swagger/index.html`
- Review server logs for error messages
- Verify environment variable configuration

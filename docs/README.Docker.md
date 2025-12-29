# Docker Build and Deployment

This document describes how to build and run the Omni API using Docker.

## Quick Start

### Build the Docker Image

```bash
# Basic build
docker build -t omni-api:latest .

# Build with version information
docker build \
  --build-arg VERSION=0.0.1 \
  --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  --build-arg COMMIT_SHA=$(git rev-parse --short HEAD) \
  -t omni-api:0.0.1 .
```

### Run the Container

```bash
# Basic run
docker run -d \
  -p 8080:8080 \
  -e OMNI_ENDPOINT="https://your-omni-instance.com" \
  -e OMNI_SERVICE_ACCOUNT_KEY="your-key" \
  --name omni-api \
  omni-api:latest

# With all environment variables
docker run -d \
  -p 8080:8080 \
  -e OMNI_ENDPOINT="https://your-omni-instance.com" \
  -e OMNI_SERVICE_ACCOUNT_KEY="your-service-account-key" \
  -e PORT=8080 \
  -e OMNI_INSECURE=false \
  --name omni-api \
  omni-api:latest
```

### Using Docker Compose

```bash
# Create .env file with your configuration
cat > .env << EOF
OMNI_ENDPOINT=https://your-omni-instance.com
OMNI_SERVICE_ACCOUNT_KEY=your-service-account-key
PORT=8080
VERSION=0.0.1
EOF

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

## Dockerfile Variants

### Default: Distroless (Recommended)

The default `Dockerfile` uses Google's distroless image, which provides:
- **Minimal attack surface**: No shell, package manager, or unnecessary tools
- **Small size**: Very minimal base image
- **Security**: Only the application binary and essential runtime libraries

**Note**: Distroless images don't include a shell, so debugging requires copying the binary out or using a debug image.

### Alternative: Alpine Linux

Use `Dockerfile.alpine` if you need:
- Shell access for debugging
- Additional tools (wget, curl, etc.)
- Easier troubleshooting

```bash
docker build -f Dockerfile.alpine -t omni-api:alpine .
```

## Build Arguments

The Dockerfile accepts the following build arguments:

- `VERSION`: Application version (default: `0.0.1`)
- `BUILD_DATE`: Build timestamp (ISO 8601 format)
- `COMMIT_SHA`: Git commit hash

Example:

```bash
docker build \
  --build-arg VERSION=0.0.1 \
  --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  --build-arg COMMIT_SHA=$(git rev-parse --short HEAD) \
  -t omni-api:0.0.1 .
```

## Environment Variables

The container requires the following environment variables:

### Required

- `OMNI_ENDPOINT`: The Omni API endpoint URL

### Authentication (choose one)

**Service Account:**
- `OMNI_SERVICE_ACCOUNT` or `OMNI_SERVICE_ACCOUNT_KEY`

**PGP Authentication:**
- `OMNI_CONTEXT`
- `OMNI_IDENTITY`
- `OMNI_KEYS_DIR` (optional)

### Optional

- `PORT`: HTTP server port (default: `8080`)
- `OMNI_INSECURE`: Set to `true` to skip TLS verification (not recommended for production)

## Security Best Practices

The Dockerfile follows security best practices:

1. **Multi-stage build**: Reduces final image size
2. **Non-root user**: Runs as `nonroot` user (uid:gid = 65532:65532)
3. **Minimal base image**: Uses distroless for minimal attack surface
4. **Static binary**: CGO disabled, no C dependencies
5. **Read-only filesystem**: Can be enabled with `--read-only` flag
6. **Health checks**: Built-in health check endpoint

### Running with Additional Security

```bash
docker run -d \
  --read-only \
  --tmpfs /tmp \
  --security-opt no-new-privileges:true \
  --cap-drop ALL \
  -p 8080:8080 \
  -e OMNI_ENDPOINT="https://your-omni-instance.com" \
  -e OMNI_SERVICE_ACCOUNT_KEY="your-key" \
  omni-api:latest
```

## Health Checks

### Distroless Images

Distroless images don't include shell utilities (wget, curl, nc), so health checks must be done externally:

```bash
# External health check
curl http://localhost:8080/health

# Or use orchestrator health checks (Kubernetes, Docker Swarm)
```

### Alpine Images

Alpine images include health check support:

```bash
# Using docker-compose with Alpine
docker-compose -f docker-compose.alpine.yml up -d

# Check container health
docker ps

# Inspect health status
docker inspect --format='{{.State.Health.Status}}' omni-api
```

## Troubleshooting

### View Logs

```bash
# Container logs
docker logs omni-api

# Follow logs
docker logs -f omni-api

# Last 100 lines
docker logs --tail 100 omni-api
```

### Debugging Distroless Container

Since distroless containers don't have a shell, debugging requires:

1. **Copy binary out**:
   ```bash
   docker cp omni-api:/usr/local/bin/omni-api ./omni-api-debug
   ```

2. **Use debug image**:
   ```bash
   docker build -f Dockerfile.alpine -t omni-api:debug .
   docker run -it --rm omni-api:debug sh
   ```

### Common Issues

**Connection refused to Omni:**
- Verify `OMNI_ENDPOINT` is correct
- Check network connectivity from container
- Verify authentication credentials

**Port already in use:**
- Change the host port: `-p 8081:8080`
- Or stop the conflicting service

**Permission denied:**
- Ensure the container runs as non-root (default)
- Check file permissions if mounting volumes

## Production Deployment

For production deployments, consider:

1. **Use specific version tags** instead of `latest`
2. **Set resource limits** (CPU, memory)
3. **Use secrets management** for sensitive credentials
4. **Enable read-only filesystem**
5. **Use orchestration** (Kubernetes, Docker Swarm)
6. **Set up monitoring** and logging
7. **Use HTTPS** with reverse proxy (nginx, traefik)

### Kubernetes Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: omni-api
spec:
  replicas: 2
  selector:
    matchLabels:
      app: omni-api
  template:
    metadata:
      labels:
        app: omni-api
    spec:
      containers:
      - name: omni-api
        image: omni-api:0.0.1
        ports:
        - containerPort: 8080
        env:
        - name: OMNI_ENDPOINT
          valueFrom:
            secretKeyRef:
              name: omni-secrets
              key: endpoint
        - name: OMNI_SERVICE_ACCOUNT_KEY
          valueFrom:
            secretKeyRef:
              name: omni-secrets
              key: service-account-key
        resources:
          limits:
            memory: "512Mi"
            cpu: "500m"
          requests:
            memory: "128Mi"
            cpu: "100m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 65532
          readOnlyRootFilesystem: true
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

## Image Size Comparison

- **Distroless**: ~20-30 MB
- **Alpine**: ~15-25 MB (includes shell)
- **Ubuntu/Debian**: ~100+ MB

Distroless is recommended for production due to minimal attack surface.

# Deployment Guide

This document provides comprehensive instructions for deploying the Anti-Martingale backend to various environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [CI/CD Pipeline](#cicd-pipeline)
- [Deployment Methods](#deployment-methods)
  - [Docker Compose](#docker-compose)
  - [Kubernetes](#kubernetes)
  - [Cloud Platforms](#cloud-platforms)
- [Environment Configuration](#environment-configuration)
- [Database Migration](#database-migration)
- [Monitoring and Logging](#monitoring-and-logging)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Docker and Docker Compose installed
- PostgreSQL 16+ database
- Environment variables configured
- Access to deployment server/platform

## CI/CD Pipeline

### Makefile Targets

The project includes comprehensive Makefile targets for CI/CD:

```bash
# Complete CI pipeline (lint + test + build)
make ci

# Individual CI steps
make ci-lint              # Run linters and format checks
make ci-test              # Run tests with coverage
make ci-build             # Build optimized binary

# Coverage and checks
make coverage             # Generate HTML coverage report
make check                # Run pre-commit checks

# Deployment
make deploy-docker        # Deploy with Docker Compose
make deploy-production    # Build production binary
```

### GitHub Actions

The project includes a complete GitHub Actions workflow (`.github/workflows/ci.yml`) that:

1. **Test Stage**: Runs tests with PostgreSQL service
2. **Build Stage**: Builds the application binary
3. **Docker Stage**: Builds and pushes Docker images
4. **Deploy Stage**: Deploys to production (manual trigger)

**Required Secrets:**
- `DOCKER_USERNAME`: Docker Hub username
- `DOCKER_PASSWORD`: Docker Hub password/token

### GitLab CI/CD

The project includes a GitLab CI configuration (`.gitlab-ci.yml`) with:

1. **Test Stage**: Runs tests with coverage reporting
2. **Build Stage**: Builds binary and Docker image
3. **Deploy Stage**: Manual deployment to production

**Required Variables:**
- `CI_REGISTRY_USER`: Container registry username
- `CI_REGISTRY_PASSWORD`: Container registry password

## Deployment Methods

### Docker Compose

The simplest deployment method using Docker Compose:

```bash
# 1. Clone the repository
git clone <repository-url>
cd anti-martingale-backend

# 2. Configure environment (optional)
cp .env.example .env
# Edit .env with your settings

# 3. Deploy
make deploy-docker
# Or: docker-compose up -d

# 4. Check status
docker-compose ps
docker-compose logs -f app

# 5. Stop deployment
docker-compose down
```

**Production Docker Compose:**

Create a `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    restart: always
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    networks:
      - antimartingale-network

  app:
    image: your-registry/anti-martingale-backend:latest
    restart: always
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME}
      DB_SSLMODE: ${DB_SSLMODE:-require}
      SERVER_PORT: ":8080"
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    networks:
      - antimartingale-network
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/stats"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:

networks:
  antimartingale-network:
    driver: bridge
```

Deploy with:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes

**Example Kubernetes deployment:**

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: anti-martingale-backend
  labels:
    app: anti-martingale
spec:
  replicas: 3
  selector:
    matchLabels:
      app: anti-martingale
  template:
    metadata:
      labels:
        app: anti-martingale
    spec:
      containers:
      - name: app
        image: your-registry/anti-martingale-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: host
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
        - name: DB_NAME
          value: "antimartingale"
        - name: SERVER_PORT
          value: ":8080"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /stats
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /stats
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: anti-martingale-service
spec:
  selector:
    app: anti-martingale
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

Deploy:
```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

### Cloud Platforms

#### AWS (ECS/Fargate)

1. Build and push Docker image to ECR
2. Create ECS task definition
3. Create ECS service with load balancer
4. Configure RDS PostgreSQL instance
5. Set environment variables in task definition

#### Google Cloud (Cloud Run)

```bash
# Build and deploy
gcloud builds submit --tag gcr.io/PROJECT_ID/anti-martingale-backend
gcloud run deploy anti-martingale-backend \
  --image gcr.io/PROJECT_ID/anti-martingale-backend \
  --platform managed \
  --region us-central1 \
  --set-env-vars DB_HOST=CLOUD_SQL_IP,DB_NAME=antimartingale
```

#### Azure (Container Instances)

```bash
# Deploy container
az container create \
  --resource-group myResourceGroup \
  --name anti-martingale-backend \
  --image your-registry/anti-martingale-backend:latest \
  --dns-name-label anti-martingale \
  --ports 8080 \
  --environment-variables \
    DB_HOST=your-db-host \
    DB_NAME=antimartingale
```

## Environment Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` or `postgres` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `your-secure-password` |
| `DB_NAME` | Database name | `antimartingale` |
| `DB_SSLMODE` | SSL mode | `disable`, `require`, `verify-full` |
| `SERVER_PORT` | Server port | `:8080` |

### Production Configuration

For production, ensure:

1. **Use strong database passwords**
2. **Enable SSL for database connections** (`DB_SSLMODE=require`)
3. **Use environment-specific database names**
4. **Set appropriate resource limits**
5. **Configure proper logging**

## Database Migration

The application automatically runs migrations on startup. However, for manual migration:

```bash
# Using the application
./bin/server  # Migrations run automatically

# Using psql (manual)
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_init.sql
```

### Database Backup

```bash
# Backup
docker-compose exec postgres pg_dump -U postgres antimartingale > backup.sql

# Restore
docker-compose exec -T postgres psql -U postgres antimartingale < backup.sql
```

## Monitoring and Logging

### Health Checks

The application provides a stats endpoint for health checks:

```bash
curl http://localhost:8080/stats
```

### Docker Logs

```bash
# View logs
docker-compose logs -f app

# View specific number of lines
docker-compose logs --tail=100 app
```

### Application Metrics

Monitor these key metrics:

1. **Active WebSocket connections**
2. **Database connection pool usage**
3. **Request latency**
4. **Error rates**
5. **Game statistics** (available at `/stats`)

## Troubleshooting

### Database Connection Issues

```bash
# Check database is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres psql -U postgres -d antimartingale -c "SELECT 1;"
```

### Application Won't Start

```bash
# Check environment variables
docker-compose config

# Check application logs
docker-compose logs app

# Restart services
docker-compose restart
```

### Performance Issues

1. **Check database indexes**
2. **Monitor connection pool size**
3. **Review slow query logs**
4. **Check memory usage**: `docker stats`

### Common Errors

**Error: "Failed to connect to database"**
- Solution: Check DB_HOST, DB_PORT, and ensure PostgreSQL is running

**Error: "Failed to run migrations"**
- Solution: Check database permissions and schema

**Error: "Port already in use"**
- Solution: Change SERVER_PORT or stop conflicting service

## Rollback Procedure

If deployment fails:

```bash
# Docker Compose
docker-compose down
docker-compose pull  # Pull previous version
docker-compose up -d

# Kubernetes
kubectl rollout undo deployment/anti-martingale-backend
```

## Security Best Practices

1. **Never commit secrets** to version control
2. **Use secrets management** (AWS Secrets Manager, HashiCorp Vault)
3. **Enable database SSL** in production
4. **Use non-root containers** (already configured in Dockerfile)
5. **Scan images for vulnerabilities** regularly
6. **Keep dependencies updated**
7. **Implement rate limiting** at reverse proxy level
8. **Use HTTPS** with valid certificates

## Support

For issues or questions:
- Create an issue in the repository
- Check existing documentation
- Review application logs

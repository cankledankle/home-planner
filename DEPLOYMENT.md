# Deployment Guide

## Overview

- **Staging**: Auto-deployed to Dokploy when pushing to `dev` branch
- **Production**: Manual deployment to any VPS using Docker Compose
- **HTTPS**: Automatic via Caddy (Let's Encrypt)

## Staging (Dokploy)

1. Push to `dev` branch:

   ```bash
   git checkout dev
   git merge main  # or make changes
   git push origin dev
   ```

2. GitHub Actions builds and pushes `ghcr.io/cankledankle/home-planner:dev`

3. Dokploy automatically pulls and deploys the new image

**Setup in Dokploy:**

- Create a new application
- Use `docker-compose.dokploy.yml` as the compose file
- Set environment variables in Dokploy dashboard

## Production (VPS)

### One-Time Setup

1. **Provision a VPS** (Ubuntu/Debian recommended)

2. **Install Docker & Docker Compose v2**:

   ```bash
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   sudo usermod -aG docker $USER
   newgrp docker
   ```

3. **Clone the repository**:

   ```bash
   git clone https://github.com/cankledankle/home-planner.git
   cd home-planner
   ```

4. **Configure environment variables**:

   ```bash
   cp .env.example .env
   # Edit .env with your values
   nano .env
   ```

5. **Start the application**:

   ```bash
   docker compose up -d
   ```

   Caddy will automatically obtain SSL certificates from Let's Encrypt on first run.

### Updating Production

When a new version is released (tagged with `v*`):

```bash
# SSH into your VPS
cd ~/home-planner

# Pull the new image and restart only the app container
docker compose pull app
docker compose up -d --no-deps app

# Or pull latest and restart everything
docker compose pull
docker compose up -d
```

### Release Workflow

To create a new release:

```bash
# Create and push a version tag
git tag -a v1.0.1 -m "Version 1.0.1"
git push origin v1.0.1
```

This triggers GitHub Actions to build and push the image to GHCR.

## Health Monitoring

The application exposes a health check endpoint at `/health` that checks database connectivity:

- **Healthy**: Returns `200 OK` with JSON `{"status": "ok", "timestamp": "...", "version": "1.0.0"}`
- **Unhealthy**: Returns `503 Service Unavailable` if database is unreachable

**Recommended:** Set up a free uptime monitor like [UptimeRobot](https://uptimerobot.com/) to monitor this endpoint and alert you if the application goes down.

### Health Check Setup

1. Sign up for a free account at [UptimeRobot](https://uptimerobot.com/)
2. Add a new monitor:
   - **Monitor Type**: HTTP(s)
   - **Friendly Name**: Home Planner Production
   - **URL**: `https://yourdomain.com/health`
   - **Monitoring Interval**: 5 minutes (free tier)
3. Configure alert contacts (email, Slack, etc.)

## Database Backups

A backup script is included (`backup.sh`) that dumps the PostgreSQL database and uploads it to Cloudflare R2.

### Backup Features

- Automated daily PostgreSQL dumps
- Compression with gzip
- Uploads to Cloudflare R2 (if configured)
- Automatic cleanup of old backups (30 days retention)
- Support for both Docker and direct PostgreSQL connections
- Optional webhook notifications

### Setup

1. **Ensure R2 credentials are configured** in your `.env` file:

   ```bash
   R2_ACCOUNT_ID=your_account_id
   R2_ACCESS_KEY_ID=your_access_key
   R2_SECRET_ACCESS_KEY=your_secret_key
   R2_BUCKET_NAME=your_bucket_name
   ```

2. **Make the backup script executable**:

   ```bash
   chmod +x backup.sh
   ```

3. **Test the backup**:

   ```bash
   ./backup.sh
   ```

4. **Setup automated daily backups** with cron:

   ```bash
   crontab -e
   ```

   Add this line to run backups daily at 2:00 AM:

   ```
   0 2 * * * /home/youruser/home-planner/backup.sh >> /var/log/home-planner-backup.log 2>&1
   ```

5. **Verify cron is set up**:
   ```bash
   crontab -l
   ```

### Backup Configuration

Optional environment variables for backup customization:

| Variable                      | Default     | Description                            |
| ----------------------------- | ----------- | -------------------------------------- |
| `BACKUP_DIR`                  | `./backups` | Local backup directory                 |
| `BACKUP_RETENTION_DAYS`       | `30`        | Days to keep backups                   |
| `BACKUP_NOTIFICATION_WEBHOOK` | -           | Optional webhook URL for notifications |

### Manual Restore

To restore from a backup:

```bash
# Download backup from R2 (if stored there)
aws s3 cp s3://your-bucket/backups/homeplanner_backup_YYYYMMDD_HHMMSS.sql.gz . \
  --endpoint-url=https://your-account-id.r2.cloudflarestorage.com \
  --region=auto

# Decompress
gunzip homeplanner_backup_YYYYMMDD_HHMMSS.sql.gz

# Restore to database
docker compose exec -T postgres psql -U postgres -d homeplanner < homeplanner_backup_YYYYMMDD_HHMMSS.sql
```

## Environment Variables

Create a `.env` file in the project root (see `.env.example`):

### Required Variables

| Variable             | Description                                                  |
| -------------------- | ------------------------------------------------------------ |
| `POSTGRES_PASSWORD`  | Database password                                            |
| `JWT_SECRET`         | JWT signing secret (generate strong random string)           |
| `JWT_REFRESH_SECRET` | JWT refresh token secret (generate strong random string)     |
| `DOMAIN`             | Your domain (e.g., `example.com`). Caddy uses this for HTTPS |
| `FRONTEND_URL`       | Full URL including https (e.g., `https://example.com`)       |

### Optional Variables

| Variable               | Default       | Description                                     |
| ---------------------- | ------------- | ----------------------------------------------- |
| `DB_NAME`              | `homeplanner` | Database name                                   |
| `IMAGE_TAG`            | `latest`      | Docker image tag to deploy                      |
| `ADMIN_EMAIL`          | -             | Initial admin user email (created on first run) |
| `ADMIN_PASSWORD`       | -             | Initial admin user password                     |
| `ADMIN_NAME`           | -             | Initial admin user name                         |
| `R2_ACCOUNT_ID`        | -             | Cloudflare R2 account ID                        |
| `R2_ACCESS_KEY_ID`     | -             | Cloudflare R2 access key                        |
| `R2_SECRET_ACCESS_KEY` | -             | Cloudflare R2 secret key                        |
| `R2_BUCKET_NAME`       | -             | Cloudflare R2 bucket name                       |

### Database Pool Configuration

| Variable                | Default | Description                       |
| ----------------------- | ------- | --------------------------------- |
| `DB_MAX_OPEN_CONNS`     | `25`    | Maximum open database connections |
| `DB_MAX_IDLE_CONNS`     | `5`     | Maximum idle connections          |
| `DB_CONN_MAX_LIFETIME`  | `30m`   | Maximum connection lifetime       |
| `DB_CONN_MAX_IDLE_TIME` | `10m`   | Maximum idle time before closing  |

## File Structure

```
/home-planner/
├── .env                    # Your environment variables (gitignored)
├── .env.example            # Template for environment variables
├── docker-compose.yml      # Production configuration
├── docker-compose.dokploy.yml  # Dokploy staging configuration
├── docker-compose.dev.yml  # Local development (Postgres only)
├── Caddyfile               # Caddy reverse proxy config
├── backup.sh               # Database backup script
└── ...
```

## Architecture Improvements

This deployment includes several optimizations:

- **Health checks**: Postgres and app services have health checks with proper dependency ordering
- **Graceful shutdown**: Server handles SIGTERM/SIGINT signals gracefully
- **Connection pooling**: Database connections are pooled with reasonable limits
- **Automatic migrations**: Database migrations run automatically on startup
- **Static file caching**: Immutable assets cached for 1 year, SPA routes have no cache
- **Rate limiting**: API routes are rate-limited (100 requests/minute per IP)

## Troubleshooting

### View logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f app
docker compose logs -f postgres
docker compose logs -f caddy
```

### Restart services

```bash
docker compose restart
docker compose restart app
```

### Database access

```bash
docker compose exec postgres psql -U postgres -d homeplanner
```

### Update to specific version

```bash
export IMAGE_TAG=v1.0.1
docker compose pull app
docker compose up -d --no-deps app
```

### Check service status

```bash
docker compose ps
```

### Check health endpoint

```bash
curl https://yourdomain.com/health
```

### Common issues

**Caddy fails to get SSL certificate:**

- Ensure your domain DNS points to the VPS IP
- Ensure ports 80 and 443 are open in firewall
- Check Caddy logs: `docker compose logs caddy`

**App won't start:**

- Check logs: `docker compose logs app`
- Verify all required env vars are set in `.env`
- Ensure postgres is healthy: `docker compose ps`
- Check health endpoint: `docker compose exec app wget -qO- http://localhost:8080/health`

**Database connection errors:**

- Check postgres is running: `docker compose ps`
- Verify POSTGRES_PASSWORD is set correctly
- Check database logs: `docker compose logs postgres`

**Backup failures:**

- Check backup logs: `tail -f /var/log/home-planner-backup.log`
- Verify R2 credentials are correct
- Ensure AWS CLI is installed: `aws --version`

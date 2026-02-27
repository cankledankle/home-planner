# Hybrid Deployment: Staging + Production with Nginx

This setup provides two environments on the same VPS, each with a dedicated Nginx web server:

- **Staging** (`main` branch): Auto-deploys on every push to main
- **Production** (version tags): Deploys only when you push a version tag like `v1.0.0`

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                            VPS                               │
│                                                              │
│  ┌──────────────────────┐    ┌──────────────────────┐       │
│  │      STAGING         │    │     PRODUCTION       │       │
│  │                      │    │                      │       │
│  │  ┌────────────────┐  │    │  ┌────────────────┐  │       │
│  │  │     Nginx      │  │    │  │     Nginx      │  │       │
│  │  │   Port 8080    │  │    │  │   Port 8081    │  │       │
│  │  │   (Web Server) │  │    │  │   (Web Server) │  │       │
│  │  └───────┬────────┘  │    │  └───────┬────────┘  │       │
│  │          │            │    │          │            │       │
│  │          ▼            │    │          ▼            │       │
│  │  ┌────────────────┐  │    │  ┌────────────────┐  │       │
│  │  │  Go Backend    │  │    │  │  Go Backend    │  │       │
│  │  │   Port 8080    │  │    │  │   Port 8080    │  │       │
│  │  │  (Internal)    │  │    │  │  (Internal)    │  │       │
│  │  └───────┬────────┘  │    │  └───────┬────────┘  │       │
│  │          │            │    │          │            │       │
│  │          ▼            │    │          ▼            │       │
│  │  ┌────────────────┐  │    │  ┌────────────────┐  │       │
│  │  │   Postgres     │  │    │  │   Postgres     │  │       │
│  │  │ homeplanner_st │  │    │  │ homeplanner_pr │  │       │
│  │  │    aging       │  │    │  │     od         │  │       │
│  │  └────────────────┘  │    │  └────────────────┘  │       │
│  │                      │    │                      │       │
│  │  SFTP: 2022          │    │  SFTP: 2023          │       │
│  │  Admin: localhost:808│    │  Admin: localhost:808│       │
│  │        1             │    │        2             │       │
│  └──────────────────────┘    └──────────────────────┘       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## What's New: Nginx Web Server

**Nginx now sits in front of the Go backend providing:**

- **Reverse Proxy**: Routes requests to the Go backend
- **Static File Caching**: Better performance for CSS/JS/images
- **Gzip Compression**: Reduced bandwidth usage
- **Security Headers**: XSS protection, CSRF protection
- **SSL/HTTPS Support**: Easy Let's Encrypt integration
- **Load Balancing**: Ready for future scaling

## Port Mapping

| Service         | Staging        | Production     | Access                |
| --------------- | -------------- | -------------- | --------------------- |
| **Nginx HTTP**  | 8080           | 8081           | `http://vps-ip:PORT`  |
| **Nginx HTTPS** | 8443           | 8444           | `https://vps-ip:PORT` |
| **SFTP**        | 2022           | 2023           | SFTP clients          |
| **SFTP Admin**  | localhost:8081 | localhost:8082 | SSH tunnel only       |
| **Go Backend**  | Internal only  | Internal only  | Via Nginx only        |

**Note:** The Go backend no longer exposes ports directly - all traffic goes through Nginx.

## Workflow

### For Development (Staging)

```bash
# Make changes
git add .
git commit -m "Add new feature"
git push origin main

# Automatically deploys to staging
# Access at: http://your-vps-ip:8080
```

### For Production Release

```bash
# Tag a release
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# Automatically deploys to production
# Access at: http://your-vps-ip:8081
```

## Setup

### 1. GitHub Secrets

Add these in GitHub → Settings → Secrets and variables → Actions:

| Secret        | Description                           |
| ------------- | ------------------------------------- |
| `VPS_HOST`    | Your VPS IP address                   |
| `VPS_USER`    | SSH username (e.g., `root`, `ubuntu`) |
| `VPS_SSH_KEY` | Private SSH key                       |

### 2. VPS Setup

```bash
# SSH into VPS
ssh user@your-vps-ip

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Create directory
mkdir -p ~/home-planner
cd ~/home-planner

# Copy files from repo
git clone <your-repo> temp
cp temp/docker-compose.yml .
cp temp/deploy.sh .
cp -r temp/nginx .
chmod +x deploy.sh
rm -rf temp
```

### 3. Environment Configuration

Create two environment files:

**`.env.staging`**

```env
# Database
POSTGRES_PASSWORD=staging_db_password

# Application
JWT_SECRET=staging_jwt_secret_at_least_32_chars

# R2 Storage (Cloudflare)
R2_ACCOUNT_ID=your-r2-account-id
R2_ACCESS_KEY_ID=your-r2-access-key-id
R2_SECRET_ACCESS_KEY=your-r2-secret-access-key
R2_BUCKET_NAME=your-bucket-name

# Initial Admin User
ADMIN_EMAIL=staging@example.com
ADMIN_PASSWORD=staging_admin_pass
ADMIN_NAME=Staging Admin

# SFTPGo
SFTPGO_ADMIN_USERNAME=admin
SFTPGO_ADMIN_PASSWORD=your_sftpgo_password

# Nginx (optional - for custom domains)
# DOMAIN=staging.yourdomain.com
# NGINX_ENABLE_SSL=false
```

**`.env.production`**

```env
# Database
POSTGRES_PASSWORD=production_db_password

# Application
JWT_SECRET=production_jwt_secret_at_least_32_chars

# R2 Storage (Cloudflare)
R2_ACCOUNT_ID=your-r2-account-id
R2_ACCESS_KEY_ID=your-r2-access-key-id
R2_SECRET_ACCESS_KEY=your-r2-secret-access-key
R2_BUCKET_NAME=your-bucket-name

# Initial Admin User
ADMIN_EMAIL=admin@yourdomain.com
ADMIN_PASSWORD=strong_production_pass
ADMIN_NAME=Production Admin

# SFTPGo
SFTPGO_ADMIN_USERNAME=admin
SFTPGO_ADMIN_PASSWORD=your_sftpgo_password

# Nginx (optional - for custom domains)
# DOMAIN=yourdomain.com
# NGINX_ENABLE_SSL=false
```

## SSL/HTTPS with Let's Encrypt (Optional but Recommended)

### Method 1: Automated with Certbot Container

Add this to your environment override file:

```yaml
services:
  certbot:
    image: certbot/certbot
    container_name: certbot-${ENVIRONMENT}
    volumes:
      - ./nginx/ssl:/etc/letsencrypt
      - ./nginx/certbot-data:/var/lib/letsencrypt
    entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew; sleep 12h & wait $${!}; done;'"
```

Or run manually first:

```bash
# On VPS
docker run -it --rm \
  -v ~/home-planner/nginx/ssl:/etc/letsencrypt \
  -v ~/home-planner/nginx/certbot-data:/var/lib/letsencrypt \
  certbot/certbot certonly \
  --standalone \
  -d yourdomain.com \
  -d www.yourdomain.com

# Then enable SSL in .env.production
# NGINX_ENABLE_SSL=true
```

### Method 2: Manual SSL Certificates

Place your certificates in:

```
nginx/ssl/
├── fullchain.pem
└── privkey.pem
```

Then enable SSL:

```env
NGINX_ENABLE_SSL=true
```

## Deployment Triggers

### Automatic

- **Push to `main`**: Deploys to staging (port 8080)
- **Push tag `v*`** (e.g., `v1.0.0`): Deploys to production (port 8081)

### Manual

Go to GitHub → Actions → "Build and Deploy" → Run workflow → Choose environment

## Access Points

| Environment | HTTP    | HTTPS   | SFTP    | SFTP Admin\*   |
| ----------- | ------- | ------- | ------- | -------------- |
| Staging     | `:8080` | `:8443` | `:2022` | localhost:8081 |
| Production  | `:8081` | `:8444` | `:2023` | localhost:8082 |

\* SFTP Admin is bound to localhost only for security. Access via SSH tunnel.

### Access SFTPGo Admin Securely

```bash
# For staging
ssh -L 8081:localhost:8081 user@your-vps-ip
# Then open http://localhost:8081 in your browser

# For production
ssh -L 8082:localhost:8082 user@your-vps-ip
# Then open http://localhost:8082 in your browser
```

## Benefits

1. **Nginx Benefits**:
   - Better static file serving with caching
   - Gzip compression
   - Security headers
   - SSL termination
   - Easy custom domains

2. **Environment Isolation**:
   - Separate databases
   - Separate nginx instances
   - Independent scaling

3. **Development Workflow**:
   - Iterate quickly on staging
   - Controlled production releases
   - Easy rollback to previous tags

## Example Workflow

```bash
# Day 1: Start new feature
git checkout -b feature/new-ui
# ... work on feature ...
git commit -m "Add new UI components"
git push origin feature/new-ui

# Day 2: Merge to main (auto-deploys to staging)
git checkout main
git merge feature/new-ui
git push origin main
# Check staging: http://vps-ip:8080

# Day 3: Show client, get feedback
# Make fixes, push to main again

# Day 4: Client approves, release to production
git tag -a v1.2.0 -m "Add new UI components"
git push origin v1.2.0
# Check production: http://vps-ip:8081

# Day 5: Bug discovered in production!
# Quick rollback
git tag -a v1.2.1 -m "Rollback to v1.1.0"
git push origin v1.2.1
# Or manually deploy previous tag via GitHub Actions
```

## Managing Both Environments

### Check Status

```bash
# Staging
docker-compose -f docker-compose.yml -f docker-compose.staging.yml ps

# Production
docker-compose -f docker-compose.yml -f docker-compose.production.yml ps
```

### View Logs

```bash
# Staging app logs
docker logs home-planner-app-staging

# Staging nginx logs
docker logs home-planner-nginx-staging

# Production logs
docker logs home-planner-app-prod
docker logs home-planner-nginx-prod
```

### Restart Services

```bash
# Restart staging
docker-compose -f docker-compose.yml -f docker-compose.staging.yml restart

# Restart only nginx
docker-compose -f docker-compose.yml -f docker-compose.staging.yml restart nginx
```

### Stop an Environment

```bash
# Stop staging
docker-compose -f docker-compose.yml -f docker-compose.staging.yml down

# Stop production
docker-compose -f docker-compose.yml -f docker-compose.production.yml down
```

## Security Considerations

1. **SFTP Admin**: Bound to localhost only (not exposed publicly)
2. **SSL**: Enable for production (`NGINX_ENABLE_SSL=true`)
3. **Firewall**: Only open ports 8080, 8081 (and 443 for SSL)
   ```bash
   sudo ufw allow 8080/tcp
   sudo ufw allow 8081/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```
4. **Different Secrets**: Never reuse passwords between environments

## Troubleshooting

### Nginx Not Starting

```bash
# Check nginx config
docker-compose -f docker-compose.yml -f docker-compose.staging.yml logs nginx

# Test nginx config manually
docker exec home-planner-nginx-staging nginx -t
```

### Port Conflicts

If you see "port already allocated":

```bash
# Check what's using the port
sudo lsof -i :8080

# Kill process or change port in .env.staging
```

### SSL Certificate Issues

```bash
# Check certificate files
ls -la nginx/ssl/

# Generate self-signed cert for testing
cd nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout privkey.pem -out fullchain.pem \
  -subj "/CN=localhost"
```

### Go Backend Not Accessible

```bash
# Check if app is healthy
curl http://localhost:8080/health

# Check app logs
docker logs home-planner-app-staging
```

## Next Steps

1. ✅ Set up GitHub Secrets
2. ✅ Create environment files on VPS
3. ✅ Copy nginx directory to VPS
4. ✅ Push to main (deploys staging)
5. ✅ Create a tag (deploys production)
6. 🎯 Set up custom domain with SSL (optional)
7. 🎯 Set up monitoring (optional)

## Questions?

- **Q: Why Nginx instead of serving directly from Go?**
  - A: Nginx is better at static files, SSL, and provides security headers

- **Q: Can I use the same domain for both environments?**
  - A: Use subdomains: staging.yourdomain.com and yourdomain.com

- **Q: How do I update Nginx config?**
  - A: Edit `nginx/nginx.conf.template`, commit, push, and redeploy

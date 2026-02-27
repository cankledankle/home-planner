#!/bin/bash
# Usage: ./deploy.sh [staging|latest|v1.2.3]

set -e

# Configuration
IMAGE_TAG=${1:-staging}
REGISTRY="ghcr.io"
REPO="${GITHUB_REPOSITORY:-cankledankle/home-planner}"
ENVIRONMENT="${ENVIRONMENT:-staging}"

# Determine environment from tag
if [[ "$IMAGE_TAG" == v* ]]; then
    ENVIRONMENT="production"
fi

echo "🚀 Starting deployment..."
echo "📦 Image: ${REGISTRY}/${REPO}:${IMAGE_TAG}"
echo "🌍 Environment: ${ENVIRONMENT}"

# Navigate to app directory
cd ~/home-planner

# Create environment-specific .env file if it doesn't exist
ENV_FILE=".env.${ENVIRONMENT}"
if [ ! -f "$ENV_FILE" ] && [ ! -f ".env" ]; then
    echo "⚠️  No environment file found!"
    echo "   Create either .env or .env.${ENVIRONMENT}"
    echo "   Example files: .env.staging.example, .env.production.example"
    exit 1
fi

# Set environment-specific variables
export APP_IMAGE="${REGISTRY}/${REPO}:${IMAGE_TAG}"
export ENVIRONMENT="${ENVIRONMENT}"

if [ "$ENVIRONMENT" == "production" ]; then
    export APP_NAME="home-planner-app-prod"
    export NGINX_NAME="home-planner-nginx-prod"
    export DB_NAME="homeplanner_prod"
    export SFTPGO_NAME="home-planner-sftpgo-prod"
    export SFTPGO_PORT=2023
    export SFTPGO_ADMIN_PORT=8082
    export NGINX_HTTP_PORT=8081
    export NGINX_HTTPS_PORT=8444
    export DOMAIN="${DOMAIN:-localhost}"
else
    export APP_NAME="home-planner-app-staging"
    export NGINX_NAME="home-planner-nginx-staging"
    export DB_NAME="homeplanner_staging"
    export SFTPGO_NAME="home-planner-sftpgo-staging"
    export SFTPGO_PORT=2022
    export SFTPGO_ADMIN_PORT=8081
    export NGINX_HTTP_PORT=8080
    export NGINX_HTTPS_PORT=8443
    export DOMAIN="${DOMAIN:-localhost}"
fi

echo "🔧 Configuration:"
echo "   App: ${APP_NAME}"
echo "   Nginx: ${NGINX_NAME}"
echo "   Database: ${DB_NAME}"
echo "   HTTP Port: ${NGINX_HTTP_PORT}"
echo "   HTTPS Port: ${NGINX_HTTPS_PORT}"
echo "   Domain: ${DOMAIN}"

# Pull the latest image
echo "⬇️  Pulling image..."
docker pull "${APP_IMAGE}"

# Create environment-specific docker-compose override
cat > docker-compose.${ENVIRONMENT}.yml << EOF
services:
  postgres:
    container_name: ${DB_NAME}
    environment:
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - ${DB_NAME}_data:/var/lib/postgresql/data

  sftpgo:
    container_name: ${SFTPGO_NAME}
    ports:
      - "${SFTPGO_PORT}:2022"
      - "127.0.0.1:${SFTPGO_ADMIN_PORT}:8080"
    volumes:
      - ${SFTPGO_NAME}_data:/var/lib/sftpgo

  app:
    image: ${APP_IMAGE}
    container_name: ${APP_NAME}
    environment:
      DATABASE_URL: postgres://postgres:\${POSTGRES_PASSWORD:-postgres}@postgres:5432/${DB_NAME}?sslmode=disable
      FRONTEND_URL: http://${DOMAIN}

  nginx:
    container_name: ${NGINX_NAME}
    ports:
      - "${NGINX_HTTP_PORT}:80"
      - "${NGINX_HTTPS_PORT}:443"
    environment:
      NGINX_SERVER_NAME: ${DOMAIN}
    depends_on:
      ${APP_NAME}:
        condition: service_healthy

volumes:
  ${DB_NAME}_data:
  ${SFTPGO_NAME}_data:
EOF

# Ensure nginx config directory exists
mkdir -p nginx/ssl

# Stop and remove old containers
echo "🛑 Stopping old containers..."
docker-compose -f docker-compose.yml -f docker-compose.${ENVIRONMENT}.yml down || true

# Start with new image
echo "▶️  Starting services..."
if [ -f ".env.${ENVIRONMENT}" ]; then
    docker-compose -f docker-compose.yml -f docker-compose.${ENVIRONMENT}.yml --env-file .env.${ENVIRONMENT} up -d --build
else
    docker-compose -f docker-compose.yml -f docker-compose.${ENVIRONMENT}.yml up -d --build
fi

# Wait for health check
echo "🏥 Checking health..."
sleep 5
for i in {1..12}; do
    if curl -sf http://localhost:${NGINX_HTTP_PORT}/health > /dev/null 2>&1; then
        echo "✅ Application is healthy!"
        break
    fi
    echo "   Waiting for app to start... (${i}/12)"
    sleep 5
done

# Cleanup old images (keep last 3)
echo "🧹 Cleaning up old images..."
docker images ${REGISTRY}/${REPO} --format "{{.Repository}}:{{.Tag}} {{.ID}}" |
  grep -v "${IMAGE_TAG}" |
  tail -n +4 |
  awk '{print $2}' |
  xargs -r docker rmi -f 2>/dev/null || true

docker image prune -f

echo ""
echo "✨ Deployment complete!"
echo ""
echo "📊 Status:"
docker-compose -f docker-compose.yml -f docker-compose.${ENVIRONMENT}.yml ps
echo ""
echo "🌐 Access your app:"
VPS_IP=$(curl -s ifconfig.me 2>/dev/null || echo "your-vps-ip")
if [ "$ENVIRONMENT" == "production" ]; then
    echo "   Production: http://${VPS_IP}:${NGINX_HTTP_PORT}"
    echo "   (or https://${VPS_IP}:${NGINX_HTTPS_PORT} if SSL enabled)"
else
    echo "   Staging: http://${VPS_IP}:${NGINX_HTTP_PORT}"
    echo "   (or https://${VPS_IP}:${NGINX_HTTPS_PORT} if SSL enabled)"
fi
echo ""
echo "⚠️  Security Note: SFTPGo Admin is bound to localhost only"
echo "   Access via SSH tunnel: ssh -L ${SFTPGO_ADMIN_PORT}:localhost:${SFTPGO_ADMIN_PORT} user@${VPS_IP}"

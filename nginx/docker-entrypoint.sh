#!/bin/sh
# Nginx entrypoint script
# Replaces environment variables in nginx.conf.template

set -e

# Default values
NGINX_SERVER_NAME=${NGINX_SERVER_NAME:-localhost}
NGINX_APP_HOST=${NGINX_APP_HOST:-app}
NGINX_APP_PORT=${NGINX_APP_PORT:-8080}
NGINX_ENABLE_SSL=${NGINX_ENABLE_SSL:-false}

# Create SSL directory if it doesn't exist
mkdir -p /etc/nginx/ssl

# Generate self-signed certificate for development if SSL is enabled but certs don't exist
if [ "$NGINX_ENABLE_SSL" = "true" ] && [ ! -f /etc/nginx/ssl/fullchain.pem ]; then
    echo "Generating self-signed SSL certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /etc/nginx/ssl/privkey.pem \
        -out /etc/nginx/ssl/fullchain.pem \
        -subj "/CN=${NGINX_SERVER_NAME}" \
        2>/dev/null || true
fi

# Replace variables in template
envsubst '$NGINX_SERVER_NAME $NGINX_APP_HOST $NGINX_APP_PORT' \
    < /etc/nginx/nginx.conf.template \
    > /etc/nginx/conf.d/default.conf

# If SSL is not enabled, remove SSL server block
if [ "$NGINX_ENABLE_SSL" != "true" ]; then
    sed -i '/^server {/,/^}/{
        /listen 443/d
    }' /etc/nginx/conf.d/default.conf
    
    # Remove empty SSL server blocks
    awk '/^server \{/{found=1; buf=$0; next} 
         found{buf=buf ORS $0} 
         /^\}/{if(found && buf ~ /listen 443/) {found=0; buf=""} else {print buf; found=0; buf=""}} 
         !found{print}' /etc/nginx/conf.d/default.conf > /tmp/nginx.conf.tmp
    mv /tmp/nginx.conf.tmp /etc/nginx/conf.d/default.conf
fi

echo "Nginx configuration generated:"
echo "  Server Name: ${NGINX_SERVER_NAME}"
echo "  App Host: ${NGINX_APP_HOST}:${NGINX_APP_PORT}"
echo "  SSL Enabled: ${NGINX_ENABLE_SSL}"

# Test nginx configuration
nginx -t

# Start nginx
exec nginx -g 'daemon off;'

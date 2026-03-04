#!/bin/bash
# Database Backup Script
# Backs up PostgreSQL database and uploads to Cloudflare R2
#
# Setup as a daily cron job:
#   crontab -e
#   Add: 0 2 * * * /path/to/backup.sh >> /var/log/home-planner-backup.log 2>&1
#   This runs daily at 2:00 AM

set -e

# Load environment variables from .env file
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/.env" ]; then
    export $(grep -v '^#' "$SCRIPT_DIR/.env" | xargs)
fi

# Configuration
DB_NAME="${DB_NAME:-homeplanner}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="homeplanner_backup_${TIMESTAMP}.sql"
COMPRESSED_FILE="${BACKUP_FILE}.gz"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"

# R2 Configuration
R2_ACCOUNT_ID="${R2_ACCOUNT_ID}"
R2_ACCESS_KEY_ID="${R2_ACCESS_KEY_ID}"
R2_SECRET_ACCESS_KEY="${R2_SECRET_ACCESS_KEY}"
R2_BUCKET_NAME="${R2_BUCKET_NAME}"
R2_ENDPOINT="https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting database backup..."

# Check if running inside Docker or directly on host
if docker compose ps | grep -q "postgres"; then
    echo "Postgres is running in Docker, using docker exec..."
    
    # Create backup using docker exec
    docker compose exec -T postgres pg_dump \
        -U postgres \
        -d "$DB_NAME" \
        --no-owner \
        --no-privileges \
        --clean \
        --if-exists \
        > "$BACKUP_DIR/$BACKUP_FILE"
else
    echo "Postgres is not running in Docker, attempting direct connection..."
    
    # Check if pg_dump is available
    if ! command -v pg_dump &> /dev/null; then
        echo "Error: pg_dump is not installed. Please install PostgreSQL client tools."
        exit 1
    fi
    
    # Create backup using direct connection
    pg_dump \
        -h localhost \
        -U postgres \
        -d "$DB_NAME" \
        --no-owner \
        --no-privileges \
        --clean \
        --if-exists \
        > "$BACKUP_DIR/$BACKUP_FILE"
fi

if [ $? -ne 0 ]; then
    echo "Error: Database backup failed!"
    exit 1
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Database dumped successfully: $BACKUP_FILE"

# Compress the backup
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Compressing backup..."
gzip -f "$BACKUP_DIR/$BACKUP_FILE"

if [ $? -ne 0 ]; then
    echo "Error: Compression failed!"
    exit 1
fi

BACKUP_SIZE=$(du -h "$BACKUP_DIR/$COMPRESSED_FILE" | cut -f1)
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup compressed: $COMPRESSED_FILE ($BACKUP_SIZE)"

# Upload to R2 if credentials are configured
if [ -n "$R2_ACCOUNT_ID" ] && [ -n "$R2_ACCESS_KEY_ID" ] && [ -n "$R2_SECRET_ACCESS_KEY" ] && [ -n "$R2_BUCKET_NAME" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Uploading to Cloudflare R2..."
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        echo "Warning: AWS CLI is not installed. Installing..."
        # Try to install AWS CLI
        if command -v apt-get &> /dev/null; then
            apt-get update && apt-get install -y awscli
        elif command -v yum &> /dev/null; then
            yum install -y awscli
        else
            echo "Error: Could not install AWS CLI. Please install it manually."
            exit 1
        fi
    fi
    
    # Configure AWS CLI for R2
    export AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID"
    export AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY"
    
    # Upload to R2
    aws s3 cp "$BACKUP_DIR/$COMPRESSED_FILE" \
        "s3://$R2_BUCKET_NAME/backups/$COMPRESSED_FILE" \
        --endpoint-url="$R2_ENDPOINT" \
        --region=auto
    
    if [ $? -eq 0 ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup uploaded to R2: s3://$R2_BUCKET_NAME/backups/$COMPRESSED_FILE"
        
        # Remove local backup after successful upload
        rm "$BACKUP_DIR/$COMPRESSED_FILE"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Local backup removed after successful upload"
    else
        echo "Warning: Failed to upload to R2. Keeping local backup."
    fi
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] R2 credentials not configured. Backup kept locally."
fi

# Cleanup old local backups (keep only last RETENTION_DAYS days)
if [ -d "$BACKUP_DIR" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaning up old local backups (keeping $RETENTION_DAYS days)..."
    find "$BACKUP_DIR" -name "homeplanner_backup_*.sql.gz" -type f -mtime +$RETENTION_DAYS -delete
fi

# Cleanup old R2 backups (keep only last RETENTION_DAYS days) if R2 is configured
if [ -n "$R2_ACCOUNT_ID" ] && [ -n "$R2_ACCESS_KEY_ID" ] && [ -n "$R2_SECRET_ACCESS_KEY" ] && [ -n "$R2_BUCKET_NAME" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaning up old R2 backups (keeping $RETENTION_DAYS days)..."
    
    CUTOFF_DATE=$(date -d "$RETENTION_DAYS days ago" +%Y%m%d 2>/dev/null || date -v-${RETENTION_DAYS}d +%Y%m%d)
    
    # List and delete old backups from R2
    aws s3 ls "s3://$R2_BUCKET_NAME/backups/" \
        --endpoint-url="$R2_ENDPOINT" \
        --region=auto |
    while read -r line; do
        FILE_DATE=$(echo "$line" | awk '{print $1}' | tr -d '-')
        FILE_NAME=$(echo "$line" | awk '{print $4}')
        
        if [ "$FILE_DATE" -lt "$CUTOFF_DATE" ] 2>/dev/null; then
            echo "Deleting old backup: $FILE_NAME"
            aws s3 rm "s3://$R2_BUCKET_NAME/backups/$FILE_NAME" \
                --endpoint-url="$R2_ENDPOINT" \
                --region=auto
        fi
    done
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup completed successfully!"

# Send notification if configured (optional)
if [ -n "$BACKUP_NOTIFICATION_WEBHOOK" ]; then
    curl -X POST "$BACKUP_NOTIFICATION_WEBHOOK" \
        -H "Content-Type: application/json" \
        -d "{\"text\":\"✅ Home Planner backup completed: $COMPRESSED_FILE ($BACKUP_SIZE)\"}" \
        > /dev/null 2>&1 || true
fi

exit 0
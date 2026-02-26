# Docker & Deployment

## Natural Element Homes — Home Planner

**Version:** 1.0
**Status:** Draft

---

## 1. Overview

The application runs as a set of Docker containers orchestrated with Docker Compose. The entire stack can be started with a single command on any machine with Docker installed.

**Containers:**

| Service    | Image                      | Purpose                                   |
| ---------- | -------------------------- | ----------------------------------------- |
| `app`      | Custom (Go + built Svelte) | API server + serves frontend static files |
| `postgres` | postgres:16                | Primary database                          |
| `sftpgo`   | drakkan/sftpgo             | SFTP access to R2 bucket                  |

Files are stored in Cloudflare R2 — no file storage container needed.

---

## 2. Repository Structure

```
home-planner/
├── backend/
│   ├── Dockerfile
│   └── ...
├── frontend/
│   ├── Dockerfile
│   └── ...
├── docker-compose.yml
├── docker-compose.dev.yml
├── .env.example
└── .env                  ← not committed, created on each server
```

---

## 3. Environment Variables

Copy `.env.example` to `.env` and fill in values before running.

`.env.example`:

```bash
# -----------------------------------------------
# Server
# -----------------------------------------------
PORT=8080
ENV=production

# -----------------------------------------------
# Auth
# -----------------------------------------------
# Generate with: openssl rand -base64 32
JWT_SECRET=
JWT_REFRESH_SECRET=

# -----------------------------------------------
# Database
# -----------------------------------------------
POSTGRES_USER=homeplanner
POSTGRES_PASSWORD=
POSTGRES_DB=homeplanner
DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}

# -----------------------------------------------
# Cloudflare R2
# -----------------------------------------------
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=
R2_ENDPOINT=https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com

# -----------------------------------------------
# Frontend
# -----------------------------------------------
PUBLIC_API_URL=http://localhost:8080

# -----------------------------------------------
# SFTPGo
# -----------------------------------------------
SFTPGO_API_URL=http://sftpgo:8081
SFTPGO_API_KEY=
```

---

## 4. Docker Compose

`docker-compose.yml`:

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - '8080:8080'
    environment:
      - PORT=${PORT}
      - ENV=${ENV}
      - JWT_SECRET=${JWT_SECRET}
      - JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}
      - DATABASE_URL=${DATABASE_URL}
      - R2_ACCOUNT_ID=${R2_ACCOUNT_ID}
      - R2_ACCESS_KEY_ID=${R2_ACCESS_KEY_ID}
      - R2_SECRET_ACCESS_KEY=${R2_SECRET_ACCESS_KEY}
      - R2_BUCKET_NAME=${R2_BUCKET_NAME}
      - R2_ENDPOINT=${R2_ENDPOINT}
      - SFTPGO_API_URL=${SFTPGO_API_URL}
      - SFTPGO_API_KEY=${SFTPGO_API_KEY}
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

  postgres:
    image: postgres:16
    environment:
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=${POSTGRES_DB}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}']
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  sftpgo:
    image: drakkan/sftpgo:latest
    ports:
      - '2022:2022' # SFTP port
      - '8081:8080' # SFTPGo admin API (internal only)
    environment:
      - SFTPGO_DATA_PROVIDER__DRIVER=memory
      - SFTPGO_HTTPD__BINDINGS__0__PORT=8080
      - SFTPGO_SFTPD__BINDINGS__0__PORT=2022
    volumes:
      - sftpgo_data:/var/lib/sftpgo
    restart: unless-stopped

volumes:
  postgres_data:
  sftpgo_data:
```

---

## 5. Dockerfile

Single multi-stage Dockerfile that builds the Svelte frontend, builds the Go binary, and produces a minimal final image.

`Dockerfile` (in repo root):

```dockerfile
# -----------------------------------------------
# Stage 1 — Build frontend
# -----------------------------------------------
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# -----------------------------------------------
# Stage 2 — Build backend
# -----------------------------------------------
FROM golang:1.23-alpine AS backend-builder

WORKDIR /backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# -----------------------------------------------
# Stage 3 — Final image
# -----------------------------------------------
FROM alpine:3.20

WORKDIR /app

# Copy Go binary
COPY --from=backend-builder /backend/server ./server

# Copy migrations
COPY --from=backend-builder /backend/migrations ./migrations

# Copy built frontend
COPY --from=frontend-builder /frontend/build ./static

EXPOSE 8080

CMD ["./server"]
```

The Go server serves the built SvelteKit static files from `./static` in addition to the API routes.

---

## 6. Development Setup

For local development the frontend and backend run separately with hot reload.

`docker-compose.dev.yml`:

```yaml
services:
  postgres:
    image: postgres:16
    environment:
      - POSTGRES_USER=homeplanner
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=homeplanner
    ports:
      - '5432:5432'
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U homeplanner -d homeplanner']
      interval: 5s
      timeout: 5s
      retries: 5

  sftpgo:
    image: drakkan/sftpgo:latest
    ports:
      - '2022:2022'
      - '8081:8080'
    volumes:
      - sftpgo_data:/var/lib/sftpgo

volumes:
  postgres_data:
  sftpgo_data:
```

Start just the dependencies:

```bash
docker compose -f docker-compose.dev.yml up -d
```

Then run frontend and backend separately:

```bash
# Terminal 1 — backend
cd backend
go run cmd/server/main.go

# Terminal 2 — frontend
cd frontend
npm run dev
```

---

## 7. Running in Production

First time setup:

```bash
# 1. Clone the repo on the server
git clone https://github.com/yourorg/home-planner.git
cd home-planner

# 2. Create and fill in environment variables
cp .env.example .env
nano .env

# 3. Build and start
docker compose up -d --build

# 4. Check logs
docker compose logs -f app
```

Updating:

```bash
git pull
docker compose up -d --build
```

Migrations run automatically on startup — no manual migration step needed.

---

## 8. Postgres Data

The Postgres data directory is persisted in a named Docker volume `postgres_data`. This volume survives container restarts and image rebuilds.

**Backup:**

```bash
docker exec home-planner-postgres-1 \
  pg_dump -U homeplanner homeplanner \
  > backup_$(date +%Y%m%d_%H%M%S).sql
```

**Restore:**

```bash
cat backup.sql | docker exec -i home-planner-postgres-1 \
  psql -U homeplanner homeplanner
```

---

## 9. Ports

| Port   | Service  | Description                                                |
| ------ | -------- | ---------------------------------------------------------- |
| `8080` | app      | API + frontend — expose this to the web                    |
| `5432` | postgres | Database — internal only, do not expose                    |
| `2022` | sftpgo   | SFTP connections — expose if SFTP access needed externally |
| `8081` | sftpgo   | SFTPGo admin API — internal only, do not expose            |

In production, put a reverse proxy (nginx or Caddy) in front of port `8080` to handle HTTPS.

---

## 10. HTTPS

Not handled by Docker Compose directly. Use a reverse proxy on the host.

**Caddy (recommended — auto HTTPS):**

```
home-planner.example.com {
    reverse_proxy localhost:8080
}
```

**nginx:**

```nginx
server {
    listen 443 ssl;
    server_name home-planner.example.com;

    ssl_certificate /etc/letsencrypt/live/home-planner.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/home-planner.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 11. Health Check

The app exposes a health endpoint:

```
GET /health
→ 200 { "status": "ok" }
```

Use this for uptime monitoring and load balancer health checks.

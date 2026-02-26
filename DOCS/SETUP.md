# Home Planner - Local Development Setup

A web application for managing home plans with file storage, user management, and SFTP access.

## Quick Start (Development Mode)

### 1. Prerequisites

- Docker Desktop
- Node.js 20+ (for frontend dev server)
- Go 1.23+ (for backend development)

### 2. Start Infrastructure Services

```bash
# Start PostgreSQL and SFTPGo containers
docker-compose -f docker-compose.dev.yml up -d

# Wait for services to be healthy (check with:)
docker-compose -f docker-compose.dev.yml ps
```

### 3. Configure Environment

```bash
# Copy the example env file
cp .env.example .env

# Edit .env with your settings:
# - JWT_SECRET: Generate a secure random string
# - R2_*: Cloudflare R2 credentials (optional for local dev)
# - ADMIN_EMAIL/PASSWORD: Credentials for first admin user
```

### 4. Run Backend (Go)

```bash
cd backend

# Install dependencies
go mod download

# Run migrations and start server
go run cmd/server/main.go

# Backend will be available at http://localhost:8080
# Health check: http://localhost:8080/health
```

### 5. Run Frontend (SvelteKit)

```bash
# In project root (separate terminal)
npm install
npm run dev

# Frontend will be available at http://localhost:5173
```

### 6. Login

Navigate to http://localhost:5173 and login with the admin credentials from your `.env` file.

---

## Alternative: Full Docker Production Build

To test the production build locally:

```bash
# Make sure .env has all required variables
docker-compose up --build

# App will be available at http://localhost:8080
```

---

## Environment Variables Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `JWT_SECRET` | Yes | Secret key for JWT tokens |
| `R2_ACCOUNT_ID` | No* | Cloudflare R2 account ID |
| `R2_ACCESS_KEY_ID` | No* | R2 access key |
| `R2_SECRET_ACCESS_KEY` | No* | R2 secret key |
| `R2_BUCKET_NAME` | No* | R2 bucket name |
| `ADMIN_EMAIL` | No | Auto-create admin user |
| `ADMIN_PASSWORD` | No | Admin user password |
| `ADMIN_NAME` | No | Admin user display name |

*Required for file uploads and SFTP to work

---

## Available Ports

| Service | Port | Description |
|---------|------|-------------|
| Frontend Dev | 5173 | Vite dev server |
| Backend API | 8080 | Go server |
| PostgreSQL | 5432 | Database |
| SFTPGo Admin | 8081 | SFTPGo web UI |
| SFTP | 2022 | SFTP access |

---

## Common Commands

```bash
# View logs
docker-compose logs -f

# Reset database (WARNING: deletes all data)
docker-compose down -v
docker-compose up -d

# Run backend tests
cd backend && go test ./...

# Build frontend for production
npm run build

# Check code formatting
cd backend && gofmt -l .
npm run lint
```

---

## Architecture

- **Frontend**: SvelteKit + Tailwind CSS + shadcn-svelte
- **Backend**: Go (Fiber framework) + PostgreSQL
- **File Storage**: Cloudflare R2 (S3-compatible)
- **SFTP Access**: SFTPGo with R2 backend

## Project Structure

```
.
├── backend/              # Go backend
│   ├── cmd/server/       # Entry point
│   ├── internal/         # Internal packages
│   └── migrations/       # Database migrations
├── src/                  # SvelteKit frontend
│   ├── lib/              # Shared components & utilities
│   └── routes/           # Page routes
├── docker-compose.yml    # Production stack
├── docker-compose.dev.yml # Development services
└── Dockerfile            # Multi-stage build
```

## Troubleshooting

**Port already in use:**
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9
```

**Database connection refused:**
- Ensure PostgreSQL container is running: `docker-compose -f docker-compose.dev.yml ps`
- Check logs: `docker-compose logs postgres`

**Frontend can't connect to backend:**
- Verify `FRONTEND_URL` in `.env` matches your frontend URL
- Check CORS settings in `backend/cmd/server/main.go`

**Missing R2 credentials:**
- File uploads will fail without valid R2 credentials
- SFTP functionality requires R2 to be configured

## Next Steps

1. Create your first plan at http://localhost:5173/plans
2. Upload images to website slots
3. Test CSV import/export functionality
4. Configure SFTP access in Settings
5. Review the full feature list in `PROJECT-LIST.md`
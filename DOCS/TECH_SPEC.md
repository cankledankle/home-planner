# Technical Specification

## Natural Element Homes — Home Planner

**Version:** 1.0
**Status:** Draft

---

## 1. Stack Overview

| Layer         | Technology             | Reason                                                     |
| ------------- | ---------------------- | ---------------------------------------------------------- |
| Frontend      | SvelteKit              | Lightweight, minimal boilerplate, fast to build            |
| UI Components | shadcn-svelte          | Polished components, Tailwind-based                        |
| Data Fetching | TanStack Query         | Caching, loading/error states, query invalidation          |
| Styling       | Tailwind CSS           | Utility-first, works seamlessly with shadcn                |
| Backend       | Go + Fiber             | Fast, excellent file I/O, single binary output             |
| Database      | PostgreSQL             | Relational, reliable, runs in Docker                       |
| File Storage  | Cloudflare R2          | S3-compatible, zero egress fees                            |
| Auth          | JWT (httpOnly cookies) | Stateless, works cleanly with Go + SvelteKit               |
| SFTP          | sftpgo                 | Self-hosted SFTP server, Docker-native, R2 backend support |
| Deployment    | Docker Compose         | Single-command deploy, portable across environments        |

---

## 2. Repository Structure

```
home-planner/
├── frontend/                  # SvelteKit app
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/           # API client functions
│   │   │   ├── components/    # Shared components
│   │   │   ├── stores/        # Svelte stores
│   │   │   └── types/         # TypeScript types
│   │   ├── routes/            # SvelteKit pages
│   │   └── app.html
│   ├── static/
│   ├── package.json
│   └── vite.config.ts
├── backend/                   # Go API
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── auth/              # JWT logic
│   │   ├── handlers/          # Route handlers
│   │   ├── middleware/        # Auth, logging, CORS
│   │   ├── models/            # DB models
│   │   ├── storage/           # R2 client
│   │   └── db/                # DB connection, migrations
│   ├── migrations/            # SQL migration files
│   └── go.mod
├── docs/
│   ├── PRD.md
│   ├── TECH_SPEC.md
│   ├── DATABASE_SCHEMA.md
│   ├── API_SPEC.md
│   ├── UI_SPEC.md
│   └── DOCKER.md
└── docker-compose.yml
```

---

## 3. Architecture Overview

```
┌─────────────────────────────────────────────┐
│                  Browser                     │
│            SvelteKit Frontend                │
└──────────────────┬──────────────────────────┘
                   │ HTTP/JSON (REST)
                   │ JWT in httpOnly cookie
┌──────────────────▼──────────────────────────┐
│              Go + Fiber API                  │
│                                              │
│  ┌─────────────┐  ┌────────────────────┐    │
│  │  Auth       │  │  Route Handlers    │    │
│  │  Middleware │  │  Plans / Files     │    │
│  │  JWT verify │  │  Users / Export    │    │
│  └─────────────┘  └────────────────────┘    │
│                                              │
│  ┌─────────────┐  ┌────────────────────┐    │
│  │  Postgres   │  │  R2 Storage Client │    │
│  │  Client     │  │  AWS SDK v2        │    │
│  └──────┬──────┘  └─────────┬──────────┘    │
└─────────┼───────────────────┼───────────────┘
          │                   │
┌─────────▼──────┐  ┌─────────▼──────────────┐
│   PostgreSQL   │  │    Cloudflare R2        │
│   Container    │  │    Object Storage       │
└────────────────┘  └────────────────────────┘
          │
┌─────────▼──────────────┐
│      SFTPGo            │
│  SFTP Server Container  │
│  R2 backend            │
└────────────────────────┘
```

---

## 4. Authentication

### 4.1 Flow

1. User submits email + password to `POST /api/auth/login`
2. Go verifies credentials against database (bcrypt password comparison)
3. On success, Go generates two tokens:
   - **Access token** — JWT, short-lived (15 minutes)
   - **Refresh token** — JWT, long-lived (7 days)
4. Both tokens set as httpOnly, Secure, SameSite=Strict cookies
5. Frontend never touches the tokens directly — cookies are sent automatically with every request
6. On access token expiry, frontend calls `POST /api/auth/refresh`
7. Go verifies refresh token, issues new access token
8. On logout, both cookies are cleared server-side

### 4.2 JWT Payload

```json
{
	"sub": "user-uuid",
	"email": "user@example.com",
	"role": "admin | editor",
	"exp": 1234567890,
	"iat": 1234567890
}
```

### 4.3 Middleware

Every protected route passes through auth middleware that:

- Extracts JWT from cookie
- Verifies signature using server secret
- Checks expiry
- Attaches user context to request
- Returns 401 if any check fails

Admin-only routes additionally check `role == "admin"` and return 403 if not.

---

## 5. File Upload Flow

### 5.1 Website Image Upload (Slotted)

1. User selects a file and a target slot in the UI
2. Frontend sends `POST /api/plans/:id/files/website` with file and slot name as multipart form data
3. Go validates file type (images only), max size (50MB)
4. Go generates a storage key: `plans/{plan-slug}/website/{slot}.{ext}`
5. Go uploads file to R2 using AWS SDK v2
6. Go upserts file record in database (insert or replace existing slot)
7. Go recalculates plan completeness status and updates plan record
8. Go returns updated file record
9. TanStack Query invalidates `plans` and `plan:{id}` query keys
10. UI updates automatically

### 5.2 Other File Upload (Unslotted)

1. User selects one or more files and a category in the UI
2. Frontend sends `POST /api/plans/:id/files` with files and category as multipart form data
3. Go validates file size (500MB max per file)
4. Go generates storage key: `plans/{plan-slug}/{category}/{original-filename}`
5. Go uploads to R2
6. Go inserts file record in database
7. Go returns file records
8. UI updates automatically

### 5.3 Presigned URL Flow

Files are never served directly through the Go backend. Instead:

1. Frontend requests a presigned URL via `GET /api/files/:id/url`
2. Go generates a presigned R2 URL valid for 60 minutes
3. Frontend uses URL directly for image preview src or download href
4. For ZIP downloads, Go streams the ZIP directly — fetching files from R2 and writing to the response stream without buffering the whole ZIP in memory

---

## 6. Database

### 6.1 Connection

- Driver: `pgx` (Go PostgreSQL driver)
- Connection pooling via `pgxpool`
- Connection string from environment variable `DATABASE_URL`
- Migrations run automatically on server startup using `golang-migrate`

### 6.2 Migration Strategy

- Migrations live in `/backend/migrations/`
- Named sequentially: `001_initial_schema.up.sql`, `001_initial_schema.down.sql`
- Run automatically on server start
- Never edit a migration that has been deployed — always add a new one

---

## 7. R2 Storage

### 7.1 Client

- AWS SDK v2 for Go (`github.com/aws/aws-sdk-go-v2`)
- R2 is S3-compatible — point the SDK at the R2 endpoint
- Credentials and endpoint from environment variables

### 7.2 Environment Variables

```
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=
R2_ENDPOINT=https://{account_id}.r2.cloudflarestorage.com
```

### 7.3 Storage Structure

```
plans/
  {plan-slug}/
    website/
      render-front.jpg
      elevation-front.jpg
      elevation-left.jpg
      elevation-rear.jpg
      elevation-right.jpg
      floor-plan-main.jpg
      floor-plan-upper.jpg
      floor-plan-lower.jpg
      poster.jpg
    reference/
      {original-filename}
    technical/
      {original-filename}
    3d/
      {original-filename}
    other/
      {original-filename}
```

### 7.4 Presigned URLs

- Generated server-side using AWS SDK v2 presigner
- Valid for 60 minutes
- Frontend never receives permanent R2 URLs
- Presigned URLs requested on demand, not stored

---

## 8. SFTP

### 8.1 Implementation

SFTPGo runs as a dedicated Docker container with R2 configured as its storage backend. This means SFTP users are reading and writing directly to the R2 bucket — no local disk involved.

### 8.2 Credential Management

- SFTP users created and managed through the Go API
- SFTPGo admin API called internally by Go when creating/updating/deleting SFTP credentials
- SFTP credentials stored in SFTPGo's own database (separate from the main Postgres instance)
- Admins can view, rotate, and revoke SFTP credentials per user from the app UI

### 8.3 Permissions

- Read/write — full access to bucket
- Read-only — can browse and download, cannot upload or delete

---

## 9. Export

### 9.1 CSV Export

- Generated server-side in Go using `encoding/csv`
- Streamed directly to response — no temp file written to disk
- Two presets:
  - **WP All Import** — columns and filenames formatted to match WP All Import template
  - **General** — all fields, human-readable column names
- Custom export — user selects which fields to include

### 9.2 ZIP Export

- Generated server-side in Go using `archive/zip`
- Files fetched from R2 concurrently using goroutines
- Written to response stream progressively — no full ZIP buffered in memory
- Folder structure mirrors R2 storage structure

---

## 10. Error Handling

### 10.1 API Errors

All API errors return a consistent JSON shape:

```json
{
	"error": {
		"code": "PLAN_NOT_FOUND",
		"message": "No plan found with the given ID"
	}
}
```

### 10.2 HTTP Status Codes

| Status | When                             |
| ------ | -------------------------------- |
| 200    | Success                          |
| 201    | Created                          |
| 400    | Bad request / validation error   |
| 401    | Not authenticated                |
| 403    | Authenticated but not authorized |
| 404    | Resource not found               |
| 409    | Conflict (e.g. duplicate slug)   |
| 500    | Internal server error            |

### 10.3 Frontend Error Handling

- TanStack Query surfaces loading and error states automatically
- Global error boundary catches unexpected errors
- Toast notifications for user-facing errors and success confirmations
- Form validation client-side before submission, server errors shown inline

---

## 11. Environment Variables

### Backend

```
# Server
PORT=8080
ENV=development | production
JWT_SECRET=
JWT_REFRESH_SECRET=

# Database
DATABASE_URL=postgres://user:password@postgres:5432/homeplanner

# R2
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=
R2_ENDPOINT=

# SFTPGo
SFTPGO_API_URL=http://sftpgo:8081
SFTPGO_API_KEY=
```

### Frontend

```
PUBLIC_API_URL=http://localhost:8080
```

---

## 12. Performance Considerations

- Postgres queries use indexes on `plans.slug`, `plans.status`, `files.plan_id`, `files.slot`
- File uploads streamed directly to R2 — not buffered in Go memory
- ZIP generation uses concurrent R2 fetches via goroutines with a worker pool to avoid overwhelming R2
- Presigned URLs cached client-side by TanStack Query for their validity window
- Plans list paginated server-side — default 50 per page
- Full-text search on plan name uses Postgres `tsvector` index

---

## 13. Security

- Passwords hashed with bcrypt (cost factor 12)
- JWT secrets minimum 256-bit random strings
- All cookies httpOnly, Secure, SameSite=Strict
- File uploads validated for type and size before touching R2
- R2 bucket is private — no public access, all access via presigned URLs
- CORS restricted to frontend origin only
- SQL queries use parameterized statements throughout — no string concatenation
- Admin routes protected by role middleware in addition to auth middleware

# Project Phases & Task List

## Natural Element Homes — Home Planner

---

## Phase 1 — Backend Foundation

### 1.1 Database & Migrations

- [x] Wire golang-migrate into main.go — run migrations on startup
- [x] Confirm migration runs cleanly against local Postgres
- [x] Seed script — create first admin user from env variables on first run

### 1.2 Project Structure

- [x] Stub out all internal packages — handlers, middleware, models, storage
- [x] Confirm project compiles cleanly with empty stubs

### 1.3 Auth

- [x] User model — struct matching users table
- [x] Password hashing with bcrypt
- [x] JWT access token generation (15 min)
- [x] JWT refresh token generation (7 days)
- [x] `POST /api/auth/login` handler
- [x] `POST /api/auth/refresh` handler
- [x] `POST /api/auth/logout` handler
- [x] `GET /api/auth/me` handler
- [x] Auth middleware — verify JWT from cookie, attach user to context
- [x] Role middleware — admin-only route protection
- [x] Refresh token storage in database
- [x] Expired refresh token cleanup on startup

### 1.4 Users

- [x] `GET /api/users` — list all users (admin)
- [x] `POST /api/users` — create user (admin)
- [x] `PUT /api/users/:id` — update user (admin)
- [x] `PUT /api/users/:id/password` — update password (admin)
- [x] `DELETE /api/users/:id` — delete user (admin)

---

## Phase 2 — Plans API

- [x] Plan model — struct matching plans table
- [x] Slug generation from plan name
- [x] `GET /api/plans` — list with search, filter, sort, pagination
- [x] `GET /api/plans/:id` — single plan with files
- [x] `POST /api/plans` — create plan
- [x] `PUT /api/plans/:id` — update plan metadata
- [x] `DELETE /api/plans/:id` — soft delete (admin)
- [x] `POST /api/plans/:id/restore` — restore soft deleted plan (admin)
- [x] `POST /api/plans/:id/duplicate` — duplicate metadata only
- [x] `PUT /api/plans/:id/flag` — flag plan
- [x] `PUT /api/plans/:id/unflag` — unflag plan
- [x] Completeness calculation — recalculate and update status on file change
- [x] Activity log writes on all plan mutations

---

## Phase 3 — File Handling & R2

### 3.1 R2 Client

- [x] R2 storage client setup using AWS SDK v2
- [x] Upload file to R2
- [x] Delete file from R2
- [x] Generate presigned URL (60 min expiry)

### 3.2 File API

- [x] File model — struct matching files table
- [x] `POST /api/plans/:id/files/website` — upload slotted image
  - [x] Validate file type (images only)
  - [x] Validate file size (50MB max)
  - [x] Validate slot name
  - [x] Upsert slot — replace existing if slot already filled
  - [x] Trigger completeness recalculation
- [x] `POST /api/plans/:id/files` — upload unslotted files
  - [x] Validate file size (500MB max)
  - [x] Validate category is not website
  - [x] Support multiple file upload
- [x] `GET /api/plans/:id/files` — list all files for plan
- [x] `GET /api/files/:id/url` — get presigned URL
- [x] `DELETE /api/files/:id` — delete file from DB and R2
- [x] Activity log writes on all file mutations

---

## Phase 4 — Export & Import

### 4.1 Export

- [x] `GET /api/export/csv` — CSV export
  - [x] WP All Import preset
  - [x] General preset
  - [x] Custom field selection
  - [x] Filter by plan IDs
  - [x] Stream response directly
- [x] `GET /api/export/zip` — ZIP export
  - [x] Concurrent R2 fetches via goroutine worker pool
  - [x] Stream ZIP to response without buffering in memory
  - [x] Filter by plan IDs and categories
  - [x] Maintain folder structure in ZIP

### 4.2 Import

- [x] `POST /api/import/csv/preview` — parse CSV, return preview + suggested mapping
- [x] `POST /api/import/csv` — bulk create/update plans from CSV
  - [x] Column mapping support
  - [x] Create only / Update only / Upsert modes
  - [x] Per-row error reporting
  - [x] Skip bad rows, import good ones

---

## Phase 5 — Activity Log

- [x] `GET /api/activity` — global activity log (admin)
- [x] `GET /api/plans/:id/activity` — per-plan activity log
- [x] Filter by user, action, plan
- [x] Pagination

---

## Phase 6 — Frontend Foundation

### 6.1 Setup

- [x] SvelteKit project confirmed running
- [x] Tailwind configured
- [x] shadcn-svelte initialized
- [x] TanStack Query wired into root layout
- [x] API client module — base fetch wrapper with cookie handling and error parsing
- [x] TypeScript types matching API response shapes

### 6.2 Auth

- [x] Login page — `/login`
- [x] Auth store — current user, login state
- [x] Route guard — redirect to login if not authenticated
- [x] Logout functionality
- [x] JWT expiry handling — auto refresh on 401

### 6.3 Layout

- [x] Authenticated shell layout — sidebar + main content
- [x] Sidebar with nav items, user info, logout
- [x] Mobile sidebar — hamburger + drawer
- [x] Toast notification system

---

## Phase 7 — Frontend Pages

### 7.1 Dashboard

- [x] Stats cards — total, complete, incomplete, flagged
- [x] Recent activity feed
- [x] Recent plans list
- [x] Quick action buttons

### 7.2 Plans List

- [x] Table view with all columns
- [x] Grid view with poster image
- [x] Live search (debounced)
- [x] Filter panel — all filter options
- [x] Sort controls
- [x] Bulk select + bulk action bar
- [x] Pagination
- [x] Empty states
- [x] New Plan modal

### 7.3 Plan Detail

- [x] Header — name, status badge, actions
- [x] Overview tab
  - [x] Specs section — view and edit mode
  - [x] Website images grid — all 9 slots
  - [x] Slot cards — filled and empty states
- [x] Image preview modal - [x] Slot upload, replace, delete
- [x] Files tab
  - [x] Sub-tabs per category
  - [x] File list per category
  - [x] Drag and drop upload zone
  - [x] Upload progress
  - [x] File download and delete
- [x] Activity tab
  - [x] Chronological activity list
  - [x] Pagination

### 7.4 Import Page

- [x] Step 1 — CSV upload
- [x] Step 2 — Column mapping UI
- [x] Step 3 — Review and confirm
- [x] Step 4 — Result summary

### 7.5 Activity Log Page

- [x] Table with filters
- [x] Pagination

### 7.6 Settings Page

- [x] Users management — list, create, edit, delete
- [ ] SFTP credentials per user — generate, revoke, permission level
- [x] Export preset reference

---

## Phase 8 — Export UI

- [x] Export modal — preset selector, field picker, scope selector
- [x] CSV download trigger
- [x] ZIP download trigger
- [x] Bulk export from plans list selection

---

## Phase 9 — SFTPGo Integration

- [ ] SFTPGo container in Docker Compose
- [ ] SFTPGo configured with R2 as storage backend
- [ ] Go service layer wrapping SFTPGo admin API
- [ ] Create SFTP user when app user is created
- [ ] Delete SFTP user when app user is deleted
- [ ] Generate / rotate credentials endpoint
- [ ] Revoke credentials endpoint
- [ ] Permission level (read / read+write) management
- [ ] SFTP credentials display in Settings UI

---

## Phase 10 — Docker & Deployment

- [ ] Backend Dockerfile
- [ ] Frontend Dockerfile
- [ ] Multi-stage Dockerfile combining both
- [ ] docker-compose.yml — production
- [ ] docker-compose.dev.yml — development dependencies only
- [ ] .env.example — all required variables documented
- [ ] Go server configured to serve built SvelteKit static files
- [ ] Health check endpoint confirmed working in container
- [ ] Test full production build locally
- [ ] Confirm migrations run on container startup

---

## Phase 11 — Polish & QA

- [ ] All empty states implemented
- [ ] All error states implemented
- [ ] All loading states / skeletons implemented
- [ ] Confirm all destructive actions have confirmation dialogs
- [ ] Unsaved changes warnings on plan edit
- [ ] Mobile layout tested
- [ ] Test full CSV import with real data
- [ ] Test full CSV export — WP All Import preset verified
- [ ] Test ZIP export with large file set
- [ ] Test SFTP connection end to end
- [ ] Test auth flow — login, refresh, logout, expired token
- [ ] Test role restrictions — editor cannot access admin routes
- [ ] Cross-browser check — Chrome, Firefox, Safari

---

## Suggested Build Order

1. Phase 1 — get auth working end to end, confirmed with a real login
2. Phase 2 — plans CRUD working in Postman/curl
3. Phase 3 — file upload and R2 confirmed working
4. Phase 4 — export CSV confirmed matching expected format
5. Phase 6 — frontend auth and layout shell
6. Phase 7 — pages built against working API
7. Phase 4 import — after frontend exists to build the UI against
8. Phase 5 — activity log (low priority, add as you go)
9. Phase 8 — export UI
10. Phase 9 — SFTPGo (last backend feature, most isolated)
11. Phase 10 — Docker, test full build
12. Phase 11 — polish and QA

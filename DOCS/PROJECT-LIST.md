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
    - [x] Includes plan metadata fields
    - [x] Includes image slot columns (Render Front, Elevation Front, Elevation Left, Elevation Rear, Elevation Right, Floor Plan Main, Floor Plan Upper, Floor Plan Lower, Poster)
    - [x] Matches format of neh-home-plans.csv
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
- [x] SFTP credentials per user — generate, revoke, permission level
- [x] Export preset reference

---

## Phase 8 — Export UI

- [x] Export modal — preset selector, field picker, scope selector
- [x] CSV download trigger
- [x] ZIP download trigger
- [x] Bulk export from plans list selection

---

## Phase 9 — SFTPGo Integration

- [x] SFTPGo container in Docker Compose
- [x] SFTPGo configured with R2 as storage backend
- [x] Go service layer wrapping SFTPGo admin API
- [x] Create SFTP user when app user is created
- [x] Delete SFTP user when app user is deleted
- [x] Generate / rotate credentials endpoint
- [x] Revoke credentials endpoint
- [x] Permission level (read / read+write) management
- [x] SFTP credentials display in Settings UI

---

## Phase 10 — Docker & Deployment

- [x] Backend Dockerfile
- [x] Frontend Dockerfile
- [x] Multi-stage Dockerfile combining both
- [x] docker-compose.yml — production
- [x] docker-compose.dev.yml — development dependencies only
- [x] .env.example — all required variables documented
- [x] Go server configured to serve built SvelteKit static files
- [x] Health check endpoint confirmed working in container
- [x] Test full production build locally
- [x] Confirm migrations run on container startup

---

## Phase 11 — Polish & QA

- [x] All empty states implemented
- [x] All error states implemented
- [x] All loading states / skeletons implemented
- [x] Confirm all destructive actions have confirmation dialogs
- [x] Unsaved changes warnings on plan edit
- [x] Mobile layout tested
- [ ] Test full CSV import with real data
- [ ] Test full CSV export — WP All Import preset verified
- [ ] Test ZIP export with large file set
- [ ] Test SFTP connection end to end
- [ ] Test auth flow — login, refresh, logout, expired token
- [ ] Test role restrictions — editor cannot access admin routes
- [ ] Cross-browser check — Chrome, Firefox, Safari

---

## Phase 12 — Post-Import Bulk Image Assignment

### 12.1 Backend API

- [x] Modify `POST /api/import/csv` to return created plan IDs in response
- [x] Create `GET /api/import/recent` endpoint — fetch recently imported plans (last 24h)
  - [x] Create `POST /api/plans/bulk-files` endpoint — upload images to multiple plans
  - [x] Accept array of `{plan_id, slot, file}` objects
  - [x] Validate all plan IDs exist
  - [x] Validate slot names are valid website slots
  - [x] Process each file with image optimization pipeline
  - [x] Return per-file success/failure results

### 12.2 Frontend - Import Flow Enhancement

- [x] Add Step 5 "Upload Images" to import wizard
- [x] Create bulk image assignment page component
  - [x] Grid view of all imported plans from CSV
  - [x] Each plan card shows: name, slug, status, empty slot placeholders
  - [x] Bulk drag-drop upload zone at top
  - [x] Smart filename matching (optional) — match `{plan-slug}--{slot}` pattern
  - [x] Manual slot assignment per uploaded file
  - [x] Progress tracking per plan
  - [x] Completion summary

### 12.3 Smart Matching Logic

- [x] Parse filenames for pattern: `{plan-slug}--{slot-type}--{view}` or `{plan-slug}--{slot}`
- [x] Auto-suggest slot assignments based on filename
- [x] Show match confidence score
- [x] Allow manual override of auto-matches
- [x] Support batch assignment — select multiple files, assign to same slot across different plans

---

## Phase 13 — Image Optimization & Processing

### 13.1 Image Processing Service

- [x] Add image processing library to Go backend
  - [x] Add `github.com/disintegration/imaging` dependency
- [x] Create `internal/processing/image.go` service
  - [x] `ProcessWebsiteImage()` — resize, compress, format conversion
  - [x] `ProcessReferenceFile()` — basic validation only
  - [x] `DetectTransparency()` — check if PNG has alpha channel
  - [x] `ConvertPNGToJPEG()` — white background replacement

### 13.2 Size Limits & Constraints

**Website Images (render, elevations, floor plans, poster):**

- [x] Max file size: 5MB (down from 50MB)
- [x] Max dimensions: 4000px width/height (4K display)
- [x] Output format: JPEG only (no WebP)
- [x] JPEG quality: 90%
- [x] Strip all metadata (EXIF)

**Poster Images:**

- [x] Max file size: 5MB
- [x] Target: 8x12 print @ 300dpi = 2400x3600px
- [x] Max dimensions: 4000px
- [x] JPEG quality: 90%

**Reference/Technical/3D Files:**

- [x] Max file size: 50MB (down from 500MB)
- [x] No image processing, just validation

### 13.3 Naming Constraints & Standardization

- [x] Create filename validation utility
  - [x] Pattern: `{plan-slug}--{slot-type}--{view}.{ext}` or `{plan-slug}--{slot}.{ext}`
  - [x] Valid extensions: .jpg, .jpeg, .png
  - [x] Max filename length: 100 characters
  - [x] Lowercase only, no spaces
  - [x] Allowed characters: a-z, 0-9, hyphen, underscore
- [x] Auto-rename uploaded files to match convention
- [x] Storage key format: `plans/{slug}/website/{slot}.jpg`

### 13.4 PNG/JPEG Duplicate Handling

- [x] When PNG uploaded for website slot:
  - [x] Detect if transparency exists
  - [x] If no transparency: convert to JPEG with white background, use JPEG for website
  - [x] If transparency exists: reject for website slots (show error)
  - [ ] Always keep original PNG in "other" category if uploaded (deferred to Phase 12)
- [ ] When JPEG exists and PNG uploaded:
  - [ ] Use JPEG for website slot
  - [ ] Store PNG in "other" category
  - [ ] Show info message: "Using JPEG for website, PNG saved to Other Files"
- [x] No auto-deletion — keep both formats in storage

### 13.5 Processing Pipeline

**Website Image Upload Flow:**

1. [x] Receive file
2. [x] Validate file type (image/jpeg, image/png only)
3. [x] Validate file size (≤5MB after processing)
4. [x] Detect format and transparency (PNG only)
5. [x] Resize if >4000px dimension
6. [x] Convert to JPEG:
   - [x] PNG with transparency → white background
   - [x] PNG without transparency → direct conversion
   - [x] JPEG → re-encode at 90% quality
7. [x] Strip metadata
8. [x] Validate output size ≤5MB
9. [x] Upload to R2
10. [x] Save file record

**Reference File Upload Flow:**

1. [x] Receive file
2. [x] Validate file size ≤50MB
3. [x] Upload to R2 as-is
4. [x] Save file record

### 13.6 Frontend Updates

- [x] Update upload components to show new size limits
- [x] Show processing progress indicator
- [x] Display warnings for:
  - [x] Files >5MB will be compressed
  - [x] PNGs with transparency cannot be used for website (will be stored as "other")
  - [x] Original filename will be standardized
- [x] Add file size preview before upload
- [x] Show optimization savings (original vs processed size)

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
13. Phase 13 — Image optimization (new uploads only, affects Phase 3/12)
14. Phase 12 — Post-import bulk image assignment

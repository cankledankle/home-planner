# Product Requirements Document

## Natural Element Homes — Home Planner

**Version:** 1.0
**Status:** Draft

---

## 1. Overview

### 1.1 Purpose

Home Planner is an internal web application for Natural Element Homes to manage their home plan catalog. It replaces a fragmented workflow involving Excel spreadsheets and Dropbox with a single, structured tool that any team member can use regardless of technical ability.

### 1.2 Background

Natural Element Homes maintains a catalog of 230+ home plan models. Each plan has associated metadata (square footage, bed/bath counts, etc.) and a large library of files including rendered images, elevation drawings, floor plans, technical drawings, reference photos, and 3D model assets. Previously this data lived across an Excel file and a Dropbox folder exceeding 500GB. Keeping the two in sync was manual, error-prone, and effectively limited to one person.

### 1.3 Goals

- Give the full team a simple, structured way to manage home plan data
- Enforce file organization without requiring technical knowledge
- Make it easy to find, preview, and download any file associated with any plan
- Support export workflows for publishing plans to their website
- Be deployable on a VPS or internal server with minimal configuration

### 1.4 Non-Goals

- Direct WordPress integration
- Public-facing website functionality
- AI-assisted image classification
- Mobile app
- Multi-tenant / multi-company support

---

## 2. Users

### 2.1 Roles

**Admin**

- Full access to all features
- Can create, edit, and delete plans
- Can manage users
- Can access SFTP credentials and system settings

**Editor**

- Can create and edit plans
- Can upload and delete files
- Cannot manage users or access system settings
- Cannot delete plans

### 2.2 User Stories

**As an editor, I want to:**

- Search for a home plan by name so I can find it quickly
- See at a glance which plans are missing required images
- Upload files to a plan by dragging and dropping them
- Assign uploaded images to the correct slot (elevation front, floor plan main, etc.)
- Edit a plan's specs without affecting its files
- Download all files for a plan as a ZIP
- See a preview of any image without downloading it

**As an admin, I want to:**

- Import an existing CSV to seed the database without manual data entry
- Export a CSV of all plans formatted for WP All Import
- Add and remove team members
- See who uploaded or changed what and when
- Access plan files via SFTP for bulk operations
- Download all files across all plans at once

---

## 3. Features

### 3.1 Authentication

- Email and password login
- JWT-based auth, tokens stored in httpOnly cookies
- Session expiry with refresh token support
- Password reset via email (phase 2)
- No self-registration — accounts created by admin only

### 3.2 Plan Management

**Plan record contains:**

- Name
- Slug (auto-generated from name, editable)
- Plan Type (Single Level / Multi-Level)
- Plan Style (Cabin / Lodge / Modern / Ranch / Farmhouse)
- Bedrooms
- Bathrooms
- Half Bathrooms
- Main Level SF
- Upper Level SF
- Lower Level SF
- Porch & Deck SF
- Garage SF
- Garage Apartment SF
- Unfinished SF
- Heated SF
- Total SF
- Status (Complete / Incomplete / Flagged)
- Notes (free text)
- Created at
- Updated at
- Created by
- Updated by

**Operations:**

- Create plan
- Edit plan
- Delete plan (admin only, soft delete)
- Duplicate plan (copies metadata, not files)
- Flag / unflag plan
- View plan detail

### 3.3 File Management

**File categories:**

- Website Images — slotted, required slots tracked
- Reference — unstructured, any file type
- Technical Drawings — unstructured, PDFs and images
- 3D Assets — unstructured, any file type
- Other — catch-all

**Website Image Slots:**

- `render-front` — required
- `elevation-front` — required
- `elevation-left` — required
- `elevation-rear` — required
- `elevation-right` — required
- `floor-plan-main` — required
- `floor-plan-upper` — optional
- `floor-plan-lower` — optional
- `poster` — required

**File operations:**

- Upload single or multiple files
- Drag and drop upload
- Assign image to slot on upload or after
- Replace a slot image
- Preview images inline
- Download individual file
- Download all files for a plan as ZIP
- Download all files across all plans as ZIP (admin)
- Delete file
- View file metadata (name, size, type, uploaded by, uploaded at)

**Storage:**

- Files stored in Cloudflare R2
- Accessed via presigned URLs for browser previews and downloads
- Storage path structure: `/plans/{plan-slug}/{category}/{filename}`

### 3.4 Completeness Tracking

A plan's completeness status is calculated automatically based on whether all required website image slots are filled:

- **Complete** — all required slots filled
- **Incomplete** — one or more required slots empty
- **Flagged** — manually marked for attention regardless of slot status

Status is always visible on the plans list and plan detail views.

### 3.5 Search and Filtering

**Search:**

- Full-text search by plan name

**Filters:**

- Status (Complete / Incomplete / Flagged)
- Plan Type (Single Level / Multi-Level)
- Plan Style
- Bedrooms (min/max)
- Bathrooms (min/max)
- Heated SF (min/max)
- Has / missing specific file slots

**Sorting:**

- Name (A-Z / Z-A)
- Heated SF
- Total SF
- Date created
- Date updated

### 3.6 Export

**CSV Export:**

- Export all plans or selected plans
- Choose which fields to include
- Preset export format for WP All Import
- Preset export format for general use
- Download immediately

**Bulk File Download:**

- Download all files for selected plans as a single ZIP
- Download all files for all plans as a single ZIP
- ZIP maintains folder structure: `/{plan-name}/{category}/`

### 3.7 CSV Import

- Upload a CSV to bulk-create or update plan records
- Column mapping UI — map CSV columns to plan fields
- Preview before committing
- Errors shown per row — skip bad rows, import good ones
- Does not import files — metadata only

### 3.8 SFTP Access

- SFTP server runs as a separate Docker service
- Provides direct read/write access to the R2 bucket contents
- Credentials managed in admin settings
- Each user can have their own SFTP credentials
- Read-only and read/write permission levels
- Credentials rotatable by admin at any time

### 3.9 Activity Log

- Every create, update, delete, and file upload action is logged
- Log shows user, action type, plan affected, timestamp
- Viewable per plan on the plan detail page
- Viewable globally in admin settings
- Not editable or deletable

---

## 4. Views / Pages

### 4.1 Login

- Email and password form
- Error messaging for invalid credentials
- No self-registration link

### 4.2 Dashboard

- Total plan count
- Count by status — Complete / Incomplete / Flagged
- Recently updated plans (last 10)
- Recently uploaded files (last 10)
- Quick action buttons — New Plan, Import CSV, Export CSV

### 4.3 Plans List

- Paginated table of all plans
- Columns: Name, Type, Style, Beds, Baths, Heated SF, Status, Last Updated
- Inline status badge per row
- Search bar (live, by name)
- Filter panel — Status, Type, Style, Beds, Baths, Heated SF range, slot completeness
- Sort by any column
- Bulk select with actions — Export CSV, Download Files as ZIP, Flag, Unflag
- Click row to open Plan Detail
- New Plan button

### 4.4 Plan Detail

Divided into clear sections:

**Header**

- Plan name (editable inline)
- Slug
- Status badge
- Flag / unflag button
- Duplicate plan button
- Delete plan button (admin only)
- Last updated by / at

**Specs**

- All metadata fields in an editable form
- Save and Cancel buttons
- Unsaved changes warning on navigation

**Website Images**

- Grid of all 9 slots
- Filled slots show thumbnail, filename, file size, upload date
- Empty required slots highlighted with a clear upload prompt
- Empty optional slots shown with a softer empty state
- Click slot to replace image
- Click image to preview full size
- Download button per slot

**Other Files**

- Tabbed by category — Reference / Technical / 3D Assets / Other
- File list per tab showing name, size, type, uploaded by, uploaded at
- Upload button per tab

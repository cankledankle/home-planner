# UI Specification

## Natural Element Homes — Home Planner

**Version:** 1.0
**Status:** Draft

---

## 1. General

### 1.1 Design Principles

- Clean, professional, minimal — this is a work tool not a consumer app
- Status and completeness always visible — never hunt for what's missing
- Errors and empty states are always explicit — no blank spaces without context
- Destructive actions always require confirmation
- Unsaved changes always warn before navigation

### 1.2 Tech

- SvelteKit
- Tailwind CSS
- shadcn-svelte components
- TanStack Query for all data fetching

### 1.3 Layout

Two layout modes:

**Unauthenticated** — centered card, no navigation
**Authenticated** — fixed left sidebar + main content area

### 1.4 Color System

- Use shadcn default theme
- Status colors:
  - Complete — green
  - Incomplete — yellow/amber
  - Flagged — red

### 1.5 Typography

- Plan names — semibold, larger
- Labels — muted, smaller
- Metadata values — regular weight

### 1.6 Notifications

- Toast notifications for all async actions (success and error)
- Positioned bottom-right
- Auto-dismiss after 4 seconds
- Errors stay until dismissed

---

## 2. Navigation

### 2.1 Sidebar

Fixed left sidebar, always visible when authenticated.

**Top section:**

- App logo / name — "Home Planner"
- Natural Element Homes branding

**Nav items:**

- Dashboard
- Plans
- Import (admin only)
- Activity (admin only)
- Settings (admin only)

**Bottom section:**

- Current user name + role badge
- Logout button

### 2.2 Mobile

- Sidebar collapses to a hamburger menu on small screens
- Overlay drawer on open

---

## 3. Pages

### 3.1 Login — `/login`

**Layout:** Centered card, full-height page

**Contents:**

- App name + logo at top
- "Natural Element Homes" subtitle
- Email input
- Password input
- Sign In button
- Error message inline below form on failed login

**Behavior:**

- On success → redirect to Dashboard
- On error → show inline error message, clear password field
- Email field focused on load
- Enter key submits form

---

### 3.2 Dashboard — `/`

**Layout:** Authenticated shell

**Contents:**

**Stats row — 4 cards:**

- Total Plans (count)
- Complete (count + % of total)
- Incomplete (count)
- Flagged (count)

Each card is clickable and navigates to Plans filtered to that status.

**Recent Activity — last 10 activity log entries:**

- User avatar/initials
- Action description e.g. "Jane uploaded render-front to Abilene"
- Timestamp (relative e.g. "2 hours ago")
- Link to the plan if applicable

**Recent Plans — last 10 updated plans:**

- Plan name
- Status badge
- Last updated time
- Click to open plan detail

**Quick Actions:**

- New Plan button → opens New Plan modal
- Import CSV button (admin only) → navigates to Import page
- Export CSV button → opens Export modal

---

### 3.3 Plans List — `/plans`

**Layout:** Authenticated shell, full width

**Header:**

- Page title "Plans"
- New Plan button (right side)

**Toolbar:**

- Search input — live search by name, debounced 300ms
- Filter button → opens filter panel
- Sort dropdown — Name, Heated SF, Total SF, Date Created, Date Updated + Asc/Desc toggle
- View toggle — Table / Grid (grid shows poster image if available)
- Export button → opens Export modal
- Selected count + bulk action bar (appears when rows selected)

**Filter Panel (slide-out or inline):**

- Status — checkbox group (Complete, Incomplete, Flagged)
- Type — checkbox group (Single Level, Multi-Level)
- Style — checkbox group (Cabin, Lodge, Modern, Ranch, Farmhouse)
- Beds — min/max number inputs
- Baths — min/max number inputs
- Heated SF — min/max number inputs
- Missing Slot — dropdown of slot names
- Clear All button
- Apply button

**Table View columns:**

- Checkbox (for bulk select)
- Name (sortable, links to plan detail)
- Type
- Style
- Beds
- Baths
- Heated SF (sortable)
- Total SF (sortable)
- Status badge
- Last Updated (sortable)
- Actions menu (⋯) — Edit, Duplicate, Flag/Unflag, Delete (admin)

**Grid View:**

- Card per plan
- Poster image or placeholder if missing
- Plan name
- Status badge
- Beds / Baths / Heated SF summary
- Click card to open plan detail

**Bulk Actions Bar (appears on selection):**

- "{n} plans selected"
- Export CSV
- Download as ZIP
- Flag
- Unflag
- Clear selection

**Empty States:**

- No plans exist → "No plans yet. Import a CSV or create your first plan."
- No results for search/filter → "No plans match your filters." + Clear Filters button

**Pagination:**

- Bottom of list
- Previous / Next buttons
- Page indicator "Page 1 of 5"
- Records per page selector (25, 50, 100)

---

### 3.4 Plan Detail — `/plans/:id`

**Layout:** Authenticated shell

**Header:**

- Back button → Plans list
- Plan name (editable inline — click to edit)
- Status badge
- Flag / Unflag button
- Actions dropdown — Duplicate, Delete (admin)
- Last updated by + timestamp

**Tabs:**

- Overview
- Files
- Activity

---

#### 3.4.1 Overview Tab

**Specs section:**

- All plan metadata fields in a two-column grid
- Edit button → switches fields to editable inputs
- Save / Cancel buttons when editing
- Unsaved changes warning on tab switch or navigation

**Fields displayed:**

- Plan Type
- Plan Style
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
- Notes (full width textarea)

**Website Images section:**

- Grid of 9 slots, 3 per row
- Each slot card shows:
  - Slot label (e.g. "Elevation — Front")
  - Required / Optional badge
  - If filled: thumbnail image, filename, file size, upload date, Download button, Replace button, Delete button
  - If empty + required: amber/yellow background, "Missing" label, Upload button
  - If empty + optional: muted background, "Not uploaded" label, Upload button
- Click thumbnail → opens full-size image preview modal

---

#### 3.4.2 Files Tab

**Four sub-tabs:**

- Reference
- Technical
- 3D Assets
- Other

Each sub-tab:

- Upload button (drag and drop zone at top, or click to browse)
- File list table:
  - Filename
  - File type
  - Size
  - Uploaded by
  - Uploaded at
  - Download button
  - Delete button
- Empty state: "No files uploaded yet."
- Multiple file upload supported
- Upload progress shown per file

---

#### 3.4.3 Activity Tab

- Chronological list of all activity for this plan
- Each entry:
  - User initials avatar
  - Action description
  - Detail if relevant (e.g. "replaced elevation-front")
  - Timestamp (relative + absolute on hover)
- Paginated, 50 per page

---

### 3.5 Import — `/import`

**Auth:** Admin only

**Layout:** Authenticated shell

**Step 1 — Upload CSV:**

- Large drag and drop zone
- Or click to browse
- Accepts `.csv` only
- On upload → preview shown automatically

**Step 2 — Map Columns:**

- Table showing detected CSV columns on the left
- Dropdown per column to map to a plan field
- Suggested mapping auto-applied based on column name matching
- Unmapped columns shown in muted style — will be ignored
- Preview of first 5 rows shown below

**Step 3 — Review:**

- Import mode selector — Create Only / Update Only / Upsert
- Summary: "{n} rows detected, {n} will be created, {n} will be updated"
- Any detected issues shown as warnings per row
- Import button

**Step 4 — Result:**

- Created: {n}
- Updated: {n}
- Skipped: {n}
- Errors list (if any) — row number + reason
- "Go to Plans" button

---

### 3.6 Activity Log — `/activity`

**Auth:** Admin only

**Layout:** Authenticated shell

**Header:**

- Page title "Activity Log"

**Filters:**

- User dropdown
- Action type dropdown
- Date range picker

**Table:**

- User
- Action
- Plan (linked)
- Detail
- Timestamp

**Pagination:**

- 50 per page

---

### 3.7 Settings — `/settings`

**Auth:** Admin only

**Layout:** Authenticated shell

**Sub-sections:**

**Users:**

- Table of all users — Name, Email, Role, Created
- Invite User button → modal with name, email, password, role fields
- Edit button per user → modal
- Delete button per user → confirmation dialog
- Cannot delete own account

**SFTP:**

- Instructions for connecting via SFTP
- Host, port displayed
- Table of SFTP credentials per user
- Generate Credentials button per user
- Revoke button per user
- Permission level toggle (Read / Read+Write) per user

**Export Presets:**

- View the WP All Import column mapping
- View the General export column mapping
- Phase 2 — allow custom presets to be saved

---

## 4. Modals

### 4.1 New Plan Modal

- Triggered from Dashboard quick action or Plans list header button
- Fields: Name (required), Type, Style, Beds, Baths, Half Baths
- Optional fields collapsed under "Add more details" toggle
- Create button → creates plan, redirects to Plan Detail
- Cancel button

### 4.2 Export Modal

- Triggered from Dashboard or Plans list
- Preset selector — WP All Import / General / Custom
- If Custom → field checklist
- Scope — All Plans / Selected Plans (if triggered from bulk select)
- Download button → triggers CSV download immediately
- Option to also download files as ZIP

### 4.3 Image Preview Modal

- Triggered by clicking a thumbnail in the website images grid
- Full-size image centered in modal
- Slot label and filename shown below
- Download button
- Replace button
- Close button / click outside to close

### 4.4 Confirm Delete Modal

- Triggered before any delete action
- Title: "Delete {resource name}?"
- Body: clear explanation of what will be deleted and whether it is recoverable
- Cancel button (default focus)
- Delete button (destructive red)

### 4.5 Upload Progress Modal

- Shown during file uploads
- Progress bar per file
- Filename and size shown
- Cancel button (cancels in-flight uploads)
- Auto-closes on completion, replaced by success toast

---

## 5. Empty States

| Context                        | Message                                                 |
| ------------------------------ | ------------------------------------------------------- |
| Plans list — no plans          | "No plans yet. Import a CSV or create your first plan." |
| Plans list — no results        | "No plans match your filters." + Clear Filters button   |
| File tab — no files            | "No files uploaded yet." + Upload button                |
| Activity tab — no activity     | "No activity recorded for this plan yet."               |
| Activity log — no results      | "No activity found for the selected filters."           |
| Dashboard — no recent activity | "No recent activity."                                   |

---

## 6. Error States

| Context               | Behavior                                      |
| --------------------- | --------------------------------------------- |
| API request fails     | Toast notification with error message         |
| Form validation fails | Inline error below each invalid field         |
| Login fails           | Inline error in login card, password cleared  |
| File upload fails     | Error shown in upload progress modal per file |
| Page not found        | 404 page with back button                     |
| Unauthorized          | Redirect to login                             |
| Forbidden             | 403 page with back button                     |
| Server error          | 500 page with retry button                    |

---

## 7. Loading States

- Plans list — skeleton rows while loading
- Plan detail — skeleton layout while loading
- Images — skeleton placeholder per slot while loading
- File list — skeleton rows while loading
- Stats cards on dashboard — skeleton while loading
- All buttons show loading spinner while their action is in flight and are disabled to prevent double submission

# Database Schema

## Natural Element Homes — Home Planner

**Version:** 1.0
**Status:** Draft

---

## 1. Overview

- **Database:** PostgreSQL 16
- **Driver:** pgx/v5
- **Connection:** pgxpool (connection pooling)
- **Migrations:** golang-migrate, run automatically on server startup
- **UUIDs:** Generated using `gen_random_uuid()` via `pgcrypto` extension
- **Timestamps:** All stored as `TIMESTAMPTZ` in UTC
- **Soft Deletes:** Plans are soft deleted via `deleted_at`. All other records are hard deleted.

---

## 2. Tables

### 2.1 `users`

Stores all application users. Accounts are created by admins only — no self-registration.

| Column          | Type        | Nullable | Default           | Description           |
| --------------- | ----------- | -------- | ----------------- | --------------------- |
| `id`            | UUID        | NO       | gen_random_uuid() | Primary key           |
| `name`          | TEXT        | NO       |                   | Full name             |
| `email`         | TEXT        | NO       |                   | Email address, unique |
| `password_hash` | TEXT        | NO       |                   | bcrypt hash           |
| `role`          | TEXT        | NO       | `editor`          | `admin` or `editor`   |
| `created_at`    | TIMESTAMPTZ | NO       | NOW()             |                       |
| `updated_at`    | TIMESTAMPTZ | NO       | NOW()             |                       |

**Constraints:**

- `role` IN (`admin`, `editor`)
- `email` UNIQUE

**Indexes:**

- Primary key on `id`
- Unique index on `email`

---

### 2.2 `plans`

Core table. One row per home plan. Soft deleted via `deleted_at`.

| Column                | Type        | Nullable | Default           | Description                                      |
| --------------------- | ----------- | -------- | ----------------- | ------------------------------------------------ |
| `id`                  | UUID        | NO       | gen_random_uuid() | Primary key                                      |
| `name`                | TEXT        | NO       |                   | Display name e.g. "Abilene"                      |
| `slug`                | TEXT        | NO       |                   | URL-safe unique identifier e.g. "abilene"        |
| `type`                | TEXT        | YES      |                   | `single_level` or `multi_level`                  |
| `style`               | TEXT        | YES      |                   | `cabin`, `lodge`, `modern`, `ranch`, `farmhouse` |
| `status`              | TEXT        | NO       | `incomplete`      | `complete`, `incomplete`, `flagged`              |
| `beds`                | INTEGER     | YES      |                   | Bedroom count                                    |
| `baths`               | INTEGER     | YES      |                   | Full bathroom count                              |
| `half_baths`          | INTEGER     | YES      |                   | Half bathroom count                              |
| `main_sf`             | INTEGER     | YES      |                   | Main level square footage                        |
| `upper_sf`            | INTEGER     | YES      |                   | Upper level square footage                       |
| `lower_sf`            | INTEGER     | YES      |                   | Lower level square footage                       |
| `porch_deck_sf`       | INTEGER     | YES      |                   | Porch and deck square footage                    |
| `garage_sf`           | INTEGER     | YES      |                   | Garage square footage                            |
| `garage_apartment_sf` | INTEGER     | YES      |                   | Garage apartment square footage                  |
| `unfinished_sf`       | INTEGER     | YES      | 0                 | Unfinished square footage                        |
| `heated_sf`           | INTEGER     | YES      |                   | Heated square footage                            |
| `total_sf`            | INTEGER     | YES      |                   | Total square footage                             |
| `notes`               | TEXT        | YES      |                   | Free text notes                                  |
| `deleted_at`          | TIMESTAMPTZ | YES      | NULL              | Soft delete timestamp                            |
| `created_at`          | TIMESTAMPTZ | NO       | NOW()             |                                                  |
| `updated_at`          | TIMESTAMPTZ | NO       | NOW()             |                                                  |
| `created_by`          | UUID        | YES      |                   | FK → users.id                                    |
| `updated_by`          | UUID        | YES      |                   | FK → users.id                                    |
| `search_vector`       | TSVECTOR    | NO       | generated         | Generated from `name` for full-text search       |

**Constraints:**

- `type` IN (`single_level`, `multi_level`)
- `style` IN (`cabin`, `lodge`, `modern`, `ranch`, `farmhouse`)
- `status` IN (`complete`, `incomplete`, `flagged`)
- `slug` UNIQUE
- `created_by` REFERENCES `users(id)` ON DELETE SET NULL
- `updated_by` REFERENCES `users(id)` ON DELETE SET NULL

**Indexes:**

- Primary key on `id`
- Unique index on `slug`
- Index on `status`
- Index on `deleted_at`
- GIN index on `search_vector`

---

### 2.3 `files`

One row per uploaded file. Website image files have a `slot` value. All other files have `slot` NULL.

| Column        | Type        | Nullable | Default           | Description                                                      |
| ------------- | ----------- | -------- | ----------------- | ---------------------------------------------------------------- |
| `id`          | UUID        | NO       | gen_random_uuid() | Primary key                                                      |
| `plan_id`     | UUID        | NO       |                   | FK → plans.id                                                    |
| `category`    | TEXT        | NO       |                   | `website`, `reference`, `technical`, `3d`, `other`               |
| `slot`        | TEXT        | YES      | NULL              | Website image slot name, NULL for non-website files              |
| `filename`    | TEXT        | NO       |                   | Original filename                                                |
| `storage_key` | TEXT        | NO       |                   | Full R2 object key e.g. `plans/abilene/website/render-front.jpg` |
| `file_type`   | TEXT        | NO       |                   | MIME type e.g. `image/jpeg`                                      |
| `size_bytes`  | BIGINT      | NO       |                   | File size in bytes                                               |
| `uploaded_at` | TIMESTAMPTZ | NO       | NOW()             |                                                                  |
| `uploaded_by` | UUID        | YES      |                   | FK → users.id                                                    |

**Constraints:**

- `category` IN (`website`, `reference`, `technical`, `3d`, `other`)
- `slot` IN (`render-front`, `elevation-front`, `elevation-left`, `elevation-rear`, `elevation-right`, `floor-plan-main`, `floor-plan-upper`, `floor-plan-lower`, `poster`)
- `storage_key` UNIQUE
- `plan_id` REFERENCES `plans(id)` ON DELETE CASCADE
- `uploaded_by` REFERENCES `users(id)` ON DELETE SET NULL
- When `category = 'website'`, `slot` must NOT be NULL (enforced at application level)
- When `category != 'website'`, `slot` must be NULL (enforced at application level)

**Indexes:**

- Primary key on `id`
- Index on `plan_id`
- Index on `slot`
- Index on `category`
- Unique index on `storage_key`
- Compound index on `(plan_id, slot)` for fast slot lookups

---

### 2.4 `activity_log`

Append-only log of all user actions. Never updated or deleted.

| Column       | Type        | Nullable | Default           | Description                              |
| ------------ | ----------- | -------- | ----------------- | ---------------------------------------- |
| `id`         | UUID        | NO       | gen_random_uuid() | Primary key                              |
| `user_id`    | UUID        | YES      |                   | FK → users.id, NULL if user deleted      |
| `plan_id`    | UUID        | YES      |                   | FK → plans.id, NULL for non-plan actions |
| `action`     | TEXT        | NO       |                   | Action identifier (see below)            |
| `detail`     | JSONB       | YES      |                   | Additional context about the action      |
| `created_at` | TIMESTAMPTZ | NO       | NOW()             |                                          |

**Action values:**

- `plan.created`
- `plan.updated`
- `plan.deleted`
- `plan.restored`
- `plan.flagged`
- `plan.unflagged`
- `plan.duplicated`
- `file.uploaded`
- `file.deleted`
- `file.slot_changed`
- `user.created`
- `user.updated`
- `user.deleted`
- `auth.login`
- `auth.logout`
- `export.csv`
- `export.zip`

**Detail examples:**

```json
// plan.updated
{ "fields_changed": ["beds", "baths", "main_sf"] }

// file.uploaded
{ "filename": "abilene--render--front.jpg", "slot": "render-front", "size_bytes": 204800 }

// export.csv
{ "plan_count": 231, "preset": "wp_all_import" }
```

**Indexes:**

- Primary key on `id`
- Index on `plan_id`
- Index on `user_id`
- Index on `created_at`

---

### 2.5 `refresh_tokens`

Stores hashed refresh tokens for JWT refresh flow. Expired tokens are cleaned up periodically.

| Column       | Type        | Nullable | Default           | Description                      |
| ------------ | ----------- | -------- | ----------------- | -------------------------------- |
| `id`         | UUID        | NO       | gen_random_uuid() | Primary key                      |
| `user_id`    | UUID        | NO       |                   | FK → users.id                    |
| `token_hash` | TEXT        | NO       |                   | bcrypt hash of the refresh token |
| `expires_at` | TIMESTAMPTZ | NO       |                   | Token expiry                     |
| `created_at` | TIMESTAMPTZ | NO       | NOW()             |                                  |

**Constraints:**

- `token_hash` UNIQUE
- `user_id` REFERENCES `users(id)` ON DELETE CASCADE

**Indexes:**

- Primary key on `id`
- Unique index on `token_hash`
- Index on `user_id`
- Index on `expires_at` (for cleanup queries)

---

## 3. Relationships

```
users ──────────────────────────────────────────────────┐
  │                                                      │
  │ created_by / updated_by                              │
  ▼                                                      │
plans ──────────────────┐                               │
  │                     │                               │
  │ plan_id             │ plan_id                       │
  ▼                     ▼                               │
files            activity_log                           │
                         │                              │
                         │ user_id                      │
                         └──────────────────────────────┘

users ──────────────┐
                    │ user_id
                    ▼
             refresh_tokens
```

---

## 4. Completeness Logic

A plan's `status` is set to `complete` automatically when all required website image slots are filled. This is calculated and updated by the backend whenever a file is uploaded or deleted.

**Required slots:**

- `render-front`
- `elevation-front`
- `elevation-left`
- `elevation-rear`
- `elevation-right`
- `floor-plan-main`
- `poster`

**Optional slots:**

- `floor-plan-upper`
- `floor-plan-lower`

**Status rules:**

- If `status = 'flagged'` — never automatically changed, only changed manually
- If all required slots filled → `status = 'complete'`
- If any required slot missing → `status = 'incomplete'`

---

## 5. Notes

- All queries use parameterized statements — no string concatenation
- `deleted_at` is always included in WHERE clauses for plan queries: `WHERE deleted_at IS NULL`
- The `search_vector` column is generated and maintained automatically by Postgres — never written to directly
- `activity_log` rows are never updated or deleted — insert only
- Refresh tokens older than their `expires_at` should be periodically purged — a cleanup job runs on server startup and removes all expired rows

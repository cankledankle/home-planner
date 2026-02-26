# API Specification

## Natural Element Homes — Home Planner

**Version:** 1.0
**Status:** Draft

---

## 1. General

### Base URL

```
http://localhost:8080/api
```

### Authentication

All endpoints except `POST /auth/login` require a valid JWT access token stored in an httpOnly cookie named `access_token`. Requests without a valid token receive `401 Unauthorized`.

Admin-only endpoints additionally require `role == "admin"`. Requests from editor-role users receive `403 Forbidden`.

### Request Format

- All request bodies are JSON unless the endpoint accepts file uploads, in which case `multipart/form-data`
- All responses are JSON

### Response Format

All responses follow one of two shapes:

**Success:**

```json
{
  "data": { ... }
}
```

**Error:**

```json
{
	"error": {
		"code": "ERROR_CODE",
		"message": "Human readable message"
	}
}
```

### Pagination

List endpoints that return multiple records support pagination via query parameters:

| Param   | Type    | Default | Description               |
| ------- | ------- | ------- | ------------------------- |
| `page`  | integer | 1       | Page number               |
| `limit` | integer | 50      | Records per page, max 100 |

Paginated responses include a `meta` object:

```json
{
  "data": [...],
  "meta": {
    "page": 1,
    "limit": 50,
    "total": 231,
    "total_pages": 5
  }
}
```

### Common Error Codes

| Code               | Status | Description                         |
| ------------------ | ------ | ----------------------------------- |
| `UNAUTHORIZED`     | 401    | Missing or invalid access token     |
| `FORBIDDEN`        | 403    | Authenticated but insufficient role |
| `NOT_FOUND`        | 404    | Resource does not exist             |
| `VALIDATION_ERROR` | 400    | Request body failed validation      |
| `CONFLICT`         | 409    | Unique constraint violation         |
| `INTERNAL_ERROR`   | 500    | Unexpected server error             |

---

## 2. Auth

### `POST /api/auth/login`

Authenticate a user and issue JWT tokens.

**Auth required:** No

**Request:**

```json
{
	"email": "user@example.com",
	"password": "plaintext-password"
}
```

**Response `200`:**

```json
{
	"data": {
		"user": {
			"id": "uuid",
			"name": "Jane Smith",
			"email": "user@example.com",
			"role": "admin"
		}
	}
}
```

Sets two httpOnly cookies:

- `access_token` — JWT, expires in 15 minutes
- `refresh_token` — JWT, expires in 7 days

**Errors:**

| Code                  | Status | When                                  |
| --------------------- | ------ | ------------------------------------- |
| `INVALID_CREDENTIALS` | 401    | Email not found or password incorrect |
| `VALIDATION_ERROR`    | 400    | Missing email or password             |

---

### `POST /api/auth/refresh`

Issue a new access token using a valid refresh token.

**Auth required:** No (uses `refresh_token` cookie)

**Request:** No body

**Response `200`:**

```json
{
	"data": {
		"message": "Token refreshed"
	}
}
```

Sets a new `access_token` cookie.

**Errors:**

| Code            | Status | When                                       |
| --------------- | ------ | ------------------------------------------ |
| `INVALID_TOKEN` | 401    | Refresh token missing, invalid, or expired |

---

### `POST /api/auth/logout`

Clear auth cookies and invalidate refresh token.

**Auth required:** Yes

**Request:** No body

**Response `200`:**

```json
{
	"data": {
		"message": "Logged out"
	}
}
```

Clears `access_token` and `refresh_token` cookies. Deletes refresh token record from database.

---

### `GET /api/auth/me`

Return the currently authenticated user.

**Auth required:** Yes

**Response `200`:**

```json
{
	"data": {
		"id": "uuid",
		"name": "Jane Smith",
		"email": "user@example.com",
		"role": "admin",
		"created_at": "2025-01-01T00:00:00Z"
	}
}
```

---

## 3. Users

### `GET /api/users`

List all users.

**Auth required:** Yes — Admin only

**Response `200`:**

```json
{
	"data": [
		{
			"id": "uuid",
			"name": "Jane Smith",
			"email": "user@example.com",
			"role": "admin",
			"created_at": "2025-01-01T00:00:00Z",
			"updated_at": "2025-01-01T00:00:00Z"
		}
	]
}
```

---

### `POST /api/users`

Create a new user.

**Auth required:** Yes — Admin only

**Request:**

```json
{
	"name": "John Doe",
	"email": "john@example.com",
	"password": "initial-password",
	"role": "editor"
}
```

**Response `201`:**

```json
{
	"data": {
		"id": "uuid",
		"name": "John Doe",
		"email": "john@example.com",
		"role": "editor",
		"created_at": "2025-01-01T00:00:00Z"
	}
}
```

**Errors:**

| Code               | Status | When                      |
| ------------------ | ------ | ------------------------- |
| `CONFLICT`         | 409    | Email already in use      |
| `VALIDATION_ERROR` | 400    | Missing or invalid fields |

---

### `PUT /api/users/:id`

Update a user.

**Auth required:** Yes — Admin only

**Request:**

```json
{
	"name": "John Doe",
	"email": "john@example.com",
	"role": "admin"
}
```

**Response `200`:**

```json
{
	"data": {
		"id": "uuid",
		"name": "John Doe",
		"email": "john@example.com",
		"role": "admin",
		"updated_at": "2025-01-01T00:00:00Z"
	}
}
```

**Errors:**

| Code        | Status | When                 |
| ----------- | ------ | -------------------- |
| `NOT_FOUND` | 404    | User does not exist  |
| `CONFLICT`  | 409    | Email already in use |

---

### `PUT /api/users/:id/password`

Update a user's password.

**Auth required:** Yes — Admin only

**Request:**

```json
{
	"password": "new-password"
}
```

**Response `200`:**

```json
{
	"data": {
		"message": "Password updated"
	}
}
```

---

### `DELETE /api/users/:id`

Delete a user.

**Auth required:** Yes — Admin only

**Response `200`:**

```json
{
	"data": {
		"message": "User deleted"
	}
}
```

**Errors:**

| Code        | Status | When                           |
| ----------- | ------ | ------------------------------ |
| `NOT_FOUND` | 404    | User does not exist            |
| `FORBIDDEN` | 403    | Cannot delete your own account |

---

## 4. Plans

### `GET /api/plans`

List all plans. Supports search, filter, sort, and pagination.

**Auth required:** Yes

**Query Parameters:**

| Param           | Type    | Description                                                 |
| --------------- | ------- | ----------------------------------------------------------- |
| `search`        | string  | Full-text search on plan name                               |
| `status`        | string  | Filter by `complete`, `incomplete`, `flagged`               |
| `type`          | string  | Filter by `single_level`, `multi_level`                     |
| `style`         | string  | Filter by style value                                       |
| `beds_min`      | integer | Minimum bedroom count                                       |
| `beds_max`      | integer | Maximum bedroom count                                       |
| `baths_min`     | integer | Minimum bathroom count                                      |
| `baths_max`     | integer | Maximum bathroom count                                      |
| `heated_sf_min` | integer | Minimum heated SF                                           |
| `heated_sf_max` | integer | Maximum heated SF                                           |
| `missing_slot`  | string  | Filter plans missing a specific slot                        |
| `sort`          | string  | `name`, `heated_sf`, `total_sf`, `created_at`, `updated_at` |
| `order`         | string  | `asc`, `desc` — default `asc`                               |
| `page`          | integer | Default 1                                                   |
| `limit`         | integer | Default 50, max 100                                         |

**Response `200`:**

```json
{
	"data": [
		{
			"id": "uuid",
			"name": "Abilene",
			"slug": "abilene",
			"type": "single_level",
			"style": "cabin",
			"status": "complete",
			"beds": 3,
			"baths": 2,
			"half_baths": 1,
			"heated_sf": 1478,
			"total_sf": 1862,
			"updated_at": "2025-01-01T00:00:00Z"
		}
	],
	"meta": {
		"page": 1,
		"limit": 50,
		"total": 231,
		"total_pages": 5
	}
}
```

---

### `GET /api/plans/:id`

Get a single plan with all fields and files.

**Auth required:** Yes

**Response `200`:**

```json
{
  "data": {
    "id": "uuid",
    "name": "Abilene",
    "slug": "abilene",
    "type": "single_level",
    "style": "cabin",
    "status": "complete",
    "beds": 3,
    "baths": 2,
    "half_baths": 1,
    "main_sf": 1478,
    "upper_sf": 0,
    "lower_sf": 0,
    "porch_deck_sf": 384,
    "garage_sf": 0,
    "garage_apartment_sf": 0,
    "unfinished_sf": 0,
    "heated_sf": 1478,
    "total_sf": 1862,
    "notes": null,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z",
    "created_by": {
      "id": "uuid",
      "name": "Jane Smith"
    },
    "updated_by": {
      "id": "uuid",
      "name": "Jane Smith"
    },
    "files": {
      "website": {
        "render-front": {
          "id": "uuid",
          "filename": "abilene--render--front.jpg",
          "storage_key": "plans/abilene/website/render-front.jpg",
          "file_type": "image/jpeg",
          "size_bytes": 204800,
          "uploaded_at": "2025-01-01T00:00:00Z"
        },
        "elevation-front": { "..." },
        "elevation-left": { "..." },
        "elevation-rear": { "..." },
        "elevation-right": { "..." },
        "floor-plan-main": { "..." },
        "floor-plan-upper": null,
        "floor-plan-lower": null,
        "poster": { "..." }
      },
      "reference": ["..."],
      "technical": ["..."],
      "3d": ["..."],
      "other": ["..."]
    }
  }
}
```

**Errors:**

| Code        | Status | When                              |
| ----------- | ------ | --------------------------------- |
| `NOT_FOUND` | 404    | Plan does not exist or is deleted |

---

### `POST /api/plans`

Create a new plan.

**Auth required:** Yes

**Request:**

```json
{
	"name": "Abilene",
	"type": "single_level",
	"style": "cabin",
	"beds": 3,
	"baths": 2,
	"half_baths": 1,
	"main_sf": 1478,
	"upper_sf": 0,
	"lower_sf": 0,
	"porch_deck_sf": 384,
	"garage_sf": 0,
	"garage_apartment_sf": 0,
	"unfinished_sf": 0,
	"heated_sf": 1478,
	"total_sf": 1862,
	"notes": null
}
```

**Response `201`:**

```json
{
  "data": {
    "id": "uuid",
    "name": "Abilene",
    "slug": "abilene",
    "status": "incomplete",
    "..."
  }
}
```

**Errors:**

| Code               | Status | When                                    |
| ------------------ | ------ | --------------------------------------- |
| `CONFLICT`         | 409    | Plan with same name/slug already exists |
| `VALIDATION_ERROR` | 400    | Missing required fields                 |

---

### `PUT /api/plans/:id`

Update a plan's metadata.

**Auth required:** Yes

**Request:** Any subset of plan fields

```json
{
	"beds": 4,
	"heated_sf": 1600
}
```

**Response `200`:** Full updated plan object (same shape as `GET /api/plans/:id`)

**Errors:**

| Code        | Status | When                |
| ----------- | ------ | ------------------- |
| `NOT_FOUND` | 404    | Plan does not exist |
| `CONFLICT`  | 409    | Slug already in use |

---

### `DELETE /api/plans/:id`

Soft delete a plan.

**Auth required:** Yes — Admin only

**Response `200`:**

```json
{
	"data": {
		"message": "Plan deleted"
	}
}
```

**Errors:**

| Code        | Status | When                |
| ----------- | ------ | ------------------- |
| `NOT_FOUND` | 404    | Plan does not exist |

---

### `POST /api/plans/:id/restore`

Restore a soft deleted plan.

**Auth required:** Yes — Admin only

**Response `200`:**

```json
{
	"data": {
		"message": "Plan restored"
	}
}
```

---

### `POST /api/plans/:id/duplicate`

Duplicate a plan's metadata. Files are not copied.

**Auth required:** Yes

**Request:**

```json
{
	"name": "Abilene B"
}
```

**Response `201`:** Full new plan object

---

### `PUT /api/plans/:id/flag`

Flag a plan for review.

**Auth required:** Yes

**Response `200`:**

```json
{
	"data": {
		"message": "Plan flagged"
	}
}
```

---

### `PUT /api/plans/:id/unflag`

Remove flag from a plan.

**Auth required:** Yes

**Response `200`:**

```json
{
	"data": {
		"message": "Plan unflagged"
	}
}
```

---

## 5. Files

### `POST /api/plans/:id/files/website`

Upload a website image to a specific slot.

**Auth required:** Yes

**Content-Type:** `multipart/form-data`

**Form fields:**

| Field  | Type   | Required | Description      |
| ------ | ------ | -------- | ---------------- |
| `file` | file   | Yes      | Image file       |
| `slot` | string | Yes      | Target slot name |

**Validation:**

- File must be an image (`image/jpeg`, `image/png`, `image/webp`)
- Max file size: 50MB
- Slot must be a valid slot name

**Response `201`:**

```json
{
	"data": {
		"id": "uuid",
		"plan_id": "uuid",
		"category": "website",
		"slot": "render-front",
		"filename": "abilene--render--front.jpg",
		"storage_key": "plans/abilene/website/render-front.jpg",
		"file_type": "image/jpeg",
		"size_bytes": 204800,
		"uploaded_at": "2025-01-01T00:00:00Z",
		"uploaded_by": {
			"id": "uuid",
			"name": "Jane Smith"
		}
	}
}
```

**Errors:**

| Code               | Status | When                             |
| ------------------ | ------ | -------------------------------- |
| `NOT_FOUND`        | 404    | Plan does not exist              |
| `VALIDATION_ERROR` | 400    | Invalid file type, size, or slot |

---

### `POST /api/plans/:id/files`

Upload one or more files to a non-website category.

**Auth required:** Yes

**Content-Type:** `multipart/form-data`

**Form fields:**

| Field      | Type   | Required | Description                             |
| ---------- | ------ | -------- | --------------------------------------- |
| `files`    | file[] | Yes      | One or more files                       |
| `category` | string | Yes      | `reference`, `technical`, `3d`, `other` |

**Validation:**

- Max file size: 500MB per file
- Category must not be `website`

**Response `201`:**

```json
{
	"data": [
		{
			"id": "uuid",
			"plan_id": "uuid",
			"category": "reference",
			"slot": null,
			"filename": "site-photo.jpg",
			"storage_key": "plans/abilene/reference/site-photo.jpg",
			"file_type": "image/jpeg",
			"size_bytes": 512000,
			"uploaded_at": "2025-01-01T00:00:00Z"
		}
	]
}
```

---

### `GET /api/plans/:id/files`

List all files for a plan.

**Auth required:** Yes

**Response `200`:**

```json
{
  "data": {
    "website": {
      "render-front": { "..." },
      "elevation-front": { "..." },
      "elevation-left": { "..." },
      "elevation-rear": { "..." },
      "elevation-right": { "..." },
      "floor-plan-main": { "..." },
      "floor-plan-upper": null,
      "floor-plan-lower": null,
      "poster": { "..." }
    },
    "reference": ["..."],
    "technical": ["..."],
    "3d": ["..."],
    "other": ["..."]
  }
}
```

---

### `GET /api/files/:id/url`

Get a presigned URL for a file.

**Auth required:** Yes

**Response `200`:**

```json
{
	"data": {
		"url": "https://r2-presigned-url...",
		"expires_at": "2025-01-01T01:00:00Z"
	}
}
```

**Errors:**

| Code        | Status | When                |
| ----------- | ------ | ------------------- |
| `NOT_FOUND` | 404    | File does not exist |

---

### `DELETE /api/files/:id`

Delete a file from both the database and R2.

**Auth required:** Yes

**Response `200`:**

```json
{
	"data": {
		"message": "File deleted"
	}
}
```

**Errors:**

| Code        | Status | When                |
| ----------- | ------ | ------------------- |
| `NOT_FOUND` | 404    | File does not exist |

---

## 6. Export

### `GET /api/export/csv`

Export plan data as a CSV file.

**Auth required:** Yes

**Query Parameters:**

| Param    | Type   | Description                                                 |
| -------- | ------ | ----------------------------------------------------------- |
| `preset` | string | `wp_all_import`, `general`, `custom`                        |
| `fields` | string | Comma-separated field names (required when `preset=custom`) |
| `ids`    | string | Comma-separated plan UUIDs — omit to export all             |

**Response `200`:**

- `Content-Type: text/csv`
- `Content-Disposition: attachment; filename="home-plans.csv"`
- CSV file streamed directly

---

### `GET /api/export/zip`

Download all files for selected plans as a ZIP.

**Auth required:** Yes

**Query Parameters:**

| Param        | Type   | Description                                          |
| ------------ | ------ | ---------------------------------------------------- |
| `ids`        | string | Comma-separated plan UUIDs — omit for all plans      |
| `categories` | string | Comma-separated categories to include — omit for all |

**Response `200`:**

- `Content-Type: application/zip`
- `Content-Disposition: attachment; filename="home-plans-files.zip"`
- ZIP streamed directly, folder structure: `/{plan-name}/{category}/`

---

## 7. Import

### `POST /api/import/csv/preview`

Parse and preview a CSV before importing.

**Auth required:** Yes — Admin only

**Content-Type:** `multipart/form-data`

**Form fields:**

| Field  | Type | Required | Description |
| ------ | ---- | -------- | ----------- |
| `file` | file | Yes      | CSV file    |

**Response `200`:**

```json
{
	"data": {
		"row_count": 231,
		"columns": ["Name", "Beds", "Baths", "..."],
		"preview_rows": [{ "Name": "Abilene", "Beds": "3", "...": "..." }],
		"suggested_mapping": {
			"Name": "name",
			"Beds": "beds",
			"Baths": "baths"
		}
	}
}
```

---

### `POST /api/import/csv`

Import plans from a CSV.

**Auth required:** Yes — Admin only

**Content-Type:** `multipart/form-data`

**Form fields:**

| Field     | Type        | Required | Description                            |
| --------- | ----------- | -------- | -------------------------------------- |
| `file`    | file        | Yes      | CSV file                               |
| `mapping` | JSON string | Yes      | Column-to-field mapping                |
| `mode`    | string      | Yes      | `create_only`, `update_only`, `upsert` |

**Mapping example:**

```json
{
	"Name": "name",
	"Beds": "beds",
	"Baths": "baths",
	"Half Baths": "half_baths",
	"Main Level SF": "main_sf"
}
```

**Response `200`:**

```json
{
	"data": {
		"created": 220,
		"updated": 10,
		"skipped": 1,
		"errors": [
			{
				"row": 45,
				"message": "Duplicate plan name: Abilene"
			}
		]
	}
}
```

---

## 8. Activity Log

### `GET /api/activity`

Get the global activity log.

**Auth required:** Yes — Admin only

**Query Parameters:**

| Param     | Type    | Description           |
| --------- | ------- | --------------------- |
| `user_id` | UUID    | Filter by user        |
| `plan_id` | UUID    | Filter by plan        |
| `action`  | string  | Filter by action type |
| `page`    | integer | Default 1             |
| `limit`   | integer | Default 50            |

**Response `200`:**

```json
{
  "data": [
    {
      "id": "uuid",
      "user": {
        "id": "uuid",
        "name": "Jane Smith"
      },
      "plan": {
        "id": "uuid",
        "name": "Abilene"
      },
      "action": "file.uploaded",
      "detail": {
        "filename": "abilene--render--front.jpg",
        "slot": "render-front"
      },
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "meta": { "..." }
}
```

---

### `GET /api/plans/:id/activity`

Get the activity log for a specific plan.

**Auth required:** Yes

**Query Parameters:**

| Param   | Type    | Description |
| ------- | ------- | ----------- |
| `page`  | integer | Default 1   |
| `limit` | integer | Default 50  |

**Response `200`:** Same shape as `GET /api/activity`

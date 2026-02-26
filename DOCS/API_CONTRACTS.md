# API Contract Guidelines

## The Problem

We had bugs where:
1. Frontend called `/api/export` but backend route was `/api/export/csv`
2. Frontend sent `wp-all-import` (kebab-case) but backend expected `wp_all_import` (snake_case)
3. Frontend had field names like `poster_url` but backend exported `poster`

These bugs **passed TypeScript compilation** because:
- TypeScript types don't validate runtime API contracts
- String literals were scattered across multiple files
- No single source of truth existed

## The Solution

### 1. Single Source of Truth

All API contracts are defined in **`src/lib/api/contracts.ts`**:

```typescript
// This is the ONLY place where these values are defined
export const EXPORT_PRESETS = {
  wpAllImport: 'wp_all_import',  // Must match backend exactly
  general: 'general',
  // ...
} as const;
```

### 2. Use Constants, Never String Literals

**❌ BAD - String literals scattered everywhere:**
```typescript
// ExportModal.svelte
const preset = 'wp-all-import';  // Wrong! Should be 'wp_all_import'

// Backend
if preset == "wp_all_import" {  // Hardcoded string
```

**✅ GOOD - Use constants:**
```typescript
// ExportModal.svelte
import { EXPORT_PRESETS } from '$lib/api/contracts';
const preset = EXPORT_PRESETS.wpAllImport;

// Backend
if preset == ExportPresetWPAllImport {  // Uses constant
```

### 3. Backend Validation

Always validate incoming parameters against allowed values:

```go
var validExportPresets = map[string]bool{
    ExportPresetWPAllImport: true,
    ExportPresetGeneral:     true,
    // ...
}

func validateExportPreset(preset string) error {
    if !validExportPresets[preset] {
        return fmt.Errorf("invalid preset: %s", preset)
    }
    return nil
}
```

### 4. Runtime Type Checking

For critical API boundaries, add runtime validation:

```typescript
import { isValidExportPreset } from '$lib/api/contracts';

if (!isValidExportPreset(preset)) {
  throw new Error(`Invalid preset: ${preset}`);
}
```

## Contract File Structure

```typescript
// src/lib/api/contracts.ts

// 1. Define the constants
export const EXPORT_ENDPOINTS = {
  csv: '/api/export/csv',
  zip: '/api/export/zip'
} as const;

// 2. Derive types from constants
export type ExportEndpoint = typeof EXPORT_ENDPOINTS[keyof typeof EXPORT_ENDPOINTS];

// 3. Validation functions
export function isValidExportEndpoint(value: string): value is ExportEndpoint {
  return Object.values(EXPORT_ENDPOINTS).includes(value as ExportEndpoint);
}
```

## Checklist for New Features

When adding API endpoints or parameters:

- [ ] Define constants in `contracts.ts` FIRST
- [ ] Update types in `contracts.ts` using `typeof`
- [ ] Use constants in frontend (no string literals)
- [ ] Use constants in backend (no string literals)
- [ ] Add backend validation for all parameters
- [ ] Add runtime validation functions
- [ ] Update both frontend and backend tests
- [ ] Update API documentation

## Detecting Mismatches

### TypeScript Compilation
```bash
npm run check  # Won't catch API contract mismatches
```

### Manual Testing
Always test the full flow:
1. Frontend makes API call
2. Backend receives and validates
3. Backend returns expected data
4. Frontend handles response

### Common Causes
1. **Copy-pasting** values instead of importing constants
2. **Guessing** what the backend expects
3. **Inconsistent naming** (kebab-case vs snake_case)
4. **Not updating** both sides when changing values

## Prevention Checklist

Add this to your PR template:

```markdown
## API Contract Checklist
- [ ] I used constants from `contracts.ts` instead of string literals
- [ ] I updated both frontend and backend
- [ ] I added backend validation for new parameters
- [ ] I tested the full API flow end-to-end
- [ ] I updated API documentation
```

## Example: Adding a New Export Preset

1. **Add to contracts.ts:**
```typescript
export const EXPORT_PRESETS = {
  // ... existing presets
  newPreset: 'new_preset'
} as const;

export const EXPORT_PRESET_FIELDS = {
  // ... existing presets
  [EXPORT_PRESETS.newPreset]: [
    EXPORT_FIELDS.name,
    EXPORT_FIELDS.newField
  ]
};
```

2. **Update backend constants:**
```go
const ExportPresetNewPreset = "new_preset"

var validExportPresets = map[string]bool{
    // ... existing presets
    ExportPresetNewPreset: true,
}
```

3. **Use in frontend:**
```typescript
import { EXPORT_PRESETS } from '$lib/api/contracts';

const preset = EXPORT_PRESETS.newPreset;
```

4. **Validate in backend:**
```go
if err := validateExportPreset(preset); err != nil {
    return error response
}
```

## Summary

**Golden Rule:** If you're typing a string literal that represents an API value, you're doing it wrong. Import it from `contracts.ts` instead.

**Why this works:**
1. TypeScript will error if constant doesn't exist
2. Backend validation catches mismatches at runtime
3. Single source of truth means one change updates everywhere
4. Self-documenting code - you can see all valid values in one place

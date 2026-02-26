// This file is the SINGLE SOURCE OF TRUTH for API contracts
// Both frontend and backend must stay in sync with this file
// When updating, update BOTH sides and increment the version

export const API_VERSION = '1.0.0';

// Export functionality
export const EXPORT_ENDPOINTS = {
	csv: '/api/export/csv',
	zip: '/api/export/zip'
} as const;

export const EXPORT_PRESETS = {
	wpAllImport: 'wp_all_import',  // Must match backend exactly
	general: 'general',
	minimal: 'minimal',
	custom: 'custom'
} as const;

export type ExportPreset = typeof EXPORT_PRESETS[keyof typeof EXPORT_PRESETS];

// Field names for CSV export - must match backend database column names
export const EXPORT_FIELDS = {
	// Plan metadata
	id: 'id',
	name: 'name',
	slug: 'slug',
	status: 'status',
	type: 'type',
	style: 'style',
	beds: 'beds',
	baths: 'baths',
	halfBaths: 'half_baths',
	mainSf: 'main_sf',
	upperSf: 'upper_sf',
	lowerSf: 'lower_sf',
	porchDeckSf: 'porch_deck_sf',
	garageSf: 'garage_sf',
	garageApartmentSf: 'garage_apartment_sf',
	unfinishedSf: 'unfinished_sf',
	heatedSf: 'heated_sf',
	totalSf: 'total_sf',
	notes: 'notes',
	createdAt: 'created_at',
	updatedAt: 'updated_at',
	
	// Image slots - must match database slot names
	renderFront: 'render_front',
	elevationFront: 'elevation_front',
	elevationLeft: 'elevation_left',
	elevationRear: 'elevation_rear',
	elevationRight: 'elevation_right',
	floorPlanMain: 'floor_plan_main',
	floorPlanUpper: 'floor_plan_upper',
	floorPlanLower: 'floor_plan_lower',
	poster: 'poster'
} as const;

export type ExportField = typeof EXPORT_FIELDS[keyof typeof EXPORT_FIELDS];

// Preset field configurations - must match backend
export const EXPORT_PRESET_FIELDS: Record<ExportPreset, ExportField[]> = {
	[EXPORT_PRESETS.wpAllImport]: [
		EXPORT_FIELDS.name,
		EXPORT_FIELDS.slug,
		EXPORT_FIELDS.type,
		EXPORT_FIELDS.style,
		EXPORT_FIELDS.beds,
		EXPORT_FIELDS.baths,
		EXPORT_FIELDS.halfBaths,
		EXPORT_FIELDS.mainSf,
		EXPORT_FIELDS.upperSf,
		EXPORT_FIELDS.lowerSf,
		EXPORT_FIELDS.porchDeckSf,
		EXPORT_FIELDS.garageSf,
		EXPORT_FIELDS.garageApartmentSf,
		EXPORT_FIELDS.unfinishedSf,
		EXPORT_FIELDS.heatedSf,
		EXPORT_FIELDS.totalSf,
		EXPORT_FIELDS.notes,
		EXPORT_FIELDS.status,
		EXPORT_FIELDS.renderFront,
		EXPORT_FIELDS.elevationFront,
		EXPORT_FIELDS.elevationLeft,
		EXPORT_FIELDS.elevationRear,
		EXPORT_FIELDS.elevationRight,
		EXPORT_FIELDS.floorPlanMain,
		EXPORT_FIELDS.floorPlanUpper,
		EXPORT_FIELDS.floorPlanLower,
		EXPORT_FIELDS.poster
	],
	[EXPORT_PRESETS.general]: [
		EXPORT_FIELDS.id,
		EXPORT_FIELDS.name,
		EXPORT_FIELDS.slug,
		EXPORT_FIELDS.status,
		EXPORT_FIELDS.type,
		EXPORT_FIELDS.style,
		EXPORT_FIELDS.beds,
		EXPORT_FIELDS.baths,
		EXPORT_FIELDS.heatedSf,
		EXPORT_FIELDS.totalSf,
		EXPORT_FIELDS.notes
	],
	[EXPORT_PRESETS.minimal]: [
		EXPORT_FIELDS.id,
		EXPORT_FIELDS.name,
		EXPORT_FIELDS.slug,
		EXPORT_FIELDS.status
	],
	[EXPORT_PRESETS.custom]: [] // User selects fields
};

// Website slot names - must match backend validWebsiteSlots
export const WEBSITE_SLOTS = {
	renderFront: 'render-front',
	elevationFront: 'elevation-front',
	elevationLeft: 'elevation-left',
	elevationRear: 'elevation-rear',
	elevationRight: 'elevation-right',
	floorPlanMain: 'floor-plan-main',
	floorPlanUpper: 'floor-plan-upper',
	floorPlanLower: 'floor-plan-lower',
	poster: 'poster'
} as const;

export type WebsiteSlot = typeof WEBSITE_SLOTS[keyof typeof WEBSITE_SLOTS];

// Slot to CSV column name mapping
export const SLOT_TO_CSV_COLUMN: Record<WebsiteSlot, string> = {
	[WEBSITE_SLOTS.renderFront]: 'render_front',
	[WEBSITE_SLOTS.elevationFront]: 'elevation_front',
	[WEBSITE_SLOTS.elevationLeft]: 'elevation_left',
	[WEBSITE_SLOTS.elevationRear]: 'elevation_rear',
	[WEBSITE_SLOTS.elevationRight]: 'elevation_right',
	[WEBSITE_SLOTS.floorPlanMain]: 'floor_plan_main',
	[WEBSITE_SLOTS.floorPlanUpper]: 'floor_plan_upper',
	[WEBSITE_SLOTS.floorPlanLower]: 'floor_plan_lower',
	[WEBSITE_SLOTS.poster]: 'poster'
};

// Validation helpers
export function isValidExportPreset(value: string): value is ExportPreset {
	return Object.values(EXPORT_PRESETS).includes(value as ExportPreset);
}

export function isValidExportField(value: string): value is ExportField {
	return Object.values(EXPORT_FIELDS).includes(value as ExportField);
}

export function isValidWebsiteSlot(value: string): value is WebsiteSlot {
	return Object.values(WEBSITE_SLOTS).includes(value as WebsiteSlot);
}

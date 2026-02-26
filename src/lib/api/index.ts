import { api } from './client';
export { api };

import type {
	User,
	Plan,
	PlanWithFiles,
	File,
	Activity,
	PaginatedResponse,
	LoginCredentials,
	SFTPCredentials,
	SFTPStatus
} from '$lib/types';

export async function login(credentials: LoginCredentials): Promise<User> {
	return api.login(credentials);
}

export async function logout(): Promise<void> {
	return api.logout();
}

export async function getMe(): Promise<User | null> {
	return api.me();
}

export async function getUsers(): Promise<User[]> {
	return api.get<User[]>('/users');
}

export async function createUser(
	user: Omit<User, 'id' | 'created_at' | 'updated_at'> & { password: string }
): Promise<User & { sftp_credentials?: SFTPCredentials }> {
	return api.post<User & { sftp_credentials?: SFTPCredentials }>('/users', user);
}

export async function updateUser(id: string, user: Partial<User>): Promise<User> {
	return api.put<User>(`/users/${id}`, user);
}

export async function deleteUser(id: string): Promise<void> {
	return api.delete<void>(`/users/${id}`);
}

interface ListPlansParams {
	search?: string;
	status?: string;
	type?: string;
	style?: string;
	beds_min?: number;
	beds_max?: number;
	baths_min?: number;
	baths_max?: number;
	heated_sf_min?: number;
	heated_sf_max?: number;
	missing_slot?: string;
	sort?: string;
	order?: 'asc' | 'desc';
	page?: number;
	limit?: number;
}

export interface DashboardStats {
	total: number;
	complete: number;
	incomplete: number;
	flagged: number;
}

export async function getDashboardStats(): Promise<DashboardStats> {
	return api.get<DashboardStats>('/plans/stats');
}

export async function getRecentPlans(limit: number = 10): Promise<Plan[]> {
	return api.get<Plan[]>(`/plans/recent?limit=${limit}`);
}

export async function getPlans(params: ListPlansParams = {}): Promise<PaginatedResponse<Plan>> {
	const searchParams = new URLSearchParams();
	Object.entries(params).forEach(([key, value]) => {
		if (value !== undefined && value !== null) {
			searchParams.append(key, String(value));
		}
	});
	const query = searchParams.toString();
	// Use raw fetch to get full paginated response with meta
	const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
	const response = await fetch(`${API_BASE}/plans${query ? `?${query}` : ''}`, {
		credentials: 'include'
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.error?.message || 'Failed to load plans');
	}
	const data = await response.json();
	return data;
}

export async function getPlan(id: string): Promise<PlanWithFiles> {
	return api.get<PlanWithFiles>(`/plans/${id}`);
}

export async function createPlan(
	plan: Omit<Plan, 'id' | 'slug' | 'status' | 'created_at' | 'updated_at'>
): Promise<Plan> {
	return api.post<Plan>('/plans', plan);
}

export async function updatePlan(id: string, plan: Partial<Plan>): Promise<Plan> {
	return api.put<Plan>(`/plans/${id}`, plan);
}

export async function deletePlan(id: string): Promise<void> {
	return api.delete<void>(`/plans/${id}`);
}

export async function restorePlan(id: string): Promise<void> {
	return api.post<void>(`/plans/${id}/restore`, {});
}

export async function duplicatePlan(id: string, name: string): Promise<Plan> {
	return api.post<Plan>(`/plans/${id}/duplicate`, { name });
}

export async function flagPlan(id: string): Promise<void> {
	return api.put<void>(`/plans/${id}/flag`, {});
}

export async function unflagPlan(id: string): Promise<void> {
	return api.put<void>(`/plans/${id}/unflag`, {});
}

interface ListActivitiesParams {
	user_id?: string;
	plan_id?: string;
	action?: string;
	page?: number;
	limit?: number;
}

export async function getActivities(
	params: ListActivitiesParams = {}
): Promise<PaginatedResponse<Activity>> {
	const searchParams = new URLSearchParams();
	Object.entries(params).forEach(([key, value]) => {
		if (value !== undefined && value !== null) {
			searchParams.append(key, String(value));
		}
	});
	const query = searchParams.toString();
	// Use raw fetch to get full paginated response with meta
	const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
	const response = await fetch(`${API_BASE}/activity${query ? `?${query}` : ''}`, {
		credentials: 'include'
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.error?.message || 'Failed to load activities');
	}
	const data = await response.json();
	return data;
}

export async function getPlanActivities(
	planId: string,
	page: number = 1,
	limit: number = 50
): Promise<PaginatedResponse<Activity>> {
	// Use raw fetch to get full paginated response with meta
	const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
	const response = await fetch(
		`${API_BASE}/plans/${planId}/activity?page=${page}&limit=${limit}`,
		{ credentials: 'include' }
	);
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.error?.message || 'Failed to load plan activities');
	}
	const data = await response.json();
	return data;
}

export async function getPlanFiles(planId: string): Promise<{
	website: Record<string, File | null>;
	reference: File[];
	technical: File[];
	'3d': File[];
	other: File[];
}> {
	return api.get(`/plans/${planId}/files`);
}

export async function deleteFile(fileId: string): Promise<void> {
	return api.delete<void>(`/files/${fileId}`);
}

export async function getFileUrl(fileId: string): Promise<{ url: string; expires_at: string }> {
	return api.get<{ url: string; expires_at: string }>(`/files/${fileId}/url`);
}

export async function uploadWebsiteFile(planId: string, slot: string, file: globalThis.File): Promise<File> {
	const formData = new FormData();
	formData.append('file', file);
	formData.append('slot', slot);

	const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
	const response = await fetch(`${API_BASE}/plans/${planId}/files/website`, {
		method: 'POST',
		credentials: 'include',
		body: formData
	});

	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.error?.message || 'Upload failed');
	}

	const data = await response.json();
	return data.data;
}

export async function uploadFiles(
	planId: string,
	category: string,
	files: globalThis.File[]
): Promise<File[]> {
	const formData = new FormData();
	files.forEach((file) => {
		formData.append('files', file);
	});
	formData.append('category', category);

	const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
	const response = await fetch(`${API_BASE}/plans/${planId}/files`, {
		method: 'POST',
		credentials: 'include',
		body: formData
	});

	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.error?.message || 'Upload failed');
	}

	const data = await response.json();
	return data.data;
}

// Import API
export interface ImportPreview {
	columns: string[];
	preview: Record<string, string>[];
	suggested_mapping: Record<string, string>;
	total_rows: number;
}

export interface ImportColumnMapping {
	[column: string]: string | null;
}

export interface ImportOptions {
	mode: 'create' | 'update' | 'upsert';
	id_column?: string;
	mapping: ImportColumnMapping;
}

export interface ImportResult {
	imported: number;
	updated: number;
	errors: Array<{
		row: number;
		message: string;
	}>;
}

export async function previewCsvImport(file: globalThis.File): Promise<ImportPreview> {
	const formData = new FormData();
	formData.append('file', file);

	const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
	const response = await fetch(`${API_BASE}/import/csv/preview`, {
		method: 'POST',
		credentials: 'include',
		body: formData
	});

	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.error?.message || 'Preview failed');
	}

	const data = await response.json();
	return data.data;
}

export async function importCsv(file: globalThis.File, options: ImportOptions): Promise<ImportResult> {
	const formData = new FormData();
	formData.append('file', file);
	formData.append('mode', options.mode);
	formData.append('mapping', JSON.stringify(options.mapping));
	if (options.id_column) {
		formData.append('id_column', options.id_column);
	}

	const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
	const response = await fetch(`${API_BASE}/import/csv`, {
		method: 'POST',
		credentials: 'include',
		body: formData
	});

	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.error?.message || 'Import failed');
	}

	const data = await response.json();
	return data.data;
}

// Export API
import {
	EXPORT_ENDPOINTS,
	EXPORT_PRESETS,
	EXPORT_FIELDS,
	EXPORT_PRESET_FIELDS,
	isValidExportPreset,
	type ExportPreset,
	type ExportField
} from './contracts';

export type ExportType = 'csv' | 'zip';
export type { ExportPreset, ExportField };
export type ExportScope = 'all' | 'selected' | 'filtered';

// Re-export contract constants
export { EXPORT_ENDPOINTS, EXPORT_PRESETS, EXPORT_FIELDS, EXPORT_PRESET_FIELDS, isValidExportPreset };

export interface ExportOptions {
	type: ExportType;
	preset?: ExportPreset;
	fields?: string[];
	scope: ExportScope;
	plan_ids?: string[];
	categories?: string[];
	filters?: {
		search?: string;
		status?: string;
		type?: string;
		style?: string;
	};
}

export async function exportData(options: ExportOptions): Promise<Blob> {
	const params = new URLSearchParams();
	params.append('type', options.type);
	params.append('scope', options.scope);

	if (options.preset) {
		params.append('preset', options.preset);
	}
	if (options.fields?.length) {
		params.append('fields', options.fields.join(','));
	}
	if (options.plan_ids?.length) {
		params.append('plan_ids', options.plan_ids.join(','));
	}
	if (options.categories?.length) {
		params.append('categories', options.categories.join(','));
	}
	if (options.filters?.search) {
		params.append('search', options.filters.search);
	}
	if (options.filters?.status) {
		params.append('status', options.filters.status);
	}
	if (options.filters?.type) {
		params.append('type', options.filters.type);
	}
	if (options.filters?.style) {
		params.append('style', options.filters.style);
	}

	const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
	const endpoint = options.type === 'csv' ? '/export/csv' : '/export/zip';
	const response = await fetch(`${API_BASE}${endpoint}?${params.toString()}`, {
		method: 'GET',
		credentials: 'include'
	});

	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.error?.message || 'Export failed');
	}

	return response.blob();
}

// SFTP API
export async function getSFTPStatus(): Promise<SFTPStatus> {
	return api.get<SFTPStatus>('/sftp/status');
}

export async function getUserSFTP(userId: string): Promise<SFTPCredentials | null> {
	try {
		return await api.get<SFTPCredentials>(`/users/${userId}/sftp`);
	} catch (error: any) {
		if (error?.error?.code === 'NOT_FOUND') {
			return null;
		}
		throw error;
	}
}

export async function generateSFTP(userId: string, permission: 'read' | 'readwrite'): Promise<SFTPCredentials> {
	return api.post<SFTPCredentials>(`/users/${userId}/sftp`, { permission });
}

export async function rotateSFTP(userId: string): Promise<SFTPCredentials> {
	return api.put<SFTPCredentials>(`/users/${userId}/sftp/rotate`, {});
}

export async function revokeSFTP(userId: string): Promise<void> {
	return api.put<void>(`/users/${userId}/sftp/revoke`, {});
}

export async function updateSFTPPermission(userId: string, permission: 'read' | 'readwrite'): Promise<void> {
	return api.put<void>(`/users/${userId}/sftp/permission`, { permission });
}

export async function deleteSFTP(userId: string): Promise<void> {
	return api.delete<void>(`/users/${userId}/sftp`);
}

export async function listAllSFTP(): Promise<SFTPCredentials[]> {
	return api.get<SFTPCredentials[]>('/sftp/credentials');
}

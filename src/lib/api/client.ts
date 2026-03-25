import { goto } from '$app/navigation';
import type { ApiError, LoginCredentials, User } from '$lib/types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

export class ApiClient {
	private baseUrl: string;

	constructor(baseUrl: string = API_BASE_URL) {
		this.baseUrl = baseUrl;
	}

	private async fetch(
		endpoint: string,
		options: RequestInit = {},
		skipAuthRetry: boolean = false
	): Promise<Response> {
		const url = `${this.baseUrl}${endpoint}`;

		// Do not set Content-Type for FormData — the browser must set the multipart boundary.
		const isFormData = options.body instanceof FormData;
		const headers: Record<string, string> = {
			...(options.headers as Record<string, string>)
		};
		if (!isFormData) {
			headers['Content-Type'] = 'application/json';
		}

		const response = await fetch(url, {
			credentials: 'include',
			...options,
			headers
		});

		// Only try to refresh token for non-auth endpoints
		if (response.status === 401 && !skipAuthRetry) {
			const refreshed = await this.refreshToken();
			if (refreshed) {
				return this.fetch(endpoint, options);
			} else {
				auth.set(null);
				goto('/login');
				throw new Error('Session expired');
			}
		}

		return response;
	}

	private async refreshToken(): Promise<boolean> {
		try {
			const response = await fetch(`${this.baseUrl}/auth/refresh`, {
				method: 'POST',
				credentials: 'include'
			});
			return response.ok;
		} catch {
			return false;
		}
	}

	async login(credentials: LoginCredentials): Promise<User> {
		const response = await this.fetch(
			'/auth/login',
			{
				method: 'POST',
				body: JSON.stringify(credentials)
			},
			true
		); // Skip auth retry for login

		if (!response.ok) {
			const error: ApiError = await response.json();
			throw new Error(error.error.message);
		}

		const data = await response.json();
		return data.data.user;
	}

	async logout(): Promise<void> {
		try {
			await this.fetch('/auth/logout', {
				method: 'POST'
			});
		} finally {
			auth.set(null);
			goto('/login');
		}
	}

	async me(): Promise<User | null> {
		try {
			const response = await this.fetch('/auth/me');
			if (!response.ok) {
				return null;
			}
			const data = await response.json();
			return data.data;
		} catch {
			return null;
		}
	}

	async get<T>(endpoint: string): Promise<T> {
		const response = await this.fetch(endpoint);

		if (!response.ok) {
			const error: ApiError = await response.json();
			throw new Error(error.error.message);
		}

		const data = await response.json();
		return data.data;
	}

	async post<T>(endpoint: string, body: unknown): Promise<T> {
		const response = await this.fetch(endpoint, {
			method: 'POST',
			body: JSON.stringify(body)
		});

		if (!response.ok) {
			const error: ApiError = await response.json();
			throw new Error(error.error.message);
		}

		const data = await response.json();
		return data.data;
	}

	async put<T>(endpoint: string, body: unknown): Promise<T> {
		const response = await this.fetch(endpoint, {
			method: 'PUT',
			body: JSON.stringify(body)
		});

		if (!response.ok) {
			const error: ApiError = await response.json();
			throw new Error(error.error.message);
		}

		const data = await response.json();
		return data.data;
	}

	async delete<T>(endpoint: string): Promise<T> {
		const response = await this.fetch(endpoint, {
			method: 'DELETE'
		});

		if (!response.ok) {
			const error: ApiError = await response.json();
			throw new Error(error.error.message);
		}

		const data = await response.json();
		return data.data;
	}

	// getRaw returns the full JSON envelope without unwrapping .data.
	// Use for paginated endpoints that return { data: [...], meta: {...} }.
	async getRaw<T>(endpoint: string): Promise<T> {
		const response = await this.fetch(endpoint);
		if (!response.ok) {
			const error: ApiError = await response.json();
			throw new Error(error.error?.message || 'Request failed');
		}
		return response.json();
	}

	// postFormData sends a FormData body and returns the unwrapped .data field.
	// Content-Type is intentionally omitted so the browser sets the multipart boundary.
	async postFormData<T>(endpoint: string, formData: FormData): Promise<T> {
		const response = await this.fetch(endpoint, {
			method: 'POST',
			body: formData
		});
		if (!response.ok) {
			const error: ApiError = await response.json();
			throw new Error(error.error?.message || 'Request failed');
		}
		const data = await response.json();
		return data.data;
	}

	// getBlob fetches a binary response (e.g., CSV or ZIP export).
	async getBlob(endpoint: string): Promise<Blob> {
		const response = await this.fetch(endpoint);
		if (!response.ok) {
			const error: ApiError = await response.json();
			throw new Error(error.error?.message || 'Request failed');
		}
		return response.blob();
	}
}

import { auth } from '$lib/stores/auth';

export const api = new ApiClient();

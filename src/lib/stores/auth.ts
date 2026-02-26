import type { User } from '$lib/types';
import { writable } from 'svelte/store';

export type AuthState = User | null;

function createAuthStore() {
	const { subscribe, set, update } = writable<AuthState>(null);

	return {
		subscribe,
		set,
		update,
		clear: () => set(null),
		isAuthenticated: (state: AuthState): state is User => state !== null,
		isAdmin: (state: AuthState): boolean => state?.role === 'admin'
	};
}

export const auth = createAuthStore();

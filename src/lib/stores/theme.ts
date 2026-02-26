import { writable, derived, type Readable } from 'svelte/store';

export type Theme = 'system' | 'light' | 'dark';

interface ThemeState {
	theme: Theme;
	resolvedTheme: 'light' | 'dark';
}

const STORAGE_KEY = 'theme-preference';

function getSystemTheme(): 'light' | 'dark' {
	if (typeof window === 'undefined') return 'light';
	return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function getInitialTheme(): Theme {
	if (typeof window === 'undefined') return 'system';
	const stored = localStorage.getItem(STORAGE_KEY);
	if (stored === 'light' || stored === 'dark' || stored === 'system') {
		return stored;
	}
	return 'system';
}

function createThemeStore(): Readable<ThemeState> & {
	setTheme: (theme: Theme) => void;
	cycleTheme: () => void;
} {
	const { subscribe, set } = writable<ThemeState>({
		theme: 'system',
		resolvedTheme: 'light'
	});

	let currentTheme: Theme = 'system';
	let mediaQuery: MediaQueryList | null = null;

	function updateResolvedTheme() {
		const resolved = currentTheme === 'system' ? getSystemTheme() : currentTheme;
		set({ theme: currentTheme, resolvedTheme: resolved });

		if (typeof document !== 'undefined') {
			if (resolved === 'dark') {
				document.documentElement.classList.add('dark');
			} else {
				document.documentElement.classList.remove('dark');
			}
		}
	}

	function handleSystemChange() {
		if (currentTheme === 'system') {
			updateResolvedTheme();
		}
	}

	function init() {
		if (typeof window === 'undefined') return;

		currentTheme = getInitialTheme();
		updateResolvedTheme();

		mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
		mediaQuery.addEventListener('change', handleSystemChange);
	}

	function setTheme(theme: Theme) {
		currentTheme = theme;
		if (typeof localStorage !== 'undefined') {
			localStorage.setItem(STORAGE_KEY, theme);
		}
		updateResolvedTheme();
	}

	function cycleTheme() {
		const themes: Theme[] = ['system', 'light', 'dark'];
		const currentIndex = themes.indexOf(currentTheme);
		const nextTheme = themes[(currentIndex + 1) % themes.length];
		setTheme(nextTheme);
	}

	if (typeof window !== 'undefined') {
		init();
	}

	return {
		subscribe,
		setTheme,
		cycleTheme
	};
}

export const theme = createThemeStore();

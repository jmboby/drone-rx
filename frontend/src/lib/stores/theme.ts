import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'dark' | 'light';

const STORAGE_KEY = 'dronerx-theme';

function getInitialTheme(): Theme {
	if (browser) {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored === 'light' || stored === 'dark') return stored;
	}
	return 'dark';
}

function createThemeStore() {
	const { subscribe, update } = writable<Theme>(getInitialTheme());

	return {
		subscribe,
		toggle() {
			update((current) => {
				const next: Theme = current === 'dark' ? 'light' : 'dark';
				if (browser) {
					localStorage.setItem(STORAGE_KEY, next);
					document.body.dataset.theme = next;
				}
				return next;
			});
		},
		init() {
			// Apply current theme to body on mount
			if (browser) {
				const stored = localStorage.getItem(STORAGE_KEY);
				const theme = stored === 'light' ? 'light' : 'dark';
				document.body.dataset.theme = theme;
			}
		}
	};
}

export const theme = createThemeStore();

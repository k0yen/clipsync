import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Default to 'light' if not in browser
const initialTheme = browser ? (localStorage.getItem('theme') ?? 'light') : 'light';

export const theme = writable<'light' | 'dark'>(initialTheme as 'light');

// Subscribe to changes and update the HTML tag
theme.subscribe((value) => {
	if (browser) {
		localStorage.setItem('theme', value);
		if (value === 'dark') {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	}
});

export function toggleTheme() {
	theme.update((current) => (current === 'light' ? 'dark' : 'light'));
}

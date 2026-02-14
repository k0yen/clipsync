import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { theme, toggleTheme } from './theme';

// Mock the $app/environment module
vi.mock('$app/environment', () => ({
	browser: true
}));

describe('Store: Theme', () => {
	// Mock LocalStorage
	const localStorageMock = (() => {
		let store: Record<string, string> = {};
		return {
			getItem: vi.fn((key: string) => store[key] || null),
			setItem: vi.fn((key: string, value: string) => {
				store[key] = value.toString();
			}),
			clear: () => {
				store = {};
			}
		};
	})();

	beforeEach(() => {
		// Reset everything before each test
		Object.defineProperty(window, 'localStorage', {
			value: localStorageMock
		});
		localStorageMock.clear();

		// Reset the store to default 'light'
		theme.set('light');
	});

	it('initializes with default light theme', () => {
		expect(get(theme)).toBe('light');
	});

	it('toggles from light to dark', () => {
		toggleTheme();
		expect(get(theme)).toBe('dark');
	});

	it('persists preference to localStorage', () => {
		toggleTheme(); // Switch to dark
		expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'dark');
	});

	it('toggles back from dark to light', () => {
		theme.set('dark');
		toggleTheme();
		expect(get(theme)).toBe('light');
	});
});

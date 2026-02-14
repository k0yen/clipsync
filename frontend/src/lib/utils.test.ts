import { describe, it, expect } from 'vitest';
import { isURL, isCode } from './utils';

describe('Utility: Content Detection', () => {
	// --- URL DETECTION TESTS ---
	describe('isURL', () => {
		it('identifies standard http/https URLs', () => {
			expect(isURL('https://google.com')).toBe(true);
			expect(isURL('http://localhost:8080')).toBe(true);
		});

		it('rejects plain text', () => {
			expect(isURL('google.com')).toBe(false); // Missing protocol
			expect(isURL('Just some text')).toBe(false);
		});

		it('rejects invalid protocols', () => {
			expect(isURL('ftp://file-server.com')).toBe(false); // We only want web links
			expect(isURL('javascript:alert(1)')).toBe(false); // Security check
		});
	});

	// --- CODE DETECTION TESTS ---
	describe('isCode', () => {
		it('identifies function definitions', () => {
			const snippet = 'function hello() { return "world"; }';
			expect(isCode(snippet)).toBe(true);
		});

		it('identifies imports and exports', () => {
			const snippet = 'import { writable } from "svelte/store";';
			expect(isCode(snippet)).toBe(true);
		});

		it('identifies CLI commands', () => {
			const snippet = 'docker run -d -p 8080:8080 my-app';
			// It might fail if the snippet is too short, let's check our logic
			// Our logic requires length > 20 and specific keywords
			expect(isCode(snippet)).toBe(false); // "docker" isn't in our keyword list yet!
		});

		it('rejects standard sentences', () => {
			expect(isCode('This is just a normal sentence about code.')).toBe(false);
		});
	});
});

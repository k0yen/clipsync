import type { Snippet } from './socket';

export function startMockMode(clipboardItems: any) {
	console.log('🛠️ Mock Mode Active');

	// Initial Data
	clipboardItems.set([
		{
			id: '1',
			content: '🔒 This is a mock encrypted message.',
			timestamp: new Date().toISOString()
		},
		{
			id: '2',
			content: 'https://github.com/your-repo',
			timestamp: new Date().toISOString()
		}
	]);

	// Simulate incoming messages every 30 seconds
	const interval = setInterval(() => {
		const newSnippet: Snippet = {
			id: Math.random().toString(36).substring(7),
			content: 'Periodic mock update: ' + new Date().toLocaleTimeString(),
			timestamp: new Date().toISOString()
		};
		clipboardItems.update((items) => [newSnippet, ...items]);
	}, 30000);

	return () => clearInterval(interval);
}

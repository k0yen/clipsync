import { writable } from 'svelte/store';
import { PUBLIC_WS_URL, PUBLIC_USE_MOCK } from '$env/static/public';
import { encryptMessage, decryptMessage, importKey } from './crypto';
import { startMockMode } from './mockSocket';

export interface Snippet {
	id: string;
	content: string;
	timestamp: string;
}

export const clipboardItems = writable<Snippet[]>([]);
export const isConnected = writable(false);

let socket: WebSocket | null = null;
let activeKey: CryptoKey | null = null;

export async function connectToRoom(roomID: string, keyString: string) {
	// 1. Setup Encryption Key
	try {
		activeKey = await importKey(keyString);
	} catch (e) {
		console.error('Encryption Key Error:', e);
		return;
	}

	// 2. Handle Mock Mode Switch
	if (PUBLIC_USE_MOCK === 'true') {
		isConnected.set(true);
		startMockMode(clipboardItems);
		return;
	}

	// 3. Real WebSocket Connection
	if (socket) socket.close();

	socket = new WebSocket(`${PUBLIC_WS_URL}/${roomID}`);

	socket.onopen = () => {
		isConnected.set(true);
		console.log('📡 Connected to Backend');
	};

	socket.onmessage = async (event) => {
		const data = JSON.parse(event.data);

		// Handle both single snippets and history arrays
		const payloads = Array.isArray(data) ? data : [data];

		for (const payload of payloads) {
			if (activeKey && payload.iv && payload.content) {
				const decrypted = await decryptMessage(payload.iv, payload.content, activeKey);

				const snippet: Snippet = {
					id: payload.id,
					timestamp: payload.timestamp,
					content: decrypted
				};

				clipboardItems.update((items) => {
					// Prevent duplicates (especially during history replay)
					if (items.find((i) => i.id === snippet.id)) return items;
					return [snippet, ...items];
				});
			}
		}
	};

	socket.onclose = () => isConnected.set(false);
}

export async function sendSnippet(text: string) {
	if (PUBLIC_USE_MOCK === 'true') {
		const mockSnippet = {
			id: Date.now().toString(),
			content: text,
			timestamp: new Date().toISOString()
		};
		clipboardItems.update((items) => [mockSnippet, ...items]);
		return;
	}

	if (socket?.readyState === WebSocket.OPEN && activeKey) {
		const { iv, ciphertext } = await encryptMessage(text, activeKey);
		socket.send(
			JSON.stringify({
				content: ciphertext,
				iv: iv
			})
		);
	}
}

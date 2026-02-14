// Generates a random 32-byte key and encodes it as a URL-safe string
export async function generateKey(): Promise<string> {
	const keyBytes = new Uint8Array(32); // 256-bit key
	window.crypto.getRandomValues(keyBytes);
	// Convert to Base64Url (safe for browser address bars)
	return btoa(String.fromCharCode(...keyBytes))
		.replace(/\+/g, '-')
		.replace(/\//g, '_')
		.replace(/=+$/, '');
}

// Generates a short, readable Room ID
export function generateRoomID(): string {
	return Math.random().toString(36).substring(2, 9); // e.g. "x9f2k3q"
}

// Convert the URL-safe base64 string back to a binary CryptoKey
export async function importKey(keyStr: string): Promise<CryptoKey> {
	if (!window.crypto || !window.crypto.subtle) {
		throw new Error(
			'Crypto API is not available. This usually happens because you are ' +
				'accessing the site over HTTP instead of HTTPS or localhost.'
		);
	}

	// Fix URL-safe replacements
	const base64 = keyStr.replace(/-/g, '+').replace(/_/g, '/');
	// Pad with '=' if needed
	const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=');

	const binary = atob(padded);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}

	return await window.crypto.subtle.importKey('raw', bytes, 'AES-GCM', true, [
		'encrypt',
		'decrypt'
	]);
}

// ENCRYPT: Text -> { iv, ciphertext }
export async function encryptMessage(
	text: string,
	key: CryptoKey
): Promise<{ iv: string; ciphertext: string }> {
	const encoder = new TextEncoder();
	const data = encoder.encode(text);

	// Generate a random 12-byte IV (Initialization Vector) - CRITICAL for security
	const iv = window.crypto.getRandomValues(new Uint8Array(12));

	const encryptedBuffer = await window.crypto.subtle.encrypt(
		{ name: 'AES-GCM', iv: iv },
		key,
		data
	);

	// Convert buffers to Base64 to send over JSON
	return {
		iv: btoa(String.fromCharCode(...iv)),
		ciphertext: btoa(String.fromCharCode(...new Uint8Array(encryptedBuffer)))
	};
}

// DECRYPT: { iv, ciphertext } -> Text
export async function decryptMessage(
	ivStr: string,
	ciphertextStr: string,
	key: CryptoKey
): Promise<string> {
	try {
		// Convert Base64 back to Uint8Array
		const iv = Uint8Array.from(atob(ivStr), (c) => c.charCodeAt(0));
		const ciphertext = Uint8Array.from(atob(ciphertextStr), (c) => c.charCodeAt(0));

		const decryptedBuffer = await window.crypto.subtle.decrypt(
			{ name: 'AES-GCM', iv: iv },
			key,
			ciphertext
		);

		return new TextDecoder().decode(decryptedBuffer);
	} catch (e) {
		console.error('Decryption failed:', e);
		return '⚠️ Failed to decrypt message (Wrong Key?)';
	}
}

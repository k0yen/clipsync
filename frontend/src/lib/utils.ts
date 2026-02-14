export function isURL(text: string): boolean {
	try {
		const url = new URL(text);
		// Ensure it has a protocol to avoid false positives like "hello.world"
		return ['http:', 'https:', 'ftp:'].includes(url.protocol);
	} catch {
		return false;
	}
}

export function isCode(text: string): boolean {
	const keywords = [
		'function',
		'const',
		'let',
		'var',
		'import',
		'export',
		'class',
		'return',
		'if',
		'else',
		'for',
		'while',
		'docker',
		'npm',
		'bun',
		'sudo',
		'git',
		'echo',
		'cat',
		'ssh'
	];

	const lines = text.split('\n');

	// Logic: If it has keywords OR code-like symbols ({, }, ;) AND isn't just a short sentence
	const hasKeyword = keywords.some((k) => text.includes(k));
	const hasSymbols = /[{};<>$]/.test(text);

	return (hasKeyword || hasSymbols) && text.length > 10;
}

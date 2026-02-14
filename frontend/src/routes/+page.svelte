<script lang="ts">
	import { goto } from '$app/navigation';
	import { generateKey, generateRoomID } from '$lib/crypto';
	import { ArrowRight, ShieldCheck, Zap, Lock } from 'lucide-svelte';
	import { theme, toggleTheme } from '$lib/theme';
	import { Moon, Sun } from 'lucide-svelte';

	let joinInput = '';
	let isCreating = false;

	async function createRoom() {
		isCreating = true;
		// 1. Generate unique ID and Encryption Key
		const roomId = generateRoomID();
		const key = await generateKey();

		// 2. Redirect to the room.
		// The key is in the HASH (#), so the server never sees it.
		goto(`/${roomId}#${key}`);
	}

	function handleJoin() {
		if (!joinInput.trim()) return;

		// 1. Handle "room#key" or full URL "http://.../room#key"
		if (joinInput.includes('#')) {
			const parts = joinInput.split('#');
			// If it was a URL, the first part might be a path. Clean it.
			const roomPart = parts[0].includes('/') ? parts[0].split('/').pop() : parts[0];
			const keyPart = parts[1];

			goto(`/${roomPart}#${keyPart}`);
			return;
		}

		// 2. Fallback: Treat as raw Room ID (no key provided)
		const cleanID = joinInput.replace(/[^a-zA-Z0-9-]/g, '');
		goto(`/${cleanID}`);
	}
</script>

<div
	class="relative min-h-screen overflow-hidden bg-gray-50 font-sans text-gray-900 transition-colors duration-500 dark:bg-black dark:text-gray-100"
>
	<div class="pointer-events-none absolute top-0 left-0 h-full w-full overflow-hidden">
		<div
			class="absolute -top-[20%] -left-[10%] h-[50vw] w-[50vw] animate-pulse rounded-full bg-indigo-500/20 blur-[100px]"
		></div>
		<div
			class="absolute top-[40%] -right-[10%] h-[40vw] w-[40vw] animate-pulse rounded-full bg-purple-500/20 blur-[100px] delay-1000"
		></div>
	</div>

	<nav class="relative z-10 mx-auto flex w-full max-w-5xl items-center justify-between p-6">
		<div class="flex items-center gap-2 text-xl font-bold tracking-tighter">
			<div class="rounded-lg bg-indigo-600 p-2 text-white shadow-lg shadow-indigo-500/30">
				<Zap size={20} fill="currentColor" />
			</div>
			<span>ClipSync</span>
		</div>

		<button
			on:click={toggleTheme}
			class="rounded-full p-2 transition-colors hover:bg-black/5 dark:hover:bg-white/10"
		>
			{#if $theme === 'dark'}
				<Sun size={20} />
			{:else}
				<Moon size={20} />
			{/if}
		</button>
	</nav>

	<main
		class="relative z-10 mx-auto flex min-h-[80vh] max-w-3xl flex-col items-center justify-center space-y-8 px-4 text-center"
	>
		<div class="space-y-4">
			<div
				class="inline-flex items-center gap-2 rounded-full border border-indigo-100 bg-indigo-50 px-3 py-1 text-xs font-semibold tracking-wider text-indigo-600 uppercase dark:border-indigo-500/20 dark:bg-indigo-900/30 dark:text-indigo-400"
			>
				<ShieldCheck size={12} />
				<span>End-to-End Encrypted</span>
			</div>

			<h1
				class="bg-gradient-to-r from-gray-900 to-gray-600 bg-clip-text pb-2 text-5xl font-extrabold tracking-tight text-transparent md:text-7xl dark:from-white dark:to-gray-400"
			>
				Sync clipboards <br /> securely.
			</h1>

			<p class="mx-auto max-w-xl text-lg leading-relaxed text-gray-600 dark:text-gray-400">
				Real-time, peer-encrypted clipboard sharing. No signups, no logs. Your data never leaves
				your browser unencrypted.
			</p>
		</div>

		<div
			class="w-full max-w-md rounded-2xl border border-gray-200 bg-white/80 p-2 shadow-xl backdrop-blur-xl transition-all duration-300 hover:shadow-2xl dark:border-white/10 dark:bg-neutral-900/60"
		>
			<button
				on:click={createRoom}
				disabled={isCreating}
				class="group relative flex w-full items-center justify-center gap-3 overflow-hidden rounded-xl bg-indigo-600 py-4 font-semibold text-white transition-all hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-70"
			>
				<div
					class="absolute inset-0 h-full w-full -translate-x-full bg-gradient-to-r from-transparent via-white/20 to-transparent group-hover:animate-[shimmer_1.5s_infinite]"
				></div>

				{#if isCreating}
					<span>Creating secure room...</span>
				{:else}
					<Zap size={20} />
					<span>Create Secure Room</span>
					<ArrowRight size={18} class="transition-transform group-hover:translate-x-1" />
				{/if}
			</button>

			<div class="relative py-4">
				<div class="absolute inset-0 flex items-center">
					<div class="w-full border-t border-gray-200 dark:border-white/10"></div>
				</div>
				<div class="relative flex justify-center text-xs uppercase">
					<span class="bg-white px-2 text-gray-500 dark:bg-neutral-900">or join existing</span>
				</div>
			</div>

			<form on:submit|preventDefault={handleJoin} class="flex gap-2">
				<input
					bind:value={joinInput}
					type="text"
					placeholder="Enter Join Code..."
					class="flex-1 rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm placeholder-gray-400 transition-all outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500 dark:border-white/10 dark:bg-black/50 dark:text-white"
				/>
				<button
					type="submit"
					class="rounded-xl bg-gray-100 p-3 text-gray-900 transition-colors hover:bg-gray-200 dark:bg-white/10 dark:text-white dark:hover:bg-white/20"
					aria-label="Join Room"
				>
					<ArrowRight size={20} />
				</button>
			</form>
		</div>

		<div
			class="grid grid-cols-1 gap-6 pt-8 text-sm text-gray-500 md:grid-cols-3 dark:text-gray-400"
		>
			<div class="flex flex-col items-center gap-2">
				<Lock size={20} class="text-indigo-500" />
				<span class="font-medium">AES-256 Encryption</span>
			</div>
			<div class="flex flex-col items-center gap-2">
				<Zap size={20} class="text-indigo-500" />
				<span class="font-medium">Real-time WebSocket</span>
			</div>
			<div class="flex flex-col items-center gap-2">
				<ShieldCheck size={20} class="text-indigo-500" />
				<span class="font-medium">No Server Logs</span>
			</div>
		</div>
	</main>
</div>

<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation'; // Now we actually use this!
	import { connectToRoom, sendSnippet, clipboardItems, isConnected } from '$lib/socket';
	import { theme, toggleTheme } from '$lib/theme';
	import SnippetCard from '$lib/components/SnippetCard.svelte';
	import { Link, Share2, Check, Send, Moon, Sun } from 'lucide-svelte';

	let roomID = $page.params.room;
	let inputText = '';

	onMount(() => {
		// 1. Extract the Encryption Key from the URL Hash (#)
		const hash = window.location.hash.substring(1); // Remove the '#' char

		// 2. Security Check: If no key, you can't read anything. Go home.
		if (!hash) {
			alert('⚠️ Encryption Key Missing!\nYou cannot join a secure room without the full link.');
			goto('/'); // Redirect to Landing Page
			return;
		}

		// 3. Connect using the key
		connectToRoom(roomID, hash);
	});

	function handleSend() {
		if (!inputText.trim()) return;
		sendSnippet(inputText);
		inputText = '';
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.ctrlKey && e.key === 'Enter') {
			handleSend();
		}
	}

	let showCopyFeedback = false;

	function copyJoinCode() {
		const hash = window.location.hash.substring(1);
		const joinCode = `${roomID}#${hash}`;
		navigator.clipboard.writeText(joinCode);

		// Simple toggle for visual feedback
		showCopyFeedback = true;
		setTimeout(() => (showCopyFeedback = false), 2000);
	}
</script>

<div
	class="min-h-screen bg-gray-50 pb-20 font-sans text-gray-900 transition-colors duration-300 dark:bg-black dark:text-gray-100"
>
	<header
		class="sticky top-0 z-10 border-b border-gray-200 bg-white/80 backdrop-blur-xl transition-colors duration-300 dark:border-white/10 dark:bg-black/70"
	>
		<div class="mx-auto flex max-w-3xl items-center justify-between px-4 py-4">
			<div class="flex items-center gap-3">
				<div class="rounded-lg bg-indigo-600 p-2 text-white shadow-md shadow-indigo-500/20">
					<Link size={20} />
				</div>
				<div>
					<h1 class="text-lg leading-tight font-bold tracking-tight">ClipSync</h1>
					<p class="font-mono text-xs tracking-widest text-gray-500 uppercase dark:text-gray-400">
						{roomID}
					</p>
				</div>
			</div>

			<div class="flex items-center gap-1 sm:gap-2">
				<div
					class="mr-2 flex items-center gap-2 rounded-full border border-gray-200 bg-gray-50 px-2 py-1 dark:border-white/10 dark:bg-white/5"
				>
					<span class="relative flex h-2 w-2">
						{#if $isConnected}
							<span
								class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"
							></span>
							<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
						{:else}
							<span class="relative inline-flex h-2 w-2 rounded-full bg-red-500"></span>
						{/if}
					</span>
				</div>

				<button
					on:click={copyJoinCode}
					class="rounded-lg p-2 text-gray-500 transition-all hover:bg-gray-100 active:scale-95 dark:text-gray-400 dark:hover:bg-white/10"
					title="Copy Join Code (room#key)"
				>
					{#if showCopyFeedback}
						<Check size={20} class="text-green-500" />
					{:else}
						<Share2 size={20} />
					{/if}
				</button>

				<button
					on:click={toggleTheme}
					class="rounded-lg p-2 text-gray-500 transition-all hover:bg-gray-100 active:scale-95 dark:text-gray-400 dark:hover:bg-white/10"
				>
					{#if $theme === 'dark'}
						<Sun size={20} />
					{:else}
						<Moon size={20} />
					{/if}
				</button>
			</div>
		</div>
	</header>

	<main class="mx-auto mt-8 max-w-3xl space-y-8 px-4">
		<section
			class="rounded-2xl border border-gray-200 bg-white p-1 shadow-sm backdrop-blur-md transition-colors duration-300 dark:border-white/10 dark:bg-neutral-900/50"
		>
			<div class="group relative">
				<textarea
					bind:value={inputText}
					on:keydown={handleKeydown}
					rows="3"
					class="block w-full resize-none rounded-xl border-transparent bg-transparent p-4 text-gray-900 placeholder-gray-400 focus:border-indigo-500 focus:ring-0 dark:text-white"
					placeholder="Paste text here..."
				></textarea>

				<div class="absolute right-2 bottom-2 flex items-center gap-2">
					<span class="hidden text-xs text-gray-400 sm:inline-block dark:text-gray-500"
						>Ctrl + Enter to send</span
					>
					<button
						on:click={handleSend}
						class="rounded-lg bg-indigo-600 p-2 text-white shadow-lg shadow-indigo-500/30 transition-all hover:bg-indigo-700 hover:shadow-indigo-500/50"
					>
						<Send size={18} />
					</button>
				</div>
			</div>
		</section>

		<section class="space-y-4 pb-20">
			{#if $clipboardItems.length === 0}
				<div class="py-12 text-center text-gray-400 dark:text-gray-500">
					<p>No clips yet. Start copying!</p>
				</div>
			{/if}

			{#each $clipboardItems as item (item.id)}
				<SnippetCard content={item.content} timestamp={item.timestamp} />
			{/each}
		</section>
	</main>
</div>

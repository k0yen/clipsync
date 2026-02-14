<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { Copy, Check, ExternalLink, Code } from 'lucide-svelte';
	import { isURL, isCode } from '$lib/utils';

	export let content: string;
	export let timestamp: string;

	let copied = false;

	const isLink = isURL(content);
	const isCodeBlock = !isLink && isCode(content);

	async function copyToClipboard() {
		try {
			await navigator.clipboard.writeText(content);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch (err) {
			console.error('Failed to copy!', err);
		}
	}
</script>

<div
	in:fly={{ y: 20, duration: 300 }}
	out:fade
	class="group relative rounded-xl border border-gray-200 bg-white p-5 shadow-sm backdrop-blur-lg transition-all duration-300 hover:shadow-md dark:border-white/10 dark:bg-neutral-900/40 hover:dark:border-white/20"
>
	<div class="mb-3 flex items-center justify-between">
		<div class="flex items-center gap-2">
			{#if isLink}
				<span class="rounded bg-blue-50 p-1 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400">
					<ExternalLink size={14} />
				</span>
				<span class="text-xs font-medium text-blue-600 uppercase dark:text-blue-400">Link</span>
			{:else if isCodeBlock}
				<span class="rounded bg-gray-100 p-1 text-gray-600 dark:bg-white/5 dark:text-gray-300">
					<Code size={14} />
				</span>
				<span class="text-xs font-medium text-gray-600 uppercase dark:text-gray-400">Code</span>
			{:else}
				<span class="text-xs font-medium tracking-wider text-gray-400 uppercase dark:text-gray-500"
					>{timestamp}</span
				>
			{/if}
		</div>

		<button
			on:click={copyToClipboard}
			class="rounded-lg p-2 text-gray-400 transition-colors hover:bg-indigo-50 hover:text-indigo-600 dark:hover:bg-white/10 dark:hover:text-indigo-400"
			title="Copy to clipboard"
		>
			{#if copied}
				<Check size={18} class="text-green-500" />
			{:else}
				<Copy size={18} />
			{/if}
		</button>
	</div>

	<div class="prose max-w-none break-words prose-slate dark:prose-invert">
		{#if isLink}
			<a
				href={content}
				target="_blank"
				rel="noopener noreferrer"
				class="break-all text-blue-600 hover:underline dark:text-blue-400"
			>
				{content}
			</a>
		{:else if isCodeBlock}
			<pre
				class="overflow-x-auto rounded-lg border border-gray-800 bg-gray-900 p-4 font-mono text-sm text-gray-100 dark:border-white/10 dark:bg-black/80">
        <code>{content}</code>
      </pre>
		{:else}
			<p class="leading-relaxed whitespace-pre-wrap text-gray-700 dark:text-gray-300">{content}</p>
		{/if}
	</div>
</div>

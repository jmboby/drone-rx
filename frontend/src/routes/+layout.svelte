<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { getUpdates } from '$lib/api';
	import type { UpdateInfo } from '$lib/types';

	let { children } = $props();

	let latestUpdate = $state<UpdateInfo | null>(null);
	let bannerDismissed = $state(false);

	let showBanner = $derived(latestUpdate !== null && !bannerDismissed);

	onMount(async () => {
		try {
			const updates = await getUpdates();
			if (updates && updates.length > 0) {
				latestUpdate = updates[0];
			}
		} catch {
			// silent — update banner is non-critical
		}
	});
</script>

{#if showBanner && latestUpdate}
	<div class="relative z-50 border-b border-amber-500/30 bg-amber-500/10 backdrop-blur-sm">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 py-2.5 flex items-center gap-3">
			<svg class="w-4 h-4 text-amber-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
			</svg>
			<p class="flex-1 text-sm text-amber-200">
				<span class="font-semibold text-amber-300">Update available:</span>
				<span class="ml-1.5 font-mono text-amber-400">{latestUpdate.versionLabel}</span>
				{#if latestUpdate.releaseNotes}
					<span class="text-amber-300/70 ml-2 hidden sm:inline">&mdash; {latestUpdate.releaseNotes}</span>
				{/if}
			</p>
			<button
				onclick={() => { bannerDismissed = true; }}
				class="shrink-0 text-amber-400/70 hover:text-amber-300 transition-colors p-1 rounded"
				aria-label="Dismiss update banner"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>
	</div>
{/if}

<div class="min-h-screen bg-navy-950 grid-bg">
	{@render children()}
</div>

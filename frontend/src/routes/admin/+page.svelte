<script lang="ts">
	import { generateSupportBundle } from '$lib/api';
	import DroneIcon from '$lib/components/DroneIcon.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { getContext } from 'svelte';
	import type { Writable } from 'svelte/store';

	const lightModeEnabled = getContext<Writable<boolean>>('lightModeEnabled');

	let loading = $state(false);
	let result = $state<{ status: string; message: string } | null>(null);
	let showConfirm = $state(false);

	async function handleGenerate() {
		showConfirm = false;
		loading = true;
		result = null;

		try {
			const response = await generateSupportBundle();
			result = response;
		} catch (err) {
			result = {
				status: 'error',
				message: err instanceof Error ? err.message : 'An unexpected error occurred',
			};
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Admin — DroneRx</title>
</svelte:head>

<header class="sticky top-0 z-20 border-b border-navy-700/60 bg-navy-900/80 backdrop-blur-xl">
	<div class="max-w-6xl mx-auto px-4 sm:px-6 py-3.5 flex items-center justify-between">
		<div class="flex items-center gap-2.5">
			<span class="text-cyan-glow"><DroneIcon size="w-7 h-7" /></span>
			<span class="text-xl font-bold tracking-tight text-white">DroneRx</span>
			<span class="text-xs text-navy-300 hidden sm:inline ml-1 font-medium">Admin</span>
		</div>
		<nav class="flex items-center gap-4">
			<a href="/" class="text-sm font-medium text-navy-200 hover:text-cyan-glow transition-colors">
				Back to Store
			</a>
			{#if $lightModeEnabled}
				<span class="text-navy-600">|</span>
				<ThemeToggle />
			{/if}
		</nav>
	</div>
</header>

<main class="max-w-2xl mx-auto px-4 sm:px-6 py-12">
	<h1 class="text-2xl font-bold text-white mb-2">Admin</h1>
	<p class="text-navy-300 text-sm mb-8">Operational tools for DroneRx administrators.</p>

	<div class="glass-card rounded-xl border border-navy-700/60 p-6">
		<h2 class="text-lg font-semibold text-white mb-1">Support Bundle</h2>
		<p class="text-navy-300 text-sm mb-5">
			Collect diagnostic data from this cluster and upload it to the Vendor Portal for troubleshooting.
		</p>

		{#if loading}
			<div class="flex items-center gap-3 text-cyan-400">
				<svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
					<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
				</svg>
				<span class="text-sm font-medium">Generating support bundle... This may take a minute.</span>
			</div>
		{:else if showConfirm}
			<div class="bg-amber-500/10 border border-amber-500/30 rounded-lg p-4 mb-4">
				<p class="text-amber-200 text-sm mb-3">
					This will collect diagnostic data from this cluster and upload it to the vendor. Continue?
				</p>
				<div class="flex gap-3">
					<button
						onclick={handleGenerate}
						class="bg-cyan-glow/15 hover:bg-cyan-glow/25 text-cyan-glow text-sm font-semibold px-4 py-2 rounded-lg border border-cyan-glow/30 transition-all"
					>
						Yes, generate
					</button>
					<button
						onclick={() => { showConfirm = false; }}
						class="text-navy-300 hover:text-navy-100 text-sm font-medium px-4 py-2 rounded-lg border border-navy-600 transition-all"
					>
						Cancel
					</button>
				</div>
			</div>
		{:else}
			<button
				onclick={() => { showConfirm = true; }}
				class="bg-cyan-glow/15 hover:bg-cyan-glow/25 text-cyan-glow text-sm font-semibold px-4 py-2 rounded-lg border border-cyan-glow/30 transition-all"
			>
				Generate Support Bundle
			</button>
		{/if}

		{#if result}
			<div class="mt-4 rounded-lg p-4 {result.status === 'ok' ? 'bg-emerald-500/10 border border-emerald-500/30' : 'bg-red-500/10 border border-red-500/30'}">
				<p class="text-sm {result.status === 'ok' ? 'text-emerald-300' : 'text-red-300'}">
					{result.message}
				</p>
			</div>
		{/if}
	</div>
</main>

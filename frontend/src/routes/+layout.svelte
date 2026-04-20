<script lang="ts">
	import '../app.css';
	import { onMount, setContext } from 'svelte';
	import { getUpdates, getLicenseStatus, getUIConfig } from '$lib/api';
	import type { UpdateInfo, LicenseStatus } from '$lib/types';
	import { theme } from '$lib/stores/theme';
	import { writable } from 'svelte/store';

	let { children } = $props();

	let latestUpdate = $state<UpdateInfo | null>(null);
	let license = $state<LicenseStatus | null>(null);
	let bannerDismissed = $state(false);

	let showUpdateBanner = $derived(latestUpdate !== null && !bannerDismissed);
	let showLicenseWarning = $derived(license !== null && (license.expired || !license.valid));

	// UI toggles sourced from plain config (not license-gated).
	const lightModeEnabled = writable(false);
	const adminLinkVisible = writable(false);
	setContext('lightModeEnabled', lightModeEnabled);
	setContext('adminLinkVisible', adminLinkVisible);

	onMount(async () => {
		theme.init();

		try {
			const [updates, licenseStatus, uiConfig] = await Promise.all([
				getUpdates().catch(() => []),
				getLicenseStatus().catch(() => null),
				getUIConfig().catch(() => null),
			]);
			if (updates && updates.length > 0) {
				latestUpdate = updates[0];
			}
			if (licenseStatus) {
				license = licenseStatus;
			}
			if (uiConfig) {
				lightModeEnabled.set(uiConfig.light_mode_enabled ?? false);
				adminLinkVisible.set(uiConfig.admin_link_visible ?? false);
			}
		} catch {
			// silent — banners are non-critical
		}
	});
</script>

{#if showUpdateBanner && latestUpdate}
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

<div class="min-h-screen bg-navy-950 grid-bg {showLicenseWarning ? 'blur-sm pointer-events-none select-none' : ''}">
	{@render children()}
</div>

{#if showLicenseWarning}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-navy-950/60 backdrop-blur-sm">
		<div class="glass-card rounded-2xl border border-red-500/30 p-8 max-w-md mx-4 text-center shadow-2xl shadow-red-500/10">
			<div class="w-16 h-16 rounded-full bg-red-500/15 border border-red-500/30 flex items-center justify-center mx-auto mb-5">
				<svg class="w-8 h-8 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
				</svg>
			</div>
			<h2 class="text-xl font-bold text-white mb-2">License Expired</h2>
			<p class="text-navy-300 text-sm leading-relaxed mb-6">
				Your DroneRx license has expired. Please contact your administrator to renew your license and restore access.
			</p>
			{#if license?.expiration_date}
				<p class="text-xs text-navy-500 font-mono">
					Expired: {new Date(license.expiration_date).toLocaleDateString()}
				</p>
			{/if}
		</div>
	</div>
{/if}

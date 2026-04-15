<script lang="ts">
	import { theme } from '$lib/stores/theme';
</script>

<button
	onclick={() => theme.toggle()}
	class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-full border transition-all text-sm
		{$theme === 'light'
			? 'bg-amber-glow/10 border-amber-glow/30 text-amber-glow'
			: 'bg-navy-700/50 border-navy-600 text-navy-300 hover:border-navy-400 hover:text-navy-200'}"
	aria-label="Toggle {$theme === 'dark' ? 'light' : 'dark'} mode"
	title="{$theme === 'dark' ? 'Light' : 'Dark'} mode"
>
	{#if $theme === 'dark'}
		<!-- Sun icon -->
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
		</svg>
	{:else}
		<!-- Moon icon -->
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
		</svg>
	{/if}
</button>

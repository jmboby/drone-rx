<script lang="ts">
	import type { PageData } from './$types';
	import MedicineCard from '$lib/components/MedicineCard.svelte';
	import DroneIcon from '$lib/components/DroneIcon.svelte';
	import { cartCount } from '$lib/stores/cart';

	let { data }: { data: PageData } = $props();

	let selectedCategory = $state('All');

	let filteredMedicines = $derived(
		selectedCategory === 'All'
			? data.medicines
			: data.medicines.filter((m) => m.category === selectedCategory)
	);
</script>

<svelte:head>
	<title>DroneRx — Medicine Delivered by Drone</title>
</svelte:head>

<!-- Header -->
<header class="sticky top-0 z-20 border-b border-navy-700/60 bg-navy-900/80 backdrop-blur-xl">
	<div class="max-w-6xl mx-auto px-4 sm:px-6 py-3.5 flex items-center justify-between">
		<div class="flex items-center gap-2.5">
			<span class="text-cyan-glow"><DroneIcon size="w-7 h-7" /></span>
			<span class="text-xl font-bold tracking-tight text-white">DroneRx</span>
			<span class="text-xs text-navy-300 hidden sm:inline ml-1 font-medium">Aerial Pharmacy</span>
		</div>
		<nav class="flex items-center gap-4">
			<a
				href="/orders"
				class="text-sm font-medium text-navy-200 hover:text-cyan-glow transition-colors"
			>
				My Orders
			</a>
			<a
				href="/order"
				class="relative flex items-center gap-1.5 bg-cyan-glow/10 hover:bg-cyan-glow/20 text-cyan-glow text-sm font-semibold px-4 py-2 rounded-lg border border-cyan-glow/30 transition-all"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 100 4 2 2 0 000-4z" />
				</svg>
				<span>Cart</span>
				{#if $cartCount > 0}
					<span class="absolute -top-2 -right-2 bg-amber-glow text-navy-950 text-xs font-bold w-5 h-5 rounded-full flex items-center justify-center shadow-lg shadow-amber-glow/30">
						{$cartCount}
					</span>
				{/if}
			</a>
		</nav>
	</div>
</header>

<!-- Hero -->
<div class="relative overflow-hidden">
	<div class="absolute inset-0 bg-gradient-to-br from-navy-800 via-navy-900 to-navy-950"></div>
	<div class="absolute inset-0 opacity-30" style="background: radial-gradient(ellipse at 70% 50%, rgba(0, 229, 255, 0.15), transparent 60%);"></div>
	<div class="relative max-w-6xl mx-auto px-4 sm:px-6 py-14">
		<div class="flex items-center gap-4 mb-4">
			<span class="text-cyan-glow/80"><DroneIcon size="w-10 h-10" animated /></span>
			<div>
				<h1 class="text-3xl sm:text-4xl font-bold text-white tracking-tight">Medicine Delivered Fast</h1>
				<p class="text-navy-200 text-lg mt-1">
					Order prescription and OTC medicines. Our drone fleet delivers straight to your door.
				</p>
			</div>
		</div>
	</div>
</div>

<!-- Main content -->
<main class="max-w-6xl mx-auto px-4 sm:px-6 py-8">
	<!-- Category filter -->
	<div class="flex flex-wrap gap-2 mb-8">
		<button
			onclick={() => selectedCategory = 'All'}
			class="px-4 py-1.5 rounded-full text-sm font-medium transition-all
				{selectedCategory === 'All'
					? 'bg-cyan-glow/15 text-cyan-glow border border-cyan-glow/40 shadow-sm shadow-cyan-glow/10'
					: 'text-navy-300 border border-navy-600 hover:border-cyan-glow/30 hover:text-cyan-300'}"
		>
			All
		</button>
		{#each data.categories as category}
			<button
				onclick={() => selectedCategory = category}
				class="px-4 py-1.5 rounded-full text-sm font-medium transition-all
					{selectedCategory === category
						? 'bg-cyan-glow/15 text-cyan-glow border border-cyan-glow/40 shadow-sm shadow-cyan-glow/10'
						: 'text-navy-300 border border-navy-600 hover:border-cyan-glow/30 hover:text-cyan-300'}"
			>
				{category}
			</button>
		{/each}
	</div>

	<!-- Results count -->
	<p class="text-sm text-navy-400 mb-4 font-medium">
		{filteredMedicines.length} medicine{filteredMedicines.length !== 1 ? 's' : ''}
		{selectedCategory !== 'All' ? `in ${selectedCategory}` : 'available'}
	</p>

	<!-- Medicine grid -->
	{#if filteredMedicines.length === 0}
		<div class="text-center py-16 text-navy-400">
			<svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
			</svg>
			<p class="text-lg font-medium">No medicines found</p>
			<p class="text-sm mt-1 text-navy-500">Try a different category</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each filteredMedicines as medicine, i (medicine.id)}
				<div class="animate-slide-up" style="animation-delay: {Math.min(i * 50, 300)}ms">
					<MedicineCard {medicine} />
				</div>
			{/each}
		</div>
	{/if}
</main>

<script lang="ts">
	import type { PageData } from './$types';
	import MedicineCard from '$lib/components/MedicineCard.svelte';
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

<div class="min-h-screen bg-slate-50">
	<!-- Header -->
	<header class="bg-white border-b border-slate-200 sticky top-0 z-10 shadow-sm">
		<div class="max-w-6xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
			<div class="flex items-center gap-2">
				<span class="text-2xl">🚁</span>
				<span class="text-xl font-bold text-teal-700 tracking-tight">DroneRx</span>
				<span class="text-xs text-slate-400 hidden sm:inline ml-1">Medicine by Drone</span>
			</div>
			<nav class="flex items-center gap-4">
				<a
					href="/orders"
					class="text-sm font-medium text-slate-600 hover:text-teal-700 transition-colors"
				>
					My Orders
				</a>
				<a
					href="/order"
					class="relative flex items-center gap-1.5 bg-teal-600 hover:bg-teal-700 text-white text-sm font-semibold px-4 py-2 rounded-lg transition-colors"
				>
					<span>🛒</span>
					<span>Cart</span>
					{#if $cartCount > 0}
						<span class="absolute -top-2 -right-2 bg-rose-500 text-white text-xs font-bold w-5 h-5 rounded-full flex items-center justify-center">
							{$cartCount}
						</span>
					{/if}
				</a>
			</nav>
		</div>
	</header>

	<!-- Hero -->
	<div class="bg-gradient-to-br from-teal-600 to-teal-800 text-white">
		<div class="max-w-6xl mx-auto px-4 sm:px-6 py-12">
			<h1 class="text-3xl sm:text-4xl font-bold mb-2">Medicine Delivered Fast</h1>
			<p class="text-teal-100 text-lg max-w-xl">
				Order prescription and OTC medicines. Our drone fleet delivers straight to your door.
			</p>
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
						? 'bg-teal-600 text-white shadow-sm'
						: 'bg-white text-slate-600 border border-slate-200 hover:border-teal-400 hover:text-teal-700'}"
			>
				All
			</button>
			{#each data.categories as category}
				<button
					onclick={() => selectedCategory = category}
					class="px-4 py-1.5 rounded-full text-sm font-medium transition-all
						{selectedCategory === category
							? 'bg-teal-600 text-white shadow-sm'
							: 'bg-white text-slate-600 border border-slate-200 hover:border-teal-400 hover:text-teal-700'}"
				>
					{category}
				</button>
			{/each}
		</div>

		<!-- Results count -->
		<p class="text-sm text-slate-500 mb-4">
			{filteredMedicines.length} medicine{filteredMedicines.length !== 1 ? 's' : ''}
			{selectedCategory !== 'All' ? `in ${selectedCategory}` : 'available'}
		</p>

		<!-- Medicine grid -->
		{#if filteredMedicines.length === 0}
			<div class="text-center py-16 text-slate-400">
				<p class="text-4xl mb-3">💊</p>
				<p class="text-lg font-medium">No medicines found</p>
				<p class="text-sm mt-1">Try a different category</p>
			</div>
		{:else}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each filteredMedicines as medicine (medicine.id)}
					<MedicineCard {medicine} />
				{/each}
			</div>
		{/if}
	</main>
</div>

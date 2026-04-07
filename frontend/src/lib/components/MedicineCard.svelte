<script lang="ts">
	import type { Medicine } from '$lib/types';
	import { cart } from '$lib/stores/cart';

	interface Props {
		medicine: Medicine;
	}

	let { medicine }: Props = $props();

	let added = $state(false);

	function handleAdd() {
		cart.add(medicine);
		added = true;
		setTimeout(() => { added = false; }, 1500);
	}
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-100 p-5 flex flex-col gap-3 hover:shadow-md transition-shadow duration-200">
	<div class="flex items-start justify-between gap-2">
		<span class="text-xs font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full bg-teal-50 text-teal-700 border border-teal-200">
			{medicine.category}
		</span>
		{#if medicine.in_stock}
			<span class="text-xs font-medium text-emerald-600 flex items-center gap-1">
				<span class="w-1.5 h-1.5 rounded-full bg-emerald-500 inline-block"></span>
				In Stock
			</span>
		{:else}
			<span class="text-xs font-medium text-slate-400 flex items-center gap-1">
				<span class="w-1.5 h-1.5 rounded-full bg-slate-300 inline-block"></span>
				Out of Stock
			</span>
		{/if}
	</div>

	<div class="flex-1">
		<h3 class="font-semibold text-slate-800 text-base leading-snug">{medicine.name}</h3>
		<p class="text-slate-500 text-sm mt-1 line-clamp-2">{medicine.description}</p>
	</div>

	<div class="flex items-center justify-between mt-1">
		<span class="text-lg font-bold text-slate-800">${medicine.price.toFixed(2)}</span>
		<button
			onclick={handleAdd}
			disabled={!medicine.in_stock}
			class="px-4 py-1.5 rounded-lg text-sm font-semibold transition-all duration-200
				{medicine.in_stock
					? added
						? 'bg-emerald-500 text-white scale-95'
						: 'bg-teal-600 hover:bg-teal-700 text-white active:scale-95'
					: 'bg-slate-100 text-slate-400 cursor-not-allowed'}"
		>
			{added ? '✓ Added' : 'Add to Cart'}
		</button>
	</div>
</div>

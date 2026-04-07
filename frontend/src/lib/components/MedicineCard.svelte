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

<div class="glass-card glass-card-hover rounded-xl p-5 flex flex-col gap-3 transition-all duration-200">
	<div class="flex items-start justify-between gap-2">
		<span class="text-xs font-semibold uppercase tracking-wider px-2.5 py-0.5 rounded-full bg-cyan-glow/10 text-cyan-300 border border-cyan-glow/20">
			{medicine.category}
		</span>
		{#if medicine.in_stock}
			<span class="text-xs font-medium text-emerald-400 flex items-center gap-1.5">
				<span class="w-1.5 h-1.5 rounded-full bg-emerald-400 inline-block shadow-sm shadow-emerald-400/50"></span>
				In Stock
			</span>
		{:else}
			<span class="text-xs font-medium text-navy-500 flex items-center gap-1.5">
				<span class="w-1.5 h-1.5 rounded-full bg-navy-600 inline-block"></span>
				Out of Stock
			</span>
		{/if}
	</div>

	<div class="flex-1">
		<h3 class="font-semibold text-white text-base leading-snug">{medicine.name}</h3>
		<p class="text-navy-300 text-sm mt-1.5 line-clamp-2 leading-relaxed">{medicine.description}</p>
	</div>

	<div class="flex items-center justify-between mt-1 pt-3 border-t border-navy-700/50">
		<span class="text-lg font-bold text-white">${medicine.price.toFixed(2)}</span>
		<button
			onclick={handleAdd}
			disabled={!medicine.in_stock}
			class="px-4 py-1.5 rounded-lg text-sm font-semibold transition-all duration-200
				{medicine.in_stock
					? added
						? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40 scale-95'
						: 'bg-cyan-glow/15 hover:bg-cyan-glow/25 text-cyan-glow border border-cyan-glow/30 active:scale-95'
					: 'bg-navy-800 text-navy-500 border border-navy-700 cursor-not-allowed'}"
		>
			{added ? '✓ Added' : 'Add to Cart'}
		</button>
	</div>
</div>

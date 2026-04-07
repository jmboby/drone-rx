<script lang="ts">
	import { STATUS_LABELS, STATUS_ORDER } from '$lib/types';
	import type { OrderStatus } from '$lib/types';

	interface Props {
		status: OrderStatus;
	}

	let { status }: Props = $props();

	let currentIndex = $derived(STATUS_ORDER.indexOf(status));
</script>

<div class="w-full">
	<div class="flex items-center">
		{#each STATUS_ORDER as step, i}
			{@const stepIndex = i}
			{@const isDone = stepIndex < currentIndex}
			{@const isCurrent = stepIndex === currentIndex}
			{@const isPending = stepIndex > currentIndex}

			<!-- Step circle -->
			<div class="flex flex-col items-center relative">
				<div
					class="w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold border-2 transition-all duration-300
						{isDone
							? 'bg-teal-600 border-teal-600 text-white'
							: isCurrent
								? 'bg-white border-teal-600 text-teal-600 shadow-md shadow-teal-100'
								: 'bg-white border-slate-200 text-slate-400'}"
				>
					{#if isDone}
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
						</svg>
					{:else}
						{stepIndex + 1}
					{/if}
				</div>
				<span
					class="text-xs mt-2 font-medium text-center w-20 leading-tight
						{isDone || isCurrent ? 'text-teal-700' : 'text-slate-400'}"
				>
					{STATUS_LABELS[step]}
				</span>
			</div>

			<!-- Connector line (between steps) -->
			{#if i < STATUS_ORDER.length - 1}
				<div class="flex-1 h-0.5 mx-1 mb-6 rounded-full transition-all duration-500
					{stepIndex < currentIndex ? 'bg-teal-500' : 'bg-slate-200'}">
				</div>
			{/if}
		{/each}
	</div>
</div>

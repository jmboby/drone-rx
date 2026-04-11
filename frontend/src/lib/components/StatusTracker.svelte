<script lang="ts">
	import { STATUS_LABELS, STATUS_ORDER } from '$lib/types';
	import type { OrderStatus } from '$lib/types';
	import DroneIcon from './DroneIcon.svelte';

	interface Props {
		status: OrderStatus;
		premium?: boolean;
	}

	let { status, premium = false }: Props = $props();

	let currentIndex = $derived(STATUS_ORDER.indexOf(status));

	// Track the previous index to detect transitions for premium animation
	let prevIndex = $state(-1);
	let animatingStep = $state(-1);
	let animatingConnector = $state(-1);
	let sparkleStep = $state(-1);

	$effect(() => {
		const idx = currentIndex;
		if (premium && prevIndex >= 0 && idx > prevIndex) {
			// Animate the newly completed step and connector
			animatingStep = idx;
			animatingConnector = idx - 1;
			sparkleStep = prevIndex; // sparkle on the step that just completed

			setTimeout(() => {
				animatingStep = -1;
				animatingConnector = -1;
				sparkleStep = -1;
			}, 700);
		}
		prevIndex = idx;
	});

	const stepIcons = ['clipboard', 'flask', 'drone', 'check'] as const;

	// Sparkle particle offsets
	const sparkleOffsets = [
		{ sx: '-12px', sy: '-14px', color: '#00e5ff' },
		{ sx: '14px', sy: '-10px', color: '#ffab00' },
		{ sx: '-8px', sy: '12px', color: '#4dd0e1' },
		{ sx: '10px', sy: '14px', color: '#ffca28' },
		{ sx: '-14px', sy: '4px', color: '#00e5ff' },
		{ sx: '12px', sy: '-6px', color: '#ffb300' },
	];
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
				{#if premium && sparkleStep === i}
					{#each sparkleOffsets as sp}
						<span
							class="sparkle-particle"
							style="--sx: {sp.sx}; --sy: {sp.sy}; background: {sp.color}; top: 50%; left: 50%;"
						></span>
					{/each}
				{/if}
				<div
					class="w-11 h-11 rounded-full flex items-center justify-center text-sm font-bold border-2 transition-all duration-500
						{isDone
							? 'bg-cyan-glow/20 border-cyan-glow text-cyan-glow shadow-md shadow-cyan-glow/20'
							: isCurrent
								? 'bg-navy-800 border-amber-glow text-amber-glow shadow-lg shadow-amber-glow/20 animate-progress-glow'
								: 'bg-navy-800 border-navy-600 text-navy-500'}
						{premium && animatingStep === i ? ' animate-step-complete' : ''}"
					style={isCurrent ? '--color-cyan-glow: var(--color-amber-glow)' : ''}
				>
					{#if isDone}
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
						</svg>
					{:else if step === 'in-flight' && isCurrent}
						<DroneIcon size="w-5 h-5" animated />
					{:else if step === 'in-flight'}
						<DroneIcon size="w-5 h-5" />
					{:else}
						{stepIndex + 1}
					{/if}
				</div>
				<span
					class="text-xs mt-2.5 font-medium text-center w-20 leading-tight
						{isDone ? 'text-cyan-300' : isCurrent ? 'text-amber-300' : 'text-navy-500'}"
				>
					{STATUS_LABELS[step]}
				</span>
			</div>

			<!-- Connector line (between steps) -->
			{#if i < STATUS_ORDER.length - 1}
				<div class="flex-1 h-0.5 mx-1.5 mb-7 rounded-full overflow-hidden bg-navy-700 relative
					{premium && animatingConnector === i ? ' animate-connector-flash' : ''}">
					<div
						class="absolute inset-y-0 left-0 rounded-full transition-all duration-700 ease-out
							{stepIndex < currentIndex ? 'w-full bg-cyan-glow shadow-sm shadow-cyan-glow/50' : 'w-0'}"
					></div>
				</div>
			{/if}
		{/each}
	</div>
</div>

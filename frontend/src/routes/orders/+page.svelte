<script lang="ts">
	import type { Order, OrderStatus } from '$lib/types';
	import { STATUS_LABELS } from '$lib/types';
	import { listOrders } from '$lib/api';
	import DroneIcon from '$lib/components/DroneIcon.svelte';

	let searchName = $state('');
	let orders = $state<Order[]>([]);
	let loading = $state(false);
	let error = $state('');
	let searched = $state(false);

	async function handleSearch(e: SubmitEvent) {
		e.preventDefault();
		if (!searchName.trim()) { error = 'Please enter a patient name to search.'; return; }
		loading = true;
		error = '';
		searched = true;

		try {
			orders = await listOrders(searchName.trim());
		} catch (err) {
			error = err instanceof Error ? err.message : 'Search failed. Please try again.';
			orders = [];
		} finally {
			loading = false;
		}
	}

	type StatusStyle = { bg: string; text: string; border: string };

	const statusStyles: Record<OrderStatus, StatusStyle> = {
		placed: { bg: 'bg-blue-500/10', text: 'text-blue-300', border: 'border-blue-500/30' },
		preparing: { bg: 'bg-amber-glow/10', text: 'text-amber-300', border: 'border-amber-glow/30' },
		'in-flight': { bg: 'bg-purple-500/10', text: 'text-purple-300', border: 'border-purple-500/30' },
		delivered: { bg: 'bg-emerald-500/10', text: 'text-emerald-300', border: 'border-emerald-500/30' }
	};

	function statusClass(status: OrderStatus): string {
		const s = statusStyles[status] ?? { bg: 'bg-navy-700/50', text: 'text-navy-300', border: 'border-navy-600' };
		return `${s.bg} ${s.text} ${s.border}`;
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString(undefined, {
			month: 'short', day: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}
</script>

<svelte:head>
	<title>DroneRx — Order History</title>
</svelte:head>

<!-- Header -->
<header class="border-b border-navy-700/60 bg-navy-900/80 backdrop-blur-xl">
	<div class="max-w-3xl mx-auto px-4 sm:px-6 py-3.5 flex items-center gap-3">
		<a href="/" aria-label="Back to medicines" class="text-navy-400 hover:text-cyan-glow transition-colors">
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
		</a>
		<div class="flex items-center gap-2">
			<span class="text-cyan-glow"><DroneIcon size="w-6 h-6" /></span>
			<span class="text-xl font-bold text-white">DroneRx</span>
		</div>
		<span class="text-navy-600">/</span>
		<span class="text-navy-200 font-medium">Order History</span>
	</div>
</header>

<main class="max-w-3xl mx-auto px-4 sm:px-6 py-8">
	<h1 class="text-2xl font-bold text-white mb-6">Find Your Orders</h1>

	<!-- Search form -->
	<form onsubmit={handleSearch} class="flex gap-2 mb-8">
		<input
			type="text"
			bind:value={searchName}
			placeholder="Enter patient name..."
			class="flex-1 px-4 py-2.5 bg-navy-800 border border-navy-600 rounded-lg text-sm text-white placeholder-navy-500 focus:outline-none focus:ring-2 focus:ring-cyan-glow/50 focus:border-cyan-glow/50 transition"
		/>
		<button
			type="submit"
			disabled={loading}
			class="px-5 py-2.5 rounded-lg font-semibold text-sm transition-all
				{loading
					? 'bg-navy-800 text-navy-500 border border-navy-700 cursor-not-allowed'
					: 'bg-cyan-glow/15 hover:bg-cyan-glow/25 text-cyan-glow border border-cyan-glow/40 active:scale-95'}"
		>
			{loading ? 'Searching...' : 'Search'}
		</button>
	</form>

	{#if error}
		<div class="bg-rose-500/10 border border-rose-500/30 text-rose-300 text-sm px-4 py-3 rounded-lg mb-4">
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="text-center py-12 text-navy-400">
			<div class="inline-block animate-drone-fly">
				<span class="text-cyan-glow"><DroneIcon size="w-10 h-10" /></span>
			</div>
			<p class="text-sm mt-3 font-medium">Searching orders...</p>
		</div>
	{:else if searched && orders.length === 0}
		<div class="text-center py-12 text-navy-400">
			<svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
			</svg>
			<p class="font-medium">No orders found</p>
			<p class="text-sm mt-1 text-navy-500">Try searching with a different name</p>
		</div>
	{:else if orders.length > 0}
		<div class="space-y-3">
			<p class="text-sm text-navy-400 mb-2 font-medium">
				{orders.length} order{orders.length !== 1 ? 's' : ''} found for "{searchName}"
			</p>

			{#each orders as order, i (order.id)}
				<a
					href="/order/{order.id}"
					class="block glass-card glass-card-hover rounded-xl p-4 transition-all group animate-slide-up"
					style="animation-delay: {i * 60}ms"
				>
					<div class="flex items-start justify-between gap-4">
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 mb-1.5">
								<span
									class="inline-flex items-center px-2.5 py-0.5 rounded-full border text-xs font-semibold {statusClass(order.status)}"
								>
									{STATUS_LABELS[order.status]}
								</span>
								{#if order.status === 'delivered'}
									<svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
									</svg>
								{:else if order.status === 'in-flight'}
									<span class="text-purple-300"><DroneIcon size="w-4 h-4" animated /></span>
								{/if}
							</div>
							<p class="text-sm text-navy-300 truncate">{order.address}</p>
							{#if order.items && order.items.length > 0}
								<p class="text-xs text-navy-500 mt-1">
									{order.items.length} item{order.items.length !== 1 ? 's' : ''}
								</p>
							{/if}
						</div>

						<div class="text-right shrink-0">
							<p class="text-xs text-navy-500">{formatDate(order.created_at)}</p>
							<p class="text-xs font-mono text-navy-600 mt-1">{order.id.slice(0, 8)}...</p>
							<span class="mt-2 text-xs font-medium text-cyan-glow opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-0.5 justify-end">
								Track &rarr;
							</span>
						</div>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</main>

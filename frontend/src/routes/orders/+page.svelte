<script lang="ts">
	import type { Order, OrderStatus } from '$lib/types';
	import { STATUS_LABELS } from '$lib/types';
	import { listOrders } from '$lib/api';

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
		placed: { bg: 'bg-blue-50', text: 'text-blue-700', border: 'border-blue-200' },
		preparing: { bg: 'bg-amber-50', text: 'text-amber-700', border: 'border-amber-200' },
		'in-flight': { bg: 'bg-purple-50', text: 'text-purple-700', border: 'border-purple-200' },
		delivered: { bg: 'bg-emerald-50', text: 'text-emerald-700', border: 'border-emerald-200' }
	};

	function statusClass(status: OrderStatus): string {
		const s = statusStyles[status] ?? { bg: 'bg-slate-50', text: 'text-slate-700', border: 'border-slate-200' };
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

<div class="min-h-screen bg-slate-50">
	<!-- Header -->
	<header class="bg-white border-b border-slate-200 shadow-sm">
		<div class="max-w-3xl mx-auto px-4 sm:px-6 py-4 flex items-center gap-3">
			<a href="/" aria-label="Back to medicines" class="text-slate-400 hover:text-teal-600 transition-colors">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
			</a>
			<div class="flex items-center gap-2">
				<span class="text-xl">🚁</span>
				<span class="text-xl font-bold text-teal-700">DroneRx</span>
			</div>
			<span class="text-slate-400">/</span>
			<span class="text-slate-600 font-medium">Order History</span>
		</div>
	</header>

	<main class="max-w-3xl mx-auto px-4 sm:px-6 py-8">
		<h1 class="text-2xl font-bold text-slate-800 mb-6">Find Your Orders</h1>

		<!-- Search form -->
		<form onsubmit={handleSearch} class="flex gap-2 mb-8">
			<input
				type="text"
				bind:value={searchName}
				placeholder="Enter patient name..."
				class="flex-1 px-4 py-2.5 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teal-500 focus:border-transparent bg-white"
			/>
			<button
				type="submit"
				disabled={loading}
				class="px-5 py-2.5 rounded-lg font-semibold text-sm transition-all
					{loading
						? 'bg-slate-100 text-slate-400 cursor-not-allowed'
						: 'bg-teal-600 hover:bg-teal-700 text-white active:scale-95'}"
			>
				{loading ? 'Searching…' : 'Search'}
			</button>
		</form>

		{#if error}
			<div class="bg-rose-50 border border-rose-200 text-rose-700 text-sm px-4 py-3 rounded-lg mb-4">
				{error}
			</div>
		{/if}

		{#if loading}
			<div class="text-center py-12 text-slate-400">
				<div class="w-8 h-8 border-2 border-teal-200 border-t-teal-600 rounded-full animate-spin mx-auto mb-3"></div>
				<p class="text-sm">Searching orders…</p>
			</div>
		{:else if searched && orders.length === 0}
			<div class="text-center py-12 text-slate-400">
				<p class="text-4xl mb-3">📭</p>
				<p class="font-medium">No orders found</p>
				<p class="text-sm mt-1">Try searching with a different name</p>
			</div>
		{:else if orders.length > 0}
			<div class="space-y-3">
				<p class="text-sm text-slate-500 mb-2">
					{orders.length} order{orders.length !== 1 ? 's' : ''} found for "{searchName}"
				</p>

				{#each orders as order (order.id)}
					<a
						href="/order/{order.id}"
						class="block bg-white rounded-xl border border-slate-200 p-4 hover:border-teal-300 hover:shadow-sm transition-all group"
					>
						<div class="flex items-start justify-between gap-4">
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2 mb-1">
									<span
										class="inline-flex items-center px-2.5 py-0.5 rounded-full border text-xs font-semibold {statusClass(order.status)}"
									>
										{STATUS_LABELS[order.status]}
									</span>
									{#if order.status === 'delivered'}
										<span class="text-sm">✅</span>
									{:else if order.status === 'in-flight'}
										<span class="text-sm">🚁</span>
									{/if}
								</div>
								<p class="text-sm text-slate-500 truncate">{order.address}</p>
								{#if order.items && order.items.length > 0}
									<p class="text-xs text-slate-400 mt-1">
										{order.items.length} item{order.items.length !== 1 ? 's' : ''}
									</p>
								{/if}
							</div>

							<div class="text-right shrink-0">
								<p class="text-xs text-slate-400">{formatDate(order.created_at)}</p>
								<p class="text-xs font-mono text-slate-300 mt-1">{order.id.slice(0, 8)}…</p>
								<span class="mt-2 text-xs font-medium text-teal-600 opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-0.5 justify-end">
									Track →
								</span>
							</div>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</main>
</div>

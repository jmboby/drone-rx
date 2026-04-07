<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { PageData } from './$types';
	import type { Order, TrackingEvent } from '$lib/types';
	import { STATUS_LABELS } from '$lib/types';
	import { connectTracking, getOrder } from '$lib/api';
	import StatusTracker from '$lib/components/StatusTracker.svelte';

	let { data }: { data: PageData } = $props();

	let order = $state<Order>(data.order);
	let wsConnected = $state(false);
	let pollInterval: ReturnType<typeof setInterval> | null = null;
	let ws: WebSocket | null = null;

	function formatETA(seconds?: number): string {
		if (seconds == null || seconds <= 0) return 'Arriving soon';
		if (seconds < 60) return `${seconds}s`;
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return secs > 0 ? `${mins}m ${secs}s` : `${mins}m`;
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString(undefined, {
			month: 'short', day: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}

	function stopTracking() {
		if (ws) { ws.close(); ws = null; }
		if (pollInterval) { clearInterval(pollInterval); pollInterval = null; }
	}

	async function pollOrder() {
		try {
			const updated = await getOrder(order.id);
			order = updated;
			if (updated.status === 'delivered') { stopTracking(); }
		} catch {
			// silent — keep polling
		}
	}

	function startPolling() {
		pollInterval = setInterval(pollOrder, 5000);
	}

	function startWebSocket() {
		try {
			ws = connectTracking(order.id);

			ws.onopen = () => { wsConnected = true; };

			ws.onmessage = (event) => {
				try {
					const msg: TrackingEvent = JSON.parse(event.data);
					order = { ...order, status: msg.status, updated_at: msg.updated_at };
					if (msg.estimated_delivery) {
						order = { ...order, estimated_delivery: msg.estimated_delivery };
					}
					if (order.status === 'delivered') { stopTracking(); }
				} catch {
					// ignore parse errors
				}
			};

			ws.onerror = () => {
				wsConnected = false;
				ws = null;
				// fall back to polling
				if (!pollInterval) startPolling();
			};

			ws.onclose = () => {
				wsConnected = false;
				// fall back to polling if not delivered
				if (order.status !== 'delivered' && !pollInterval) {
					startPolling();
				}
			};
		} catch {
			startPolling();
		}
	}

	onMount(() => {
		if (order.status !== 'delivered') {
			startWebSocket();
		}
	});

	onDestroy(() => {
		stopTracking();
	});

	let statusBg: Record<string, string> = {
		placed: 'bg-blue-50 text-blue-700 border-blue-200',
		preparing: 'bg-amber-50 text-amber-700 border-amber-200',
		'in-flight': 'bg-purple-50 text-purple-700 border-purple-200',
		delivered: 'bg-emerald-50 text-emerald-700 border-emerald-200'
	};
</script>

<svelte:head>
	<title>DroneRx — Order #{order.id.slice(0, 8)}</title>
</svelte:head>

<div class="min-h-screen bg-slate-50">
	<!-- Header -->
	<header class="bg-white border-b border-slate-200 shadow-sm">
		<div class="max-w-3xl mx-auto px-4 sm:px-6 py-4 flex items-center gap-3">
			<a href="/" class="text-slate-400 hover:text-teal-600 transition-colors">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
			</a>
			<div class="flex items-center gap-2">
				<span class="text-xl">🚁</span>
				<span class="text-xl font-bold text-teal-700">DroneRx</span>
			</div>
			<span class="text-slate-400">/</span>
			<span class="text-slate-600 font-medium">Track Order</span>

			<!-- Live indicator -->
			{#if wsConnected}
				<span class="ml-auto flex items-center gap-1.5 text-xs font-medium text-emerald-600">
					<span class="relative flex h-2 w-2">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
						<span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
					</span>
					Live
				</span>
			{/if}
		</div>
	</header>

	<main class="max-w-3xl mx-auto px-4 sm:px-6 py-8 space-y-6">
		<!-- Status card -->
		<div class="bg-white rounded-xl border border-slate-200 p-6 shadow-sm">
			<div class="flex items-center justify-between mb-6">
				<div>
					<p class="text-xs text-slate-400 font-medium uppercase tracking-wide mb-1">Order Status</p>
					<span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full border text-sm font-semibold {statusBg[order.status] ?? 'bg-slate-50 text-slate-700 border-slate-200'}">
						{STATUS_LABELS[order.status]}
					</span>
				</div>
				{#if order.status !== 'delivered' && order.remaining_eta_seconds != null}
					<div class="text-right">
						<p class="text-xs text-slate-400 font-medium uppercase tracking-wide mb-1">ETA</p>
						<p class="text-2xl font-bold text-teal-700 tabular-nums">{formatETA(order.remaining_eta_seconds)}</p>
					</div>
				{:else if order.status === 'delivered'}
					<div class="text-right">
						<span class="text-3xl">✅</span>
						<p class="text-xs text-emerald-600 font-medium mt-1">Delivered!</p>
					</div>
				{/if}
			</div>

			<StatusTracker status={order.status} />
		</div>

		<!-- Order details card -->
		<div class="bg-white rounded-xl border border-slate-200 p-6 shadow-sm">
			<h2 class="font-semibold text-slate-800 mb-4">Order Details</h2>

			<dl class="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
				<div>
					<dt class="text-slate-400 font-medium text-xs uppercase tracking-wide mb-0.5">Order ID</dt>
					<dd class="text-slate-700 font-mono text-xs">{order.id}</dd>
				</div>
				<div>
					<dt class="text-slate-400 font-medium text-xs uppercase tracking-wide mb-0.5">Patient</dt>
					<dd class="text-slate-700">{order.patient_name}</dd>
				</div>
				<div class="col-span-2">
					<dt class="text-slate-400 font-medium text-xs uppercase tracking-wide mb-0.5">Delivery Address</dt>
					<dd class="text-slate-700">{order.address}</dd>
				</div>
				<div>
					<dt class="text-slate-400 font-medium text-xs uppercase tracking-wide mb-0.5">Placed At</dt>
					<dd class="text-slate-700">{formatDate(order.created_at)}</dd>
				</div>
				{#if order.estimated_delivery}
					<div>
						<dt class="text-slate-400 font-medium text-xs uppercase tracking-wide mb-0.5">Est. Delivery</dt>
						<dd class="text-slate-700">{formatDate(order.estimated_delivery)}</dd>
					</div>
				{/if}
			</dl>

			{#if order.items && order.items.length > 0}
				<div class="mt-5 pt-5 border-t border-slate-100">
					<h3 class="text-sm font-semibold text-slate-700 mb-3">Items</h3>
					<ul class="space-y-2">
						{#each order.items as item (item.id)}
							<li class="flex justify-between items-center text-sm">
								<span class="text-slate-700">
									{item.name ?? item.medicine_id}
									<span class="text-slate-400 ml-1">× {item.quantity}</span>
								</span>
								{#if item.price != null}
									<span class="text-slate-600 font-medium">${(item.price * item.quantity).toFixed(2)}</span>
								{/if}
							</li>
						{/each}
					</ul>
				</div>
			{/if}
		</div>

		<!-- Links -->
		<div class="flex gap-3">
			<a href="/" class="flex-1 text-center py-2.5 rounded-lg border border-slate-200 text-sm font-medium text-slate-600 hover:border-teal-400 hover:text-teal-600 transition-colors bg-white">
				Browse Medicines
			</a>
			<a href="/orders" class="flex-1 text-center py-2.5 rounded-lg border border-slate-200 text-sm font-medium text-slate-600 hover:border-teal-400 hover:text-teal-600 transition-colors bg-white">
				All My Orders
			</a>
		</div>
	</main>
</div>

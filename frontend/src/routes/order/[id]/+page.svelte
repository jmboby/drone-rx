<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { PageData } from './$types';
	import type { Order, TrackingEvent } from '$lib/types';
	import { STATUS_LABELS } from '$lib/types';
	import { connectTracking, getOrder, getLicenseStatus } from '$lib/api';
	import StatusTracker from '$lib/components/StatusTracker.svelte';
	import DroneIcon from '$lib/components/DroneIcon.svelte';

	let { data }: { data: PageData } = $props();

	let order = $state<Order>(data.order);
	let wsConnected = $state(false);
	let trackingEnabled = $state<boolean | null>(null);
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

	onMount(async () => {
		if (order.status !== 'delivered') {
			try {
				const license = await getLicenseStatus();
				trackingEnabled = license.live_tracking_enabled;
				if (license.live_tracking_enabled) {
					startWebSocket();
				} else {
					startPolling();
				}
			} catch {
				trackingEnabled = false;
				startPolling();
			}
		}
	});

	onDestroy(() => {
		stopTracking();
	});

	type StatusStyleEntry = { bg: string; text: string; border: string; glow: string };

	let statusStyles: Record<string, StatusStyleEntry> = {
		placed: { bg: 'bg-blue-500/10', text: 'text-blue-300', border: 'border-blue-500/30', glow: '' },
		preparing: { bg: 'bg-amber-glow/10', text: 'text-amber-300', border: 'border-amber-glow/30', glow: '' },
		'in-flight': { bg: 'bg-purple-500/10', text: 'text-purple-300', border: 'border-purple-500/30', glow: 'shadow-sm shadow-purple-500/20' },
		delivered: { bg: 'bg-emerald-500/10', text: 'text-emerald-300', border: 'border-emerald-500/30', glow: '' }
	};

	let currentStyle = $derived(statusStyles[order.status] ?? { bg: 'bg-navy-700/50', text: 'text-navy-300', border: 'border-navy-600', glow: '' });
</script>

<svelte:head>
	<title>DroneRx — Order #{order.id.slice(0, 8)}</title>
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
		<span class="text-navy-200 font-medium">Track Order</span>

		<!-- Tracking indicator -->
		{#if wsConnected}
			<span class="ml-auto flex items-center gap-1.5 text-xs font-semibold text-cyan-glow">
				<span class="relative flex h-2 w-2">
					<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-glow opacity-75"></span>
					<span class="relative inline-flex rounded-full h-2 w-2 bg-cyan-glow"></span>
				</span>
				LIVE
			</span>
		{:else if trackingEnabled === false && order.status !== 'delivered'}
			<span class="ml-auto flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-amber-glow/10 border border-amber-glow/30 text-xs font-semibold text-amber-300">
				<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
				</svg>
				Premium
			</span>
		{/if}
	</div>
</header>

<main class="max-w-3xl mx-auto px-4 sm:px-6 py-8 space-y-6">
	<!-- Status card -->
	<div class="glass-card rounded-xl p-6">
		<div class="flex items-center justify-between mb-6">
			<div>
				<p class="text-xs text-navy-400 font-semibold uppercase tracking-widest mb-2">Order Status</p>
				<span class="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full border text-sm font-semibold {currentStyle.bg} {currentStyle.text} {currentStyle.border} {currentStyle.glow}">
					{#if order.status === 'in-flight'}
						<DroneIcon size="w-4 h-4" animated />
					{/if}
					{STATUS_LABELS[order.status]}
				</span>
			</div>
			{#if order.status !== 'delivered' && order.remaining_eta_seconds != null}
				<div class="text-right">
					<p class="text-xs text-navy-400 font-semibold uppercase tracking-widest mb-2">ETA</p>
					<p class="text-2xl font-bold text-cyan-glow tabular-nums font-mono">{formatETA(order.remaining_eta_seconds)}</p>
				</div>
			{:else if order.status === 'delivered'}
				<div class="text-right">
					<div class="w-10 h-10 rounded-full bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center mb-1 ml-auto">
						<svg class="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
						</svg>
					</div>
					<p class="text-xs text-emerald-400 font-semibold">Delivered</p>
				</div>
			{/if}
		</div>

		<StatusTracker status={order.status} />

		{#if trackingEnabled === false && order.status !== 'delivered'}
			<div class="mt-4 p-3 rounded-lg bg-amber-glow/5 border border-amber-glow/20 flex items-start gap-3">
				<svg class="w-4 h-4 text-amber-400 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
				<div>
					<p class="text-sm font-medium text-amber-300">Real-time tracking is a premium feature</p>
					<p class="text-xs text-amber-300/60 mt-0.5">Your order status refreshes automatically every 5 seconds. Upgrade your license to enable live drone tracking.</p>
				</div>
			</div>
		{/if}
	</div>

	<!-- Flight visualization for in-flight -->
	{#if order.status === 'in-flight'}
		<div class="glass-card rounded-xl p-5 flex items-center justify-center gap-6 overflow-hidden relative">
			<div class="absolute inset-0 opacity-5" style="background: radial-gradient(circle at 50% 50%, var(--color-cyan-glow), transparent 70%);"></div>
			<div class="flex items-center gap-3 relative">
				<svg class="w-5 h-5 text-navy-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
				</svg>
				<span class="text-xs font-medium text-navy-400 uppercase tracking-wide">Pharmacy</span>
			</div>
			<div class="flex-1 h-0.5 bg-navy-700 rounded-full relative overflow-hidden">
				<div class="absolute inset-y-0 left-0 w-2/3 bg-gradient-to-r from-cyan-glow/80 to-cyan-glow rounded-full animate-pulse"></div>
				<div class="absolute top-1/2 left-[60%] -translate-y-1/2">
					<span class="text-cyan-glow"><DroneIcon size="w-6 h-6" animated /></span>
				</div>
			</div>
			<div class="flex items-center gap-3">
				<span class="text-xs font-medium text-navy-400 uppercase tracking-wide">You</span>
				<svg class="w-5 h-5 text-navy-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
				</svg>
			</div>
		</div>
	{/if}

	<!-- Order details card -->
	<div class="glass-card rounded-xl p-6">
		<h2 class="font-semibold text-white mb-4">Order Details</h2>

		<dl class="grid grid-cols-2 gap-x-6 gap-y-4 text-sm">
			<div>
				<dt class="text-navy-500 font-semibold text-xs uppercase tracking-widest mb-1">Order ID</dt>
				<dd class="text-navy-200 font-mono text-xs">{order.id}</dd>
			</div>
			<div>
				<dt class="text-navy-500 font-semibold text-xs uppercase tracking-widest mb-1">Patient</dt>
				<dd class="text-navy-200">{order.patient_name}</dd>
			</div>
			<div class="col-span-2">
				<dt class="text-navy-500 font-semibold text-xs uppercase tracking-widest mb-1">Delivery Address</dt>
				<dd class="text-navy-200">{order.address}</dd>
			</div>
			<div>
				<dt class="text-navy-500 font-semibold text-xs uppercase tracking-widest mb-1">Placed At</dt>
				<dd class="text-navy-200">{formatDate(order.created_at)}</dd>
			</div>
			{#if order.estimated_delivery}
				<div>
					<dt class="text-navy-500 font-semibold text-xs uppercase tracking-widest mb-1">Est. Delivery</dt>
					<dd class="text-navy-200">{formatDate(order.estimated_delivery)}</dd>
				</div>
			{/if}
		</dl>

		{#if order.items && order.items.length > 0}
			<div class="mt-5 pt-5 border-t border-navy-700/50">
				<h3 class="text-sm font-semibold text-navy-300 mb-3">Items</h3>
				<ul class="space-y-2">
					{#each order.items as item (item.id)}
						<li class="flex justify-between items-center text-sm">
							<span class="text-navy-200">
								{item.name ?? item.medicine_id}
								<span class="text-navy-500 ml-1">&times; {item.quantity}</span>
							</span>
							{#if item.price != null}
								<span class="text-navy-300 font-medium tabular-nums">${(item.price * item.quantity).toFixed(2)}</span>
							{/if}
						</li>
					{/each}
				</ul>
			</div>
		{/if}
	</div>

	<!-- Links -->
	<div class="flex gap-3">
		<a href="/" class="flex-1 text-center py-2.5 rounded-lg border border-navy-600 text-sm font-medium text-navy-300 hover:border-cyan-glow/30 hover:text-cyan-glow transition-all glass-card">
			Browse Medicines
		</a>
		<a href="/orders" class="flex-1 text-center py-2.5 rounded-lg border border-navy-600 text-sm font-medium text-navy-300 hover:border-cyan-glow/30 hover:text-cyan-glow transition-all glass-card">
			All My Orders
		</a>
	</div>
</main>

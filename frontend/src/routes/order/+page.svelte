<script lang="ts">
	import { goto } from '$app/navigation';
	import { cart, cartTotal } from '$lib/stores/cart';
	import { createOrder } from '$lib/api';
	import DroneIcon from '$lib/components/DroneIcon.svelte';

	let patientName = $state('');
	let address = $state('');
	let submitting = $state(false);
	let error = $state('');

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if ($cart.length === 0) { error = 'Your cart is empty.'; return; }
		if (!patientName.trim()) { error = 'Please enter patient name.'; return; }
		if (!address.trim()) { error = 'Please enter delivery address.'; return; }

		submitting = true;
		error = '';

		try {
			const order = await createOrder({
				patient_name: patientName.trim(),
				address: address.trim(),
				items: $cart.map((item) => ({
					medicine_id: item.medicine.id,
					quantity: item.quantity
				}))
			});
			cart.clear();
			goto(`/order/${order.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to place order. Please try again.';
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>DroneRx — Place Order</title>
</svelte:head>

<!-- Header -->
<header class="border-b border-navy-700/60 bg-navy-900/80 backdrop-blur-xl">
	<div class="max-w-4xl mx-auto px-4 sm:px-6 py-3.5 flex items-center gap-3">
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
		<span class="text-navy-200 font-medium">Place Order</span>
	</div>
</header>

<main class="max-w-4xl mx-auto px-4 sm:px-6 py-8">
	<div class="grid grid-cols-1 lg:grid-cols-5 gap-8">
		<!-- Cart summary -->
		<div class="lg:col-span-3">
			<h2 class="text-lg font-semibold text-white mb-4">Your Cart</h2>

			{#if $cart.length === 0}
				<div class="glass-card rounded-xl p-8 text-center text-navy-400">
					<svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 100 4 2 2 0 000-4z" />
					</svg>
					<p class="font-medium">Your cart is empty</p>
					<a href="/" class="mt-3 inline-block text-cyan-glow hover:text-cyan-300 text-sm font-medium transition-colors">
						Browse medicines &rarr;
					</a>
				</div>
			{:else}
				<div class="glass-card rounded-xl divide-y divide-navy-700/50">
					{#each $cart as item (item.medicine.id)}
						<div class="flex items-center gap-4 p-4">
							<div class="flex-1 min-w-0">
								<p class="font-medium text-white truncate">{item.medicine.name}</p>
								<p class="text-xs text-navy-400 mt-0.5">{item.medicine.category}</p>
							</div>

							<div class="flex items-center gap-2">
								<button
									onclick={() => cart.updateQuantity(item.medicine.id, item.quantity - 1)}
									class="w-7 h-7 rounded-full border border-navy-600 text-navy-300 hover:border-cyan-glow/40 hover:text-cyan-glow flex items-center justify-center text-sm font-bold transition-colors"
								>&minus;</button>
								<span class="w-6 text-center text-sm font-medium text-white tabular-nums">{item.quantity}</span>
								<button
									onclick={() => cart.updateQuantity(item.medicine.id, item.quantity + 1)}
									class="w-7 h-7 rounded-full border border-navy-600 text-navy-300 hover:border-cyan-glow/40 hover:text-cyan-glow flex items-center justify-center text-sm font-bold transition-colors"
								>+</button>
							</div>

							<span class="text-sm font-semibold text-white w-16 text-right tabular-nums">
								${(item.medicine.price * item.quantity).toFixed(2)}
							</span>

							<button
								onclick={() => cart.remove(item.medicine.id)}
								class="text-navy-500 hover:text-rose-400 transition-colors ml-1"
								aria-label="Remove"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
							</button>
						</div>
					{/each}

					<div class="flex items-center justify-between p-4 bg-navy-800/50 rounded-b-xl">
						<span class="font-semibold text-navy-200">Total</span>
						<span class="text-xl font-bold text-cyan-glow tabular-nums">${$cartTotal.toFixed(2)}</span>
					</div>
				</div>
			{/if}
		</div>

		<!-- Delivery details form -->
		<div class="lg:col-span-2">
			<h2 class="text-lg font-semibold text-white mb-4">Delivery Details</h2>

			<form onsubmit={handleSubmit} class="glass-card rounded-xl p-5 flex flex-col gap-4">
				<div>
					<label for="patient-name" class="block text-sm font-medium text-navy-200 mb-1.5">
						Patient Name
					</label>
					<input
						id="patient-name"
						type="text"
						bind:value={patientName}
						placeholder="Full name"
						class="w-full px-3 py-2.5 bg-navy-800 border border-navy-600 rounded-lg text-sm text-white placeholder-navy-500 focus:outline-none focus:ring-2 focus:ring-cyan-glow/50 focus:border-cyan-glow/50 transition"
						required
					/>
				</div>

				<div>
					<label for="address" class="block text-sm font-medium text-navy-200 mb-1.5">
						Delivery Address
					</label>
					<textarea
						id="address"
						bind:value={address}
						placeholder="Street address, city, state, zip"
						rows="3"
						class="w-full px-3 py-2.5 bg-navy-800 border border-navy-600 rounded-lg text-sm text-white placeholder-navy-500 focus:outline-none focus:ring-2 focus:ring-cyan-glow/50 focus:border-cyan-glow/50 transition resize-none"
						required
					></textarea>
				</div>

				{#if error}
					<div class="bg-rose-500/10 border border-rose-500/30 text-rose-300 text-sm px-3 py-2 rounded-lg">
						{error}
					</div>
				{/if}

				<button
					type="submit"
					disabled={submitting || $cart.length === 0}
					class="w-full py-2.5 rounded-lg font-semibold text-sm transition-all
						{submitting || $cart.length === 0
							? 'bg-navy-800 text-navy-500 border border-navy-700 cursor-not-allowed'
							: 'bg-cyan-glow/15 hover:bg-cyan-glow/25 text-cyan-glow border border-cyan-glow/40 active:scale-[0.98]'}"
				>
					{#if submitting}
						<span class="flex items-center justify-center gap-2">
							<DroneIcon size="w-4 h-4" animated />
							Placing Order...
						</span>
					{:else}
						Place Order
					{/if}
				</button>

				<p class="text-xs text-navy-500 text-center flex items-center justify-center gap-1.5">
					<DroneIcon size="w-3.5 h-3.5" />
					Estimated delivery within 30 minutes by drone
				</p>
			</form>
		</div>
	</div>
</main>

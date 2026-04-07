<script lang="ts">
	import { goto } from '$app/navigation';
	import { cart, cartTotal } from '$lib/stores/cart';
	import { createOrder } from '$lib/api';

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

<div class="min-h-screen bg-slate-50">
	<!-- Header -->
	<header class="bg-white border-b border-slate-200 shadow-sm">
		<div class="max-w-4xl mx-auto px-4 sm:px-6 py-4 flex items-center gap-3">
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
			<span class="text-slate-600 font-medium">Place Order</span>
		</div>
	</header>

	<main class="max-w-4xl mx-auto px-4 sm:px-6 py-8">
		<div class="grid grid-cols-1 lg:grid-cols-5 gap-8">
			<!-- Cart summary -->
			<div class="lg:col-span-3">
				<h2 class="text-lg font-semibold text-slate-800 mb-4">Your Cart</h2>

				{#if $cart.length === 0}
					<div class="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-400">
						<p class="text-4xl mb-3">🛒</p>
						<p class="font-medium">Your cart is empty</p>
						<a href="/" class="mt-3 inline-block text-teal-600 hover:underline text-sm font-medium">
							Browse medicines →
						</a>
					</div>
				{:else}
					<div class="bg-white rounded-xl border border-slate-200 divide-y divide-slate-100">
						{#each $cart as item (item.medicine.id)}
							<div class="flex items-center gap-4 p-4">
								<div class="flex-1 min-w-0">
									<p class="font-medium text-slate-800 truncate">{item.medicine.name}</p>
									<p class="text-xs text-slate-400 mt-0.5">{item.medicine.category}</p>
								</div>

								<div class="flex items-center gap-2">
									<button
										onclick={() => cart.updateQuantity(item.medicine.id, item.quantity - 1)}
										class="w-7 h-7 rounded-full border border-slate-200 text-slate-600 hover:border-teal-400 hover:text-teal-600 flex items-center justify-center text-sm font-bold transition-colors"
									>−</button>
									<span class="w-6 text-center text-sm font-medium text-slate-800">{item.quantity}</span>
									<button
										onclick={() => cart.updateQuantity(item.medicine.id, item.quantity + 1)}
										class="w-7 h-7 rounded-full border border-slate-200 text-slate-600 hover:border-teal-400 hover:text-teal-600 flex items-center justify-center text-sm font-bold transition-colors"
									>+</button>
								</div>

								<span class="text-sm font-semibold text-slate-800 w-16 text-right">
									${(item.medicine.price * item.quantity).toFixed(2)}
								</span>

								<button
									onclick={() => cart.remove(item.medicine.id)}
									class="text-slate-300 hover:text-rose-400 transition-colors ml-1"
									aria-label="Remove"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
								</button>
							</div>
						{/each}

						<div class="flex items-center justify-between p-4 bg-slate-50 rounded-b-xl">
							<span class="font-semibold text-slate-700">Total</span>
							<span class="text-xl font-bold text-teal-700">${$cartTotal.toFixed(2)}</span>
						</div>
					</div>
				{/if}
			</div>

			<!-- Delivery details form -->
			<div class="lg:col-span-2">
				<h2 class="text-lg font-semibold text-slate-800 mb-4">Delivery Details</h2>

				<form onsubmit={handleSubmit} class="bg-white rounded-xl border border-slate-200 p-5 flex flex-col gap-4">
					<div>
						<label for="patient-name" class="block text-sm font-medium text-slate-700 mb-1">
							Patient Name
						</label>
						<input
							id="patient-name"
							type="text"
							bind:value={patientName}
							placeholder="Full name"
							class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teal-500 focus:border-transparent transition"
							required
						/>
					</div>

					<div>
						<label for="address" class="block text-sm font-medium text-slate-700 mb-1">
							Delivery Address
						</label>
						<textarea
							id="address"
							bind:value={address}
							placeholder="Street address, city, state, zip"
							rows="3"
							class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teal-500 focus:border-transparent transition resize-none"
							required
						></textarea>
					</div>

					{#if error}
						<div class="bg-rose-50 border border-rose-200 text-rose-700 text-sm px-3 py-2 rounded-lg">
							{error}
						</div>
					{/if}

					<button
						type="submit"
						disabled={submitting || $cart.length === 0}
						class="w-full py-2.5 rounded-lg font-semibold text-sm transition-all
							{submitting || $cart.length === 0
								? 'bg-slate-100 text-slate-400 cursor-not-allowed'
								: 'bg-teal-600 hover:bg-teal-700 text-white active:scale-95'}"
					>
						{submitting ? 'Placing Order…' : 'Place Order'}
					</button>

					<p class="text-xs text-slate-400 text-center">
						Estimated delivery within 30 minutes by drone.
					</p>
				</form>
			</div>
		</div>
	</main>
</div>

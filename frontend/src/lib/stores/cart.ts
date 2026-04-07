import { writable, derived } from 'svelte/store';
import type { Medicine } from '$lib/types';

interface CartItem { medicine: Medicine; quantity: number; }

function createCart() {
	const { subscribe, set, update } = writable<CartItem[]>([]);
	return {
		subscribe,
		add(medicine: Medicine) {
			update((items) => {
				const existing = items.find((i) => i.medicine.id === medicine.id);
				if (existing) { existing.quantity += 1; return [...items]; }
				return [...items, { medicine, quantity: 1 }];
			});
		},
		remove(medicineId: string) {
			update((items) => items.filter((i) => i.medicine.id !== medicineId));
		},
		updateQuantity(medicineId: string, quantity: number) {
			update((items) => {
				if (quantity <= 0) return items.filter((i) => i.medicine.id !== medicineId);
				const item = items.find((i) => i.medicine.id === medicineId);
				if (item) item.quantity = quantity;
				return [...items];
			});
		},
		clear() { set([]); }
	};
}

export const cart = createCart();
export const cartTotal = derived(cart, ($cart) => $cart.reduce((sum, item) => sum + item.medicine.price * item.quantity, 0));
export const cartCount = derived(cart, ($cart) => $cart.reduce((sum, item) => sum + item.quantity, 0));

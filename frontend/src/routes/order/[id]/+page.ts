import type { PageLoad } from './$types';
import type { Order } from '$lib/types';

export const load: PageLoad = async ({ fetch, params }) => {
	const order: Order = await fetch(`/api/orders/${params.id}`).then((r) => {
		if (!r.ok) throw new Error('Order not found');
		return r.json();
	});
	return { order };
};

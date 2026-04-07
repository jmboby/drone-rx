import type { PageLoad } from './$types';
import type { Medicine } from '$lib/types';

export const load: PageLoad = async ({ fetch }) => {
	const medicines: Medicine[] = await fetch('/api/medicines').then((r) => {
		if (!r.ok) throw new Error('Failed to load medicines');
		return r.json();
	});

	const categories = [...new Set(medicines.map((m) => m.category))].sort();

	return { medicines, categories };
};

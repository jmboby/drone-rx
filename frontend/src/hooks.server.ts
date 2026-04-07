import type { Handle } from '@sveltejs/kit';

const API_URL = process.env.API_URL || 'http://localhost:8080';

export const handle: Handle = async ({ event, resolve }) => {
	const { pathname } = event.url;

	if (pathname.startsWith('/api') || pathname === '/healthz') {
		const target = `${API_URL}${pathname}${event.url.search}`;

		// Pass through WebSocket upgrade requests
		const headers = new Headers(event.request.headers);
		headers.set('host', new URL(API_URL).host);

		const response = await fetch(target, {
			method: event.request.method,
			headers,
			body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
				? await event.request.text()
				: undefined,
		});

		return new Response(response.body, {
			status: response.status,
			statusText: response.statusText,
			headers: response.headers,
		});
	}

	return resolve(event);
};

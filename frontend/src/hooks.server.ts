import type { Handle } from '@sveltejs/kit';

const API_URL = process.env.API_URL || 'http://localhost:8080';

function log(level: string, msg: string, data: Record<string, unknown> = {}) {
	const entry = { time: new Date().toISOString(), level, msg, ...data };
	console.log(JSON.stringify(entry));
}

export const handle: Handle = async ({ event, resolve }) => {
	const start = Date.now();
	const { pathname } = event.url;

	if (pathname.startsWith('/api') || pathname === '/healthz') {
		const target = `${API_URL}${pathname}${event.url.search}`;

		const headers = new Headers(event.request.headers);
		headers.set('host', new URL(API_URL).host);

		try {
			const response = await fetch(target, {
				method: event.request.method,
				headers,
				body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
					? await event.request.text()
					: undefined,
			});

			const duration = Date.now() - start;
			if (pathname !== '/healthz' && pathname !== '/api/updates' && pathname !== '/api/license/status') {
				log('INFO', 'api proxy', {
					method: event.request.method,
					path: pathname,
					status: response.status,
					duration_ms: duration,
				});
			}

			return new Response(response.body, {
				status: response.status,
				statusText: response.statusText,
				headers: response.headers,
			});
		} catch (err) {
			log('ERROR', 'api proxy failed', {
				method: event.request.method,
				path: pathname,
				error: err instanceof Error ? err.message : String(err),
			});
			return new Response('Backend unavailable', { status: 502 });
		}
	}

	const response = await resolve(event);
	const duration = Date.now() - start;

	// Log page requests (skip static assets)
	if (!pathname.startsWith('/_app/') && !pathname.startsWith('/favicon') && pathname !== '/robots.txt') {
		log('INFO', 'page request', {
			method: event.request.method,
			path: pathname,
			status: response.status,
			duration_ms: duration,
		});
	}

	return response;
};

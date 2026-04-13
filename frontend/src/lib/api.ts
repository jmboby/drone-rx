import type { Medicine, Order, CreateOrderRequest, LicenseStatus, UpdateInfo, SupportBundleResponse } from './types';

const BASE_URL = '/api';

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
	const response = await fetch(url, init);
	if (!response.ok) {
		const text = await response.text();
		throw new Error(`${response.status}: ${text}`);
	}
	return response.json();
}

export async function listMedicines(): Promise<Medicine[]> {
	return fetchJSON<Medicine[]>(`${BASE_URL}/medicines`);
}

export async function getMedicine(id: string): Promise<Medicine> {
	return fetchJSON<Medicine>(`${BASE_URL}/medicines/${id}`);
}

export async function createOrder(req: CreateOrderRequest): Promise<Order> {
	return fetchJSON<Order>(`${BASE_URL}/orders`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(req)
	});
}

export async function getOrder(id: string): Promise<Order> {
	return fetchJSON<Order>(`${BASE_URL}/orders/${id}`);
}

export async function listOrders(patientName: string): Promise<Order[]> {
	return fetchJSON<Order[]>(`${BASE_URL}/orders?patient_name=${encodeURIComponent(patientName)}`);
}

export function connectTracking(orderID: string): WebSocket {
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	return new WebSocket(`${protocol}//${window.location.host}/api/orders/${orderID}/track`);
}

export async function getLicenseStatus(): Promise<LicenseStatus> {
	return fetchJSON<LicenseStatus>(`${BASE_URL}/license/status`);
}

export async function getUpdates(): Promise<UpdateInfo[]> {
	return fetchJSON<UpdateInfo[]>(`${BASE_URL}/updates`);
}

export async function generateSupportBundle(): Promise<SupportBundleResponse> {
	return fetchJSON<SupportBundleResponse>(`${BASE_URL}/admin/support-bundle`, {
		method: 'POST',
	});
}

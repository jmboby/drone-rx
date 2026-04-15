export interface Medicine {
	id: string; name: string; description: string;
	price: number; in_stock: boolean; category: string;
}

export interface Order {
	id: string; patient_name: string; address: string;
	status: OrderStatus; estimated_delivery: string | null;
	remaining_eta_seconds?: number;
	created_at: string; updated_at: string; items?: OrderItem[];
}

export interface OrderItem {
	id: string; order_id: string; medicine_id: string;
	quantity: number; name?: string; price?: number;
}

export type OrderStatus = 'placed' | 'preparing' | 'in-flight' | 'delivered';

export interface CreateOrderRequest {
	patient_name: string; address: string;
	items: { medicine_id: string; quantity: number }[];
}

export interface TrackingEvent {
	order_id: string; status: OrderStatus;
	estimated_delivery?: string; updated_at: string;
}

export const STATUS_LABELS: Record<OrderStatus, string> = {
	placed: 'Order Placed', preparing: 'Preparing',
	'in-flight': 'Drone In Flight', delivered: 'Delivered'
};

export const STATUS_ORDER: OrderStatus[] = ['placed', 'preparing', 'in-flight', 'delivered'];

export interface LicenseStatus {
	valid: boolean;
	expired: boolean;
	license_type?: string;
	expiration_date?: string;
	live_tracking_enabled: boolean;
	light_mode_enabled: boolean;
}

export interface UpdateInfo {
	versionLabel: string;
	createdAt: string;
	releaseNotes: string;
}

export interface SupportBundleResponse {
	status: string;
	message: string;
}

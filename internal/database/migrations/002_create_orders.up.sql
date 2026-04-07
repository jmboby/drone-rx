CREATE TYPE order_status AS ENUM ('placed', 'preparing', 'in-flight', 'delivered');

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_name TEXT NOT NULL,
    address TEXT NOT NULL,
    status order_status NOT NULL DEFAULT 'placed',
    estimated_delivery TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    medicine_id UUID NOT NULL REFERENCES medicines(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0)
);

CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_patient_name ON orders(patient_name);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

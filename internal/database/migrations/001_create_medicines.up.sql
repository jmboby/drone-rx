CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE medicines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    in_stock BOOLEAN NOT NULL DEFAULT true,
    category TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO medicines (name, description, price, in_stock, category) VALUES
    ('Paracetamol 500mg', 'Pain relief and fever reduction tablets', 4.99, true, 'Pain Relief'),
    ('Ibuprofen 200mg', 'Anti-inflammatory pain relief', 5.49, true, 'Pain Relief'),
    ('Amoxicillin 250mg', 'Broad-spectrum antibiotic capsules', 8.99, true, 'Antibiotics'),
    ('Cetirizine 10mg', 'Non-drowsy antihistamine for allergies', 6.29, true, 'Allergy'),
    ('Omeprazole 20mg', 'Acid reflux and heartburn relief', 7.49, true, 'Digestive'),
    ('Loratadine 10mg', 'Allergy relief tablets', 5.99, true, 'Allergy'),
    ('Vitamin D3 1000IU', 'Daily vitamin D supplement', 3.99, true, 'Supplements'),
    ('Zinc 25mg', 'Immune support supplement', 4.49, true, 'Supplements'),
    ('Salbutamol Inhaler', 'Reliever inhaler for asthma', 12.99, true, 'Respiratory'),
    ('Throat Lozenges', 'Soothing lozenges for sore throat', 3.49, true, 'Respiratory');

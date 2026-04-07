package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStatus string

const (
	StatusPlaced    OrderStatus = "placed"
	StatusPreparing OrderStatus = "preparing"
	StatusInFlight  OrderStatus = "in_flight"
	StatusDelivered OrderStatus = "delivered"
)

// Next returns the next status in the progression and whether the current status is terminal.
func (s OrderStatus) Next() (OrderStatus, bool) {
	switch s {
	case StatusPlaced:
		return StatusPreparing, false
	case StatusPreparing:
		return StatusInFlight, false
	case StatusInFlight:
		return StatusDelivered, false
	case StatusDelivered:
		return "", true
	default:
		return "", true
	}
}

// IsValid reports whether the status is one of the known values.
func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusPlaced, StatusPreparing, StatusInFlight, StatusDelivered:
		return true
	}
	return false
}

// Scan implements the sql.Scanner interface for pgx.
func (s *OrderStatus) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		*s = OrderStatus(v)
		return nil
	case []byte:
		*s = OrderStatus(string(v))
		return nil
	}
	return fmt.Errorf("cannot scan %T into OrderStatus", src)
}

// Order represents a drone delivery order.
type Order struct {
	ID                string      `json:"id"`
	PatientName       string      `json:"patient_name"`
	Address           string      `json:"address"`
	Status            OrderStatus `json:"status"`
	EstimatedDelivery *time.Time  `json:"estimated_delivery,omitempty"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	Items             []OrderItem `json:"items,omitempty"`
}

// OrderItem is one line in an order.
type OrderItem struct {
	ID         string   `json:"id"`
	OrderID    string   `json:"order_id"`
	MedicineID string   `json:"medicine_id"`
	Quantity   int      `json:"quantity"`
	Price      float64  `json:"price"`
	Medicine   *Medicine `json:"medicine,omitempty"`
}

// CreateOrderRequest is the payload for creating a new order.
type CreateOrderRequest struct {
	PatientName string             `json:"patient_name"`
	Address     string             `json:"address"`
	Items       []OrderItemRequest `json:"items"`
}

// OrderItemRequest is one line in a CreateOrderRequest.
type OrderItemRequest struct {
	MedicineID string `json:"medicine_id"`
	Quantity   int    `json:"quantity"`
}

// Validate checks that the request is well-formed.
func (r *CreateOrderRequest) Validate() error {
	if r.PatientName == "" {
		return fmt.Errorf("patient_name is required")
	}
	if r.Address == "" {
		return fmt.Errorf("address is required")
	}
	if len(r.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}
	for i, item := range r.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("item[%d]: quantity must be greater than zero", i)
		}
	}
	return nil
}

// OrderStore provides database operations for orders.
type OrderStore struct{ db *pgxpool.Pool }

// NewOrderStore creates an OrderStore backed by the given pool.
func NewOrderStore(db *pgxpool.Pool) *OrderStore { return &OrderStore{db: db} }

// Create inserts a new order and its items in a single transaction.
func (s *OrderStore) Create(ctx context.Context, req CreateOrderRequest, eta time.Time) (*Order, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var order Order
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (patient_name, address, status, estimated_delivery)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, patient_name, address, status, estimated_delivery, created_at, updated_at`,
		req.PatientName, req.Address, StatusPlaced, eta,
	).Scan(&order.ID, &order.PatientName, &order.Address, &order.Status, &order.EstimatedDelivery, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	for _, item := range req.Items {
		var oi OrderItem
		err = tx.QueryRow(ctx,
			`INSERT INTO order_items (order_id, medicine_id, quantity)
			 VALUES ($1, $2, $3)
			 RETURNING id, order_id, medicine_id, quantity`,
			order.ID, item.MedicineID, item.Quantity,
		).Scan(&oi.ID, &oi.OrderID, &oi.MedicineID, &oi.Quantity)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, oi)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByID fetches an order with its items by ID.
func (s *OrderStore) GetByID(ctx context.Context, id string) (*Order, error) {
	var order Order
	err := s.db.QueryRow(ctx,
		`SELECT id, patient_name, address, status, estimated_delivery, created_at, updated_at FROM orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.PatientName, &order.Address, &order.Status, &order.EstimatedDelivery, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx,
		`SELECT oi.id, oi.order_id, oi.medicine_id, oi.quantity,
		        m.id, m.name, m.description, m.price, m.in_stock, m.category
		 FROM order_items oi
		 JOIN medicines m ON m.id = oi.medicine_id
		 WHERE oi.order_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var oi OrderItem
		var med Medicine
		if err := rows.Scan(
			&oi.ID, &oi.OrderID, &oi.MedicineID, &oi.Quantity,
			&med.ID, &med.Name, &med.Description, &med.Price, &med.InStock, &med.Category,
		); err != nil {
			return nil, err
		}
		oi.Medicine = &med
		order.Items = append(order.Items, oi)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &order, nil
}

// ListByPatient returns all orders for a given patient name.
func (s *OrderStore) ListByPatient(ctx context.Context, patientName string) ([]Order, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, patient_name, address, status, estimated_delivery, created_at, updated_at
		 FROM orders WHERE patient_name = $1 ORDER BY created_at DESC`, patientName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

// ListByStatus returns all orders with the given status.
func (s *OrderStore) ListByStatus(ctx context.Context, status OrderStatus) ([]Order, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, patient_name, address, status, estimated_delivery, created_at, updated_at
		 FROM orders WHERE status = $1 ORDER BY created_at ASC`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

// ListNonTerminal returns all orders that have not yet been delivered.
func (s *OrderStore) ListNonTerminal(ctx context.Context) ([]Order, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, patient_name, address, status, estimated_delivery, created_at, updated_at
		 FROM orders WHERE status != $1 ORDER BY created_at ASC`, StatusDelivered)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

// AdvanceStatus moves an order to the next status in the progression.
// It returns an error if the order is already in a terminal state.
func (s *OrderStore) AdvanceStatus(ctx context.Context, id string) (*Order, error) {
	var current OrderStatus
	err := s.db.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1`, id).Scan(&current)
	if err != nil {
		return nil, err
	}

	next, terminal := current.Next()
	if terminal {
		return nil, fmt.Errorf("order %s is already in terminal status %s", id, current)
	}

	var order Order
	err = s.db.QueryRow(ctx,
		`UPDATE orders SET status = $1, updated_at = now()
		 WHERE id = $2
		 RETURNING id, patient_name, address, status, estimated_delivery, created_at, updated_at`,
		next, id,
	).Scan(&order.ID, &order.PatientName, &order.Address, &order.Status, &order.EstimatedDelivery, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// scanOrders is a helper that reads a result set of orders (no items).
func scanOrders(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}) ([]Order, error) {
	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.PatientName, &o.Address, &o.Status, &o.EstimatedDelivery, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

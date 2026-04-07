package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jwilson/dronerx/internal/handlers"
	"github.com/jwilson/dronerx/internal/models"
)

type mockOrderStore struct {
	orders []models.Order
}

func (m *mockOrderStore) Create(ctx context.Context, req models.CreateOrderRequest, eta time.Time) (*models.Order, error) {
	order := &models.Order{
		ID:          "order-1",
		PatientName: req.PatientName,
		Address:     req.Address,
		Status:      models.StatusPlaced,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	for _, item := range req.Items {
		order.Items = append(order.Items, models.OrderItem{
			ID:         "item-1",
			OrderID:    order.ID,
			MedicineID: item.MedicineID,
			Quantity:   item.Quantity,
		})
	}
	m.orders = append(m.orders, *order)
	return order, nil
}

func (m *mockOrderStore) GetByID(ctx context.Context, id string) (*models.Order, error) {
	for _, o := range m.orders {
		if o.ID == id {
			return &o, nil
		}
	}
	return nil, nil
}

func (m *mockOrderStore) ListByPatient(ctx context.Context, name string) ([]models.Order, error) {
	var result []models.Order
	for _, o := range m.orders {
		if o.PatientName == name {
			result = append(result, o)
		}
	}
	return result, nil
}

func newOrderHandler() (*handlers.OrderHandler, *mockOrderStore) {
	store := &mockOrderStore{}
	return handlers.NewOrderHandler(store, 60), store
}

func TestCreateOrder(t *testing.T) {
	h, _ := newOrderHandler()

	body := models.CreateOrderRequest{
		PatientName: "Jane Doe",
		Address:     "456 Oak Ave",
		Items:       []models.OrderItemRequest{{MedicineID: "med-1", Quantity: 2}},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var resp models.Order
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.PatientName != "Jane Doe" {
		t.Errorf("expected patient_name Jane Doe, got %s", resp.PatientName)
	}
	if resp.Status != models.StatusPlaced {
		t.Errorf("expected status placed, got %s", resp.Status)
	}
}

func TestCreateOrder_InvalidBody(t *testing.T) {
	h, _ := newOrderHandler()

	req := httptest.NewRequest("POST", "/orders", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreateOrder_ValidationFails(t *testing.T) {
	h, _ := newOrderHandler()

	// Missing patient_name
	body := models.CreateOrderRequest{
		Address: "456 Oak Ave",
		Items:   []models.OrderItemRequest{{MedicineID: "med-1", Quantity: 1}},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetOrder(t *testing.T) {
	h, store := newOrderHandler()

	// Pre-seed an order in the store.
	store.orders = []models.Order{
		{
			ID:          "order-1",
			PatientName: "Jane Doe",
			Address:     "456 Oak Ave",
			Status:      models.StatusPlaced,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	req := httptest.NewRequest("GET", "/orders/order-1", nil)
	req.SetPathValue("id", "order-1")
	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["id"] != "order-1" {
		t.Errorf("expected id order-1, got %v", resp["id"])
	}
	if _, ok := resp["remaining_eta_seconds"]; !ok {
		t.Error("expected remaining_eta_seconds in response")
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	h, _ := newOrderHandler()

	req := httptest.NewRequest("GET", "/orders/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestListOrders(t *testing.T) {
	h, store := newOrderHandler()

	store.orders = []models.Order{
		{ID: "order-1", PatientName: "Jane Doe", Address: "456 Oak Ave", Status: models.StatusPlaced, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "order-2", PatientName: "John Smith", Address: "789 Elm St", Status: models.StatusDelivered, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	req := httptest.NewRequest("GET", "/orders?patient_name=Jane+Doe", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body []models.Order
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body) != 1 {
		t.Errorf("expected 1 order for Jane Doe, got %d", len(body))
	}
	if body[0].PatientName != "Jane Doe" {
		t.Errorf("expected patient Jane Doe, got %s", body[0].PatientName)
	}
}

func TestListOrders_MissingPatientName(t *testing.T) {
	h, _ := newOrderHandler()

	req := httptest.NewRequest("GET", "/orders", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

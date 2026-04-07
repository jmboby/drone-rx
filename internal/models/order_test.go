package models_test

import (
	"testing"
	"github.com/jwilson/dronerx/internal/models"
)

func TestOrderStatusProgression(t *testing.T) {
	tests := []struct {
		current  models.OrderStatus
		expected models.OrderStatus
		terminal bool
	}{
		{models.StatusPlaced, models.StatusPreparing, false},
		{models.StatusPreparing, models.StatusInFlight, false},
		{models.StatusInFlight, models.StatusDelivered, false},
		{models.StatusDelivered, "", true},
	}
	for _, tt := range tests {
		next, isTerminal := tt.current.Next()
		if isTerminal != tt.terminal {
			t.Errorf("status %s: expected terminal=%v, got %v", tt.current, tt.terminal, isTerminal)
		}
		if !tt.terminal && next != tt.expected {
			t.Errorf("status %s: expected next=%s, got %s", tt.current, tt.expected, next)
		}
	}
}

func TestOrderStatusIsValid(t *testing.T) {
	if !models.StatusPlaced.IsValid() { t.Error("placed should be valid") }
	if models.OrderStatus("invalid").IsValid() { t.Error("invalid should not be valid") }
}

func TestCreateOrderRequest(t *testing.T) {
	req := models.CreateOrderRequest{
		PatientName: "John", Address: "123 Main St",
		Items: []models.OrderItemRequest{{MedicineID: "med-1", Quantity: 2}},
	}
	if err := req.Validate(); err != nil { t.Errorf("expected valid request, got error: %v", err) }
}

func TestCreateOrderRequest_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  models.CreateOrderRequest
	}{
		{"empty name", models.CreateOrderRequest{Address: "123 St", Items: []models.OrderItemRequest{{MedicineID: "m1", Quantity: 1}}}},
		{"empty address", models.CreateOrderRequest{PatientName: "John", Items: []models.OrderItemRequest{{MedicineID: "m1", Quantity: 1}}}},
		{"no items", models.CreateOrderRequest{PatientName: "John", Address: "123 St"}},
		{"zero quantity", models.CreateOrderRequest{PatientName: "John", Address: "123 St", Items: []models.OrderItemRequest{{MedicineID: "m1", Quantity: 0}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); err == nil { t.Error("expected validation error") }
		})
	}
}

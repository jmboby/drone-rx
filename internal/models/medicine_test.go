package models_test

import (
	"testing"
	"github.com/jwilson/dronerx/internal/models"
)

func TestMedicineFields(t *testing.T) {
	m := models.Medicine{
		ID: "test-id", Name: "Paracetamol", Description: "Pain relief",
		Price: 4.99, InStock: true, Category: "Pain Relief",
	}
	if m.ID != "test-id" { t.Errorf("expected ID test-id, got %s", m.ID) }
	if m.Name != "Paracetamol" { t.Errorf("expected Name Paracetamol, got %s", m.Name) }
	if m.Price != 4.99 { t.Errorf("expected Price 4.99, got %f", m.Price) }
}

package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/handlers"
	"github.com/jwilson/dronerx/internal/models"
)

type mockMedicineStore struct {
	medicines []models.Medicine
}

func (m *mockMedicineStore) List(ctx context.Context) ([]models.Medicine, error) {
	return m.medicines, nil
}

func (m *mockMedicineStore) GetByID(ctx context.Context, id string) (*models.Medicine, error) {
	for _, med := range m.medicines {
		if med.ID == id {
			return &med, nil
		}
	}
	return nil, nil
}

var testMedicines = []models.Medicine{
	{ID: "med-1", Name: "Amoxicillin", Description: "Antibiotic", Price: 12.50, InStock: true, Category: "antibiotics"},
	{ID: "med-2", Name: "Ibuprofen", Description: "Pain reliever", Price: 8.00, InStock: true, Category: "pain-relief"},
}

func newMedicineHandler() *handlers.MedicineHandler {
	return handlers.NewMedicineHandler(&mockMedicineStore{medicines: testMedicines})
}

func TestListMedicines(t *testing.T) {
	h := newMedicineHandler()
	req := httptest.NewRequest("GET", "/medicines", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body []models.Medicine
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body) != 2 {
		t.Errorf("expected 2 medicines, got %d", len(body))
	}
	if body[0].ID != "med-1" {
		t.Errorf("expected first medicine id med-1, got %s", body[0].ID)
	}
}

func TestGetMedicine(t *testing.T) {
	h := newMedicineHandler()
	req := httptest.NewRequest("GET", "/medicines/med-1", nil)
	req.SetPathValue("id", "med-1")
	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body models.Medicine
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.ID != "med-1" {
		t.Errorf("expected medicine id med-1, got %s", body.ID)
	}
	if body.Name != "Amoxicillin" {
		t.Errorf("expected name Amoxicillin, got %s", body.Name)
	}
}

func TestGetMedicine_NotFound(t *testing.T) {
	h := newMedicineHandler()
	req := httptest.NewRequest("GET", "/medicines/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

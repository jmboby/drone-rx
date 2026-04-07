package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jwilson/dronerx/internal/models"
)

// MedicineLister is the interface used by MedicineHandler to query medicines.
type MedicineLister interface {
	List(ctx context.Context) ([]models.Medicine, error)
	GetByID(ctx context.Context, id string) (*models.Medicine, error)
}

// MedicineHandler handles HTTP requests for medicines.
type MedicineHandler struct {
	store MedicineLister
}

// NewMedicineHandler constructs a MedicineHandler with the given store.
func NewMedicineHandler(store MedicineLister) *MedicineHandler {
	return &MedicineHandler{store: store}
}

// List handles GET /medicines — returns all in-stock medicines as a JSON array.
func (h *MedicineHandler) List(w http.ResponseWriter, r *http.Request) {
	medicines, err := h.store.List(r.Context())
	if err != nil {
		http.Error(w, "failed to fetch medicines", http.StatusInternalServerError)
		return
	}
	// Return an empty array rather than null when there are no medicines.
	if medicines == nil {
		medicines = []models.Medicine{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medicines)
}

// GetByID handles GET /medicines/{id} — returns a single medicine by ID.
func (h *MedicineHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	medicine, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to fetch medicine", http.StatusInternalServerError)
		return
	}
	if medicine == nil {
		http.Error(w, "medicine not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medicine)
}

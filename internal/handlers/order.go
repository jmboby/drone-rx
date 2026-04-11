package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jwilson/dronerx/internal/models"
)

// OrderStorer is the interface used by OrderHandler to query and create orders.
type OrderStorer interface {
	Create(ctx context.Context, req models.CreateOrderRequest, eta time.Time) (*models.Order, error)
	GetByID(ctx context.Context, id string) (*models.Order, error)
	ListByPatient(ctx context.Context, name string) ([]models.Order, error)
}

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	store          OrderStorer
	tickerInterval int // seconds between status transitions
}

// NewOrderHandler constructs an OrderHandler with the given store and ticker interval.
func NewOrderHandler(store OrderStorer, tickerInterval int) *OrderHandler {
	return &OrderHandler{store: store, tickerInterval: tickerInterval}
}

// OrderResponse embeds Order and adds the remaining ETA in seconds.
type OrderResponse struct {
	models.Order
	RemainingETA float64 `json:"remaining_eta_seconds"`
}

// Create handles POST /orders — decodes and validates the request, calculates ETA,
// creates the order, and returns 201 with the created order.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	eta := models.CalculateETA(time.Now(), h.tickerInterval)

	order, err := h.store.Create(r.Context(), req, eta)
	if err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	slog.Info("order created", "order_id", order.ID, "patient", req.PatientName, "items", len(req.Items))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// GetByID handles GET /orders/{id} — returns the order with its remaining ETA.
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	order, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to fetch order", http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	slog.Debug("order fetched", "order_id", id)

	remaining := models.RemainingETA(order.Status, order.UpdatedAt, h.tickerInterval)
	resp := OrderResponse{
		Order:        *order,
		RemainingETA: remaining.Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// List handles GET /orders — requires a patient_name query param, returns matching orders.
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	patientName := r.URL.Query().Get("patient_name")
	if patientName == "" {
		http.Error(w, "patient_name query parameter is required", http.StatusBadRequest)
		return
	}

	orders, err := h.store.ListByPatient(r.Context(), patientName)
	if err != nil {
		http.Error(w, "failed to fetch orders", http.StatusInternalServerError)
		return
	}

	// Return an empty array rather than null when there are no orders.
	if orders == nil {
		orders = []models.Order{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

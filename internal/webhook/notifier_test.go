package webhook_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/webhook"
)

func TestNotifier_SendsWebhook(t *testing.T) {
	var received map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	n := webhook.NewNotifier(server.URL)
	n.NotifyDelivered("order-1", "Alice", "42 High St")
	if received["order_id"] != "order-1" {
		t.Errorf("expected order_id order-1, got %s", received["order_id"])
	}
	if received["patient_name"] != "Alice" {
		t.Errorf("expected patient_name Alice, got %s", received["patient_name"])
	}
	if received["event"] != "delivered" {
		t.Errorf("expected event delivered, got %s", received["event"])
	}
}

func TestNotifier_EmptyURL_NoOp(t *testing.T) {
	n := webhook.NewNotifier("")
	n.NotifyDelivered("order-1", "Alice", "42 High St") // Should not panic
}

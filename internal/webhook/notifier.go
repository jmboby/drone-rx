package webhook

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Notifier struct {
	url    string
	client *http.Client
}

func NewNotifier(url string) *Notifier {
	return &Notifier{url: url, client: &http.Client{Timeout: 10 * time.Second}}
}

func (n *Notifier) NotifyDelivered(orderID, patientName, address string) {
	if n.url == "" {
		return
	}
	payload := map[string]string{
		"event":        "delivered",
		"order_id":     orderID,
		"patient_name": patientName,
		"address":      address,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("webhook failed", "order_id", orderID, "url", n.url, "error", err)
		return
	}
	resp, err := n.client.Post(n.url, "application/json", bytes.NewReader(data))
	if err != nil {
		slog.Error("webhook failed", "order_id", orderID, "url", n.url, "error", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		slog.Error("webhook failed", "order_id", orderID, "url", n.url, "status", resp.StatusCode)
		return
	}
	slog.Info("webhook delivered", "order_id", orderID, "url", n.url)
}

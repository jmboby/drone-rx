package webhook

import (
	"bytes"
	"encoding/json"
	"log"
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
		log.Printf("webhook: marshal error: %v", err)
		return
	}
	resp, err := n.client.Post(n.url, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("webhook: POST to %s failed: %v", n.url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("webhook: POST to %s returned %d", n.url, resp.StatusCode)
	}
}

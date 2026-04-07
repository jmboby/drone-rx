package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type OrderStatusEvent struct {
	OrderID           string `json:"order_id"`
	Status            string `json:"status"`
	EstimatedDelivery string `json:"estimated_delivery,omitempty"`
	UpdatedAt         string `json:"updated_at"`
}

type Publisher struct{ nc *nats.Conn }

func NewPublisher(nc *nats.Conn) *Publisher { return &Publisher{nc: nc} }

func (p *Publisher) PublishOrderStatus(orderID, status string, estimatedDelivery *time.Time, updatedAt time.Time) error {
	event := OrderStatusEvent{
		OrderID:   orderID,
		Status:    status,
		UpdatedAt: updatedAt.Format(time.RFC3339),
	}
	if estimatedDelivery != nil {
		event.EstimatedDelivery = estimatedDelivery.Format(time.RFC3339)
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	subject := fmt.Sprintf("orders.%s.status", orderID)
	if err := p.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}

func ConnectNATS(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS at %s: %w", url, err)
	}
	return nc, nil
}

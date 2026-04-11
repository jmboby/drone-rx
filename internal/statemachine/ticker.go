package statemachine

import (
	"context"
	"log/slog"
	"time"

	"github.com/jwilson/dronerx/internal/models"
)

type OrderAdvancer interface {
	ListNonTerminal(ctx context.Context) ([]models.Order, error)
	AdvanceStatus(ctx context.Context, id string) (*models.Order, error)
}

type StatusPublisher interface {
	PublishOrderStatus(orderID, status string, eta *time.Time, updatedAt time.Time) error
}

type DeliveryNotifier interface {
	NotifyDelivered(orderID, patientName, address string)
}

type Ticker struct {
	advancer    OrderAdvancer
	publisher   StatusPublisher
	notifier    DeliveryNotifier
	intervalSec int
}

func NewTicker(advancer OrderAdvancer, publisher StatusPublisher, notifier DeliveryNotifier, intervalSec int) *Ticker {
	return &Ticker{advancer: advancer, publisher: publisher, notifier: notifier, intervalSec: intervalSec}
}

func (t *Ticker) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(t.intervalSec) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.tick(ctx)
		}
	}
}

func (t *Ticker) tick(ctx context.Context) {
	orders, err := t.advancer.ListNonTerminal(ctx)
	if err != nil {
		slog.Error("ticker: list orders failed", "error", err)
		return
	}
	for _, order := range orders {
		updated, err := t.advancer.AdvanceStatus(ctx, order.ID)
		if err != nil {
			slog.Error("ticker: advance failed", "order_id", order.ID, "error", err)
			continue
		}
		if updated == nil {
			continue
		}
		slog.Info("order status changed", "order_id", updated.ID, "from", order.Status, "to", updated.Status)
		if err := t.publisher.PublishOrderStatus(updated.ID, string(updated.Status), nil, updated.UpdatedAt); err != nil {
			slog.Error("ticker: publish status failed", "order_id", updated.ID, "error", err)
		}
		if updated.Status == models.StatusDelivered {
			slog.Info("order delivered", "order_id", updated.ID, "patient", updated.PatientName)
			t.notifier.NotifyDelivered(updated.ID, updated.PatientName, updated.Address)
		}
	}
}

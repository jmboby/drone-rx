package statemachine

import (
	"context"
	"log"
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
		log.Printf("ticker: listing orders: %v", err)
		return
	}
	for _, order := range orders {
		updated, err := t.advancer.AdvanceStatus(ctx, order.ID)
		if err != nil {
			log.Printf("ticker: advancing order %s: %v", order.ID, err)
			continue
		}
		if updated == nil {
			continue
		}
		if err := t.publisher.PublishOrderStatus(updated.ID, string(updated.Status), nil, updated.UpdatedAt); err != nil {
			log.Printf("ticker: publishing status for %s: %v", updated.ID, err)
		}
		if updated.Status == models.StatusDelivered {
			t.notifier.NotifyDelivered(updated.ID, updated.PatientName, updated.Address)
		}
	}
}

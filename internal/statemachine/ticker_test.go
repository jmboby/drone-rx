package statemachine_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jwilson/dronerx/internal/models"
	"github.com/jwilson/dronerx/internal/statemachine"
)

type mockAdvancer struct {
	mu       sync.Mutex
	advanced []string
	orders   []models.Order
}

func (m *mockAdvancer) ListNonTerminal(ctx context.Context) ([]models.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.orders, nil
}

func (m *mockAdvancer) AdvanceStatus(ctx context.Context, id string) (*models.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.advanced = append(m.advanced, id)
	for i, o := range m.orders {
		if o.ID == id {
			next, terminal := o.Status.Next()
			if !terminal {
				m.orders[i].Status = next
			}
			return &m.orders[i], nil
		}
	}
	return nil, nil
}

type mockPublisher struct {
	mu        sync.Mutex
	published []string
}

func (m *mockPublisher) PublishOrderStatus(orderID, status string, eta *time.Time, updatedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, orderID+":"+status)
	return nil
}

type mockNotifier struct {
	mu       sync.Mutex
	notified []string
}

func (m *mockNotifier) NotifyDelivered(orderID, patientName, address string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notified = append(m.notified, orderID)
}

func TestTicker_AdvancesOrders(t *testing.T) {
	now := time.Now()
	advancer := &mockAdvancer{
		orders: []models.Order{
			{ID: "o1", PatientName: "Alice", Status: models.StatusPlaced, UpdatedAt: now},
		},
	}
	pub := &mockPublisher{}
	notifier := &mockNotifier{}
	ticker := statemachine.NewTicker(advancer, pub, notifier, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go ticker.Start(ctx)
	time.Sleep(2500 * time.Millisecond)
	cancel()
	advancer.mu.Lock()
	defer advancer.mu.Unlock()
	if len(advancer.advanced) < 1 {
		t.Errorf("expected at least 1 advancement, got %d", len(advancer.advanced))
	}
}

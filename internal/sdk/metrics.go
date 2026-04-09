package sdk

import (
	"context"
	"log"
	"time"

	"github.com/jwilson/dronerx/internal/models"
)

type MetricsSource interface {
	DeliveryStats(ctx context.Context) (*models.DeliveryStats, error)
}

func StartMetricsSender(ctx context.Context, client *Client, source MetricsSource, interval time.Duration) {
	// Send immediately on startup
	sendMetrics(ctx, client, source)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sendMetrics(ctx, client, source)
		}
	}
}

func sendMetrics(ctx context.Context, client *Client, source MetricsSource) {
	stats, err := source.DeliveryStats(ctx)
	if err != nil {
		log.Printf("metrics: fetching delivery stats: %v", err)
		return
	}

	data := map[string]interface{}{
		"total_orders":              stats.TotalOrders,
		"orders_completed":          stats.OrdersCompleted,
		"avg_delivery_time_seconds": stats.AvgDeliveryTimeSec,
	}

	log.Printf("metrics: sending total=%d completed=%d avg_time=%.1fs",
		stats.TotalOrders, stats.OrdersCompleted, stats.AvgDeliveryTimeSec)

	client.SendMetrics(data)
}

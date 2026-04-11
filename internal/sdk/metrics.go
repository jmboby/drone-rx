package sdk

import (
	"context"
	"log/slog"
	"time"

	"github.com/jwilson/dronerx/internal/models"
)

type MetricsSource interface {
	DeliveryStats(ctx context.Context) (*models.DeliveryStats, error)
}

func StartMetricsSender(ctx context.Context, client *Client, source MetricsSource, interval time.Duration) {
	var last *models.DeliveryStats

	// Send immediately on startup.
	last = sendMetrics(ctx, client, source, last)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			last = sendMetrics(ctx, client, source, last)
		}
	}
}

// sendMetrics fetches stats and sends them only when values have changed (or on first send).
// Returns the last successfully sent stats so the caller can track state.
func sendMetrics(ctx context.Context, client *Client, source MetricsSource, last *models.DeliveryStats) *models.DeliveryStats {
	stats, err := source.DeliveryStats(ctx)
	if err != nil {
		slog.Error("metrics: send failed", "error", err)
		return last
	}

	// Skip sending when nothing has changed.
	if last != nil &&
		stats.TotalOrders == last.TotalOrders &&
		stats.OrdersCompleted == last.OrdersCompleted &&
		stats.AvgDeliveryTimeSec == last.AvgDeliveryTimeSec {
		return last
	}

	data := map[string]interface{}{
		"total_orders":              stats.TotalOrders,
		"orders_completed":          stats.OrdersCompleted,
		"avg_delivery_time_seconds": stats.AvgDeliveryTimeSec,
	}

	if last == nil {
		slog.Info("metrics: initial send",
			"total_orders", stats.TotalOrders,
			"orders_completed", stats.OrdersCompleted,
			"avg_delivery_time_seconds", stats.AvgDeliveryTimeSec,
		)
	} else {
		slog.Info("metrics: updated",
			"total_orders", stats.TotalOrders,
			"orders_completed", stats.OrdersCompleted,
			"avg_delivery_time_seconds", stats.AvgDeliveryTimeSec,
		)
	}

	client.SendMetrics(data)
	return stats
}

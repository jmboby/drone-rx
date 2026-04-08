package sdk

import (
	"context"
	"log"
	"time"
)

// MetricsSource provides order count data for metrics reporting.
type MetricsSource interface {
	CountByStatus(ctx context.Context) (map[string]int, error)
}

// StartMetricsSender starts a goroutine that sends order metrics on the given interval.
func StartMetricsSender(ctx context.Context, client *Client, source MetricsSource, interval time.Duration) {
	go func() {
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
	}()
}

// sendMetrics fetches counts from source and posts them as custom metrics.
func sendMetrics(ctx context.Context, client *Client, source MetricsSource) {
	counts, err := source.CountByStatus(ctx)
	if err != nil {
		log.Printf("metrics: count by status: %v", err)
		return
	}
	data := make(map[string]interface{}, len(counts))
	for status, count := range counts {
		data["orders_"+status] = count
	}
	if err := client.SendMetrics(data); err != nil {
		log.Printf("metrics: send: %v", err)
	}
}

package models_test

import (
	"testing"
	"time"
	"github.com/jwilson/dronerx/internal/models"
)

func TestCalculateETA_NewOrder(t *testing.T) {
	now := time.Now()
	eta := models.CalculateETA(now, 30)
	expectedDuration := 4 * 30 * time.Second
	expected := now.Add(expectedDuration)
	diff := eta.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected ETA near %v, got %v (diff: %v)", expected, eta, diff)
	}
}

func TestCalculateRemainingETA(t *testing.T) {
	tests := []struct {
		name     string
		status   models.OrderStatus
		interval int
		minSecs  float64
		maxSecs  float64
	}{
		{"placed has 3 transitions left", models.StatusPlaced, 30, 80, 100},
		{"preparing has 2 transitions left", models.StatusPreparing, 30, 50, 70},
		{"in-flight has 1 transition left", models.StatusInFlight, 30, 20, 40},
		{"delivered has 0 remaining", models.StatusDelivered, 30, -1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining := models.RemainingETA(tt.status, time.Now(), tt.interval)
			secs := remaining.Seconds()
			if secs < tt.minSecs || secs > tt.maxSecs {
				t.Errorf("expected remaining between %v-%vs, got %vs", tt.minSecs, tt.maxSecs, secs)
			}
		})
	}
}

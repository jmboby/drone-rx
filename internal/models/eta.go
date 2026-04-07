package models

import "time"

func transitionsRemaining(status OrderStatus) int {
	switch status {
	case StatusPlaced:
		return 3
	case StatusPreparing:
		return 2
	case StatusInFlight:
		return 1
	case StatusDelivered:
		return 0
	default:
		return 0
	}
}

// CalculateETA returns the estimated delivery time for a newly placed order.
// intervalSecs is the number of seconds between each status transition.
func CalculateETA(orderCreated time.Time, intervalSecs int) time.Time {
	totalDuration := time.Duration(4*intervalSecs) * time.Second
	return orderCreated.Add(totalDuration)
}

// RemainingETA returns the duration remaining until delivery based on the
// current status. updatedAt is the time the order last changed status.
func RemainingETA(status OrderStatus, updatedAt time.Time, intervalSecs int) time.Duration {
	remaining := transitionsRemaining(status)
	if remaining == 0 {
		return 0
	}
	totalRemaining := time.Duration(remaining*intervalSecs) * time.Second
	elapsed := time.Since(updatedAt)
	eta := totalRemaining - elapsed
	if eta < 0 {
		return 0
	}
	return eta
}

package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/nats-io/nats.go"

	"github.com/jwilson/dronerx/internal/sdk"
)

type TrackingHandler struct {
	nc     *nats.Conn
	client *sdk.Client
}

func NewTrackingHandler(nc *nats.Conn, client *sdk.Client) *TrackingHandler {
	return &TrackingHandler{nc: nc, client: client}
}

func (h *TrackingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	if orderID == "" {
		http.Error(w, "order ID required", http.StatusBadRequest)
		return
	}

	if h.client != nil && !h.client.IsFeatureEnabled("live_tracking_enabled") {
		slog.Warn("tracking denied — license", "order_id", orderID)
		http.Error(w, "live tracking not enabled for this license", http.StatusForbidden)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		slog.Error("websocket accept failed", "order_id", orderID, "error", err)
		return
	}
	defer conn.CloseNow()

	slog.Info("tracking connected", "order_id", orderID)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	subject := "orders." + orderID + ".status"
	sub, err := h.nc.Subscribe(subject, func(msg *nats.Msg) {
		if err := conn.Write(ctx, websocket.MessageText, msg.Data); err != nil {
			slog.Error("websocket write failed", "order_id", orderID, "error", err)
			cancel()
		}
	})
	if err != nil {
		slog.Error("nats subscribe failed", "order_id", orderID, "error", err)
		conn.Close(websocket.StatusInternalError, "subscription failed")
		return
	}
	defer sub.Unsubscribe()

	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			break
		}
	}
	conn.Close(websocket.StatusNormalClosure, "")
	slog.Info("tracking disconnected", "order_id", orderID)
}

package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/nats-io/nats.go"
)

type TrackingHandler struct{ nc *nats.Conn }

func NewTrackingHandler(nc *nats.Conn) *TrackingHandler { return &TrackingHandler{nc: nc} }

func (h *TrackingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	if orderID == "" {
		http.Error(w, "order ID required", http.StatusBadRequest)
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		log.Printf("websocket accept: %v", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	subject := "orders." + orderID + ".status"
	sub, err := h.nc.Subscribe(subject, func(msg *nats.Msg) {
		if err := conn.Write(ctx, websocket.MessageText, msg.Data); err != nil {
			log.Printf("websocket write: %v", err)
			cancel()
		}
	})
	if err != nil {
		log.Printf("nats subscribe: %v", err)
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
}

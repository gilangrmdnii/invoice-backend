package handler

import (
	"bufio"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/internal/middleware"
	"github.com/gilangrmdnii/invoice-backend/internal/sse"
)

type SSEHandler struct {
	hub *sse.Hub
}

func NewSSEHandler(hub *sse.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

func (h *SSEHandler) Stream(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ch, cleanup := h.hub.Subscribe(userID)
		defer cleanup()

		// Send initial connection event
		fmt.Fprintf(w, "event: connected\ndata: {\"user_id\":%d}\n\n", userID)
		if err := w.Flush(); err != nil {
			return
		}

		for event := range ch {
			data, err := event.Marshal()
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			if err := w.Flush(); err != nil {
				return
			}
		}
	})

	return nil
}

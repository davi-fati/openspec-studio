package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/davidlima/openspec-studio/backend/internal/service"
)

// EventsHandler godoc
//
//	@Summary		Real-time update stream (SSE)
//	@Description	Server-Sent Events stream. Emits an "update" event (with the changed file path as data) whenever a watched project's openspec/ directory changes on disk, plus a periodic heartbeat comment. Not a JSON request/response endpoint - no response schema applies.
//	@Tags			events
//	@Produce		text/event-stream
//	@Router			/events [get]
func EventsHandler(events *service.EventBroadcaster) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		flusher, ok := c.Writer.(interface{ Flush() })
		if !ok {
			c.Status(500)
			return
		}

		sub, unsubscribe := events.Subscribe()
		defer unsubscribe()

		heartbeat := time.NewTicker(15 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case ev, ok := <-sub:
				if !ok {
					return
				}
				fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", ev.Name, ev.Data)
				flusher.Flush()
			case <-heartbeat.C:
				fmt.Fprintf(c.Writer, ": heartbeat\n\n")
				flusher.Flush()
			case <-c.Request.Context().Done():
				return
			}
		}
	}
}

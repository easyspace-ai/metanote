package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"youmind-backend-v2/internal/config"
)

// W6Event is a lightweight representation of messages coming from the W6
// WebSocket stream. The exact JSON schema is controlled by the W6 service;
// we only care about a few generic fields.
type W6Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Current json.RawMessage `json:"current,omitempty"`
}

// W6WS wraps WebSocket connectivity for a single W6 chat.
type W6WS struct {
	wssBaseURL string
	authField  string
	authValue  string
	dialer     *websocket.Dialer
}

// NewW6WS constructs a WebSocket helper from config.
func NewW6WS(cfg config.W6Config) *W6WS {
	if cfg.WSSBaseURL == "" {
		return nil
	}
	return &W6WS{
		wssBaseURL: strings.TrimRight(cfg.WSSBaseURL, "/") + "/",
		authField:  cfg.AuthHeaderKey,
		authValue:  cfg.AuthHeaderVal,
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		},
	}
}

// ConnectAndStream dials the W6 WebSocket endpoint for the given chatID and
// forwards each JSON message to the provided callback. The callback should
// return false to stop streaming early.
func (w *W6WS) ConnectAndStream(
	ctx context.Context,
	chatID string,
	onEvent func(ev W6Event) bool,
) error {
	if w == nil {
		return fmt.Errorf("w6 websocket is not configured")
	}

	header := http.Header{}
	if w.authField != "" && w.authValue != "" {
		header.Set(w.authField, w.authValue)
	}

	url := w.wssBaseURL + chatID
	conn, _, err := w.dialer.DialContext(ctx, url, header)
	if err != nil {
		return fmt.Errorf("dial w6 websocket: %w", err)
	}
	defer conn.Close()

	// Ensure we close on context cancellation.
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		case <-done:
		}
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			close(done)
			// Normal closure is not an error.
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return nil
			}
			// If context is cancelled, treat as normal exit.
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("read w6 websocket: %w", err)
		}

		var ev W6Event
		if err := json.Unmarshal(data, &ev); err != nil {
			// Skip malformed events but continue streaming.
			continue
		}

		if onEvent != nil {
			if !onEvent(ev) {
				break
			}
		}
	}

	close(done)
	return nil
}


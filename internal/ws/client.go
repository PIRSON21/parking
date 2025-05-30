package ws

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	// Conn - подключение к клиенту
	Conn *websocket.Conn
	// SendChan - канал для отправки сообщений клиенту
	SendChan chan []byte
	Done     chan struct{}
	once     sync.Once
}

// Stop закрывает соединение с клиентом.
func (c *Client) Stop() {
	c.Conn.Close()
	c.once.Do(func() {
		close(c.Done)
	})
}

// NewClient создает клиента сервера.
func NewClient(conn *websocket.Conn) *Client {
	send := make(chan []byte, 256)

	return &Client{
		Conn:     conn,
		SendChan: send,
		Done:     make(chan struct{}),
	}
}

// ReadLoop получает сообщения от клиента.
func (c *Client) ReadLoop(log *slog.Logger, onMessage func(msg []byte)) {
	const op = "ws.client.ReadLoop"
	log = log.With(slog.String("op", op))

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Error("error while reading from webSocket conn", slog.String("err", err.Error()))
			}
			close(c.SendChan)
			return
		}
		onMessage(msg)
	}
}

// WriteLoop отправляет сообщения клиенту.
func (c *Client) WriteLoop(log *slog.Logger) {
	const op = "ws.client.WriteLoop"

	log = log.With(slog.String("op", op))

	defer c.Conn.Close()
	defer c.once.Do(func() {
		close(c.Done)
	})

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.SendChan:
			if !ok {
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Error("error while sending message on webSocket", slog.String("err", err.Error()))
				}
				return
			}
		case <-ticker.C:
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Error("error while sending message on webSocket", slog.String("err", err.Error()))
				}
				return
			}
		}
	}
}

func (c *Client) Send(msg []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	select {
	case c.SendChan <- msg:
	default:
	}
}

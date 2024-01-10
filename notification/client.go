package notification

import "github.com/gorilla/websocket"

type Client struct {
	Conn *websocket.Conn
}

func (s *Client) Write(p []byte) (int, error) {
	if err := s.Conn.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *Client) Close() error {
	return s.Conn.Close()
}

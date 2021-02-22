package main

import (
	"context"
	"github.com/gorilla/websocket"
	"net/url"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	wssrv := NewWebsocketServer("3000", "/ws")
	go func() {
		if err := wssrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	m.Run()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := wssrv.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func TestWebsocketHandler_ServeHTTP(t *testing.T) {
	u := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/ws"}
	t.Logf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal("dial:", err)
	}
	defer func(c *websocket.Conn) {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}(c)
}

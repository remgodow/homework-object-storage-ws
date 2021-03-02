package main

import (
	"context"
	"github.com/gorilla/websocket"
	"net/url"
	"testing"
	"time"
)

const (
	listeningPort = "3001"
	wsPath        = "/ws"
)

type GetMethod struct {
}

func (GetMethod) GetType() string {
	return "GET"
}

func (GetMethod) Handle(_ map[string]interface{}) interface{} {
	resp := struct {
		Data string `json:"data"`
	}{
		"some data",
	}
	return resp
}

func TestMain(m *testing.M) {
	wssrv := NewWebsocketServer(listeningPort, wsPath)
	wssrv.AddMethod(&GetMethod{})
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
	u := url.URL{Scheme: "ws", Host: "localhost:" + listeningPort, Path: wsPath}
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

func TestGetHandler(t *testing.T) {
	u := url.URL{Scheme: "ws", Host: "localhost:" + listeningPort, Path: wsPath}
	t.Logf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal("dial:", err)
	}
	req := struct {
		Type string `json:"type"`
	}{
		"GET",
	}

	err = c.WriteJSON(req)
	if err != nil {
		t.Fatal(err)
	}

	response := make(map[string]interface{})
	err = c.ReadJSON(&response)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := response["data"]; !ok {
		t.Fatal("invalid response")
	}

	defer func(c *websocket.Conn) {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}(c)
}

func TestInvalidRequest(t *testing.T) {
	u := url.URL{Scheme: "ws", Host: "localhost:" + listeningPort, Path: wsPath}
	t.Logf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal("dial:", err)
	}
	req := struct {
		Something string `json:"something"`
	}{
		"GET",
	}

	err = c.WriteJSON(req)
	if err != nil {
		t.Fatal(err)
	}

	response := make(map[string]interface{})
	err = c.ReadJSON(&response)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := response["message"]; !ok {
		t.Fatal("No error for invalid request")
	}

	defer func(c *websocket.Conn) {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}(c)
}

func TestNotAJSON(t *testing.T) {
	u := url.URL{Scheme: "ws", Host: "localhost:" + listeningPort, Path: wsPath}
	t.Logf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal("dial:", err)
	}

	err = c.WriteMessage(1, []byte("something"))
	if err != nil {
		t.Fatal(err)
	}

	response := make(map[string]interface{})
	err = c.ReadJSON(&response)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := response["message"]; !ok {
		t.Fatal("No error for invalid request")
	}

	defer func(c *websocket.Conn) {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}(c)
}

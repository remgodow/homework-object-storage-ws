package main

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type WebsocketServer struct {
	Port       string
	Path       string
	upgrader   websocket.Upgrader
	httpServer http.Server
}

func NewWebsocketServer(port string, path string) *WebsocketServer {
	server := &WebsocketServer{
		Port:     port,
		Path:     path,
		upgrader: websocket.Upgrader{},
	}
	server.httpServer = http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}
	http.Handle("/ws", server)
	return server
}

func (w *WebsocketServer) ListenAndServe() error {
	if err := w.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
		return err
	}
	return nil
}

func (w *WebsocketServer) Shutdown(ctx context.Context) error {
	return w.httpServer.Shutdown(ctx)
}

func (w *WebsocketServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	_, err := w.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	//TODO: handle requests
}

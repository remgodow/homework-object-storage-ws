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
	routes     map[string]WebsocketRoute
}

type WebsocketRoute interface {
	GetType() string
	Handle(request map[string]interface{}) (interface{}, error)
}

func NewWebsocketServer(port string, path string) *WebsocketServer {
	server := &WebsocketServer{
		Port:     port,
		Path:     path,
		upgrader: websocket.Upgrader{},
		routes:   make(map[string]WebsocketRoute),
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

func (w *WebsocketServer) AddRoute(route WebsocketRoute) {
	w.routes[route.GetType()] = route
}

func (w *WebsocketServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	conn, err := w.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer func(conn *websocket.Conn) {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}(conn)

	for {
		request := make(map[string]interface{})
		err := conn.ReadJSON(&request)
		if err != nil {
			log.Println("read:", err)
			break
		}
		var response interface{}
		//check if string?
		if reqtype, ok := request["type"]; ok {
			tp := reqtype.(string)
			if handler, ok := w.routes[tp]; ok {
				resp, err := handler.Handle(request)
				if err != nil {
					log.Println("handler error", err)
					break
				}
				response = resp
			} else {
				log.Println("no handler for type", tp)
				break
			}
		}
		if response != nil {
			err = conn.WriteJSON(&response)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}

	}
}

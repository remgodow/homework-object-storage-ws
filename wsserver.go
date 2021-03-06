package main

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Code    uint16 `json:"code"`
	Message string `json:"message"`
}

type WebsocketServer struct {
	Port       string
	Path       string
	upgrader   websocket.Upgrader
	httpServer http.Server
	methods    map[string]WebsocketMethod
}

type WebsocketMethod interface {
	GetType() string
	Handle(request map[string]interface{}) interface{}
}

func NewWebsocketServer(port string, path string) *WebsocketServer {
	server := &WebsocketServer{
		Port:     port,
		Path:     path,
		upgrader: websocket.Upgrader{},
		methods:  make(map[string]WebsocketMethod),
	}
	server.httpServer = http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}
	http.Handle(path, server)
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

//Use before ListenAndServe() otherwise there is a possibility of multithread issues
func (w *WebsocketServer) AddMethod(method WebsocketMethod) {
	w.methods[method.GetType()] = method
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
		var response interface{}
		err := conn.ReadJSON(&request)
		if err != nil {
			log.Println("Request read error:", err)
			if _, ok := err.(*websocket.CloseError); ok {
				break
			}
			response = ErrorResponse{Code: 400, Message: "Request should be a JSON"}
		}

		//do handling when no error response present
		if response == nil {
			if reqtype, ok := request["type"]; ok {
				if tp, ok := reqtype.(string); ok {
					if handler, ok := w.methods[tp]; ok {
						response = handler.Handle(request)
					} else {
						log.Println("no handler for type", tp)
						response = ErrorResponse{Code: 400, Message: "Unknown request type"}
					}
				} else {
					log.Println("type field is not a string")
					response = ErrorResponse{Code: 400, Message: "type field is not a string"}
				}
			} else {
				log.Println("type field is missing")
				response = ErrorResponse{Code: 400, Message: "type field is missing"}
			}
		}

		if response != nil {
			err = conn.WriteJSON(&response)
			if err != nil {
				log.Println("Could not write response:", err)
			}
		}
	}
}

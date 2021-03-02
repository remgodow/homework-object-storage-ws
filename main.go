package main

import "log"

func main() {
	wssrv := NewWebsocketServer("3000", "/ws")
	wssrv.AddMethod(&MinioGetMethod{})
	wssrv.AddMethod(&MinioPutMethod{})
	if err := wssrv.ListenAndServe(); err != nil {
		log.Println(err)
		panic(err)
	}
}

package main

import "log"

func main() {
	wssrv := NewWebsocketServer("3000", "/ws")
	wssrv.AddRoute(&MinioGetRoute{})
	wssrv.AddRoute(&MinioPutRoute{})
	if err := wssrv.ListenAndServe(); err != nil {
		log.Println(err)
		panic(err)
	}
}

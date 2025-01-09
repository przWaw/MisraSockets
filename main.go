package main

import (
	"MisraSockets/MisraSocket"
	"MisraSockets/util"
	"sync"
)

func main() {
	listen, send, initiator := util.AppInput()

	misraSocket := MisraSocket.NewMisraSocket()

	wg := sync.WaitGroup{}
	wg.Add(1)

	if *initiator {
		go misraSocket.Listen(*listen, &wg)
		misraSocket.InitOutgoingConnection(*send)
		misraSocket.SendInitMessage()
	} else {
		misraSocket.Listen(*listen, &wg)
		misraSocket.InitOutgoingConnection(*send)
	}

	wg.Wait()

	misraSocket.HandleMessages()
}

package main

import (
	"fmt"
	"netcat/server"
	"os"
)

func main() {

	args := len(os.Args)

	switch args {
	case 1:
		server.Port = "8989"
	case 2:
		server.Port = os.Args[1]
	default:
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	server.ServerTCP()

}

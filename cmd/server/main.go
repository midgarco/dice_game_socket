package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/midgarco/dice_game_socket/client"
	"github.com/midgarco/dice_game_socket/io"
	"github.com/midgarco/dice_game_socket/manager"
)

var clients uint64

func main() {
	fmt.Println("starting server...")
	l, err := net.Listen("tcp", ":12345")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("listening on port 12345")
	manager := manager.NewClientManager()
	go manager.Start()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		atomic.AddUint64(&clients, 1)
		cid := atomic.LoadUint64(&clients)
		c := client.New(conn)
		c.ID = cid
		c.Data = make(chan []byte)

		manager.Register(c)
		go manager.Receive(c)
		go manager.Send(c)

		fmt.Println(c)
		msg, err := json.Marshal(c)
		if err != nil {
			fmt.Printf("error marshaling client data: %v", err)
			continue
		}
		c.SendResponse("CLIENT", &io.Response{Message: string(msg)})
	}
}

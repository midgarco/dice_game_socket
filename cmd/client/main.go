package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/midgarco/dice_game_socket/client"
	"github.com/midgarco/dice_game_socket/io"
)

func main() {
	fmt.Println("starting client...")
	conn, err := net.Dial("tcp", "localhost:12345")
	if err != nil {
		fmt.Println(err)
	}
	c := client.New(conn)
	go c.Receive()
	for {
		r := bufio.NewReader(os.Stdin)
		msg, _ := r.ReadString('\n')
		msg = strings.TrimRight(msg, "\n")

		if len(c.Prompt) > 0 {
			msg = c.Prompt + " " + msg
			c.Prompt = ""
		}
		req := io.CreateRequest(msg)
		b, err := json.Marshal(req)
		if err != nil {
			fmt.Fprintf(os.Stdout, "error: %v", err)
			continue
		}
		conn.Write(b)
	}
}

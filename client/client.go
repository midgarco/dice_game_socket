package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/midgarco/dice_game/game"
	"github.com/midgarco/dice_game_socket/io"
)

type Client struct {
	ID     uint64
	Socket net.Conn
	Data   chan []byte
	Game   *game.Game
}

func (c *Client) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID uint64 `json:"id"`
	}{
		ID: c.ID,
	})
}

func New(conn net.Conn) *Client {
	return &Client{Socket: conn}
}

func (c Client) HandleError(err error) {
	resp := &io.Response{}
	resp.Error = true
	resp.Message = err.Error()

	c.SendResponse("ERROR", resp)
}

func (c Client) SendResponse(cat string, msg *io.Response) {
	msg.Group = strings.ToUpper(cat)

	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Fprintf(os.Stdout, "error marshaling json response: %v", err)
		return
	}
	c.Data <- append([]byte("RESP "), b...)
}

func (c Client) SendRequest(msg *io.Request) {
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Fprintf(os.Stdout, "error marshaling json response: %v", err)
		return
	}
	c.Data <- append([]byte("REQ "), b...)
}

func (c *Client) Receive() {
	for {
		msg := make([]byte, 4096)
		length, err := c.Socket.Read(msg)
		if err != nil {
			c.Socket.Close()
			break
		}
		if length > 0 {
			msg = bytes.Trim(msg, "\x00")

			command := string(msg[:bytes.IndexAny(msg, " ")])

			if command == "REQ" {
				data := &io.Request{}
				err := json.Unmarshal(msg[bytes.IndexAny(msg, " ")+1:], data)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}

			if command == "RESP" {
				data := &io.Response{}
				err := json.Unmarshal(msg[bytes.IndexAny(msg, " ")+1:], data)
				if err != nil {
					fmt.Println(err)
					continue
				}

				switch strings.ToUpper(data.Group) {
				case "CLIENT":
					fmt.Println("successfully connected")
					err := json.Unmarshal([]byte(data.Message), c)
					if err != nil {
						fmt.Println(err)
					}
				case "GAME":
					switch strings.ToUpper(data.Message) {
					case "UPDATE":
						fmt.Println("updating game data")
						c.Game = data.GameData
					}
				case "ERROR":
					fmt.Println(data.Message)
				case "FATAL":
					fmt.Println(data.Message)
					c.Socket.Close()
					break
				default:
					fmt.Println("received: " + string(msg))
				}
			}
		}
	}
}

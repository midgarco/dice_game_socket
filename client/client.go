package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

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
	time.Sleep(10 * time.Millisecond)

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
			fmt.Println(string(msg))

			space := bytes.IndexAny(msg, " ")
			if space < 0 {
				continue
			}

			command := string(msg[:space])

			if command == "REQ" {
				req := &io.Request{}
				err := json.Unmarshal(msg[space+1:], req)
				if err != nil {
					fmt.Printf("unmarshal request: %v", err)
					continue
				}

				if req.Command == "PROMPT" {
					fmt.Fprint(os.Stdin, "\r"+req.Data+" ")
				}
			}

			if command == "RESP" {
				resp := &io.Response{}
				err := json.Unmarshal(msg[space+1:], resp)
				if err != nil {
					fmt.Printf("unmarshal response: %v", err)
					continue
				}

				switch strings.ToUpper(resp.Group) {
				case "CLIENT":
					fmt.Println("successfully connected")
					err := json.Unmarshal([]byte(resp.Message), c)
					if err != nil {
						fmt.Printf("unmarshal message: %v", err)
					}
				case "GAME":
					switch strings.ToUpper(resp.Message) {
					case "UPDATE":
						fmt.Println("updating game data")
						c.Game = resp.GameData
					}
				case "ERROR":
					fmt.Printf("response error message: %v", resp.Message)
				case "FATAL":
					fmt.Printf("response fatal message: %v", resp.Message)
					c.Socket.Close()
					break
				default:
					fmt.Println("received: " + string(msg))
				}
			}
		}
	}
}

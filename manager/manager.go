package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/midgarco/dice_game/game"
	"github.com/midgarco/dice_game_socket/client"
	"github.com/midgarco/dice_game_socket/io"
)

var games uint64

type ClientManager struct {
	clients    map[*client.Client]bool
	broadcast  chan []byte
	register   chan *client.Client
	unregister chan *client.Client
	games      map[uint64]*game.Game
}

func NewClientManager() ClientManager {
	return ClientManager{
		clients:    make(map[*client.Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *client.Client),
		unregister: make(chan *client.Client),
		games:      make(map[uint64]*game.Game),
	}
}

func (cm *ClientManager) Register(c *client.Client) {
	cm.register <- c
}

func (cm *ClientManager) Start() {
	for {
		select {
		case conn := <-cm.register:
			cm.clients[conn] = true
			fmt.Println("added new connection")
		case conn := <-cm.unregister:
			if _, ok := cm.clients[conn]; ok {
				close(conn.Data)
				delete(cm.clients, conn)
				fmt.Println("a connection has terminated")
			}
		case msg := <-cm.broadcast:
			for conn := range cm.clients {
				select {
				case conn.Data <- msg:
				default:
					close(conn.Data)
					delete(cm.clients, conn)
				}
			}
		}
	}
}

func (cm *ClientManager) Receive(c *client.Client) {
	for {
		msg := make([]byte, 4096)
		length, err := c.Socket.Read(msg)
		if err != nil {
			cm.unregister <- c
			c.Socket.Close()
			break
		}
		if length > 0 {
			msg = bytes.Trim(msg, "\x00")
			fmt.Println(string(msg))

			req := &io.Request{}
			err = json.Unmarshal(msg, req)
			if err != nil {
				c.HandleError(err)
				continue
			}

			switch req.Command {
			case "GAME":
				if len(req.Action) == 0 {
					c.HandleError(fmt.Errorf("GAME missing action"))
				}

				switch req.Action {
				case "NEW":
					// create a new game
					atomic.AddUint64(&games, 1)
					gid := atomic.LoadUint64(&games)
					cm.games[gid] = &game.Game{ID: gid, MaxScore: 10000, OpenScore: 650}

					resp := &io.Response{Message: "UPDATE", GameData: cm.games[gid]}
					c.SendResponse("GAME", resp)
					c.SendRequest(&io.Request{Command: "PROMPT", Action: "PLAYERS", Data: "How many players?"})
				default:
					c.HandleError(fmt.Errorf("GAME unknown command %s", req.Action))
				}
			default:
				fmt.Println("received: " + string(msg))
				cm.broadcast <- msg
			}
		}
	}
}

func (cm *ClientManager) Send(c *client.Client) {
	defer c.Socket.Close()
	for {
		select {
		case msg, ok := <-c.Data:
			if !ok {
				return
			}
			c.Socket.Write(msg)
		}
	}
}

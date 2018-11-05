package io

import (
	"github.com/midgarco/dice_game/game"
)

type Response struct {
	Group    string     `json:"group"`
	Message  string     `json:"msg"`
	Error    bool       `json:"error"`
	GameData *game.Game `json:"game"`
}

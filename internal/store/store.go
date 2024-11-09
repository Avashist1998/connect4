package store

import "4connect/internal/game"

var datastore = make(map[string]*game.Game)

func GetDataStore() map[string]*game.Game {
	return datastore
}

package store

import (
	"4connect/internal/models"
)

var datastore = make(map[string]*models.Match)

func GetDataStore() map[string]*models.Match {
	return datastore
}

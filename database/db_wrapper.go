package database

import (
	"github.com/bluelamar/pinoy/config"
)

// DBInterface defines an interface that is implemented by underlying DB
type DBInterface interface {
	Init(cfg *config.PinoyConfig) error
	Close(cfg *config.PinoyConfig) error
	Create(entity, key string, val interface{}) (*map[string]interface{}, error)
	Read(entity, id string) (*map[string]interface{}, error)
	ReadAll(entity string) ([]interface{}, error)
	Update(entity, id, rev string, val map[string]interface{}) (string, error)
	Delete(entity, id, rev string) error
	Find(entity, field string, value interface{}) ([]interface{}, error)

	DeleteM(entity string, rMap *map[string]interface{}) error
	UpdateM(entity, key string, rMap *map[string]interface{}) error
}

var dbInt DBInterface

// GetDB returns the DB handle set for the application
func GetDB() DBInterface {
	return dbInt
}

// SetDB will set or replace the DBInterface used by the wrapper
func SetDB(dbi DBInterface) DBInterface {
	dbiOld := dbInt
	dbInt = dbi
	return dbiOld
}

// DbwCreate creates an entry in the given entity with given key and values
func DbwCreate(entity, key string, val interface{}) (*map[string]interface{}, error) {
	return (dbInt).Create(entity, key, val)
}

// DbwInit initializes the underlying the DB
func DbwInit(cfg *config.PinoyConfig) error {
	return (dbInt).Init(cfg)
}

// DbwDelete deletes an entry in the given entity where the map contains the key
func DbwDelete(entity string, rMap *map[string]interface{}) error {
	return (dbInt).DeleteM(entity, rMap)
}

// DbwUpdate updates a matching entry for the given entity and key
func DbwUpdate(entity, key string, rMap *map[string]interface{}) error {
	return (dbInt).UpdateM(entity, key, rMap)
}

// DbwRead will read and entry from the entity with given id
func DbwRead(entity, id string) (*map[string]interface{}, error) {
	return (dbInt).Read(entity, id)
}

// DbwReadAll reads all entries for the given entity
func DbwReadAll(entity string) ([]interface{}, error) {
	return (dbInt).ReadAll(entity)
}

// DbwFind finds a list of entities with the given field containing the given value
func DbwFind(entity, field string, value interface{}) ([]interface{}, error) {
	return (dbInt).Find(entity, field, value)
}

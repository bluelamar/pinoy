package database

import (
	"github.com/bluelamar/pinoy/config"
)

type DBInterface interface {
	Init(cfg *config.PinoyConfig) error
	Close(cfg *config.PinoyConfig) error
	Create(entity, key string, val interface{}) (*map[string]interface{}, error)
	Read(entity, id string) (*map[string]interface{}, error)
	ReadAll(entity string) ([]interface{}, error)
	Update(entity, id, rev string, val map[string]interface{}) (string, error)
	Delete(entity, id, rev string) error
	Find(entity, field, value string) ([]interface{}, error)

	DeleteM(entity string, rMap *map[string]interface{}) error
	UpdateM(entity, key string, rMap *map[string]interface{}) error
}

var dbInt DBInterface

func GetDB() DBInterface {
	return dbInt
}

// SetDB will set or replace the DBInterface used by the wrapper
func SetDB(dbi DBInterface) DBInterface {
	dbiOld := dbInt
	dbInt = dbi
	return dbiOld
}

func DbwCreate(entity, key string, val interface{}) (*map[string]interface{}, error) {
	return (dbInt).Create(entity, key, val)
}
func DbwInit(cfg *config.PinoyConfig) error {
	return (dbInt).Init(cfg)
}
func DbwDelete(entity string, rMap *map[string]interface{}) error {
	return (dbInt).DeleteM(entity, rMap)
}
func DbwUpdate(entity, key string, rMap *map[string]interface{}) error {
	return (dbInt).UpdateM(entity, key, rMap)
}
func DbwRead(entity, id string) (*map[string]interface{}, error) {
	return (dbInt).Read(entity, id)
}
func DbwReadAll(entity string) ([]interface{}, error) {
	return (dbInt).ReadAll(entity)
}
func DbwFind(entity, field, value string) ([]interface{}, error) {
	return (dbInt).Find(entity, field, value)
}


package database

import (
	"github.com/bluelamar/pinoy/config"
)

type DBInterface interface {
	Init(cfg *config.PinoyConfig) error
	Create(entity, key string, val interface{}) (*map[string]interface{}, error)
	Read(entity, id string) (*map[string]interface{}, error)
	ReadAll(entity string) ([]interface{}, error)
	//Update(entity, id, rev string, val map[string]interface{}) (string, error)
	//Delete(entity, id, rev string) error
	Find(entity, field, value string) ([]interface{}, error)

	DbwDelete(entity string, rMap *map[string]interface{}) error
	DbwUpdate(entity, key string, rMap *map[string]interface{}) error
}

var dbInt *DBInterface

func GetDB() DBInterface {
	return *dbInt
}
func SetDB(dbi *DBInterface) {
	dbInt = dbi
}

func Init(dbi *DBInterface, cfg *config.PinoyConfig) error {
	return (*dbi).Init(cfg)
}
func Create(dbi DBInterface, entity, key string, val interface{}) (*map[string]interface{}, error) {
	return dbi.Create(entity, key, val)
}
func ReadAll(dbi DBInterface, entity string) ([]interface{}, error) {
	return dbi.ReadAll(entity)
}
func Read(dbi DBInterface, entity, id string) (*map[string]interface{}, error) {
	return dbi.Read(entity, id)
}
func Find(dbi DBInterface, entity, field, value string) ([]interface{}, error) {
	return dbi.Find(entity, field, value)
}

func DbwDelete(entity string, rMap *map[string]interface{}) error {
	return (*dbInt).DbwDelete(entity, rMap)
}
func DbwUpdate(entity, key string, rMap *map[string]interface{}) error {
	return (*dbInt).DbwUpdate(entity, key, rMap)
}
func DbwRead(entity, id string) (*map[string]interface{}, error) {
	return (*dbInt).Read(entity, id)
}
func DbwReadAll(entity string) ([]interface{}, error) {
	return (*dbInt).ReadAll(entity)
}
func DbwFind(entity, field, value string) ([]interface{}, error) {
	return (*dbInt).Find(entity, field, value)
}

/* FIX
func (pDb *DBInterface) DbwDelete(entity string, rMap *map[string]interface{}) error {
	id := (*rMap)["_id"].(string)
	rev := (*rMap)["_rev"].(string)
	return pDb.Delete(entity, id, rev)
} */

/*
 * Determines to update if _id is present and key is empty, else create entry
 */
/* FIX
func (pDb *DBInterface) DbwUpdate(entity, key string, rMap *map[string]interface{}) error {
	var err error
	id, ok := (*rMap)["_id"]
	if ok && key == "" {
		_, err = pDb.Update(entity, id.(string), (*rMap)["_rev"].(string), (*rMap))
	} else {
		_, err = pDb.Create(entity, key, (*rMap))
	}

	return err
} */

package database

import (
	"errors"

	"github.com/bluelamar/pinoy/config"
)

type MDBInterface struct {
	
}

func NewMDatabase() MDBInterface {
        pDbInt := MDBInterface{}
        return pDbInt
}

func (pDbInt *MDBInterface) Init(cfg *config.PinoyConfig) error {
	return errors.New("mongo db init not implemented")
}

func (pDbInt *MDBInterface) Create(entity, key string, val interface{}) (*map[string]interface{}, error) {
	return nil, errors.New("mongo db Create not implemented")
}

func (pDbInt *MDBInterface) Read(entity, id string) (*map[string]interface{}, error) {
	return nil, errors.New("mongo db Read not implemented")
}

func (pDbInt *MDBInterface) ReadAll(entity string) ([]interface{}, error) {
	return nil, errors.New("mongo db ReadAll not implemented")
}

func (pDbInt *MDBInterface) Update(entity, id, rev string, val map[string]interface{}) (string, error) {
	return "", errors.New("mongo db Update not implemented")
}

func (pDbInt *MDBInterface) Delete(entity, id, rev string) error {
	return errors.New("mongo db Delete not implemented")
}

func (pDbInt *MDBInterface) Find(entity, field, value string) ([]interface{}, error) {
	return nil, errors.New("mongo db Find not implemented")
}


func (pDbInt *MDBInterface) DeleteM(entity string, rMap *map[string]interface{}) error {
	return errors.New("mongo db DeleteM not implemented")
}

func (pDbInt *MDBInterface) UpdateM(entity, key string, rMap *map[string]interface{}) error {
	return errors.New("mongo db UpdateM not implemented")
}



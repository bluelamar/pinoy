package main

import couchdb "github.com/leesper/couchdb-golang"

type DBInterface struct {
	dbimpl *couchdb.Server
}

func NewDatabase(url, port, dbName, dbUser, dbPwd string) (*DBInterface, error) {
	svr, err := couchdb.NewServer("http://root:password@localhost:5984/")
	if err != nil {
		// TODO failed getting db log error
		return nil, err
	}
	var dbint DBInterface
	dbint.dbimpl = svr
	return &dbint, nil
}

func (*DBInterface) write(key string, val interface{}) error {
	// TODO write
	return nil
}

func (*DBInterface) read(key string) (interface{}, error) {
	// TODO read
	return nil, nil
}

func (*DBInterface) find() ([]interface{}, error) {
	// TODO find
	return nil, nil
}

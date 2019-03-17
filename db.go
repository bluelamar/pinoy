package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	// FIX couchdb "github.com/leesper/couchdb-golang"
)

type DBInterface struct {
	// FIX dbimpl *couchdb.Server
	baseUrl string
	cookies []*http.Cookie
}

func NewDatabase(cfg *PinoyConfig) (*DBInterface, error) {
	pwd, err := cfg.DecryptDbPwd()
	if err != nil {
		return nil, err
	}
	// FIX TODO build the string for the server: ex: http://localhost:5984/_session
	port := strconv.Itoa(cfg.DbPort)
	url := cfg.DbUrl + ":" + port + "/"
	var dbint DBInterface
	dbint.baseUrl = url
	url = dbint.baseUrl + "_session"
	loginCreds := "name=" + cfg.DbUser + "&password=" + pwd
	var payLoad = []byte(loginCreds)
	//`{"title":"Buy cheese and bread for breakfast."}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payLoad))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	// FIX svr, err := couchdb.NewServer("http://root:password@localhost:5984/")
	// curl -c cdbcookies -H "Accept: application/json" -H "Content-Type: application/x-www-form-urlencoded"  http://localhost:5984/_session -X POST -d "name=wsruler&password=oneringtorule"
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	cookies := resp.Cookies()
	dbint.cookies = cookies

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	//var dbint DBInterface
	//dbint.dbimpl = svr
	return &dbint, nil
}

func (*DBInterface) create(entity string, val interface{}) error {
	// TODO create: POST -H "Content-Type: application/json" -H "Accept: application/json" http://localhost:8080/v1/link/${ENTITY} -d "${DATA}"
	return nil
}

func (*DBInterface) read(entity string) (interface{}, error) {
	// TODO read: -H "Content-Type: application/json" -H "Accept: application/json" http://localhost:8080/v1/link/${ENTITY}/${DBLINK2}
	return nil, nil
}

func (*DBInterface) update(entity string, val interface{}) error {
	// TODO update: PUT -H "Content-Type: application/json" -H "Accept: application/json" http://localhost:8080/v1/${ENTITY}/${ENV1} -d "${DATA}
	return nil
}

func (*DBInterface) delete(key, entity string) error {
	// TODO delete: DELETE -H "Accept: application/json" http://localhost:8080/v1/${ENTITY}/${ID}
	return nil
}

func (*DBInterface) find() ([]interface{}, error) {
	// TODO find
	return nil, nil
}

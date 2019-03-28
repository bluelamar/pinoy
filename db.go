package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"time"
	//"golang.org/x/net/publicsuffix"
	// FIX couchdb "github.com/leesper/couchdb-golang"
)

type DBInterface struct {
	// FIX dbimpl *couchdb.Server
	baseUrl string
	// cookies []*http.Cookie
	//authCookie *http.Header
	client *http.Client
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

	timeout := time.Duration(time.Duration(cfg.Timeout) * time.Second)
	jar, err := cookiejar.New(nil) // &cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Jar:     jar,
		Timeout: timeout,
	}

	dbint.client = client

	resp, err := client.Do(req)
	// FIX svr, err := couchdb.NewServer("http://root:password@localhost:5984/")
	// curl -c cdbcookies -H "Accept: application/json" -H "Content-Type: application/x-www-form-urlencoded"  http://localhost:5984/_session -X POST -d "name=wsruler&password=oneringtorule"
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	/* FIX
	cookies := resp.Cookies()
	//dbint.cookies = cookies // []*http.Cookie
	//cookie := cookies.Get("Set-Cookie")
	if cookies != nil {
		for _, cookie := range cookies {
			if cookie != nil && cookie.Name == "Set-Cookie" {
				// HAVE: the auth session cookie
				token := cookie.Value
				//token := strings.Split(tokenPart, "=")[1]
				cookieAuthHeader := &http.Header{}
				cookieAuthHeader.Add("Cookie", fmt.Sprintf("AuthSession=%s", token))
				dbint.authCookie = cookieAuthHeader
				break
			}
		}
	} */

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	//var dbint DBInterface
	//dbint.dbimpl = svr
	return &dbint, nil
}

func (dbi *DBInterface) create(entity string, val interface{}) (*map[string]interface{}, error) {
	// create: POST -H "Content-Type: application/json" -H "Accept: application/json" http://localhost:8080/v1/link/${ENTITY} -d "${DATA}"
	url := dbi.baseUrl + entity
	bytesRepresentation, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return nil, err
	}

	resp, err := dbi.client.Do(request)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println("FIX create: ", result)

	return &result, nil
}

func (dbi *DBInterface) read(entity string) (*map[string]interface{}, error) {
	// read: -H "Content-Type: application/json" -H "Accept: application/json" http://localhost:8080/v1/link/${ENTITY}/${DBLINK2}
	url := dbi.baseUrl + entity
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := dbi.client.Do(request)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println("FIX read: ", result)

	return &result, nil
}

func (dbi *DBInterface) update(entity string, val interface{}) error {
	// update: PUT -H "Content-Type: application/json" -H "Accept: application/json" http://localhost:8080/v1/${ENTITY}/${ENV1} -d "${DATA}
	url := dbi.baseUrl + entity
	bytesRepresentation, err := json.Marshal(val)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return err
	}

	resp, err := dbi.client.Do(request)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println("FIX put: ", result)
	// FIX TODO get the status from the result
	return err
}

func (dbi *DBInterface) delete(key, entity string) error {
	// TODO delete: DELETE -H "Accept: application/json" http://localhost:8080/v1/${ENTITY}/${ID}
	// curl -H "Content-Type: application/json" http://localhost:5984/dirsvc_links/hawaii?rev="1-4b0b6f6ea78ae4b11632f2640d3a89cb" -X DELETE
	return nil
}

func (dbi *DBInterface) find() ([]interface{}, error) {
	// TODO find
	/* FIX
	String selector = "{\"selector\":{\"" + field  + "\":{\"$eq\":\"" + value + "\"}}}";

		Map<String,Object> entity = post(path+"/_find", selector, null);
		if (entity != null) {
			Object retObj = entity.get("docs");
			return (List<Object>)retObj;
		} */
	return nil, nil
}

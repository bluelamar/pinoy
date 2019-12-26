package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"time"

	"github.com/bluelamar/pinoy/config"
	//"golang.org/x/net/publicsuffix"
	// couchdb "github.com/leesper/couchdb-golang"
)

type CDBInterface struct {
	baseUrl string
	// cookies []*http.Cookie
	//authCookie *http.Header
	client *http.Client
}

func NewCDatabase() CDBInterface {
	pDbInt := CDBInterface{}
	return pDbInt
}
func (pDbInt *CDBInterface) Init(cfg *config.PinoyConfig) error {
	pwd, err := cfg.DecryptDbPwd()
	if err != nil {
		return err
	}
	// build the string for the server: ex: http://localhost:5984/_session
	port := strconv.Itoa(cfg.DbPort)
	url := cfg.DbUrl + ":" + port + "/"
	//var dbint CDBInterface
	pDbInt.baseUrl = url
	url = pDbInt.baseUrl + "_session"
	loginCreds := "name=" + cfg.DbUser + "&password=" + pwd
	var payLoad = []byte(loginCreds)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payLoad))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	timeout := time.Duration(time.Duration(cfg.DbCommTimeout) * time.Second)
	jar, err := cookiejar.New(nil) // &cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Jar:     jar,
		Timeout: timeout,
	}

	pDbInt.client = client

	resp, err := client.Do(req)
	// curl -c cdbcookies -H "Accept: application/json" -H "Content-Type: application/x-www-form-urlencoded"  http://localhost:5984/_session -X POST -d "name=wsruler&password=oneringtorule"
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ioutil.ReadAll(resp.Body)

	return nil
}

func (pDbInt *CDBInterface) Create(entity, key string, val interface{}) (*map[string]interface{}, error) {
	// create: POST -H "Content-Type: application/json" -H "Accept: application/json" http://localhost:8080/v1/link/${ENTITY} -d "${DATA}"
	url := pDbInt.baseUrl + entity
	valMap := val.(map[string]interface{})
	if key != "" {
		valMap["_id"] = key
	}

	bytesRepresentation, err := json.Marshal(valMap)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", url, bytes.NewReader(bytesRepresentation))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := pDbInt.client.Do(request)
	if err != nil {
		return nil, err
	}
	//defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	err = checkResultError(&result, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (pDbInt *CDBInterface) Read(entity, id string) (*map[string]interface{}, error) {
	// curl -v --cookie "cdbcookies" http://localhost:5984/dblnk/19b74cd4
	url := pDbInt.baseUrl + entity + "/" + id
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := pDbInt.client.Do(request)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	err = checkResultError(&result, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (pDbInt *CDBInterface) ReadAll(entity string) ([]interface{}, error) {
	// curl -v --cookie "cdbcookies" http://localhost:5984/testxyz/_all_docs
	url := pDbInt.baseUrl + entity + "/_all_docs?include_docs=true"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := pDbInt.client.Do(request)
	if err != nil {
		return nil, err
	}

	/*
	   ex output:
	   map[offset:0 rows:[map[id:70c54a580c9af21c0e698fdf2a003691 key:70c54a580c9af21c0e698fdf2a003691 value:map[rev:1-e1e73b2d88ada8d8f636cb13f2c06b71]]
	   ...

	   Create slice of id's to return
	*/
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	err = checkResultError(&result, err)
	if err != nil {
		return nil, err
	}

	var rows []interface{}
	rows = result["rows"].([]interface{})

	docs := make([]interface{}, len(rows))
	for k, v := range rows {
		vmap := v.(map[string]interface{})
		docs[k] = vmap["doc"]
	}
	return docs, nil
}

// return the new revision
func (pDbInt *CDBInterface) Update(entity, id, rev string, val map[string]interface{}) (string, error) {
	// curl --cookie "cdbcookies" -H "Content-Type: application/json" http://localhost:5984/stuff/592ccd646f8202691a77f1b1c5004496 -X PUT -d '{"name":"sam","age":42,"_rev":"1-3f12b5828db45fda239607bf7785619a"}'
	url := pDbInt.baseUrl + entity + "/" + id
	val["_rev"] = rev
	bytesRepresentation, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return "", err
	}

	resp, err := pDbInt.client.Do(request)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	err = checkResultError(&result, err)
	if err != nil {
		return "", err
	}
	// ex: map[ok:true id:3d_shapes rev:28-d2bc68f0f0132cbb483ee1196e3c482e]
	rev = result["rev"].(string)
	return rev, err
}

func (pDbInt *CDBInterface) Delete(entity, id, rev string) error {
	// curl -v --cookie "cdbcookies" http://localhost:5984/testxyz/f00dc0ba83aec8f560bd7c8036000c0a?rev=1-e1e73b2d88ada8d8f636cb13f2c06b71 -X DELETE

	url := pDbInt.baseUrl + entity + "/" + id + "?rev=" + rev
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := pDbInt.client.Do(request)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return checkResultError(&result, err)
}

func (pDbInt *CDBInterface) Find(entity, field, value string) ([]interface{}, error) {
	// curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/testxyz/_find -X POST -d $SEL
	// SEL='{"selector":{"shape":{"$eq":"pyramid"}}}'
	/*
		String selector = "{\"selector\":{\"" + field  + "\":{\"$eq\":\"" + value + "\"}}}";

			Map<String,Object> entity = post(path+"/_find", selector, null);
			if (entity != null) {
				Object retObj = entity.get("docs");
				return (List<Object>)retObj;
			} */
	entity = entity + "/_find"
	//val := `{"selector":{"` + field + `":{"$eq":"` + value + `"}}}`
	eqm := map[string]string{"$eq": value} // make(map[string]interface{})
	//eqm["$eq"] = value
	fldm := map[string]interface{}{field: eqm}
	//fldm[field] = eqm
	sel := map[string]interface{}{"selector": fldm} // make(map[string]interface{})
	//sel["selector"] = fldm
	url := pDbInt.baseUrl + entity
	bytesRepresentation, err := json.Marshal(sel)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", url, bytes.NewReader(bytesRepresentation))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := pDbInt.client.Do(request)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	err = checkResultError(&result, err)
	if err != nil {
		return nil, err
	}
	return result["docs"].([]interface{}), nil
}

func (pDbInt *CDBInterface) DeleteM(entity string, rMap *map[string]interface{}) error {
	id, ok := (*rMap)["_id"]
	if !ok {
		return errors.New("cdb:missing required id")
	}
	rev, ok := (*rMap)["_rev"]
	if !ok {
		return errors.New("cdb:missing required rev")
	}
	return pDbInt.Delete(entity, id.(string), rev.(string))
}

/*
 * Determines to update if _id is present and key is empty, else create entry
 */
func (pDbInt *CDBInterface) UpdateM(entity, key string, rMap *map[string]interface{}) error {
	var err error
	id, ok := (*rMap)["_id"]
	if ok && key == "" {
		_, err = pDbInt.Update(entity, id.(string), (*rMap)["_rev"].(string), (*rMap))
	} else {
		// remove _rev if it exists
		delete((*rMap), "_rev")
		_, err = pDbInt.Create(entity, key, (*rMap))
	}

	return err
}

func checkResultError(result *map[string]interface{}, err error) error {
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	msg, ok := (*result)["error"]
	if ok {
		// CouchDb returned an error msg
		// log.Println("db:check: result contains error=", msg)
		return errors.New(msg.(string))
	}
	return nil
}

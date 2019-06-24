package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	DBInterface
	// FIX dbimpl *couchdb.Server
	baseUrl string
	// cookies []*http.Cookie
	//authCookie *http.Header
	client *http.Client
}

func NewCDatabase() CDBInterface {
	dbint := CDBInterface{}
	return dbint
}
func (dbint *CDBInterface) Init(cfg *config.PinoyConfig) error {
	pwd, err := cfg.DecryptDbPwd()
	if err != nil {
		return err
	}
	// build the string for the server: ex: http://localhost:5984/_session
	port := strconv.Itoa(cfg.DbPort)
	url := cfg.DbUrl + ":" + port + "/"
	//var dbint CDBInterface
	dbint.baseUrl = url
	url = dbint.baseUrl + "_session"
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

	dbint.client = client

	resp, err := client.Do(req)
	// FIX svr, err := couchdb.NewServer("http://root:password@localhost:5984/")
	// curl -c cdbcookies -H "Accept: application/json" -H "Content-Type: application/x-www-form-urlencoded"  http://localhost:5984/_session -X POST -d "name=wsruler&password=oneringtorule"
	if err != nil {
		return err
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

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("db.init:FIX: response Body:", string(body))

	//var dbint DBInterface
	//dbint.dbimpl = svr
	return nil
}

func (dbi *CDBInterface) Create(entity, key string, val interface{}) (*map[string]interface{}, error) {
	// create: POST -H "Content-Type: application/json" -H "Accept: application/json" http://localhost:8080/v1/link/${ENTITY} -d "${DATA}"
	url := dbi.baseUrl + entity
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

	resp, err := dbi.client.Do(request)
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
	fmt.Println("FIX db.create: ", result)

	return &result, nil
}

func (dbi *CDBInterface) Read(entity, id string) (*map[string]interface{}, error) {
	// curl -v --cookie "cdbcookies" http://localhost:5984/dblnk/19b74cd4
	url := dbi.baseUrl + entity + "/" + id
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := dbi.client.Do(request)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	err = checkResultError(&result, err)
	if err != nil {
		return nil, err
	}
	//log.Println("FIX read: ", result)

	return &result, nil
}

//func (dbi *DBInterface) ReadAll(entity string) (*map[string]interface{}, error) {
//func (dbi *DBInterface) ReadAll(entity string) ([]string, error) {
func (dbi *CDBInterface) ReadAll(entity string) ([]interface{}, error) {
	// curl -v --cookie "cdbcookies" http://localhost:5984/testxyz/_all_docs
	url := dbi.baseUrl + entity + "/_all_docs?include_docs=true"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := dbi.client.Do(request)
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

	//log.Println("FIX readall result: ", result)
	var rows []interface{}
	rows = result["rows"].([]interface{})
	//log.Println("FIX readall rows: ", rows)
	//return rows, nil

	//ids := make([]string, len(rows))
	docs := make([]interface{}, len(rows))
	for k, v := range rows {
		vmap := v.(map[string]interface{})
		docs[k] = vmap["doc"]
		//ids[k] = vmap["id"].(string)
	}
	return docs, nil
	//log.Println("FIX readall rows: ", result["rows"])
	//log.Println("FIX readall ids: ", ids)

	//return ids, nil
	//return &result, nil
}

// return the new revision
func (dbi *CDBInterface) Update(entity, id, rev string, val map[string]interface{}) (string, error) {
	// curl --cookie "cdbcookies" -H "Content-Type: application/json" http://localhost:5984/stuff/592ccd646f8202691a77f1b1c5004496 -X PUT -d '{"name":"sam","age":42,"_rev":"1-3f12b5828db45fda239607bf7785619a"}'
	url := dbi.baseUrl + entity + "/" + id
	//val["_id"] = id
	val["_rev"] = rev
	bytesRepresentation, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return "", err
	}

	resp, err := dbi.client.Do(request)
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
	fmt.Println("FIX db.put: ", result)
	// FIX TODO get the status from the result
	rev = result["rev"].(string)
	return rev, err
}

func (dbi *CDBInterface) Delete(entity, id, rev string) error {
	// curl -v --cookie "cdbcookies" http://localhost:5984/testxyz/f00dc0ba83aec8f560bd7c8036000c0a?rev=1-e1e73b2d88ada8d8f636cb13f2c06b71 -X DELETE

	url := dbi.baseUrl + entity + "/" + id + "?rev=" + rev
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := dbi.client.Do(request)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	fmt.Println("FIX db.delete: ", result)
	return checkResultError(&result, err)
}

func (dbi *CDBInterface) Find(entity, field, value string) ([]interface{}, error) {
	// TODO find
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
	// TODO change how to make json for val
	//val := `{"selector":{"` + field + `":{"$eq":"` + value + `"}}}`
	eqm := map[string]string{"$eq": value} // make(map[string]interface{})
	//eqm["$eq"] = value
	fldm := map[string]interface{}{field: eqm}
	//fldm[field] = eqm
	sel := map[string]interface{}{"selector": fldm} // make(map[string]interface{})
	//sel["selector"] = fldm
	fmt.Println("FIX Find: entity=", entity, " :val=", sel)
	// FIX var ret *map[string]interface{}
	var err error
	//ret, err = dbi.Create(entity, "", val)
	url := dbi.baseUrl + entity
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

	resp, err := dbi.client.Do(request)
	if err != nil {
		fmt.Println("FIX find: client res=", resp, " : err=", err)
		return nil, err
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	err = checkResultError(&result, err)
	if err != nil {
		return nil, err
	}
	fmt.Println("FIX db.find: res=", result)
	//return (*ret)["docs"].([]interface{}), nil
	return result["docs"].([]interface{}), nil
}

func (pDb *CDBInterface) DbwDelete(entity string, rMap *map[string]interface{}) error {
	id, ok := (*rMap)["_id"]
	if !ok {
		return errors.New("cdb:missing required id")
	}
	rev, ok := (*rMap)["_rev"]
	if !ok {
		return errors.New("cdb:missing required rev")
	}
	return pDb.Delete(entity, id.(string), rev.(string))
}

/*
 * Determines to update if _id is present and key is empty, else create entry
 */
func (pDb *CDBInterface) DbwUpdate(entity, key string, rMap *map[string]interface{}) error {
	var err error
	id, ok := (*rMap)["_id"]
	if ok && key == "" {
		_, err = pDb.Update(entity, id.(string), (*rMap)["_rev"].(string), (*rMap))
	} else {
		_, err = pDb.Create(entity, key, (*rMap))
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
		log.Println("db:check: result contains error=", msg)
		return errors.New(msg.(string))
	}
	return nil
}

package pinoy

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
	//req.Header.Set("Content-Type", "application/json")
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

func (dbi *DBInterface) Create(entity string, val interface{}) (*map[string]interface{}, error) {
	// create: POST -H "Content-Type: application/json" -H "Accept: application/json" http://localhost:8080/v1/link/${ENTITY} -d "${DATA}"
	url := dbi.baseUrl + entity
	bytesRepresentation, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	log.Println("FIX create:body=", bytesRepresentation)
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
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println("FIX create: ", result)

	return &result, nil
}

func (dbi *DBInterface) Read(entity, id string) (*map[string]interface{}, error) {
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
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println("FIX read: ", result)

	return &result, nil
}

//func (dbi *DBInterface) ReadAll(entity string) (*map[string]interface{}, error) {
func (dbi *DBInterface) ReadAll(entity string) ([]string, error) {
	// curl -v --cookie "cdbcookies" http://localhost:5984/testxyz/_all_docs
	url := dbi.baseUrl + entity + "/_all_docs"
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
	if err != nil {
		return nil, err
	}
	//log.Println("FIX readall: ", result)
	var rows []interface{}
	rows = result["rows"].([]interface{})
	ids := make([]string, len(rows))
	for k, v := range rows {
		vmap := v.(map[string]interface{})
		ids[k] = vmap["id"].(string)
	}
	//log.Println("FIX readall rows: ", result["rows"])
	//log.Println("FIX readall ids: ", ids)

	return ids, nil
	//return &result, nil
}

// return the new revision
func (dbi *DBInterface) Update(entity, id, rev string, val map[string]interface{}) (string, error) {
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
	if err != nil {
		return "", err
	}
	// ex: map[ok:true id:3d_shapes rev:28-d2bc68f0f0132cbb483ee1196e3c482e]
	log.Println("FIX put: ", result)
	// FIX TODO get the status from the result
	rev = result["rev"].(string)
	return rev, err
}

func (dbi *DBInterface) Delete(entity, id, rev string) error {
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
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println("FIX delete: ", result)
	// FIX TODO get the status from the result
	return err
}

func (dbi *DBInterface) Find(entity, field, value string) ([]interface{}, error) {
	// TODO find
	// curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/testxyz/_find -X POST -d $SEL
	// SEL='{"selector":{"shape":{"$eq":"pyramid"}}}'
	/* FIX
	String selector = "{\"selector\":{\"" + field  + "\":{\"$eq\":\"" + value + "\"}}}";

		Map<String,Object> entity = post(path+"/_find", selector, null);
		if (entity != null) {
			Object retObj = entity.get("docs");
			return (List<Object>)retObj;
		} */
	entity = entity + "/_find"
	val := `{"selector":{"` + field + `":{"$eq":"` + value + `"}}}`
	var ret *map[string]interface{}
	var err error
	ret, err = dbi.Create(entity, val)
	if err != nil {
		return nil, err
	}
	return (*ret)["docs"].([]interface{}), nil
}

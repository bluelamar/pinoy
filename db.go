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

// curl -v --cookie "cdbcookies" http://localhost:5984/testxyz/_all_docs
// Readall

func (dbi *DBInterface) Update(entity, id, rev string, val map[string]interface{}) error {
	// curl --cookie "cdbcookies" -H "Content-Type: application/json" http://localhost:5984/stuff/592ccd646f8202691a77f1b1c5004496 -X PUT -d '{"name":"sam","age":42,"_rev":"1-3f12b5828db45fda239607bf7785619a"}'
	url := dbi.baseUrl + entity + "/" + id
	//val["_id"] = id
	val["_rev"] = rev
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

func (dbi *DBInterface) Find() ([]interface{}, error) {
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

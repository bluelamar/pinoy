package pinoy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type PinoyConfig struct {
	DbUrl   string
	DbName  string
	DbPort  int
	DbUser  string
	DbPwd   string
	Timeout int
}

func LoadConfig(fpath string) (*PinoyConfig, error) {
	content, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	// var dat map[string]interface{}
	//if err := json.Unmarshal(byt, &dat); err != nil {
	//      panic(err)
	//  }
	// num := dat["num"].(int64)
	// fmt.Println(num)
	//  strs := dat["strs"].([]interface{})
	//  str1 := strs[0].(string)
	var cfg PinoyConfig
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("FIX:loadconfig: %+v", cfg)
	return &cfg, nil
}

package main

import (

	//"fmt"
	"testing"
)

var dbInt *DBInterface
var cfg *PinoyConfig

func TestNewDB(t *testing.T) {
	//newDB := "golang-newdb"
	//server.Create(newDB)
	//defer server.Delete(newDB)
	//dbNew, err := NewDatabase(fmt.Sprintf("%s/%s", DefaultBaseURL, newDB))

	cfg = &PinoyConfig{
		DbUrl:   "http://localhost",
		DbName:  "testxyz",
		DbPort:  5984,
		DbUser:  "ruler",
		DbPwd:   "oneringtorule",
		Timeout: 5,
	}

	pwd, err := cfg.EncryptDbPwd()
	if err != nil {
		t.Error(`TestNewDb: encrypt db pwd error`, err)
	} else {
		t.Logf("TestNewDb got pwd: %q\n", pwd)
	}

	dbInt1, err := NewDatabase(cfg)
	if err != nil {
		t.Error(`TestNewDb: database error`, err)
	}
	dbInt = dbInt1
}

func TestCreate(t *testing.T) {
	doc := map[string]interface{}{"doc": "bar"}
	doc2, err := dbInt.Create("testxyz", "docbar", doc)
	if err != nil {
		t.Error(`db save error`, err)
	}
	//id := (*doc2)["_id"].(string)
	id := ""
	if id2, ok := (*doc2)["id"]; ok {
		id = id2.(string)
		t.Logf("TestCreate got id: %q\n", id)
	} else {
		t.Errorf("TestCreate: missing id")
	}
	rev := ""
	if rv, ok := (*doc2)["rev"]; ok {
		rev = rv.(string)
		t.Logf("TestCreate got rev: %q\n", rev)
	} else {
		t.Errorf("TestCreate: missing rev")
	}

	entity := "testxyz"
	ent_map, err := dbInt.Read(entity, id)
	if err != nil {
		t.Error(`db read error`, err)
	}
	t.Logf("read entity=%s id=%s val=%v\n", entity, id, ent_map)

	ent_map, err = dbInt.Read(entity, "nosuchid")
	if err != nil {
		t.Logf(`db read nosuchid error=%v\n`, err)
	} else {
		t.Logf("read entity=%s id=nosuchid val=%v\n", entity, ent_map)
		errMsg, exists := (*ent_map)["error"]
		if exists {
			t.Logf("read entity=%s id=nosuchid got error=%v\n", entity, errMsg)
		}
	}

	resArray, err := dbInt.ReadAll(entity)
	if err != nil {
		t.Error(`db readall error`, err)
	}
	t.Logf("readall entity=%s val=%v\n", entity, resArray)

	fres, err := dbInt.Find("room_rates", "RateClass", "Small Room")
	if err != nil {
		t.Error(`db find error`, err)
	}
	t.Logf("find entity=%s val=%v\n", entity, fres)

	err = dbInt.Delete(entity, id, rev)
	if err != nil {
		t.Error(`db delete error`, err)
	}

	ent_map, err = dbInt.Read(entity, "3d_shapes")
	if err != nil {
		t.Logf("db read error: %v\n", err)
	} else {
		t.Logf("db read 3d_shaps: %v\n", ent_map)
		rev3, found := (*ent_map)["_rev"].(string)
		if found {
			dbInt.Delete(entity, "3d_shapes", rev3)
		}
	}
	//doc = map[string]interface{}{"_id": "3d_shapes", "shape": "box"}
	doc = map[string]interface{}{"shape": "box"}
	doc2, err = dbInt.Create("testxyz", "3d_shapes", doc)
	if err != nil {
		t.Error(`db save error`, err)
	}
	t.Logf("create doc id=3d_shapes with shape=box: %v\n", doc2)
	id = ""
	if id2, ok := (*doc2)["id"]; ok {
		id = id2.(string)
		t.Logf("TestCreate-2: got id: %q\n", id)
	} else {
		t.Errorf("TestCreate-2: missing id")
	}
	rev = ""
	if rv, ok := (*doc2)["rev"]; ok {
		rev = rv.(string)
		t.Logf("TestCreate-2: got rev: %q\n", rev)
	} else {
		t.Errorf("TestCreate-2: missing rev")
	}
	ent_map, err = dbInt.Read(entity, id)
	if err != nil {
		t.Error(`db read error`, err)
	}
	t.Logf("read entity=%s id=%s v-type=%T val=%v\n", entity, id, ent_map, ent_map)

	var updEntity map[string]interface{}
	updEntity = *ent_map // .(*map[string]interface{})
	updEntity["shape"] = "pyramid"
	rev, err = dbInt.Update(entity, id, rev, updEntity)
	if err != nil {
		t.Error(`db update error`, err)
	}
	t.Logf("update entity=%s id=%s new-rev=%s\n", entity, id, rev)

	ent_map, err = dbInt.Read(entity, id)
	if err != nil {
		t.Error(`db read error`, err)
	}
	t.Logf("read entity=%s id=%s val=%v\n", entity, id, ent_map)

	err = dbInt.Delete(entity, id, rev)
	if err != nil {
		t.Error(`db delete error`, err)
	}
	// try again - should get error
	err = dbInt.Delete(entity, id, rev)
	if err != nil {
		t.Logf(`repeat db delete: get error=%v`, err)
	} else {
		t.Error(`repeated delete should have gotten error`)
	}
}

func TestEncrypt(t *testing.T) {

	hashed := HashIt("xyz")
	t.Logf("test-encrypt: xyz=%s\n", hashed)
}

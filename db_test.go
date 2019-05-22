package pinoy

import (

	//"fmt"
	"testing"

)

var dbInt *DBInterface

func TestNewDB(t *testing.T) {
	//newDB := "golang-newdb"
	//server.Create(newDB)
	//defer server.Delete(newDB)
	//dbNew, err := NewDatabase(fmt.Sprintf("%s/%s", DefaultBaseURL, newDB))


	cfg := PinoyConfig{
		DbUrl   : "http://localhost",
		DbName  : "testxyz",
		DbPort  : 5984,
		DbUser  : "ruler",
		DbPwd   : "oneringtorule",
		Timeout : 5,
	}

	pwd, err := cfg.EncryptDbPwd()
	if err != nil {
		t.Error(`TestNewDb: encrypt db pwd error`, err)
	} else {
		t.Logf("TestNewDb got pwd: %q\n", pwd)
	}

	dbInt1, err := NewDatabase(&cfg)
	if err != nil {
		t.Error(`TestNewDb: database error`, err)
	}
	dbInt = dbInt1
}

func TestCreate(t *testing.T) {
	doc := map[string]interface{}{"doc": "bar"}
	doc2, err := dbInt.Create("testxyz", doc)
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

	err = dbInt.Delete(entity, id, rev)
	if err != nil {
		t.Error(`db delete error`, err)
	}

	doc = map[string]interface{}{"_id":"3d_shapes","shape":"box"}
	doc2, err = dbInt.Create("testxyz", doc)
	if err != nil {
		t.Error(`db save error`, err)
	}
	id = ""
	if id2, ok := (*doc2)["id"]; ok {
		id = id2.(string)
		t.Logf("TestCreate got id: %q\n", id)
	} else {
		t.Errorf("TestCreate: missing id")
	}
	rev = ""
	if rv, ok := (*doc2)["rev"]; ok {
		rev = rv.(string)
		t.Logf("TestCreate got rev: %q\n", rev)
	} else {
		t.Errorf("TestCreate: missing rev")
	}
	ent_map, err = dbInt.Read(entity, id)
	if err != nil {
		t.Error(`db read error`, err)
	}
	t.Logf("read entity=%s id=%s val=%v\n", entity, id, ent_map)

	// TODO test Update

	err = dbInt.Delete(entity, id, rev)
	if err != nil {
		t.Error(`db delete error`, err)
	}
}



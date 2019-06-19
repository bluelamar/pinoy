package main

import (

	//"fmt"
	"fmt"
	"testing"
	"time"
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
		t.Logf("repeat db delete: get error=%v\n", err)
	} else {
		t.Error(`repeated delete should have gotten error`)
	}

	loc, err := time.LoadLocation("Singapore")
	nowStr, nowTime := TimeNow(loc)
	t.Logf("TimeNow returns loc=%v str=%s tn=%v\n", loc, nowStr, nowTime)
	checkOutTime, err := CalcCheckoutTime(nowTime, "3 Hours")
	t.Logf("CalcCheckoutTime returns=(%s) err=%v\n", checkOutTime, err)

	loc = time.FixedZone("UTC-8", -8*60*60)
	nowStr, nowTime = TimeNow(loc)
	t.Logf("TimeNow returns loc=%v str=%s tn=%v\n", loc, nowStr, nowTime)

	loc = time.FixedZone("UTC-8", 8*60*60)
	nowStr, nowTime = TimeNow(loc)
	t.Logf("TimeNow returns loc=%v str=%s tn=%v\n", loc, nowStr, nowTime)

	role := "Manager"
	if role == "Manager" {
		t.Logf("role == Manager is correct\n")
	} else {
		t.Logf("Failed: role == %s\n", role)
	}
	if role != "Manager" {
		t.Logf("Failed: role == %s\n", role)
	} else {
		t.Logf("role == Manager is correct\n")
	}

	role = "Desk"
	if role == "Manager" {
		t.Logf("Failed: role == Manager but should be Desk\n")
	} else {
		t.Logf("Success: role == %s\n", role)
	}
	if role != "Desk" {
		t.Logf("Failed: role == %s\n", role)
	} else {
		t.Logf("role == Desk is correct\n")
	}

	const longForm = "2006-01-02 15:04" // FIX :05"
	// ex clockinTime: 2019-06-11 12:49
	ci := "2019-06-11 12:49"
	clockin, err := time.ParseInLocation(longForm, ci, loc)
	if err != nil {
		t.Logf("ParseInLoc failed :err=%v\n", err)
	} else {
		t.Logf("ParseInLoc works: time=%v\n", clockin)
	}
	h, _ := time.ParseDuration("4h30m")
	hours := h.Hours()
	ihours := int(hours)
	fmt.Printf("Ive got %.1f hours of work left or rounded=%d\n", hours, ihours)
}

func TestEncrypt(t *testing.T) {

	hashed := HashIt("xyz")
	t.Logf("test-encrypt: xyz=%s\n", hashed)
}

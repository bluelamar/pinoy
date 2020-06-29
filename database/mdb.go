package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bluelamar/pinoy/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// An MDBInterface is used in methods that reperesent the mongodb interface
type MDBInterface struct {
	client *mongo.Client
	cfg    *config.PinoyConfig
}

// NewMDatabase returns a new MDBInterface
func NewMDatabase() MDBInterface {
	pDbInt := MDBInterface{}
	return pDbInt
}

// Init will initialized the DB with the given configuration
func (pDbInt *MDBInterface) Init(cfg *config.PinoyConfig) error {
	pwd, err := cfg.DecryptDbPwd()
	if err != nil {
		return err
	}
	// using SCRAM auth
	loginCreds := cfg.DbUser + ":" + pwd + "@"
	port := strconv.Itoa(cfg.DbPort)
	// use the database to auth on ubuntu-18.04 client
	// url := "mongodb://" + loginCreds + cfg.DbUrl + ":" + port + "/pinoy" // ex: mongodb://foo:bar@localhost:27017/pinoy
	url := "mongodb://" + loginCreds + cfg.DbURL + ":" + port // works on mac
	if len(cfg.DbAuthDb) > 0 {
		url = url + "/" + cfg.DbAuthDb
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		return normalizeError(err)
	}
	// ex: cfg.DbCommTimeout == 20
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.DbCommTimeout)*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return normalizeError(err)
	}
	pDbInt.client = client
	pDbInt.cfg = cfg
	return nil
}

// Close will close the database resource
func (pDbInt *MDBInterface) Close(cfg *config.PinoyConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.DbCommTimeout)*time.Second)
	defer cancel()
	return pDbInt.client.Disconnect(ctx)
}

// Create will create an entry with the given value and key
func (pDbInt *MDBInterface) Create(entity, key string, val interface{}) (*map[string]interface{}, error) {
	valMap := val.(map[string]interface{})
	if _, ok := valMap["key"]; !ok {
		valMap["key"] = key
	}

	ctx, cf := context.WithTimeout(context.Background(), time.Duration(pDbInt.cfg.DbCommTimeout)*time.Second)
	defer cf()
	coll := pDbInt.client.Database(pDbInt.cfg.DbName).Collection(entity)
	res, err := coll.InsertOne(ctx, valMap)
	if err != nil {
		return nil, normalizeError(err)
	}

	result := make(map[string]interface{})
	result["_id"] = res.InsertedID
	result["key"] = key

	return &result, nil
}

// Read will read the entry for the given entity keyed by id
func (pDbInt *MDBInterface) Read(entity, id string) (*map[string]interface{}, error) {

	ctx, cf := context.WithTimeout(context.Background(), time.Duration(pDbInt.cfg.DbCommTimeout)*time.Second)
	defer cf()
	coll := pDbInt.client.Database(pDbInt.cfg.DbName).Collection(entity)
	if coll == nil {
		return nil, errors.New("failed to find entity=" + entity)
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: 1}})
	sr := coll.FindOne(ctx, bson.D{{Key: "key", Value: id}}, opts)
	if sr == nil {
		return nil, errors.New("failed to find id")
	}

	result := make(map[string]interface{})
	err := sr.Decode(&result)
	if err != nil {
		return &result, normalizeError(err)
	}

	for key, value := range result {
		v := convertToNative(value)
		result[key] = v
	}

	return &result, nil
}

// ReadAll reads all entries of the given entry
func (pDbInt *MDBInterface) ReadAll(entity string) ([]interface{}, error) {
	return pDbInt.Find(entity, "", "")
}

// Update the entity with key in the map
func (pDbInt *MDBInterface) Update(entity, id, rev string, val map[string]interface{}) (string, error) {

	var filter bson.D
	if id == "" {
		if k, ok := val["key"].(string); ok {
			filter = bson.D{{Key: "key", Value: k}}
		} else if oid, ok := val["_id"].(primitive.ObjectID); ok {
			filter = bson.D{{Key: "_id", Value: oid}}
		} else {
			return "", errors.New("missing key field")
		}
	} else {
		filter = bson.D{{Key: "key", Value: id}}
	}
	update := bson.D{{Key: "$set", Value: val}}
	ctx, cf := context.WithTimeout(context.Background(), time.Duration(pDbInt.cfg.DbCommTimeout)*time.Second)
	defer cf()
	coll := pDbInt.client.Database(pDbInt.cfg.DbName).Collection(entity)
	opts := options.Update().SetUpsert(false)
	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return "", normalizeError(err)
	}
	if result.MatchedCount == 0 {
		fmt.Println("matched and replaced an existing document")
		return "", errors.New("no match found for entity=" + entity + " id=" + id)
	}
	return "", nil
}

// Delete the entity keyed by the given id
func (pDbInt *MDBInterface) Delete(entity, id, rev string) error {

	opts := options.Delete().SetCollation(&options.Collation{
		Locale:    "en_US",
		Strength:  1,
		CaseLevel: false,
	})

	ctx, cf := context.WithTimeout(context.Background(), time.Duration(pDbInt.cfg.DbCommTimeout)*time.Second)
	defer cf()
	coll := pDbInt.client.Database(pDbInt.cfg.DbName).Collection(entity)
	res, err := coll.DeleteOne(ctx, bson.D{{Key: "key", Value: id}}, opts)
	if err != nil {
		return err
	}

	if res.DeletedCount == 1 {
		return nil
	}
	return errors.New("failed to delete entity=" + entity + " id=" + id)
}

// Find the list of entities matching the field with the given value
func (pDbInt *MDBInterface) Find(entity, field string, value interface{}) ([]interface{}, error) {

	ctx, cf := context.WithTimeout(context.Background(), time.Duration(pDbInt.cfg.DbCommTimeout)*time.Second)
	defer cf()
	coll := pDbInt.client.Database(pDbInt.cfg.DbName).Collection(entity)

	var err error
	var cursor *mongo.Cursor
	if field == "" {
		cursor, err = coll.Find(ctx, bson.M{})
	} else {
		cursor, err = coll.Find(ctx, bson.D{{Key: field, Value: value}})
	}
	if err != nil {
		return nil, normalizeError(err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, normalizeError(err)
	}

	docs := make([]interface{}, 0)
	for _, result := range results {
		res := make(map[string]interface{})
		// must replace fields that are primitive.A with []interface{}
		respm := (primitive.M)(result)
		for key, value := range respm {
			v := convertToNative(value)
			res[key] = v
		}
		docs = append(docs, res)
	}

	return docs, nil
}

func convertToNative(value interface{}) interface{} {

	if v, ok := value.(primitive.A); ok {
		// convert into generic array
		pa := make([]interface{}, len(v))
		for i, pav := range v {
			pa[i] = convertToNative(pav)
		}
		return pa
	} else if mv, ok := value.(primitive.M); ok {
		pm := make(map[string]interface{})
		for k, v := range mv {
			pm[k] = convertToNative(v)
		}
		return pm
	}

	return value
}

// DeleteM deletes the entity keyed by the key value in the map
func (pDbInt *MDBInterface) DeleteM(entity string, rMap *map[string]interface{}) error {

	if key, ok := (*rMap)["key"].(string); ok {
		return pDbInt.Delete(entity, key, "")
	}
	if id, ok := (*rMap)["_id"]; ok {
		opts := options.Delete().SetCollation(&options.Collation{
			Locale:    "en_US",
			Strength:  1,
			CaseLevel: false,
		})
		ctx, cf := context.WithTimeout(context.Background(), time.Duration(pDbInt.cfg.DbCommTimeout)*time.Second)
		defer cf()
		coll := pDbInt.client.Database(pDbInt.cfg.DbName).Collection(entity)
		res, err := coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}}, opts)
		if err != nil {
			return err
		}

		if res.DeletedCount == 1 {
			return nil
		}
		return errors.New("failed to delete entity=" + entity + " using id")
	}

	return errors.New("missing required key")
}

// UpdateM updates the entity with the values from the map
func (pDbInt *MDBInterface) UpdateM(entity, key string, rMap *map[string]interface{}) error {

	var err error
	_, ok := (*rMap)["_id"]
	if ok && key == "" {
		_, err = pDbInt.Update(entity, "", "", (*rMap))
	} else {
		_, err = pDbInt.Create(entity, key, (*rMap))
		/*
			if err.Error().Contains("duplicate key error") {
				// this should be update and not create
				_, err = pDbInt.Update(entity, "", "", (*rMap))
			} */
	}

	return err
}

func normalizeError(err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "no documents in result") {
		return errors.New("not_found")
	}

	return err
}

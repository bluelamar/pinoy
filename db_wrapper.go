package main

func (pDb *DBInterface) DbwDelete(entity string, rMap *map[string]interface{}) error {
	id := (*rMap)["_id"].(string)
	rev := (*rMap)["_rev"].(string)
	return pDb.Delete(entity, id, rev)
}

/*
 * Determines to update if _id is present and key is empty, else create entry
 */
func (pDb *DBInterface) DbwUpdate(entity, key string, rMap *map[string]interface{}) error {
	var err error
	id, ok := (*rMap)["_id"]
	if ok && key == "" {
		_, err = pDb.Update(entity, id.(string), (*rMap)["_rev"].(string), (*rMap))
	} else {
		_, err = pDb.Create(entity, key, (*rMap))
	}

	return err
}

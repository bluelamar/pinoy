package food

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
)

const (
	foodEntity = "food"

	foodUsageEntity = "food_usage" // database entity
	foodUsageID     = "ItemID"
	foodUsageTO     = "TotOrders"

	SessFItems = "fitems"    // session key for slice of FoodItem
	SessTCost  = "totalCost" // session key for total cost of food items purchased
)

type FoodUsage struct {
	ItemID    string // FoodItem.Item + ":" +  FoodItem.Size
	TotOrders int
}
type FoodUsageTable struct {
	*psession.SessionDetails
	Title         string
	FoodUsageList []FoodUsage
	BackupTime    string
}

type FoodItem struct {
	Item     string
	Size     string
	Price    string
	Quantity int
	ItemID   string // Item + ":" +  Size
}
type FoodTable struct {
	*psession.SessionDetails
	Items []FoodItem
	Room  string
}

type FoodRecord struct {
	*psession.SessionDetails
	FoodData FoodItem
	Room     string
}

type PurchaseRecord struct {
	*psession.SessionDetails
	Items     []FoodItem
	Room      string
	TotalCost float64
}

func makeItemID(item, size string) string {
	return item + ":" + size
}

func makeFoodItem(vm map[string]interface{}) *FoodItem {

	id := ""
	name, exists := vm["ItemID"]
	if !exists {
		return nil
	}
	id = name.(string)
	if id == "" {
		// ignore this record
		return nil
	}

	item := ""
	if name, exists = vm["Item"]; exists {
		item = name.(string)
	}

	size := ""
	if name, exists = vm["Size"]; exists {
		size = name.(string)
	}

	price := ""
	if name, exists = vm["Price"]; exists {
		price = name.(string)
	}

	quant := int(0)
	if num, exists := vm["Quantity"]; exists {
		quant = int(num.(float64))
	}

	fitem := FoodItem{
		ItemID:   id,
		Item:     item,
		Size:     size,
		Price:    price,
		Quantity: quant,
	}

	return &fitem
}

func getFoodItems() ([]FoodItem, error) {
	fitems := make([]FoodItem, 0)
	resArray, err := database.DbwReadAll(foodEntity)
	if err != nil {
		return fitems, err
	}

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		foodItem := makeFoodItem(vm)
		if foodItem == nil {
			continue
		}
		/*
			 FoodItem{
				ItemID:   id,
				Item:     item,
				Size:     size,
				Price:    price,
				Quantity: quant,
			} */
		fitems = append(fitems, *foodItem)
	}

	return fitems, nil
}

func Food(w http.ResponseWriter, r *http.Request) {

	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Food Items", "Food Item page to Pinoy Lodge")
	if r.Method != "GET" {
		log.Println("food: bad http method: should only be a GET")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/food.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("food: template parse error=", err)
		sessDetails.Sess.Message = "Failed to get food items"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	// read in all food items
	fitems, err := getFoodItems()
	if err != nil {
		log.Println(`food:ERROR: db readall error`, err)
		sessDetails.Sess.Message = "Failed to get all food items"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	foodData := FoodTable{
		sessDetails,
		fitems,
		"",
	}

	err = t.Execute(w, &foodData)
	if err != nil {
		log.Println("food:ERROR: Failed to return food items: err=", err)
		sessDetails.Sess.Message = "Failed to get food items"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
	}
}

func UpdFood(w http.ResponseWriter, r *http.Request) {

	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Update Food Items", "Update Food Item page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {

		item := ""
		items, ok := r.URL.Query()["item"]
		if ok && len(items[0]) > 0 {
			item = items[0]
		}

		update := ""
		updates, ok := r.URL.Query()["update"]
		if ok && len(updates[0]) > 0 {
			update = updates[0]
		}

		deleteItem := false
		if update == "delete" {
			deleteItem = true
		}

		var rMap *map[string]interface{}
		if len(item) > 1 {
			var err error
			rMap, err = database.DbwRead(foodEntity, item)
			if err != nil {
				log.Println("upd_food: Invalid Food item=", item, " : err=", err)
				sessDetails.Sess.Message = `Internal error for food item`
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
				return
			}
		}

		if deleteItem {
			if rMap == nil {
				log.Println("upd_item:delete: Item not specified")
				sessDetails.Sess.Message = `Item not specified`
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
				return
			}

			if err := database.DbwDelete(foodEntity, rMap); err != nil {
				sessDetails.Sess.Message = "Failed to delete item: " + item
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
				return
			}

			// leave Room Status in db - it will be cleaned out when manager does backups

			http.Redirect(w, r, "/desk/food", http.StatusFound)
			return
		}

		// user wants to add or update existing item
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_food.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("upd_food:ERROR: Failed to parse template: err=", err)
			sessDetails.Sess.Message = "Failed to Update food item: " + item
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		var foodData FoodItem
		if rMap != nil {
			foodData = FoodItem{
				ItemID:   (*rMap)["ItemID"].(string),
				Item:     (*rMap)["Item"].(string),
				Size:     (*rMap)["Size"].(string),
				Price:    (*rMap)["Price"].(string),
				Quantity: 0,
			}
		}

		updData := FoodRecord{
			sessDetails,
			foodData,
			"",
		}
		err = t.Execute(w, updData)
		if err != nil {
			log.Println("upd_food:ERROR: Failed to exec template: err=", err)
			sessDetails.Sess.Message = `Internal error in Update Food Item`
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		}

		return
	}

	if r.Method != "POST" {
		log.Println("upd_food:ERROR: bad http method: should only be a POST")
		sessDetails.Sess.Message = `Bad request to Update Food Item`
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
		return
	}

	r.ParseForm()

	item, _ := r.Form["item"]
	size, _ := r.Form["item_size"]
	cost, _ := r.Form["item_price"]

	// verify all fields are set
	if len(item) < 1 || len(item[0]) == 0 || len(size) < 1 || len(size[0]) == 0 || len(cost) < 1 || len(cost[0]) == 0 {
		log.Println("upd_food:POST: Missing form data")
		sessDetails.Sess.Message = "Missing required rate class fields"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
		return
	}

	id := makeItemID(item[0], size[0])

	fitem := make(map[string]interface{})
	fitem["Item"] = item[0]
	fitem["Size"] = size[0]
	fitem["Price"] = cost[0]
	//Quantity int
	fitem["ItemID"] = id
	err := database.DbwUpdate(foodEntity, id, &fitem)
	if err != nil {
		log.Println("upd_food:ERROR: Failed to update food item=", id, " : err=", err)
		sessDetails.Sess.Message = `Internal error in Update Food Item`
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/desk/food", http.StatusFound)
}

func Purchase(w http.ResponseWriter, r *http.Request) {

	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Purchase Food Items", "Purchase Food Item page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	// item size room
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {

		room := ""
		if rooms, ok := r.URL.Query()["room"]; ok {
			room = rooms[0]
		}

		item := ""
		if items, ok := r.URL.Query()["item"]; ok {
			item = items[0]
		}

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/purchase.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("purchase:ERROR: Failed to make purchase page for room=", room, " : err=", err)
			sessDetails.Sess.Message = "Failed to make purchase page: room=" + room
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		foodMap, err := database.DbwRead(foodEntity, item)
		if err != nil {
			log.Println("purchase:ERROR: Failed to read food item=", item, " : err=", err)
			sessDetails.Sess.Message = "Failed to get food item - bad or missing item"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		foodItem := makeFoodItem(*foodMap)
		if foodItem == nil {
			log.Println("purchase:ERROR: Failed to read food item=", item, " : err=", err)
			sessDetails.Sess.Message = "Failed to get food item - missing item"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		foodData := FoodRecord{
			sessDetails,
			*foodItem,
			room,
		}
		err = t.Execute(w, foodData)
		if err != nil {
			if room == "" {
				log.Println("purchase:ERROR: Failed to execute food purchase page for item=", item, ": err=", err)
				sessDetails.Sess.Message = "Failed to make food purchase page: item=" + item
			} else {
				log.Println("purchase:ERROR: Failed to execute food purchase page for item=", item, " room=", room, ": err=", err)
				sessDetails.Sess.Message = "Failed to make food purchase page: item=" + item + " room=" + room
			}
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)

		}
	} else {
		r.ParseForm()
		item, _ := r.Form["item"]
		size, _ := r.Form["size"]
		quantity := r.Form["quantity"]
		roomNum, _ := r.Form["room_num"]

		if len(item) < 1 || len(item[0]) == 0 || len(size) < 1 || len(size[0]) == 0 || len(quantity) < 1 || len(quantity[0]) == 0 {
			log.Println("purchase:ERROR: Missing required form data")
			sessDetails.Sess.Message = `Missing required fields in Purchase Food Items`
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
			return
		}

		id := makeItemID(item[0], size[0])

		// read the food_usage record and update it - then write it back to the db
		// what to do with the room_num? oh, update its
		// now update room usage stats
		key := ""
		fue, err := database.DbwRead(foodUsageEntity, id)
		if err != nil {
			// lets make a new usage object
			fu := map[string]interface{}{
				foodUsageID: id,
				foodUsageTO: int(0),
			}
			fue = &fu
			key = id
		}

		quant, _ := strconv.Atoi(quantity[0])
		sum := misc.XtractIntField(foodUsageTO, fue) + quant
		(*fue)[foodUsageTO] = sum

		// update food_usage for the item
		err = database.DbwUpdate(foodUsageEntity, key, fue)
		if err != nil {
			log.Println("purchase:ERROR: Failed to update food usage for room=", roomNum[0], " : item=", id, " : err=", err)
		}

		http.Redirect(w, r, "/desk/purchase_summary?room="+roomNum[0]+"&item="+id+"&quantity="+quantity[0], http.StatusFound)
	}
}

func PurchaseSummary(w http.ResponseWriter, r *http.Request) {

	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Purchase Food Items", "Purchase Food Item page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	// item size room
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {

		var items []FoodItem
		totalCost := float64(0)
		sess, err := psession.GetUserSession(w, r)
		if sess != nil {
			if fitems, ok := sess.Values["fitems"].([]FoodItem); ok {
				items = fitems
			}
			if tcost, ok := sess.Values["totalCost"].(float64); ok {
				totalCost = tcost
			}
		} else {
			log.Println("purchase_summary:WARN: error from get user session: err=", err)
		}

		done := ""
		if dones, ok := r.URL.Query()["done"]; ok {
			done = dones[0]
		}
		if done == "true" {
			// cleanup shopping cart
			if sess != nil {
				delete(sess.Values, "fitems")
				delete(sess.Values, "totalCost")
				sess.Save(r, w)
			}
			http.Redirect(w, r, "/desk/food", http.StatusFound)
			return
		}

		room := ""
		if rooms, ok := r.URL.Query()["room"]; ok {
			room = rooms[0]
		}

		item := ""
		size := ""
		itemID := ""
		if items, ok := r.URL.Query()["item"]; ok {
			itemID = items[0]
			// split item into item and size
			iAndS := strings.Split(itemID, ":")
			item = iAndS[0]
			size = iAndS[1]
		}

		price := ""
		if prices, ok := r.URL.Query()["price"]; ok {
			price = prices[0]
		} else {
			// get price for this item
			fe, err := database.DbwRead(foodEntity, itemID)
			if err == nil {
				price, _ = (*fe)["Price"].(string)
			}
		}

		quantity := ""
		if quantities, ok := r.URL.Query()["quantity"]; ok {
			quantity = quantities[0]
		}

		purMap := make(map[string]interface{})
		purMap["quant"] = quantity
		purMap["price"] = misc.StripMonPrefix(price) // strip the "$" from the price

		quant := misc.XtractIntField("quant", &purMap)
		pr := misc.XtractFloatField("price", &purMap)
		cost := pr * float64(quant)
		totalCost += cost

		// create food item data for purchase summary
		purchaseItem := FoodItem{
			Item:     item,
			Size:     size,
			Price:    price,
			Quantity: quant,
			ItemID:   itemID,
		}

		if items == nil {
			items = make([]FoodItem, 1)
			items[0] = purchaseItem
		} else {
			items = append(items, purchaseItem)
		}
		if sess != nil {
			sess.Values["fitems"] = items
			sess.Values["totalCost"] = totalCost
			if err = sess.Save(r, w); err != nil {
				log.Println("purchase_summary: failed to save sess: err=", err)
			}
		}

		purchaseData := PurchaseRecord{
			sessDetails,
			items,
			room,
			totalCost,
		}

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/purchase_summary.gtpl", "static/header.gtpl")
		if err != nil {
			if room == "" {
				log.Println("purchase_summary:ERROR: Failed to make purchase summary page for item=", item, " : err=", err)
				sessDetails.Sess.Message = "Failed to make purchase summary page: item=" + item
			} else {
				log.Println("purchase_summary:ERROR: Failed to make purchase summary page for item=", item, " room=", room, " : err=", err)
				sessDetails.Sess.Message = "Failed to make purchase summary page: item=" + item + " room=" + room
			}
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, purchaseData)
		if err != nil {
			if room == "" {
				log.Println("purchase_summary:ERROR: Failed to execute food purchase page for item=", item, ": err=", err)
				sessDetails.Sess.Message = "Failed to make food purchase summary page: item=" + item
			} else {
				log.Println("purchase_summary:ERROR: Failed to execute food purchase page for item=", item, " room=", room, ": err=", err)
				sessDetails.Sess.Message = "Failed to make food purchase summary page: item=" + item + " room=" + room
			}
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)

		}
	}
}

func ReportFoodUsage(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Food Report", "Food Report page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	if r.Method != "GET" {
		log.Println("ReportFoodUsage:ERROR: bad http method: should only be a GET")
		sessDetails.Sess.Message = "Failed to get food report"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/food_report.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("ReportFoodUsage:ERROR: Failed to parse templates: err=", err)
		sessDetails.Sess.Message = "Failed to get food report"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	title := `Current Food Report`
	dbName := foodUsageEntity
	if bkups, ok := r.URL.Query()["bkup"]; ok {
		dbName = foodUsageEntity + "_" + bkups[0]
		log.Println("ReportFoodUsage: use backup db=", dbName)
		if bkups[0] == "b" {
			title = `Previous Food Report`
		} else {
			title = `Oldest Food Report`
		}
	}

	resArray, err := database.DbwReadAll(dbName)
	if err != nil {
		log.Println(`ReportFoodUsage:ERROR: db readall error`, err)
		sessDetails.Sess.Message = "Failed to get food report"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	timeStamp := ""
	usageList := make([]FoodUsage, 0)
	for _, v := range resArray {
		vm := v.(map[string]interface{})
		id := ""
		name, exists := vm[foodUsageID]
		if !exists {
			// check for timestamp record
			name, exists = vm["BackupTime"]
			if exists {
				timeStamp = name.(string)
			}
			continue
		}
		id = name.(string)
		if id == "" {
			// ignore this record
			continue
		}

		orderCnt := misc.XtractIntField(foodUsageTO, &vm)

		fusage := FoodUsage{
			ItemID:    id,
			TotOrders: orderCnt,
		}
		usageList = append(usageList, fusage)
	}

	tblData := FoodUsageTable{
		sessDetails,
		title,
		usageList,
		timeStamp,
	}

	if err = t.Execute(w, &tblData); err != nil {
		log.Println("ReportFoodUsage:ERROR: could not execute template: err=", err)
		sessDetails.Sess.Message = "Failed to report food usage"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
	}
}

func BackupFoodUsage(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	// check session expiration and authorization
	sessDetails := psession.GetSessDetails(r, "Backup and Reset Food Report", "Backup and Reset Food Report page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	toDB := foodUsageEntity + "_c"
	if err := misc.CleanupDbUsage(toDB, foodUsageID); err != nil {
		log.Println("BackupFoodUsage:ERROR: Failed to cleanup db=", toDB, " : err=", err)
	}
	fromDB := foodUsageEntity + "_b"
	if err := misc.CopyDbUsage(fromDB, toDB, foodUsageID); err != nil {
		log.Println("BackupFoodUsage:ERROR: Failed to copy usage from db=", fromDB, " to=", toDB, " : err=", err)
	}

	bkupTime, err := database.DbwRead(fromDB, "BackupTime")
	if err == nil {
		// write it to the toDB
		database.DbwUpdate(toDB, "BackupTime", bkupTime)
	} else {
		log.Println("BackupFoodUsage:WARN: Failed to copy backup time from=", fromDB, " to=", toDB, " : err=", err)
	}

	toDB = fromDB
	if err := misc.CleanupDbUsage(toDB, foodUsageID); err != nil {
		log.Println("BackupFoodUsage:ERROR: Failed to cleanup db=", toDB, " : err=", err)
	}
	if err := misc.CopyDbUsage(foodUsageEntity, toDB, foodUsageID); err != nil {
		log.Println("BackupFoodUsage:ERROR: Failed to copy usage from db=", foodUsageEntity, " to=", toDB, " : err=", err)
	}
	bkupTime, err = database.DbwRead(foodUsageEntity, "BackupTime")
	if err == nil {
		// write it to the toDB
		database.DbwUpdate(toDB, "BackupTime", bkupTime)
	} else {
		log.Println("BackupFoodUsage:ERROR: Failed to copy backup time from=", foodUsageEntity, " to=", toDB, " : err=", err)
	}

	// lastly reset the current food usage
	// 0 the TotOrders
	resArray, err := database.DbwReadAll(foodUsageEntity)
	if err != nil {
		log.Println(`BackupFoodUsage:ERROR: db readall: err=`, err)
		return
	}

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		_, exists := vm[foodUsageID]
		if !exists {
			continue
		}

		(vm)[foodUsageTO] = int(0)
		if err := database.DbwUpdate(foodUsageEntity, "", &vm); err != nil {
			log.Println(`BackupFoodUsage:ERROR: db update: err=`, err)
		}
	}

	nowStr, _ := misc.TimeNow()

	bkupTime, err = database.DbwRead(foodUsageEntity, "BackupTime")
	if err == nil {
		// write it to the toDB
		(*bkupTime)["BackupTime"] = nowStr
		if err := database.DbwUpdate(foodUsageEntity, "", bkupTime); err != nil {
			log.Println("BackupFoodUsage:ERROR: Failed to update backup time for=", foodUsageEntity, " : err=", err)
		}
	} else {
		tstamp := map[string]interface{}{"BackupTime": nowStr}
		if err := database.DbwUpdate(foodUsageEntity, "BackupTime", &tstamp); err != nil {
			log.Println("BackupFoodUsage:ERROR: Failed to create backup time for=", foodUsageEntity, " : err=", err)
		}
	}

	http.Redirect(w, r, "/manager/report_food_usage", http.StatusFound)
}

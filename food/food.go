package food

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
)

const (
	foodEntity = "food"
)

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
		/* FIX
		id := ""
		name, exists := vm["ItemID"]
		if !exists {
			continue
		}
		id = name.(string)
		if id == "" {
			// ignore this record
			continue
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
		if !ok || len(updates[0]) < 1 {
			log.Println("upd_room: missing Url param: update")
		} else {
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

			// FIX removeRoomStatus(room)

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

		// FIX http.Redirect(w, r, "/desk/food", http.StatusFound)
		return
	}

	if r.Method != "POST" {
		log.Println("upd_food: bad http method: should only be a POST")
		http.Error(w, "Bad request", http.StatusBadRequest) // FIX
		return
	}

	r.ParseForm()
	for k, v := range r.Form { // FIX
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}

	item := r.Form["item"]
	size := r.Form["item_size"]
	cost := r.Form["item_price"]

	// TODO verify all fields are set

	id := makeItemID(item[0], size[0])

	// TODO set in db
	fmt.Printf("upd_food:FIX item=%s size=%s cost=%s id=%s\n", item, size, cost, id)

	fmt.Printf("upd_food:FIX post about to redirect to food\n")
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

		size := ""
		if sizes, ok := r.URL.Query()["size"]; ok {
			size = sizes[0]
		}

		price := "" // get price for item
		if prices, ok := r.URL.Query()["price"]; ok {
			price = prices[0]
		}

		id := "" // unique id for item
		if ids, ok := r.URL.Query()["id"]; ok {
			id = ids[0]
		}

		fmt.Printf("purchase:FIX room=%s item=%s size=%s price=%s id=%s\n", room, item, size, price, id)
		if strings.Compare(id, "") == 0 {
			id = makeItemID(item, size)
		}

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/purchase.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("purchase:ERROR: Failed to make purchase page for room=", room, " : err=", err)
			sessDetails.Sess.Message = "Failed to make purchase page: room=" + room
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		foodMap, err := database.DbwRead(foodEntity, id)
		if err != nil {
			log.Println("purchase:ERROR: Failed to read food item=", id, " : err=", err)
			sessDetails.Sess.Message = "Failed to get food item - bad or missing item"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		foodItem := makeFoodItem(*foodMap)
		if foodItem == nil {
			log.Println("purchase:ERROR: Failed to read food item=", id, " : err=", err)
			sessDetails.Sess.Message = "Failed to get food item - bad or missing item"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}
		/* FIX
		foodItem := FoodItem{
			Item:     item,
			Size:     size,
			Price:    price,
			Quantity: 1,
			ItemID:   id,
		} */
		foodData := FoodRecord{
			sessDetails,
			*foodItem,
			room,
		}
		err = t.Execute(w, foodData)
		if err != nil {
			log.Println("purchase:ERROR: Failed to execute food purchase page for room=", room, ": err=", err)
			sessDetails.Sess.Message = "Failed to make food purchase page: room=" + room
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)

		}
	} else {
		fmt.Println("purchase:FIX should be post")
		r.ParseForm()
		for k, v := range r.Form { // FIX
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}

		item := r.Form["item"]
		size := r.Form["size"]
		quantity := r.Form["quantity"]
		id := r.Form["id"]
		roomNum := r.Form["room_num"]

		// TODO set in db
		fmt.Printf("purchase:FIX item=%s size=%s quantity=%s room-num=%s id=%s\n", item, size, quantity, roomNum, id)

		fmt.Printf("purchase:FIX post about to redirect to room_status\n")
		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

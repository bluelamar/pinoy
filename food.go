package pinoy

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type FoodItem struct {
	Item     string
	Size     string
	Price    string
	Quantity int
	Id       string
}
type FoodTable struct {
	*SessionDetails
	FoodData []FoodItem
	Room     string
}

type FoodRecord struct {
	*SessionDetails
	FoodData FoodItem
	Room     string
}

func food(w http.ResponseWriter, r *http.Request) {
	fmt.Println("food:method:", r.Method)
	if r.Method != "GET" {
		fmt.Printf("food: bad http method: should only be a GET\n")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	room := ""
	rooms, ok := r.URL.Query()["room"]
	if !ok || len(rooms[0]) < 1 {
		log.Println("food: Url Param 'room' is missing")
	} else {
		room = rooms[0]
	}
	fmt.Printf("food: room: %s\n", room)

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/food.gtpl", "static/header.gtpl")
	if err != nil {
		fmt.Printf("food: err: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		sessDetails := get_sess_details(r, "Food Items", "Food Item page to Pinoy Lodge")
		a1 := FoodItem{
			Item:  "San Miguel beer",
			Size:  "large",
			Price: "$2.50",
			Id:    "a1",
		}
		a2 := FoodItem{
			Item:  "Buko Pandan",
			Size:  "small",
			Price: "$4.75",
			Id:    "a2",
		}

		ftbl := make([]FoodItem, 2)
		ftbl[0] = a1
		ftbl[1] = a2

		foodData := FoodTable{
			sessDetails,
			ftbl,
			room,
		}
		err = t.Execute(w, &foodData)
		if err != nil {
			fmt.Println("food: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func upd_food(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upd_food:method:", r.Method)
	if r.Method != "POST" {
		fmt.Printf("upd_food: bad http method: should only be a POST\n")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	r.ParseForm()
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}

	item := r.Form["item"]
	size := r.Form["size"]
	id := r.Form["id"]
	cost := r.Form["cost"]

	// TODO if no id create unique id
	// verify all fields are set

	// TODO set in db
	fmt.Printf("upd_food: item=%s size=%s cost=%s id=%s\n", item, size, cost, id)

	fmt.Printf("upd_food: post about to redirect to food\n")
	http.Redirect(w, r, "/desk/food", http.StatusFound)
}

func purchase(w http.ResponseWriter, r *http.Request) {
	fmt.Println("purchase:method:", r.Method)
	// item size room
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {

		room := ""
		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("purchase: Url Param 'room' is missing")
		} else {
			room = rooms[0]
		}

		item := ""
		items, ok := r.URL.Query()["item"]
		if !ok || len(items[0]) < 1 {
			log.Println("purchase: Url Param 'item' is missing")
		} else {
			item = items[0]
		}

		size := ""
		sizes, ok := r.URL.Query()["size"]
		if !ok || len(sizes[0]) < 1 {
			log.Println("purchase: Url Param 'size' is missing")
		} else {
			size = sizes[0]
		}

		price := "" // get price for item
		prices, ok := r.URL.Query()["price"]
		if !ok || len(prices[0]) < 1 {
			log.Println("purchase: Url Param 'price' is missing")
		} else {
			price = prices[0]
		}

		id := "" // unique id for item
		ids, ok := r.URL.Query()["price"]
		if !ok || len(ids[0]) < 1 {
			log.Println("purchase: Url Param 'id' is missing")
		} else {
			id = ids[0]
		}

		fmt.Printf("purchase: room=%s item=%s size=%s price=%s id=%s\n", room, item, size, price, id)

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/purchase.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("purchase:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sessDetails := get_sess_details(r, "Purchase", "Purchase page of Pinoy Lodge")

			foodItem := FoodItem{
				Item:  item,
				Size:  size,
				Price: price,
				//Quantity: 3,
				Id: id,
			}
			foodData := FoodRecord{
				sessDetails,
				foodItem,
				room,
			}
			err = t.Execute(w, foodData)
			if err != nil {
				fmt.Println("purchase err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	} else {
		fmt.Println("purchase: should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}

		item := r.Form["item"]
		size := r.Form["size"]
		quantity := r.Form["quantity"]
		id := r.Form["id"]
		room_num := r.Form["room_num"]

		// TODO set in db
		fmt.Printf("purchase: item=%s size=%s quantity=%s room-num=%s id=%s\n", item, size, quantity, room_num, id)

		fmt.Printf("purchase: post about to redirect to room_status\n")
		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

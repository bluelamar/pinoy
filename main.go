package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

type SessionDetails struct {
	Sess   *PinoySession
	PgCont *PageContent
}

var welcTempl *template.Template

// TODO create secret using random values - should be from db? so all servers use
// the same secret
// Note: Don't store your key in your source code. Pass it via an
// environmental variable, or flag (or both), and don't accidentally commit it
// alongside your code. Ensure your key is sufficiently random - i.e. use Go's
// crypto/rand or securecookie.GenerateRandomKey(32) and persist the result
var store = sessions.NewCookieStore([]byte("something-very-secret"))

const CookieNameSID string = "PinoySID"

func welcome(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	// attention: If you do not call ParseForm method, the following data can not be obtained form
	fmt.Println(r.Form) // print information on server side.
	fmt.Println("welcome:path", r.URL.Path)
	fmt.Println("welcome:scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	//fmt.Fprintf(w, "Welcome to cd proLogue") // write data to response

	session, _ := store.Get(r, CookieNameSID)

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		//http.Error(w, "Forbidden", http.StatusForbidden)
		session.Values["authenticated"] = false
	}

	user := ""
	if sess_user, ok := session.Values["user"].(string); ok {
		user = sess_user
	}
	role := ""
	if sess_role, ok := session.Values["role"].(string); ok {
		role = sess_role
	}
	sess := &PinoySession{
		User:      user,
		Role:      role,
		SessID:    session.ID,
		CsrfToken: "",
		CsrfParam: "",
	}

	sess.welcomeSess(w, r)
}

func (sess *PinoySession) welcomeSess(w http.ResponseWriter, r *http.Request) {

	pageContent := &PageContent{
		"Welcome",
		"Welcome to Pinoy Lodge",
	}

	welc := SessionDetails{
		sess,
		pageContent,
	}
	fmt.Printf("welcsess: user=%s role=%s sessid=%s title=%s descr=%s\n", welc.Sess.User, welc.Sess.Role, welc.Sess.SessID, welc.PgCont.PageTitle, welc.PgCont.PageDescr)
	//http.Redirect(w, r, "/welcome", http.StatusFound)
	//err := welcTempl.ExecuteTemplate(w, "layout", welc)
	t, err := template.ParseFiles("static/layout.gtpl", "static/welcome.gtpl", "static/header.gtpl")
	//err := welcTempl.ExecuteTemplate(w, "welcome", welc)
	if err != nil {
		fmt.Printf("welcomesess:ExecuteTemplate error: %s", err.Error())
	} else {
		err = t.Execute(w, welc)
		if err != nil {
			fmt.Println("welcomesess err=", err)
		}
	}
}

// signout revokes authentication for a user
func signout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, CookieNameSID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Options.MaxAge = -1
	session.Values["authenticated"] = false
	session.Values["user"] = nil
	session.Values["role"] = nil

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func signin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("signin:method:", r.Method)
	if r.Method == "GET" {

		fmt.Println("login:get: parse tmpls login and header")
		t, err := template.ParseFiles("static/login.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("signin:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {

			sess := sess_attrs(r)
			pageContent := &PageContent{
				"Login",
				"Login for Pinoy Lodge",
			}

			loginPage := SessionDetails{
				sess,
				pageContent,
			}
			fmt.Println("signin:get: exec login")
			err = t.Execute(w, loginPage)
			if err != nil {
				fmt.Printf("signin:exec:err: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		fmt.Println("signin: else should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}
		username := r.Form["user_id"]
		password := r.Form["user_password"]
		// logic part of log in
		fmt.Println("username:", username)
		fmt.Println("password:", password)
		// verify user in db and set cookie et al

		session, err := store.Get(r, CookieNameSID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["authenticated"] = true
		session.Values["user"] = username[0]
		session.Values["role"] = "Manager" // "manager" "desk" "staff"
		session.Save(r, w)

		fmt.Printf("signin: post about to redirect to frontpage: auth=%t\n", session.Values["authenticated"].(bool))
		http.Redirect(w, r, "/frontpage", http.StatusFound)
	}

	fmt.Printf("signin:method=%s DONE", r.Method)
}

func frontpage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("frontpage:method:", r.Method)

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/frontpage.gtpl", "static/header.gtpl")
	if err != nil {
		fmt.Printf("frontpage:err: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		sessDetails := get_sess_details(r, "Front page", "Front page to Pinoy Lodge")
		err = t.Execute(w, sessDetails)
		if err != nil {
			fmt.Println("frontpage err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type RoomData struct {
	*SessionDetails
	RoomTable []RoomState
}

func room_status(w http.ResponseWriter, r *http.Request) {
	fmt.Println("room_status:method:", r.Method)

	open_rooms_only := false
	open_rooms, ok := r.URL.Query()["register"]
	if !ok || len(open_rooms[0]) < 1 {
		log.Println("register: Url Param 'register' is missing")
	}

	if len(open_rooms) > 0 {
		open_room := open_rooms[0]
		if open_room == "yes" {
			open_rooms_only = true
		}
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/room_status.gtpl", "static/header.gtpl")
	if err != nil {
		fmt.Printf("room_status: err: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		sessDetails := get_sess_details(r, "Room status", "Room status page to Pinoy Lodge")
		a1 := RoomState{
			Num:         122,
			Status:      "open",
			GuestInfo:   "robby",
			CheckinTime: "mar 1, 2019: 08:42am",
		}
		a2 := RoomState{
			Num:         101,
			Status:      "occupied",
			GuestInfo:   "ray",
			CheckinTime: "feb 9, 2019: 14:30pm",
		}
		a3 := RoomState{
			Num:         117,
			Status:      "open",
			GuestInfo:   "rich",
			CheckinTime: "feb 19, 2019: 22:30pm",
		}

		num_rooms := 3
		if open_rooms_only {
			num_rooms = 2
		}

		rtbl := make([]RoomState, num_rooms)
		rtbl[0] = a1
		if open_rooms_only {
			rtbl[1] = a3
		} else {
			rtbl[1] = a2
			rtbl[2] = a3
		}

		roomData := RoomData{
			sessDetails,
			rtbl,
			/* RoomTable: []RoomState{
				a1, a2,
			}, */
		}
		err = t.Execute(w, &roomData)
		if err != nil {
			fmt.Println("room_status: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type RegisterData struct {
	*SessionDetails
	RoomNum string
}

func register(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	fmt.Printf("register:method=%s time=%s\n", r.Method, t.Local())

	if r.Method == "GET" {

		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("register: Url Param 'room' is missing")
			return
		}
		// Query()["room"] will return an array of items, we only want the single item.
		room := rooms[0]

		fmt.Printf("register: room=%s\n", room)

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/register.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("register:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sessDetails := get_sess_details(r, "Registration", "Register page to Pinoy Lodge")
			regData := RegisterData{
				sessDetails,
				room,
			}
			err = t.Execute(w, regData)
			if err != nil {
				fmt.Println("register err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	} else {
		fmt.Println("register: should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}
		fname := r.Form["first_name"]
		lname := r.Form["last_name"]
		duration := r.Form["duration"]
		room_num := r.Form["room_num"]

		// TODO set in db
		fmt.Printf("register: first-name=%s last-name=%s room-num=%s duration=%s\n", fname, lname, room_num, duration)

		fmt.Printf("register: post about to redirect to room_hop for room=%s\n", room_num)
		http.Redirect(w, r, "/desk/room_hop?room="+room_num[0], http.StatusFound)
	}
}

type FoodItem struct {
	Item     string
	Size     string
	Price    string
	Quantity int
}
type FoodTable struct {
	*SessionDetails
	FoodData []FoodItem
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
		}
		a2 := FoodItem{
			Item:  "Buko Pandan",
			Size:  "small",
			Price: "$4.75",
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

type FoodRecord struct {
	*SessionDetails
	FoodData FoodItem
	Room     string
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

		fmt.Printf("purchase: room=%s item=%s size=%s price=%s\n", room, item, size, price)

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
		room_num := r.Form["room_num"]

		// TODO set in db
		fmt.Printf("purchase: item=%s size=%s quantity=%s room-num=%s\n", item, size, quantity, room_num)

		fmt.Printf("purchase: post about to redirect to room_status\n")
		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

func room_hop(w http.ResponseWriter, r *http.Request) {
	fmt.Println("room_hop:method:", r.Method)
	// item size room
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {

		room := ""
		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("room_hop: Url Param 'room' is missing")
		} else {
			room = rooms[0]
		}

		fmt.Printf("room_hop: room=%s\n", room)

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/room_hop.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("room_hop:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sessDetails := get_sess_details(r, "Room Bell Hop", "Bell Hop page of Pinoy Lodge")
			regData := RegisterData{
				sessDetails,
				room,
			}
			err = t.Execute(w, regData)
			if err != nil {
				fmt.Println("room_hop err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	} else {
		fmt.Println("room_hop: should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}

		bell_hop_pin := r.Form["bell_hop_pin"]
		room_num := r.Form["room_num"]

		// TODO set in db
		fmt.Printf("room_hop: bell-hop-pin=%s room-num=%s\n", bell_hop_pin, room_num)

		fmt.Printf("room_hop: post about to redirect to room_status\n")
		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

type UpdateRoom struct {
	*SessionDetails
	RoomDetails
}

func upd_room(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upd_room:method:", r.Method)
	// item size room
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {

		room := ""
		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("upd_room: Url Param 'room' is missing")
		} else {
			room = rooms[0]
		}

		update := ""
		updates, ok := r.URL.Query()["update"]
		if !ok || len(updates[0]) < 1 {
			log.Println("upd_room: Url Param 'update' is missing")
		} else {
			update = updates[0]
		}

		delete_room := false
		if update == "delete" {
			delete_room = true
		}

		fmt.Printf("upd_room: room=%s update=%s\n", room, update)

		if delete_room {
			// TODO delete specified room - error if room is not set
			if room == "" {
				http.Error(w, "Room number not specified", http.StatusBadRequest)
			}
			fmt.Printf("upd_room: delete room=%s\n", room)
			http.Redirect(w, r, "/desk/room_status", http.StatusFound)
		}

		// TODO get the room details from the db

		// user wants to add or update existing room
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_room.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("upd_room:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sessDetails := get_sess_details(r, "Update Room", "Update Room page of Pinoy Lodge")
			roomData := RoomDetails{
				room,
				1,
				"queen",
			}
			updData := UpdateRoom{
				sessDetails,
				roomData,
			}
			err = t.Execute(w, updData)
			if err != nil {
				fmt.Println("upd_room err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	} else {
		fmt.Println("upd_room: should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}

		bell_hop_pin := r.Form["bell_hop_pin"]
		room_num := r.Form["room_num"]

		// TODO set in db
		fmt.Printf("upd_room: bell-hop-pin=%s room-num=%s\n", bell_hop_pin, room_num)

		fmt.Printf("upd_room: post about to redirect to room_status\n")
		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

func staff(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upd_room:method:", r.Method)
}

func get_sess_details(r *http.Request, title, desc string) *SessionDetails {

	sess := sess_attrs(r)

	pageContent := &PageContent{
		title,
		desc,
	}

	sessDetails := SessionDetails{
		sess,
		pageContent,
	}
	return &sessDetails
}

func sess_attrs(r *http.Request) *PinoySession {

	session, err := store.Get(r, CookieNameSID)
	if err != nil {
		fmt.Printf("sess_attrs: err=%v\n", err)
	} else {
		fmt.Printf("sess_attrs: sess=%v\n", session)
	}
	for k, v := range session.Values {
		fmt.Printf("s-key: %v", k)
		fmt.Printf(" : s-val: %v\n", v)
	}
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		//http.Error(w, "Forbidden", http.StatusForbidden)
		session.Values["authenticated"] = false
		fmt.Printf("sess_attrs: set auth=%t\n", session.Values["authenticated"].(bool))
	}
	fmt.Printf("sess_attrs: auth=%t\n", session.Values["authenticated"].(bool))

	user := ""
	if sess_user, ok := session.Values["user"].(string); ok {
		user = sess_user
	}
	role := ""
	if sess_role, ok := session.Values["role"].(string); ok {
		role = sess_role
	}
	sess := &PinoySession{
		Auth:      session.Values["authenticated"].(bool),
		User:      user,
		Role:      role,
		SessID:    session.ID,
		CsrfToken: "",
		CsrfParam: "",
	}
	return sess
}

func main() {
	/* FIX TODO setup templates first
	 */

	// setup session options
	storeOptions := store.Options
	fmt.Printf("store options: path=%s domain=%s httponly=%t maxage=%d secure=%t\n",
		storeOptions.Path, storeOptions.Domain, storeOptions.HttpOnly, storeOptions.MaxAge, storeOptions.Secure)

	// set session expiry to 12 hours
	storeOptions.MaxAge = 12 * 60 * 60
	store.Options = storeOptions

	// setup routes
	http.HandleFunc("/", frontpage) // FIX welcome) // setting router rule
	http.HandleFunc("/frontpage", frontpage)
	http.HandleFunc("/signin", signin)
	http.HandleFunc("/signout", signout)
	http.HandleFunc("/desk/room_status", room_status)
	http.HandleFunc("/desk/register", register)
	http.HandleFunc("/desk/room_hop", room_hop)
	http.HandleFunc("/desk/food", food)
	http.HandleFunc("/desk/purchase", purchase)
	http.HandleFunc("/manager/staff", staff)
	http.HandleFunc("/manager/upd_room", upd_room)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	err := http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux)) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

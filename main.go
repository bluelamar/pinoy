package pinoy

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

// TODO create secret using random values - should be from db? so all servers use
// the same secret
// Note: Don't store your key in your source code. Pass it via an
// environmental variable, or flag (or both), and don't accidentally commit it
// alongside your code. Ensure your key is sufficiently random - i.e. use Go's
// crypto/rand or securecookie.GenerateRandomKey(32) and persist the result
var store = sessions.NewCookieStore([]byte("something-very-secret"))

var PCfg *PinoyConfig
var PDb *DBInterface

const CookieNameSID string = "PinoySID"

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
		fmt.Println("password:", password) // FIX
		// verify user in db and set cookie et al
		entity := "staff/" + username[0]
		umap, err := PDb.Read(entity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// get the role from the map
		role, ok := (*umap)["role"]
		if !ok {
			role = "Bypasser"
		}

		session, err := store.Get(r, CookieNameSID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["authenticated"] = true
		session.Values["user"] = username[0]
		session.Values["role"] = role // "Manager" "desk" "staff"
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

func main() {
	/* FIX TODO setup templates first
	 */

	cfg, err := LoadConfig("/etc/pinoy/config.json")
	if err != nil {
		log.Printf("pinoy:main: config load error: %v", err)
	} else {
		log.Printf("pinoy:main: cfg: %v", *cfg)
		PCfg = cfg
	}

	// initialize DB
	db, err := NewDatabase(PCfg)
	if err != nil {
		log.Printf("pinoy:main: db init error: %v", err)
	} else {
		log.Printf("pinoy:main: db init success")
		PDb = db
	}

	// setup session options
	storeOptions := store.Options
	fmt.Printf("store options: path=%s domain=%s httponly=%t maxage=%d secure=%t\n",
		storeOptions.Path, storeOptions.Domain, storeOptions.HttpOnly, storeOptions.MaxAge, storeOptions.Secure)

	// set session expiry to 12 hours
	storeOptions.MaxAge = 12 * 60 * 60
	store.Options = storeOptions

	// setup routes
	http.HandleFunc("/", frontpage)
	http.HandleFunc("/frontpage", frontpage)
	http.HandleFunc("/signin", signin)
	http.HandleFunc("/signout", signout)
	http.HandleFunc("/desk/room_status", room_status)
	http.HandleFunc("/desk/register", register)
	http.HandleFunc("/desk/room_hop", room_hop)
	http.HandleFunc("/desk/food", food)
	http.HandleFunc("/desk/purchase", purchase)
	http.HandleFunc("/manager/upd_food", upd_food)
	http.HandleFunc("/manager/staff", staff)
	http.HandleFunc("/manager/upd_staff", upd_staff)
	http.HandleFunc("/manager/upd_room", upd_room)
	http.HandleFunc("/manager/room_rates", room_rates)
	http.HandleFunc("/manager/upd_room_rate", upd_room_rate)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	err = http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux)) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

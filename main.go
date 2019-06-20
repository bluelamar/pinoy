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

// TODO create secret using random values - should be from db? so all servers use
// the same secret
// Note: Don't store your key in your source code. Pass it via an
// environmental variable, or flag (or both), and don't accidentally commit it
// alongside your code. Ensure your key is sufficiently random - i.e. use Go's
// crypto/rand or securecookie.GenerateRandomKey(32) and persist the result
var store = sessions.NewCookieStore([]byte("something-very-secret"))

var PCfg *PinoyConfig
var PDb *DBInterface
var Locale *time.Location

const CookieNameSID string = "PinoySID"

// signout revokes authentication for a user
func signout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, CookieNameSID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// update the employee report record
	UpdateEmployeeHours(session.Values["user"].(string), false, sess_attrs(r))

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
	fmt.Println("signin:FIX:method:", r.Method)
	if r.Method == "GET" {

		fmt.Println("login:FIX:get: parse tmpls login and header")
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
			fmt.Println("signin:FIX:get: exec login")
			err = t.Execute(w, loginPage)
			if err != nil {
				fmt.Printf("signin:exec:err: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		fmt.Println("signin:FIX else should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("FIX key:", k)
			fmt.Println("FIX val:", strings.Join(v, ""))
		}
		username := r.Form["user_id"]
		password := r.Form["user_password"]
		// logic part of log in
		fmt.Println("username:", username)
		fmt.Println("password:", password)
		// verify user in db and set cookie et al]
		entity := StaffEntity
		umap, err := PDb.Read(entity, username[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		log.Println("signin:FIX read user=", username[0], " entry=", umap)

		passwd, ok := (*umap)["Pwd"]
		if !ok {
			http.Error(w, "Not authorized", http.StatusInternalServerError) // FIX
			return
		}
		// use hash only for user password
		pwd := HashIt(password[0])

		log.Println("signin:FIX db.pwd=", passwd, " form.pwd=", password[0], " hash=", pwd)
		if passwd != pwd {
			http.Error(w, "Not authorized", http.StatusUnauthorized) // FIX
			return
		}

		// get the role from the map
		role, ok := (*umap)["Role"]
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

		// update the employee report record - dont wait for it
		go UpdateEmployeeHours(username[0], true, sess_attrs(r))

		fmt.Printf("signin:FIX: post about to redirect to frontpage: auth=%t\n", session.Values["authenticated"].(bool))
		http.Redirect(w, r, "/frontpage", http.StatusFound)
	}

	fmt.Printf("signin:method=%s DONE\n", r.Method)
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

func TimeNow(locale *time.Location) (string, time.Time) {
	var now time.Time
	if locale == nil {
		secondsEastOfUTC := int((8 * time.Hour).Seconds())

		maynila := time.FixedZone("Maynila Time", secondsEastOfUTC)
		now = time.Now().In(maynila)
	} else {
		// TODO use alternative to subtract time from utc
		now = time.Now().In(locale)
	}
	fmt.Println("timeNow:FIX got singapore now=", now)
	nowStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute())
	return nowStr, now
}

func main() {
	/* FIX TODO   setup templates first
	 */

	cfg, err := LoadConfig("/etc/pinoy/config.json")
	if err != nil {
		log.Printf("main: config load error: %v", err)
	} else {
		log.Printf("main: cfg: %v", *cfg)
		PCfg = cfg
	}

	loc, err := time.LoadLocation("Singapore")
	if err != nil {
		log.Println("main:WARN: Failed to load singapore time location: Use default locale: +0800 UTC-8: err=", err)
		//Locale = time.FixedZone("UTC-8", 8*60*60)
		secondsEastOfUTC := int((8 * time.Hour).Seconds())
		Locale = time.FixedZone("Maynila Time", secondsEastOfUTC)
	} else {
		Locale = loc
	}

	// initialize DB
	db, err := NewDatabase(PCfg)
	if err != nil {
		log.Printf("main: db init error: %v", err)
		log.Fatal("Failed to create db: ", err)
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
	http.HandleFunc("/desk/report_staff_hours", report_staff_hours)
	http.HandleFunc("/manager/upd_food", upd_food)
	http.HandleFunc("/manager/staff", staff)
	http.HandleFunc("/manager/upd_staff", upd_staff)
	http.HandleFunc("/manager/add_staff", add_staff)
	http.HandleFunc("/manager/rooms", rooms)
	http.HandleFunc("/manager/upd_room", upd_room)
	http.HandleFunc("/manager/room_rates", room_rates)
	http.HandleFunc("/manager/upd_room_rate", upd_room_rate)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	err = http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux)) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

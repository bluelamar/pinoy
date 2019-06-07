package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	RoomStatusEntity = "room_status"
)

type RoomState struct {
	RoomNum     string
	Status      string
	GuestInfo   string
	CheckinTime string
	Rate        string
}

type RoomStateTable struct {
	*SessionDetails
	Rooms     []RoomState
	OpenRooms string
}
type RoomStateEntry struct {
	*SessionDetails
	Room RoomState
}

type RegisterData struct {
	*SessionDetails
	RoomNum         string
	DurationOptions []string
}

func room_status(w http.ResponseWriter, r *http.Request) {
	fmt.Println("room_status:method:", r.Method)

	open_rooms_only := false
	open_rooms, ok := r.URL.Query()["register"]
	if !ok || len(open_rooms[0]) < 1 {
		log.Println("register: Url Param 'register' is missing so list all rooms")
	}

	openRoom := ""
	if len(open_rooms) > 0 {
		openRoom = open_rooms[0]
		if openRoom == "open" {
			open_rooms_only = true
		}
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/room_status.gtpl", "static/header.gtpl")
	if err != nil {
		log.Printf("room_status: template error: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {

		var roomStati []interface{}
		// FIX if open_rooms_only then Find else ReadAll

		if open_rooms_only {
			roomStati, err = PDb.Find(RoomStatusEntity, "Status", "open")
		} else {
			roomStati, err = PDb.Find(RoomStatusEntity, "Status", "booked")
		}
		if err != nil {
			log.Printf("room_status: Find %s rooms: err: %s\n", openRoom, err.Error())
			// FIX http.Error(w, err.Error(), http.StatusInternalServerError)
			// FIX return
		}
		fmt.Println("room_status:FIX: got=", roomStati)

		rtbl := make([]RoomState, len(roomStati))
		for k, v := range roomStati {
			val := v.(map[string]interface{})
			rs := RoomState{
				RoomNum:     val["RoomNum"].(string),
				Status:      val["Status"].(string),
				GuestInfo:   val["GuestInfo"].(string),
				CheckinTime: val["CheckinTime"].(string),
				Rate:        val["Rate"].(string),
			}
			rtbl[k] = rs
		}

		sessDetails := get_sess_details(r, "Room status", "Room status page to Pinoy Lodge")

		roomData := RoomStateTable{
			sessDetails,
			rtbl,
			openRoom,
		}
		err = t.Execute(w, &roomData)
		if err != nil {
			fmt.Println("room_status: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	fmt.Printf("register:method=%s time=%s\n", r.Method, t.Local())

	if r.Method == "GET" {

		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("register: Missing required 'room' param")
			return // FIX TODO return to frontpage with error
		}
		// Query()["room"] will return an array of items, we only want the single item.
		room := rooms[0]

		rates, ok := r.URL.Query()["rate"]
		if !ok || len(rates[0]) < 1 {
			log.Println("register: Missing required 'rates' param")
			return // FIX TODO return to frontpage with error
		}
		rate := rates[0]
		// get the RateClass for the room in order to get the options
		rateMap, err := PDb.Read(RoomRatesEntity, rate)
		if err != nil {
			log.Printf("register: Failed to read rate class=%s :err=%s\n", rate, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rrs2, ok := (*rateMap)["Rates"]
		if !ok {
			fmt.Printf("register: failed to get rates\n")
			http.Error(w, "No rates", http.StatusInternalServerError)
			return
		}
		rrs3 := rrs2.([]interface{})
		fmt.Printf("register: rates=%v\n", rrs2)
		durations := make([]string, len(rrs3))
		for k, v := range rrs3 {
			fmt.Printf("register: k=%d v=%v\n", k, v)

			v2 := v.(map[string]interface{})
			durations[k] = v2["TUnit"].(string)
		}

		fmt.Printf("register: room=%s rates=%v\n", room, durations)

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/register.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("register:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessDetails := get_sess_details(r, "Registration", "Register page of Pinoy Lodge")
		regData := RegisterData{
			sessDetails,
			room,
			durations,
		}
		err = t.Execute(w, regData)
		if err != nil {
			fmt.Println("register err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
				nil,
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

		// TODO set in db - date + timestamp + bell_hop_pin + room_num
		fmt.Printf("room_hop: bell-hop-pin=%s room-num=%s\n", bell_hop_pin, room_num)

		fmt.Printf("room_hop: post about to redirect to room_status\n")
		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

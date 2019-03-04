package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type RoomState struct {
	Num         int
	Status      string
	GuestInfo   string
	CheckinTime string
	Rate        string
}

type RoomDetails struct {
	RoomNum string
	NumBeds int
	BedSize string
	Rate    string
}

type UpdateRoom struct {
	*SessionDetails
	RoomDetails
}

type RoomRateData struct {
	Class string
	Hour3 string
	Hour6 string
	Extra string
}

type RateDataTable struct {
	*SessionDetails
	RateData []RoomRateData
}

type RateDataEntry struct {
	*SessionDetails
	RateData RoomRateData
}

func room_rates(w http.ResponseWriter, r *http.Request) {
	fmt.Println("room_rates:method:", r.Method)
	if r.Method != "GET" {
		fmt.Printf("room_rates: bad http method: should only be a GET\n")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/room_rates.gtpl", "static/header.gtpl")
	if err != nil {
		fmt.Printf("room_rates: err: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		sessDetails := get_sess_details(r, "Room Rates", "Room Rates page to Pinoy Lodge")
		a1 := RoomRateData{
			Class: "C",
			Hour3: "$10.00",
			Hour6: "$18.00",
			Extra: "$3.00",
		}
		a2 := RoomRateData{
			Class: "B",
			Hour3: "$9.00",
			Hour6: "$16.00",
			Extra: "$2.50",
		}

		rrd := make([]RoomRateData, 2)
		rrd[0] = a1
		rrd[1] = a2

		tblData := RateDataTable{
			sessDetails,
			rrd,
		}
		err = t.Execute(w, &tblData)
		if err != nil {
			fmt.Println("room_rates: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func upd_room_rate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upd_room_rate:method:", r.Method)
	if r.Method == "GET" {

		rate_class := ""
		rate_classes, ok := r.URL.Query()["rate_class"]
		if !ok || len(rate_classes[0]) < 1 {
			log.Println("upd_room_rate: Url Param 'rate_class' is missing")
		} else {
			rate_class = rate_classes[0]
		}

		hour3 := ""
		hour3s, ok := r.URL.Query()["hour3"]
		if !ok || len(hour3s[0]) < 1 {
			log.Println("upd_room_rate: Url Param 'hour3' is missing")
		} else {
			hour3 = hour3s[0]
		}

		hour6 := ""
		hour6s, ok := r.URL.Query()["hour6"]
		if !ok || len(hour6s[0]) < 1 {
			log.Println("upd_room_rate: Url Param 'hour6' is missing")
		} else {
			hour6 = hour6s[0]
		}

		extra := ""
		extras, ok := r.URL.Query()["extra"]
		if !ok || len(extras[0]) < 1 {
			log.Println("upd_room_rate: Url Param 'extra' is missing")
		} else {
			extra = extras[0]
		}

		update := ""
		updates, ok := r.URL.Query()["update"]
		if !ok || len(updates[0]) < 1 {
			log.Println("upd_room_rate: Url Param 'update' is missing")
		} else {
			update = updates[0]
		}

		delete_rate := false
		if update == "delete" {
			delete_rate = true
		}

		fmt.Printf("upd_room_rate: rate-class=%s update=%s hour3=%s hour6=%s extra=%s\n", rate_class, update, hour3, hour6, extra)

		if delete_rate {
			// TODO delete specified room - error if room is not set
			if rate_class == "" {
				http.Error(w, "Rate class not specified", http.StatusBadRequest)
			}
			fmt.Printf("upd_room_rate: delete room-rate=%s\n", rate_class)
			http.Redirect(w, r, "/manager/room_rates", http.StatusFound)
		}

		// TODO get the room details from the db

		// user wants to add or update existing room
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_room_rate.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("upd_room_rate:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sessDetails := get_sess_details(r, "Update Room Rate", "Update Room Rate page of Pinoy Lodge")
			roomData := RoomRateData{
				rate_class,
				hour3,
				hour6,
				extra,
			}
			updData := RateDataEntry{
				sessDetails,
				roomData,
			}
			err = t.Execute(w, updData)
			if err != nil {
				fmt.Println("upd_room_rate err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	} else {
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}

		rate_class := r.Form["rate_class"]
		hour3 := r.Form["hour3"]
		hour6 := r.Form["hour6"]
		extra := r.Form["extra"]

		// TODO if no id create unique id
		// verify all fields are set

		// TODO set in db
		fmt.Printf("upd_room_rate: rate_class=%s hour3=%s hour6=%s extra=%s\n", rate_class, hour3, hour6, extra)

		fmt.Printf("upd_room_rate: post about to redirect to food\n")
		http.Redirect(w, r, "/manager/room_rates", http.StatusFound)
	}
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
				"C", // room rate class
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

		num_beds := r.Form["num_beds"]
		room_num := r.Form["room_num"]
		bed_size := r.Form["bed_size"]
		room_rate := r.Form["room_rate"]

		// TODO set in db
		fmt.Printf("upd_room: room-num=%s num-beds=%s bed-size=%s room-rate=%s\n", room_num, num_beds, bed_size, room_rate)

		fmt.Printf("upd_room: post about to redirect to room_status\n")
		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

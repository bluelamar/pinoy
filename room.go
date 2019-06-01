package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

const (
	RoomRatesEntity = "room_rates"
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

type RoomData struct {
	*SessionDetails
	RoomTable []RoomState
}

type UpdateRoom struct {
	*SessionDetails
	RoomDetails
}

type RoomRate struct {
	TUnit string // time unit, ex: "3 Hours"
	Cost  string // cost of the TUnit, ex: "$10"
}
type RoomRateData struct {
	RateClass string // ex: "Small Room"
	Rates     []RoomRate
	/* FIX
	Hour3 string
	Hour6 string
	Extra string */
}

type RateDataTable struct {
	*SessionDetails
	RateData []RoomRateData
}

type RateDataEntry struct {
	*SessionDetails
	RateData RoomRateData
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
			Rate:        "C",
		}
		a2 := RoomState{
			Num:         101,
			Status:      "occupied",
			GuestInfo:   "ray",
			CheckinTime: "feb 9, 2019: 14:30pm",
			Rate:        "B",
		}
		a3 := RoomState{
			Num:         117,
			Status:      "open",
			GuestInfo:   "rich",
			CheckinTime: "feb 19, 2019: 22:30pm",
			Rate:        "B",
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
		/* FIX
		a1 := RoomRateData{
			RateClass: "C",
			Hour3:     "$10.00",
			Hour6:     "$18.00",
			Extra:     "$3.00",
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
		*/
		// []interface{}, error
		rrs, err := PDb.ReadAll(RoomRatesEntity)
		if err != nil {
			log.Println("room_rates: Failed to read room rates: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		rrds := make([]RoomRateData, len(rrs))
		for k, v := range rrs {
			log.Println("FIX got k=", k, " v=", v)
			val := v.(map[string]interface{})
			rs := val["Rates"].([]interface{})
			rates := make([]RoomRate, len(rs))
			for k2, v2 := range rs {
				val2 := v2.(map[string]interface{})
				rr := RoomRate{
					val2["TUnit"].(string),
					val2["Cost"].(string),
				}
				rates[k2] = rr
			}
			rrd := RoomRateData{
				val["RateClass"].(string),
				rates,
			}
			rrds[k] = rrd
		}

		tblData := RateDataTable{
			sessDetails,
			rrds,
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
		/* FIX
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
		*/
		update := ""
		updates, ok := r.URL.Query()["update"]
		if !ok || len(updates[0]) < 1 {
			log.Println("upd_room_rate: Url Param 'update' is missing")
		} else {
			update = updates[0]
		}

		deleteRate := false
		if update == "delete" {
			deleteRate = true
		}

		fmt.Printf("upd_room_rate: rate-class=%s update=%s\n", rate_class, update)
		var rateMap *map[string]interface{}
		if len(rate_class) > 1 {
			var err error
			rateMap, err = PDb.Read(RoomRatesEntity, rate_class)
			if err != nil {
				http.Error(w, "Invalid Rate class specified", http.StatusBadRequest)
				return
			}
		}

		if deleteRate {
			// delete specified room rate - error if room rate is not set
			if rate_class == "" {
				http.Error(w, "Rate class not specified", http.StatusBadRequest)
				return
			}
			fmt.Printf("upd_room_rate: delete room-rate=%s\n", rate_class)

			id := (*rateMap)["_id"].(string)
			rev := (*rateMap)["_rev"].(string)
			err := PDb.Delete(RoomRatesEntity, id, rev)
			if err != nil {
				http.Error(w, "Failed to delete room rate: "+rate_class, http.StatusConflict)
			} else {
				http.Redirect(w, r, "/manager/room_rates", http.StatusFound)
			}
			return
		}

		// user wants to add or update existing room rate
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_room_rate.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("upd_room_rate:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sessDetails := get_sess_details(r, "Update Room Rate", "Update Room Rate page of Pinoy Lodge")

			var rrs []RoomRate
			if rateMap != nil {
				fmt.Printf("upd_room_rate: rate-map=%v\n", (*rateMap))
				rrs2, ok := (*rateMap)["Rates"]
				if !ok {
					fmt.Printf("upd_room_rate: failed to get rates\n")
					http.Error(w, "No rates", http.StatusInternalServerError)
					return
				}
				rrs3 := rrs2.([]interface{})
				fmt.Printf("upd_room_rate: rates=%v\n", rrs2)
				rrs = make([]RoomRate, len(rrs3))
				for k, v := range rrs3 {
					fmt.Printf("upd_room_rate: k=%d v=%v\n", k, v)

					v2 := v.(map[string]interface{})
					rrs[k] = RoomRate{
						TUnit: v2["TUnit"].(string),
						Cost:  v2["Cost"].(string),
					}
				}
			} else {
				rrs = nil
			}
			roomData := RoomRateData{
				rate_class,
				rrs,
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
		newTimeUnit := r.Form["new_rate_time_unit"]
		newCost := r.Form["new_rate_cost"]
		fmt.Printf("upd_room_rate: rate_class=%s time-unit=%s cost=%s\n", rate_class, newTimeUnit, newCost)

		// validate incoming form fields
		if len(rate_class[0]) == 0 || len(newTimeUnit[0]) == 0 || len(newCost[0]) == 0 {
			log.Println("upd_room_rate:POST: Missing form data")
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		update := true
		var rateMap *map[string]interface{}
		var err error
		rateMap, err = PDb.Read(RoomRatesEntity, rate_class[0])
		if err != nil {
			log.Println("upd_room_rate:POST: err=", err)

			// no such entry so the rate-class must be new
			rm := make(map[string]interface{})
			rateMap = &rm
			(*rateMap)["RateClass"] = rate_class[0]
			(*rateMap)["Rates"] = make([]interface{}, 0)
			update = false
		}

		rates := (*rateMap)["Rates"]
		fmt.Printf("upd_room_rate: rate_class=%s rates=%v\n", rate_class, rates)
		// if rates has TUnit entry matching newTimeUnit, remove it since it will be replaced
		newRates := make([]map[string]interface{}, 0)
		rts := rates.([]interface{})
		for _, v := range rts {
			v2 := v.(map[string]interface{})
			tu := v2["TUnit"].(string)
			if newTimeUnit[0] == tu {
				continue
			}
			newRates = append(newRates, v2)
		}

		newRate := make(map[string]interface{})
		newRate["TUnit"] = newTimeUnit[0]
		newRate["Cost"] = newCost[0]

		newRates = append(newRates, newRate)
		(*rateMap)["Rates"] = newRates

		// set in db
		fmt.Printf("upd_room_rate:FIX rate_class=%s newrates=%v\n", rate_class, newRates)
		if update {
			_, err = PDb.Update(RoomRatesEntity, (*rateMap)["_id"].(string), (*rateMap)["_rev"].(string), (*rateMap))
			fmt.Printf("upd_room_rate:FIX update rate_class=%s val=%v\n", rate_class, (*rateMap))
		} else {
			_, err = PDb.Create(RoomRatesEntity, rate_class[0], (*rateMap))
		}
		if err != nil {
			log.Println("upd_room_rate:POST: Failed to create or updated rate=", rate_class[0], " :err=", err)
			http.Error(w, "Failed to create or update rate="+rate_class[0], http.StatusInternalServerError)
			return
		}

		fmt.Printf("upd_room_rate: post about to redirect to room rates\n")
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

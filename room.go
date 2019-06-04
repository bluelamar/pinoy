package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	RoomsEntity = "rooms"
)

type RoomDetails struct {
	RoomNum   string
	NumBeds   int
	BedSize   string
	RateClass string
}

type RoomDetailDataTable struct {
	*SessionDetails
	Rooms []RoomDetails
}
type RoomDetailEntry struct {
	*SessionDetails
	Room        RoomDetails
	RateClasses []string
}

func rooms(w http.ResponseWriter, r *http.Request) {
	fmt.Println("rooms:method:", r.Method)
	if r.Method != "GET" {
		fmt.Printf("rooms: bad http method: should only be a GET\n")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/rooms.gtpl", "static/header.gtpl")
	if err != nil {
		fmt.Printf("rooms: err: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		sessDetails := get_sess_details(r, "Rooms", "Rooms page to Pinoy Lodge")
		// []interface{}, error
		rrs, err := PDb.ReadAll(RoomsEntity)
		if err != nil {
			log.Println("rooms: Failed to read rooms: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		rrds := make([]RoomDetails, len(rrs))
		for k, v := range rrs {
			log.Println("rooms:FIX got k=", k, " v=", v)
			val := v.(map[string]interface{})
			nbs, err := strconv.Atoi(val["NumBeds"].(string))
			if err != nil {
				log.Println("rooms: Failed to comvert num rooms: err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			rrd := RoomDetails{
				RoomNum:   val["RoomNum"].(string),
				NumBeds:   nbs,
				BedSize:   val["BedSize"].(string),
				RateClass: val["RateClass"].(string),
			}
			rrds[k] = rrd
		}

		tblData := RoomDetailDataTable{
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

func upd_room(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upd_room:method:", r.Method)
	// item size room
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {

		room := ""
		rooms, ok := r.URL.Query()["room_num"]
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

		deleteRoom := false
		if update == "delete" {
			deleteRoom = true
		}

		fmt.Printf("upd_room: room=%s update=%s\n", room, update)
		var rMap *map[string]interface{}
		if len(room) > 1 {
			var err error
			rMap, err = PDb.Read(RoomsEntity, room)
			if err != nil {
				http.Error(w, "FIX: Invalid Room specified", http.StatusBadRequest)
				return
			}
		}

		if deleteRoom {
			// delete specified room - error if room is not set
			if room == "" {
				http.Error(w, "FIX: Room number not specified", http.StatusBadRequest)
				return
			}
			fmt.Printf("upd_room: delete room=%s\n", room)

			id := (*rMap)["_id"].(string)
			rev := (*rMap)["_rev"].(string)
			err := PDb.Delete(RoomsEntity, id, rev)
			if err != nil {
				sessDetails := get_sess_details(r, "Update Room", "Update Room page of Pinoy Lodge")
				sessDetails.Sess.Message = "Failed to delete room: " + room
				_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
			} else {
				http.Redirect(w, r, "/desk/room_status", http.StatusFound)
			}
			return
		}

		sessDetails := get_sess_details(r, "Update Room", "Update Room page of Pinoy Lodge")
		// user wants to add or update existing room
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_room.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("upd_room_:err: %s", err.Error())
			sessDetails.Sess.Message = "Failed to Update room: " + room
			err = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			if err != nil {
				return
			}
		} else {

			var roomData RoomDetails
			if rMap != nil {
				fmt.Printf("upd_room: r-map=%v\n", (*rMap))
				nbs, err := strconv.Atoi((*rMap)["NumBeds"].(string))
				if err != nil {
					log.Println("upd_room: Failed to comvert num rooms: err=", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				roomData = RoomDetails{
					RoomNum:   (*rMap)["RoomNum"].(string),
					NumBeds:   nbs,
					BedSize:   (*rMap)["BedSize"].(string),
					RateClass: (*rMap)["RateClass"].(string),
				}
			}

			// read the rate classes and create slice of strings
			rrs, err := PDb.ReadAll(RoomRatesEntity)
			if err != nil {
				log.Println("upd_room: Failed to read room rates: err=", err)
				sessDetails.Sess.Message = "Please Add or Update Room Rates"
				_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
				return
			}
			fmt.Println("upd_room:FIX got rates=", rrs)
			if len(rrs) == 0 {
				log.Println("upd_room: No room rates - ask user to update rates")
				sessDetails.Sess.Message = `Please Add or Update Room Rates`
				_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNoContent)
				return
			}

			rateClasses := make([]string, len(rrs))
			for k, v := range rrs {
				log.Println("upd_room:FIX got k=", k, " v=", v)
				val := v.(map[string]interface{})
				rateClasses[k] = val["RateClass"].(string)
			}

			updData := RoomDetailEntry{
				sessDetails,
				roomData,
				rateClasses,
			}
			err = t.Execute(w, updData)
			if err != nil {
				log.Println("upd_room: err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError) // FIX
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

		fmt.Printf("upd_room: room-num=%s num-beds=%s bed-size=%s room-rate=%s\n", room_num, num_beds, bed_size, room_rate)
		// validate incoming form fields
		if len(num_beds[0]) == 0 || len(room_num[0]) == 0 || len(bed_size[0]) == 0 || len(room_rate[0]) == 0 {
			log.Println("upd_room:POST: Missing form data")
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		update := true
		var rMap *map[string]interface{}
		var err error
		rMap, err = PDb.Read(RoomsEntity, room_num[0])
		if err != nil {
			log.Println("upd_room:POST: err=", err)

			// no such entry so the room must be new
			rm := make(map[string]interface{})
			rMap = &rm
			(*rMap)["NumBeds"] = num_beds[0]
			(*rMap)["RoomNum"] = room_num[0]
			(*rMap)["RateClass"] = room_rate[0]
			(*rMap)["BedSize"] = bed_size[0]
			update = false
		} else {
			(*rMap)["NumBeds"] = num_beds[0]
			(*rMap)["RateClass"] = room_rate[0]
			(*rMap)["BedSize"] = bed_size[0]
		}

		if update {
			_, err = PDb.Update(RoomsEntity, (*rMap)["_id"].(string), (*rMap)["_rev"].(string), (*rMap))
			fmt.Printf("upd_room_rate:FIX update room_num=%s val=%v\n", room_num, (*rMap))
		} else {
			_, err = PDb.Create(RoomsEntity, room_num[0], (*rMap))
		}
		if err != nil {
			log.Println("upd_room:POST: Failed to create or update room=", room_num[0], " :err=", err)
			http.Error(w, "Failed to create or update room="+room_num[0], http.StatusInternalServerError)
			return
		}

		fmt.Printf("upd_room: post about to redirect to rooms\n")
		http.Redirect(w, r, "/manager/rooms", http.StatusFound)
	}
}

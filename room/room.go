package room

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
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
	*psession.SessionDetails
	Rooms []RoomDetails
}
type RoomDetailEntry struct {
	*psession.SessionDetails
	Room        RoomDetails
	RateClasses []string
}

func Rooms(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Rooms", "Rooms page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	if r.Method != "GET" {
		log.Println("rooms: bad http method: should only be a GET")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/rooms.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("rooms:ERROR: Failed to parse template: err=", err)
		sessDetails.Sess.Message = "Internal error"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	// []interface{}, error
	rrs, err := database.DbwReadAll(RoomsEntity)
	if err != nil {
		log.Println("rooms:ERROR: Failed to read rooms: err=", err)
		sessDetails.Sess.Message = `Failed to read rooms`
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}
	rrds := make([]RoomDetails, len(rrs))
	for k, v := range rrs {
		val := v.(map[string]interface{})
		nbs, err := strconv.Atoi(val["NumBeds"].(string))
		if err != nil {
			log.Println("rooms:ERROR: Failed to convert num rooms: err=", err)
			sessDetails.Sess.Message = `Failed to convert number of rooms`
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
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
		log.Println("room_rates:ERROR: Failed to exec template: err=", err)
		sessDetails.Sess.Message = `Internal error for room rates`
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
	}
}

func UpdRoom(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Update Room", "Update Room page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {

		room := ""
		rooms, ok := r.URL.Query()["room_num"]
		if ok && len(rooms[0]) > 0 {
			room = rooms[0]
		}

		update := ""
		updates, ok := r.URL.Query()["update"]
		if !ok || len(updates[0]) < 1 {
			log.Println("upd_room: missing Url param: update")
		} else {
			update = updates[0]
		}

		deleteRoom := false
		if update == "delete" {
			deleteRoom = true
		}

		var rMap *map[string]interface{}
		if len(room) > 1 {
			var err error
			rMap, err = database.DbwRead(RoomsEntity, room)
			if err != nil {
				log.Println("upd_room: Invalid Room=", room, " : err=", err)
				sessDetails.Sess.Message = `Internal error for rooms`
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
				return
			}
		}

		if deleteRoom {
			if rMap == nil {
				log.Println("upd_room:delete: Room number not specified")
				sessDetails.Sess.Message = `Room number not specified`
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
				return
			}

			if err := database.DbwDelete(RoomsEntity, rMap); err != nil {
				sessDetails.Sess.Message = "Failed to delete room: " + room
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
				return
			}

			removeRoomStatus(room)

			http.Redirect(w, r, "/desk/room_status", http.StatusFound)
			return
		}

		// user wants to add or update existing room
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_room.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("upd_room:ERROR: Failed to parse template: err=", err)
			sessDetails.Sess.Message = "Failed to Update room: " + room
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		var roomData RoomDetails
		if rMap != nil {
			nbs, err := strconv.Atoi((*rMap)["NumBeds"].(string))
			if err != nil {
				log.Println("upd_room:ERROR: Failed to convert num beds: err=", err)
				sessDetails.Sess.Message = "Failed to Update room: " + room
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
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
		rrs, err := database.DbwReadAll(RoomRatesEntity)
		if err != nil {
			log.Println("upd_room:ERROR: Failed to read room rates: err=", err)
			sessDetails.Sess.Message = "Please Add or Update Room Rates"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
			return
		}
		if len(rrs) == 0 {
			log.Println("upd_room: No room rates - ask user to update rates")
			sessDetails.Sess.Message = `Please Add or Update Room Rates`
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNoContent)
			return
		}

		rateClasses := make([]string, len(rrs))
		for k, v := range rrs {
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
			log.Println("upd_room:ERROR: Failed to exec template: err=", err)
			sessDetails.Sess.Message = `Internal error in Update Room`
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		}
	} else {
		r.ParseForm()

		numBeds := r.Form["num_beds"]
		roomNum := r.Form["room_num"]
		bedSize := r.Form["bed_size"]
		roomRate := r.Form["room_rate"]

		// validate incoming form fields
		if len(numBeds[0]) == 0 || len(roomNum[0]) == 0 || len(bedSize[0]) == 0 || len(roomRate[0]) == 0 {
			log.Println("upd_room:POST: Missing form data")
			sessDetails.Sess.Message = `Missing required fields in Update Room Rates`
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
			return
		}

		key := ""
		update := true
		var rMap *map[string]interface{}
		var err error
		rMap, err = database.DbwRead(RoomsEntity, roomNum[0])
		if err != nil {
			log.Println("upd_room:POST: room num=", roomNum[0], " :err=", err)

			// no such entry so the room must be new
			rm := make(map[string]interface{})
			rMap = &rm
			(*rMap)["RoomNum"] = roomNum[0]
			(*rMap)["NumBeds"] = numBeds[0]
			(*rMap)["RateClass"] = roomRate[0]
			(*rMap)["BedSize"] = bedSize[0]
			update = false
			key = roomNum[0]
		} else {
			(*rMap)["NumBeds"] = numBeds[0]
			(*rMap)["RateClass"] = roomRate[0]
			(*rMap)["BedSize"] = bedSize[0]
		}

		err = database.DbwUpdate(RoomsEntity, key, rMap)
		if err != nil {
			log.Println("upd_room:POST:ERROR: Failed to create or update room=", roomNum[0], " :err=", err)
			sessDetails.Sess.Message = "Failed to create or update room=" + roomNum[0]
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		// if existing room updated - update the status in case the rate was changed
		// else if new room then create a new status record
		var roomStatus *map[string]interface{}
		key = ""
		if update {
			roomStatus, err = database.DbwRead(RoomStatusEntity, roomNum[0])
			if err == nil {
				(*roomStatus)["Rate"] = roomRate[0]
			}
		}
		if roomStatus == nil {
			// create the room status record for this room
			rs := make(map[string]interface{})
			rs["RoomNum"] = roomNum[0]
			rs["Status"] = OpenStatus
			rs["GuestInfo"] = ""
			rs["CheckinTime"] = ""
			rs["CheckoutTime"] = ""
			rs["Rate"] = roomRate[0]
			roomStatus = &rs
			key = roomNum[0]
		}
		err = database.DbwUpdate(RoomStatusEntity, key, roomStatus)
		if err != nil {
			log.Println("upd_room:POST:ERROR: Failed to create or update room status=", roomNum[0], " :err=", err)

			sessDetails.Sess.Message = "Failed to update room status: room=" + roomNum[0]
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}
		// update in-memory room_status
		putNewRoomStatus(*roomStatus)

		http.Redirect(w, r, "/manager/rooms", http.StatusFound)
	}
}

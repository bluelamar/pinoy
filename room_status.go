package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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
	Rooms []RoomState
}
type RoomStateEntry struct {
	*SessionDetails
	Room RoomState
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
			RoomNum:     "122",
			Status:      "open",
			GuestInfo:   "robby",
			CheckinTime: "mar 1, 2019: 08:42am",
			Rate:        "C",
		}
		a2 := RoomState{
			RoomNum:     "101",
			Status:      "occupied",
			GuestInfo:   "ray",
			CheckinTime: "feb 9, 2019: 14:30pm",
			Rate:        "B",
		}
		a3 := RoomState{
			RoomNum:     "117",
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

		roomData := RoomStateTable{
			sessDetails,
			rtbl,
		}
		err = t.Execute(w, &roomData)
		if err != nil {
			fmt.Println("room_status: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

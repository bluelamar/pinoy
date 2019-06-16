package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

const (
	BellHopEntity  = "bellhops"
	HopShiftEntity = "hop_shift"
)

type BellHop struct {
	UserID     string
	Room       string
	TimeStamp  string
	LoginValid bool // true=valid or false
	// key= UserID + ":" + Room + ":" + TimeStamp
}

type BellHopsTable struct {
	*SessionDetails
	BellHops []BellHop
}

type HopperTable struct {
	*SessionDetails
	RoomNum     string
	CheckinTime string
	Repeat      string
}

// currently clocked in
type HopShift struct {
	UserID  string
	ClockIn string
}

type HopShiftTable struct {
	*SessionDetails
	Shift []HopShift
}

func room_hop(w http.ResponseWriter, r *http.Request) {
	fmt.Println("room_hop:method:", r.Method)
	sessDetails := get_sess_details(r, "Room Bell Hop", "Bell Hop page of Pinoy Lodge")
	if sessDetails.Sess.Role != ROLE_MGR && sessDetails.Sess.Role != ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {

		room := ""
		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("room_hop: Url Param 'room' is missing")
		} else {
			room = rooms[0]
		}

		citime := ""
		citimes, ok := r.URL.Query()["citime"]
		if !ok || len(citimes[0]) < 1 {
			log.Println("room_hop: Url Param 'citime' is missing")
		} else {
			citime = citimes[0]
		}

		repeat := ""
		repeats, ok := r.URL.Query()["repeat"]
		if !ok || len(repeats[0]) < 1 {
			log.Println("room_hop: Url Param 'repeat' is missing")
		} else {
			repeat = repeats[0]
		}
		fmt.Printf("room_hop: room=%s checkin=%s repeat=%s\n", room, citime, repeat)

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/room_hop.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("room_hop:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			regData := HopperTable{
				sessDetails,
				room,
				citime,
				repeat,
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
		citime := r.Form["citime"]
		repeat := r.Form["repeat"] // if true then can repeat upon failure

		// FIX TODO read the scheduled hop shift - ReadAll HopShiftEntity
		// - if no hops returned, redirect to update_shift with all the above params
		// -- the update_shift impl should use those params if they exist to redirect
		//    to the room_hop page again
		// for each UserID read the staff record to get the passwd
		// get pwd for each hop and compare to the entered pin
		// if matched then have the userid - if no match then invalid pin and record no-match
		// or return error? repeat the GET one more time if repeat=true
		// ie. http.Redirect(w, r, "/desk/room_hop?room="+room_num[0]+"&citime="+nowStr+"&repeat=false", http.StatusFound)

		/* TODO set in db: BellHopEntity
		type BellHop struct {
		- userid : UserID
		- room : Room
		- time stamp : TimeStamp
		- success login : LoginValid = true or false
		- key= UserID + ":" + Room + ":" + TimeStamp
		*/
		fmt.Printf("room_hop: bell-hop-pin=%s room-num=%s citime=%s repeat=%s\n", bell_hop_pin, room_num, citime, repeat)

		fmt.Printf("room_hop: post about to redirect to room_status\n")
		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	RoomStatusEntity = "room_status"
)

type RoomState struct {
	RoomNum      string
	Status       string
	GuestInfo    string
	CheckinTime  string
	CheckoutTime string
	Rate         string
}

type RoomStateTable struct {
	*SessionDetails
	Rooms         []RoomState
	OpenRoomsOnly bool
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

	if len(open_rooms) > 0 {
		if open_rooms[0] == "open" {
			open_rooms_only = true
		}
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/room_status.gtpl", "static/header.gtpl")
	if err != nil {
		log.Printf("room_status: template error: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {

		var roomStati []interface{}

		if open_rooms_only {
			roomStati, err = PDb.Find(RoomStatusEntity, "Status", "open")
		} else {
			roomStati, err = PDb.Find(RoomStatusEntity, "Status", "booked")
		}
		if err != nil {
			log.Printf("room_status: Find %s rooms: err: %s\n", open_rooms[0], err.Error())
			// FIX http.Error(w, err.Error(), http.StatusInternalServerError)
			// FIX return
		}
		fmt.Println("room_status:FIX: got=", roomStati)

		rtbl := make([]RoomState, len(roomStati))
		for k, v := range roomStati {
			fmt.Println("room_status:FIX k=", k, " :v=", v)
			val := v.(map[string]interface{})
			rs := RoomState{
				RoomNum:      val["RoomNum"].(string),
				Status:       val["Status"].(string),
				GuestInfo:    val["GuestInfo"].(string),
				CheckinTime:  val["CheckinTime"].(string),
				CheckoutTime: val["CheckoutTime"].(string),
				Rate:         val["Rate"].(string),
			}
			rtbl[k] = rs
		}

		sessDetails := get_sess_details(r, "Room status", "Room status page to Pinoy Lodge")

		roomData := RoomStateTable{
			sessDetails,
			rtbl,
			open_rooms_only,
		}
		err = t.Execute(w, &roomData)
		if err != nil {
			fmt.Println("room_status: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func CalcCheckoutTime(checkinTime, duration string) (string, error) {
	// ex checkinTime: 2019-06-11 12:49
	ci := strings.Split(checkinTime, " ")
	date, hourMin := ci[0], ci[1]
	dateSlice := strings.Split(date, "-")
	hm := strings.Split(hourMin, ":")
	hourStr := hm[0]
	min := hm[1]
	hourNum, err := strconv.Atoi(hourStr)
	if err != nil {
		return "", err
	}

	/* FIX
	start := time.Date(2009, 1, 1, 12, 0, 0, 0, time.UTC)
		afterTenMinutes := start.Add(time.Minute * 10)
		afterTenHours := start.Add(time.Hour * 10)
		afterTenDays := start.Add(time.Hour * 24 * 10)
	*/
	// ex duration: 3 Hours
	dur := strings.Split(duration, " ")
	dnum := dur[0]
	dunit := dur[1]
	durNum, err := strconv.Atoi(dnum)
	if err != nil {
		return "", err
	}
	dnumUnit := 1
	if dunit == "Days" {
		dnumUnit = 24
	}

	newDate := date
	newHourMin := ""
	// Handle crossover midnight
	newHour := hourNum + (durNum * dnumUnit)
	if newHour >= 24 {
		// crossed to a following day
		nHour := newHour % 24
		numDays := newHour / 24
		newHour = nHour
		// parse to get new day from the date, ex: 2019-06-11
		day, err := strconv.Atoi(dateSlice[2])
		if err != nil {
			return "", err
		}
		day += numDays
		if day > 30 { // FIX TODO need to check calendar feb=28, etc
			// Handle crossover to next month - and crossover to the next year
		}
		dayStr := strconv.Itoa(day)
		newDate = dateSlice[0] + "-" + dateSlice[1] + "-" + dayStr
	}
	fmt.Println("register:FIX newdate=", newDate)

	hourStr = strconv.Itoa(newHour)
	newHourMin = hourStr + ":" + min
	fmt.Println("register:FIX hourstr=", hourStr, " :min=", min, " :newhour=", newHour)
	return newDate + " " + newHourMin, nil
}

func checkoutRoom(room string, w http.ResponseWriter, r *http.Request) error {
	rs, err := PDb.Read(RoomStatusEntity, room)
	if err != nil {
		log.Println("checkout: Failed to read room status for room=", room, " :err=", err)
		sessDetails := get_sess_details(r, "Checkout", "Register page of Pinoy Lodge")
		sessDetails.Sess.Message = "Failed to checkout: room=" + room
		err = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return err
	}
	(*rs)["Status"] = "open"
	(*rs)["GuestInfo"] = ""
	(*rs)["CheckinTime"] = ""
	(*rs)["CheckoutTime"] = ""
	err = PDb.DbwUpdate(RoomStatusEntity, "", rs)
	if err != nil {
		log.Println("checkout: Failed to update room status for room=", room, " :err=", err)
		sessDetails := get_sess_details(r, "Checkout", "Register page of Pinoy Lodge")
		sessDetails.Sess.Message = "Failed to checkout: room=" + room
		err = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return err
	}
	return nil
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
		room := rooms[0]

		regAction := ""
		regActions, ok := r.URL.Query()["reg"]
		if ok && len(regActions[0]) > 1 {
			regAction = regActions[0]
		}

		if regAction == "checkout" {
			err := checkoutRoom(room, w, r)
			if err == nil {
				http.Redirect(w, r, "/desk/room_status?register=open", http.StatusTemporaryRedirect)
			}
			return
		}

		rate := ""
		rates, ok := r.URL.Query()["rate"]
		if ok {
			rate = rates[0]
		}

		// get the RateClass for the room in order to get the options
		rateMap, err := PDb.Read(RoomRatesEntity, rate)
		if err != nil {
			log.Printf("register: Failed to read rate class=%s :err=%s\n", rate, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rrs2, ok := (*rateMap)["Rates"]
		if !ok {
			// TODO SendErrorPage
			fmt.Printf("register:FIX: failed to get rates\n")
			http.Error(w, "No rates", http.StatusInternalServerError)
			return
		}
		rrs3 := rrs2.([]interface{})
		fmt.Printf("register: rates=%v\n", rrs2)
		durations := make([]string, len(rrs3))
		for k, v := range rrs3 {
			fmt.Printf("register:FIX: k=%d v=%v\n", k, v)

			v2 := v.(map[string]interface{})
			durations[k] = v2["TUnit"].(string)
		}

		fmt.Printf("register: room=%s rates=%v\n", room, durations)

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/register.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("register:FIX:err: %s", err.Error())
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("register: Failed to make register page for room=", room, " :err=", err)
			sessDetails := get_sess_details(r, "Registration", "Register page of Pinoy Lodge")
			sessDetails.Sess.Message = "Failed to make register page: room=" + room
			err = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
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
			fmt.Println("register:FIX err=", err)
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("register: Failed to read room status for room=", room, " :err=", err)
			sessDetails.Sess.Message = "Failed to make register page: room=" + room
			err = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
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

		// set in db
		fmt.Printf("register: first-name=%s last-name=%s room-num=%s duration=%s\n", fname, lname, room_num, duration)
		// read room status record and reset as booked with customers name
		rs, err := PDb.Read(RoomStatusEntity, room_num[0])
		if err != nil {
			log.Println("register: Failed to read room status for room=", room_num[0], " :err=", err)
			sessDetails := get_sess_details(r, "Registration", "Register page of Pinoy Lodge")
			sessDetails.Sess.Message = "Failed to read room status: room=" + room_num[0]
			err = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}
		fmt.Println("register:FIX read room num=", room_num[0], " :room=", rs)

		(*rs)["Status"] = "booked"
		(*rs)["GuestInfo"] = fname[0] + " " + lname[0]

		nowStr := TimeNow()
		(*rs)["CheckinTime"] = nowStr
		checkOutTime, err := CalcCheckoutTime(nowStr, duration[0])
		(*rs)["CheckoutTime"] = checkOutTime
		fmt.Println("register:FIX got singapore nowStr=", nowStr, " :checkout=", checkOutTime)

		// put status record back into db
		// FIX TODO record for the customer ?
		err = PDb.DbwUpdate(RoomStatusEntity, "", rs)
		if err != nil {
			log.Println("register: Failed to update room status for room=", room_num[0], " :err=", err)
			sessDetails := get_sess_details(r, "Registration", "Register page of Pinoy Lodge")
			sessDetails.Sess.Message = "Failed to read room status: room=" + room_num[0]
			err = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		fmt.Printf("register:FIX: post about to redirect to room_hop for room=%s\n", room_num)
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

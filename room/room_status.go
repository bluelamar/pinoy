package room

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
	"github.com/bluelamar/pinoy/staff"
)

const (
	RoomStatusEntity = "room_status" // database entity

	// room status
	BookedStatus = "booked"
	OpenStatus   = "open"
	LimboStatus  = "limbo"
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
	*psession.SessionDetails
	Rooms         []RoomState
	OpenRoomsOnly bool
}
type RoomStateEntry struct {
	*psession.SessionDetails
	Room RoomState
}

type RegisterData struct {
	*psession.SessionDetails
	RoomNum         string
	DurationOptions []string
}

// keep all room status in memory for fast lookup
var roomStateCurrent map[string]RoomState = make(map[string]RoomState)
var mutex sync.Mutex

func xlateToRoomStatus(val map[string]interface{}) *RoomState {
	rn := ""
	if str, exists := val["RoomNum"]; exists {
		rn = str.(string)
	} else {
		return nil
	}
	st := ""
	if str, exists := val["Status"]; exists {
		st = str.(string)
	}
	gi := ""
	if str, exists := val["GuestInfo"]; exists {
		gi = str.(string)
	}
	ci := ""
	if str, exists := val["CheckinTime"]; exists {
		ci = str.(string)
	}
	co := ""
	if str, exists := val["CheckoutTime"]; exists {
		co = str.(string)
	}
	rt := ""
	if str, exists := val["Rate"]; exists {
		rt = str.(string)
	}
	rs := RoomState{
		RoomNum:      rn,
		Status:       st,
		GuestInfo:    gi,
		CheckinTime:  ci,
		CheckoutTime: co,
		Rate:         rt,
	}
	return &rs
}

func loadRoomsState() error {
	var once sync.Once
	var err error
	onceBody := func() {
		var roomStati []interface{}
		if len(roomStateCurrent) > 0 {
			// other task already loaded so nothing to do
			return
		}
		roomStati, err = database.DbwReadAll(RoomStatusEntity)
		if err != nil {
			return
		}

		var roomsMap map[string]RoomState = make(map[string]RoomState)
		for _, v := range roomStati {
			val := v.(map[string]interface{})
			rs := xlateToRoomStatus(val)
			if rs == nil {
				continue
			}
			rn := val["RoomNum"].(string)
			/*
				rn := ""
				if str, exists := val["RoomNum"]; exists {
					rn = str.(string)
				} else {
					continue
				}
				st := ""
				if str, exists := val["Status"]; exists {
					st = str.(string)
				}
				gi := ""
				if str, exists := val["GuestInfo"]; exists {
					gi = str.(string)
				}
				ci := ""
				if str, exists := val["CheckinTime"]; exists {
					ci = str.(string)
				}
				co := ""
				if str, exists := val["CheckoutTime"]; exists {
					co = str.(string)
				}
				rt := ""
				if str, exists := val["Rate"]; exists {
					rt = str.(string)
				}
				rs := RoomState{
					RoomNum:      rn,
					Status:       st,
					GuestInfo:    gi,
					CheckinTime:  ci,
					CheckoutTime: co,
					Rate:         rt,
				} */
			roomsMap[rn] = *rs
		}
		roomStateCurrent = roomsMap
	}
	once.Do(onceBody)
	return err
}

func PutNewRoomStatus(roomStatus map[string]interface{}) error {
	rs := xlateToRoomStatus(roomStatus)
	if rs == nil {
		log.Println("PutNewRoomStatus: Failed to translate map to room status")
		return errors.New("room status translation")
	}

	mutex.Lock()
	roomStateCurrent[roomStatus["RoomNum"].(string)] = *rs
	mutex.Unlock()
	return nil
}

func GetRoomStati(roomStatus string, durLimit time.Duration) ([]RoomState, error) {
	// process roomStateCurrent and create list of rooms according to status
	// within	cfg.RoomStatusMonitorInterval minutes
	if len(roomStateCurrent) == 0 {
		if err := loadRoomsState(); err != nil {
			log.Println("room_status: Cannot load room status: error=", err)
			return nil, err
		}
	}

	_, nowTime := misc.TimeNow()

	var keys []string
	for k := range roomStateCurrent {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rtbl := make([]RoomState, 0)
	for _, kvalue := range keys {
		rs := roomStateCurrent[kvalue]
		if strings.Compare(roomStatus, rs.Status) == 0 {
			if durLimit > 0 {
				checkoutTime, err := time.ParseInLocation(staff.DateTimeLongForm, rs.CheckoutTime, misc.GetLocale())
				if err == nil {
					dur := checkoutTime.Sub(nowTime)
					if dur.Minutes() > durLimit.Minutes() {
						continue
					}
				} else {
					log.Println("GetRoomStati:ERROR: bad checkout time in record=", rs, " :err=", err)
				}
			}
			rtbl = append(rtbl, rs)
		}
	}
	return rtbl, nil
}

func RoomStatus(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.Get_sess_details(r, "Room Status", "Room Status page for Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	openRoomsOnly := false
	openRooms, _ := r.URL.Query()["register"]
	if len(openRooms) > 0 {
		if openRooms[0] == OpenStatus {
			openRoomsOnly = true
		}
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/room_status.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("room_status: template parse error=", err)
		sessDetails.Sess.Message = "Failed to get room status"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	// fill the map since service probably just started up
	if len(roomStateCurrent) == 0 {
		if err := loadRoomsState(); err != nil {
			log.Println("room_status: Cannot load room status: error=", err)
			sessDetails.Sess.Message = "Failed to get room status"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}
	}

	var rtbl []RoomState
	dur := time.Duration(0)
	if openRoomsOnly {
		rtbl, err = GetRoomStati(OpenStatus, dur)
	} else {
		rtbl, err = GetRoomStati(BookedStatus, dur)
	}
	if err != nil {
		log.Println("room_status:ERROR: Find ", openRooms[0], " rooms: err=", err)
		sessDetails.Sess.Message = "Failed to get room status"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	roomData := RoomStateTable{
		sessDetails,
		rtbl,
		openRoomsOnly,
	}
	err = t.Execute(w, &roomData)
	if err != nil {
		log.Println("room_status:ERROR: Failed to return room status: err=", err)
		sessDetails.Sess.Message = "Failed to get room status"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
	}
}

func CalcCheckoutTime(ciTime time.Time, duration string) (string, error) {
	// ex checkinTime: 2019-06-11 12:49
	/*
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

	newDate := ciTime.Add(time.Hour * time.Duration(durNum) * time.Duration(dnumUnit))

	nowStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d",
		newDate.Year(), newDate.Month(), newDate.Day(),
		newDate.Hour(), newDate.Minute())

	return nowStr, nil
}

func checkoutRoom(room string, w http.ResponseWriter, r *http.Request, sessDetails *psession.SessionDetails) error {
	rs, err := database.DbwRead(RoomStatusEntity, room)
	if err != nil {
		log.Println("checkout:ERROR: Failed to read room status for room=", room, " : err=", err)
		sessDetails.Sess.Message = "Failed to checkout: room=" + room
		err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return err
	}
	(*rs)["Status"] = OpenStatus
	(*rs)["GuestInfo"] = ""
	(*rs)["CheckinTime"] = ""
	(*rs)["CheckoutTime"] = ""
	err = database.DbwUpdate(RoomStatusEntity, "", rs)
	if err != nil {
		log.Println("checkout:ERROR: Failed to update room status for room=", room, " : err=", err)
		sessDetails.Sess.Message = "Failed to checkout: room=" + room
		err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return err
	}

	var once sync.Once
	onceBody := func() {
		delete(roomStateCurrent, room)
	}
	once.Do(onceBody)

	return nil
}

func Register(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()

	// check session expiration and authorization
	sessDetails := psession.Get_sess_details(r, "Registration", "Register page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {

		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("register: Missing required room param")
			sessDetails.Sess.Message = "Failed to register - missing room number"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
			return
		}
		room := rooms[0]

		regAction := ""
		regActions, ok := r.URL.Query()["reg"]
		if ok && len(regActions[0]) > 1 {
			regAction = regActions[0]
		}

		if regAction == "checkout" {
			err := checkoutRoom(room, w, r, sessDetails)
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
		rateMap, err := database.DbwRead(RoomRatesEntity, rate)
		if err != nil {
			log.Println("register:ERROR: Failed to read rate class=", rate, " : err=", rate, err)
			sessDetails.Sess.Message = "Failed to register - bad or missing room rate"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		rrs2, ok := (*rateMap)["Rates"]
		if !ok {
			log.Println("register:ERROR: failed to get rates: err=", err)
			sessDetails.Sess.Message = "Failed to register - rates missing"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}
		rrs3 := rrs2.([]interface{})
		durations := make([]string, len(rrs3))
		for k, v := range rrs3 {
			v2 := v.(map[string]interface{})
			durations[k] = v2["TUnit"].(string)
		}

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/register.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("register:ERROR: Failed to make register page for room=", room, " : err=", err)
			sessDetails.Sess.Message = "Failed to make register page: room=" + room
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		regData := RegisterData{
			sessDetails,
			room,
			durations,
		}
		err = t.Execute(w, regData)
		if err != nil {
			log.Println("register:ERROR: Failed to read room status for room=", room, " : err=", err)
			sessDetails.Sess.Message = "Failed to make register page: room=" + room
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		}

	} else {
		r.ParseForm()

		fname := r.Form["first_name"]
		lname := r.Form["last_name"]
		duration := r.Form["duration"]
		roomNum := r.Form["room_num"]

		// set in db
		// read room status record and reset as booked with customers name
		rs, err := database.DbwRead(RoomStatusEntity, roomNum[0])
		if err != nil {
			log.Println("register:ERROR: Failed to read room status for room=", roomNum[0], " : err=", err)
			sessDetails.Sess.Message = "Failed to read room status: room=" + roomNum[0]
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		(*rs)["Status"] = BookedStatus
		guestInfo := fname[0] + " " + lname[0]
		(*rs)["GuestInfo"] = guestInfo

		nowStr, nowTime := misc.TimeNow()
		(*rs)["CheckinTime"] = nowStr
		checkOutTime, err := CalcCheckoutTime(nowTime, duration[0])
		(*rs)["CheckoutTime"] = checkOutTime

		// put status record back into db
		// TODO record for the customer ?
		err = database.DbwUpdate(RoomStatusEntity, "", rs)
		if err != nil {
			log.Println("register:ERROR: Failed to update room status for room=", roomNum[0], " : err=", err)
			sessDetails.Sess.Message = "Failed to read room status: room=" + roomNum[0]
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		roomState := RoomState{
			RoomNum:      roomNum[0],
			Status:       BookedStatus,
			GuestInfo:    guestInfo,
			CheckinTime:  nowStr,
			CheckoutTime: checkOutTime,
			Rate:         (*rs)["Rate"].(string),
		}
		var once sync.Once
		onceBody := func() {
			roomStateCurrent[roomNum[0]] = roomState
		}
		once.Do(onceBody)

		http.Redirect(w, r, "/desk/room_hop?room="+roomNum[0]+"&citime="+nowStr+"&repeat=true", http.StatusFound)
	}
}

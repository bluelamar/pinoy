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
	roomUsageEntity  = "room_usage"  // database entity

	// room status
	BookedStatus = "booked"
	OpenStatus   = "open"
	LimboStatus  = "limbo"

	roomUsageRN  = "RoomNum"
	roomUsageTNG = "TotNumGuests"
	roomUsageTH  = "TotHours"
	roomUsageNTO = "NumTimesOccupied"
)

type RoomUsage struct {
	RoomNum          string
	TotNumGuests     int
	TotHours         float64
	NumTimesOccupied int
}
type RoomUsageTable struct {
	*psession.SessionDetails
	Title         string
	RoomUsageList []RoomUsage
	BackupTime    string
}

type RoomState struct {
	RoomNum        string
	Status         string
	GuestInfo      string
	NumGuests      int
	Duration       string
	CheckinTime    string
	CheckoutTime   string
	Rate           string
	NumExtraGuests int
	ExtraRate      string
	Purchases      string // append food purchases here
	PurchaseTotal  string // how many pesos spent
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
	if str, exists := val[roomUsageRN]; exists {
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
	numg := misc.XtractIntField("NumGuests", &val)
	dur := ""
	if str, exists := val["Duration"]; exists {
		dur = str.(string)
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
	neg := misc.XtractIntField("NumExtraGuests", &val)
	er := ""
	if str, exists := val["ExtraRate"]; exists {
		er = str.(string)
	}
	pur := ""
	if str, exists := val["Purchase"]; exists {
		pur = str.(string)
	}
	pt := ""
	if str, exists := val["PurchaseTotal"]; exists {
		pt = str.(string)
	}
	rs := RoomState{
		RoomNum:        rn,
		Status:         st,
		GuestInfo:      gi,
		NumGuests:      numg,
		Duration:       dur,
		CheckinTime:    ci,
		CheckoutTime:   co,
		Rate:           rt,
		NumExtraGuests: neg,
		ExtraRate:      er,
		Purchases:      pur,
		PurchaseTotal:  pt,
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

		roomsMap := make(map[string]RoomState)
		for _, v := range roomStati {
			val := v.(map[string]interface{})
			rs := xlateToRoomStatus(val)
			if rs == nil {
				continue
			}
			rn := val[roomUsageRN].(string)
			roomsMap[rn] = *rs
		}
		roomStateCurrent = roomsMap
	}
	once.Do(onceBody)
	return err
}

func putNewRoomStatus(roomStatus map[string]interface{}) error {
	rs := xlateToRoomStatus(roomStatus)
	if rs == nil {
		log.Println("putNewRoomStatus:ERROR: Failed to translate map to room status")
		return errors.New("room status translation")
	}

	updateRoomStatus(roomStatus[roomUsageRN].(string), *rs)
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
	mutex.Lock()
	for k := range roomStateCurrent {
		keys = append(keys, k)
	}
	mutex.Unlock()
	sort.Strings(keys)

	rtbl := make([]RoomState, 0)
	for _, kvalue := range keys {
		rs, ok := roomStateCurrent[kvalue]
		if ok && strings.Compare(roomStatus, rs.Status) == 0 {
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
	sessDetails := psession.GetSessDetails(r, "Room Status", "Room Status page for Pinoy Lodge")
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

func removeRoomStatus(room string) {
	mutex.Lock()
	delete(roomStateCurrent, room)
	mutex.Unlock()
}
func updateRoomStatus(room string, rs RoomState) {
	mutex.Lock()
	roomStateCurrent[room] = rs
	mutex.Unlock()
}

// returns <real usage hours>, <expected usage hours>
func calcDiffTime(ciTime, coTime string) (float64, float64) {

	_, nowTime := misc.TimeNow() // real checkout time

	// convert each string to Time and get the diff

	// expected check out time
	clockoutTime, err := time.ParseInLocation(staff.DateTimeLongForm, coTime, misc.GetLocale())
	if err != nil {
		return 0, 0
	}

	clockinTime, err := time.ParseInLocation(staff.DateTimeLongForm, ciTime, misc.GetLocale())
	if err != nil {
		return 0, 0
	}

	expDur := clockoutTime.Sub(clockinTime)
	realDur := nowTime.Sub(clockinTime)

	// check if real checkout time significantly different than expected checkout time
	gracePeriodMinutes := int(15) // TODO should be configurable
	durDiff := realDur - expDur
	diffMinutes := int(durDiff.Minutes())
	if diffMinutes > gracePeriodMinutes {
		log.Println("room_status.caldDiffTime:WARN: Checkout time surpassed grace period=",
			gracePeriodMinutes, " : actual minutes past expected checkout time=", diffMinutes)
		return realDur.Hours(), expDur.Hours()
	}

	return realDur.Hours(), expDur.Hours()
}

func checkoutRoom(room string, w http.ResponseWriter, r *http.Request, sessDetails *psession.SessionDetails) error {
	rs, err := database.DbwRead(RoomStatusEntity, room)
	if err != nil {
		log.Println("checkout:ERROR: Failed to read room status for room=", room, " : err=", err)
		sessDetails.Sess.Message = "Failed to checkout: room=" + room
		err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return err
	}

	// get the number guests and room usage times
	num := misc.XtractIntField("NumGuests", rs)
	ciTime, _ := (*rs)["CheckinTime"].(string)
	coTime, _ := (*rs)["CheckoutTime"].(string)

	(*rs)["Status"] = OpenStatus
	(*rs)["GuestInfo"] = ""
	(*rs)["NumGuests"] = int(0)
	(*rs)["NumExtraGuests"] = int(0)
	(*rs)["ExtraRate"] = ""
	(*rs)["Duration"] = ""
	(*rs)["CheckinTime"] = ""
	(*rs)["CheckoutTime"] = ""
	(*rs)["Purchase"] = ""
	(*rs)["PurchaseTotal"] = ""
	err = database.DbwUpdate(RoomStatusEntity, "", rs)
	if err != nil {
		log.Println("checkout:ERROR: Failed to update room status for room=", room, " : err=", err)
		sessDetails.Sess.Message = "Failed to checkout: room=" + room
		err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return err
	}

	roomState := xlateToRoomStatus(*rs)
	if roomState != nil {
		updateRoomStatus(room, *roomState)
	} else {
		log.Println("checkout:ERROR: Failed to update in-mem checkout room=", room)
		// TODO set flag force reload?
	}

	// now update room usage stats
	key := ""
	rs, err = database.DbwRead(roomUsageEntity, room)
	if err != nil {
		// lets make a new usage object
		ru := map[string]interface{}{
			roomUsageRN:  room,
			roomUsageTNG: int(0),
			roomUsageTH:  float64(0),
			roomUsageNTO: int(0),
		}
		rs = &ru
		key = room
	}

	totGuestCnt := num + misc.XtractIntField(roomUsageTNG, rs)
	(*rs)[roomUsageTNG] = totGuestCnt

	(*rs)[roomUsageNTO] = 1 + misc.XtractIntField(roomUsageNTO, rs)

	// calculate the hours of room usage
	hours, _ := calcDiffTime(ciTime, coTime)
	(*rs)[roomUsageTH] = hours + misc.XtractFloatField(roomUsageTH, rs)

	err = database.DbwUpdate(roomUsageEntity, key, rs)
	if err != nil {
		log.Println("checkout:ERROR: Failed to update room usage for room=", room, " : err=", err)
	}

	return nil
}

func Register(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()

	// check session expiration and authorization
	sessDetails := psession.GetSessDetails(r, "Registration", "Register page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {

		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("register:ERROR: Missing required room param")
			sessDetails.Sess.Message = "Failed to register - missing room number"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
			return
		}
		room := rooms[0]

		regAction := ""
		if regActions, ok := r.URL.Query()["reg"]; ok {
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
		if rates, ok := r.URL.Query()["rate"]; ok {
			rate = rates[0]
		}

		// get the RateClass for the room in order to get the options
		rateMap, err := database.DbwRead(RoomRatesEntity, rate)
		if err != nil {
			log.Println("register:ERROR: Failed to read rate class=", rate, " : err=", err)
			sessDetails.Sess.Message = "Failed to register - bad or missing room rate"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		rrs2, ok := (*rateMap)["Rates"]
		if !ok {
			log.Println("register:ERROR: Failed to get rates: err=", err)
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
			log.Println("register:ERROR: Failed to exec register page for room=", room, " : err=", err)
			sessDetails.Sess.Message = "Failed to make register page: room=" + room
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		}

	} else {
		r.ParseForm()

		fname := r.Form["first_name"]
		lname := r.Form["last_name"]
		duration := r.Form["duration"]
		roomNum := r.Form["room_num"]
		numGuests := r.Form["num_guests"]
		family := r.Form["family"]
		//purchases := r.Form["purchases"]
		//purchaseTotal := r.Form["purchase_total"]

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

		num, _ := strconv.Atoi(numGuests[0])
		(*rs)["NumGuests"] = num

		// calc guest overage
		numExtraGuests := int(0)
		extraRate := ""
		if family[0] == "no" {
			// get the roomSleepsNum for the room
			roomSleepsNum := num
			rMap, err := database.DbwRead(RoomsEntity, roomNum[0])
			if err == nil {
				nsleeps := misc.XtractIntField("NumSleeps", rMap)
				roomSleepsNum = nsleeps
				extraRate, _ = (*rMap)["ExtraRate"].(string)
				(*rs)["ExtraRate"] = extraRate
				fmt.Println("FIX register: read room=", rMap)
			}
			numExtraGuests = num - roomSleepsNum
			if numExtraGuests < 0 {
				numExtraGuests = 0
			}
		}
		(*rs)["NumExtraGuests"] = numExtraGuests

		(*rs)["Duration"] = duration[0]

		nowStr, nowTime := misc.TimeNow()
		(*rs)["CheckinTime"] = nowStr
		checkOutTime, err := CalcCheckoutTime(nowTime, duration[0])
		(*rs)["CheckoutTime"] = checkOutTime

		//(*rs)["Purchases"] = purchases[0]
		//(*rs)["PurchaseTotal"] = purchaseTotal[0]

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
			RoomNum:       roomNum[0],
			Status:        BookedStatus,
			GuestInfo:     guestInfo,
			NumGuests:     num,
			Duration:      duration[0],
			CheckinTime:   nowStr,
			CheckoutTime:  checkOutTime,
			Rate:          (*rs)["Rate"].(string),
			Purchases:     "", // (*rs)["Purchases"].(string),
			PurchaseTotal: "", // (*rs)["PurchaseTotal"].(string),
		}
		updateRoomStatus(roomNum[0], roomState)

		http.Redirect(w, r, "/desk/room_hop?room="+roomNum[0]+"&citime="+nowStr+"&repeat=true", http.StatusFound)
	}
}

func ReportRoomUsage(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Room Usage", "Room Usage page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	if r.Method != "GET" {
		log.Println("ReportRoomUsage:ERROR: bad http method: should only be a GET")
		sessDetails.Sess.Message = "Failed to get room usage"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/room_usage.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("ReportRoomUsage:ERROR: Failed to parse templates: err=", err)
		sessDetails.Sess.Message = "Failed to get all room usage"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	title := `Current Room Usage`
	dbName := roomUsageEntity
	if bkups, ok := r.URL.Query()["bkup"]; ok {
		dbName = staff.ComposeDbName(roomUsageEntity, bkups[0])
		log.Println("ReportRoomUsage: use backup db=", dbName)
		if bkups[0] == "b" {
			title = `Previous Room Usage`
		} else {
			title = `Oldest Room Usage`
		}
	}

	resArray, err := database.DbwReadAll(dbName)
	if err != nil {
		log.Println(`ReportRoomUsage:ERROR: db readall error`, err)
		sessDetails.Sess.Message = "Failed to get all room usage"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	timeStamp := ""
	usageList := make([]RoomUsage, 0)
	for _, v := range resArray {
		vm := v.(map[string]interface{})
		id := ""
		name, exists := vm[roomUsageRN]
		if !exists {
			// check for timestamp record
			name, exists = vm["BackupTime"]
			if exists {
				timeStamp = name.(string)
			}
			continue
		}
		id = name.(string)
		if id == "" {
			// ignore this record
			continue
		}

		totHours := misc.XtractFloatField(roomUsageTH, &vm)

		guestCnt := misc.XtractIntField(roomUsageTNG, &vm)

		numTimesOcc := misc.XtractIntField(roomUsageNTO, &vm)

		rusage := RoomUsage{
			RoomNum:          id,
			TotNumGuests:     guestCnt,
			TotHours:         totHours,
			NumTimesOccupied: numTimesOcc,
		}
		usageList = append(usageList, rusage)
	}

	tblData := RoomUsageTable{
		sessDetails,
		title,
		usageList,
		timeStamp,
	}

	if err = t.Execute(w, &tblData); err != nil {
		log.Println("ReportRoomUsage:ERROR: could not execute template: err=", err)
		sessDetails.Sess.Message = "Failed to report room usage"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
	}
}

func filterRoomUsageFromMap(ruMap map[string]interface{}) *map[string]interface{} {
	id := ""
	name, exists := ruMap[roomUsageRN]
	if !exists {
		return nil
	}
	id = name.(string)
	if id == "" {
		// ignore this record
		return nil
	}

	return &ruMap
}
func cleanupUsage(dbName string) error {
	// remove all entities from specified db
	resArray, err := database.DbwReadAll(dbName)
	if err != nil {
		log.Println(`room.cleanupUsage:ERROR: db readall: err=`, err)
		return err
	}

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		ru := filterRoomUsageFromMap(vm)
		if ru == nil {
			continue
		}
		if err := database.DbwDelete(dbName, ru); err != nil {
			log.Println(`room.cleanupUsage:ERROR: db delete: err=`, err, ` : room-usage=`, ru)
		}
	}
	return nil
}
func copyUsage(fromDB, toDB string) error {
	// copy each entity from fromDB to the toDB
	resArray, err := database.DbwReadAll(fromDB)
	if err != nil {
		log.Println(`room.Usage:ERROR: db readall: err=`, err)
		return err
	}
	for _, v := range resArray {
		vm := v.(map[string]interface{})
		ru := filterRoomUsageFromMap(vm)
		if ru == nil {
			continue
		}
		err = database.DbwUpdate(toDB, (*ru)[roomUsageRN].(string), ru)
		if err != nil {
			log.Println("room.copyUsage:ERROR: Failed to update db=", toDB, " : for room-usage=", ru, " : err=", err)
			return err
		}
	}

	return nil
}
func BackupRoomUsage(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	// check session expiration and authorization
	sessDetails := psession.GetSessDetails(r, "Backup and Reset Room Usage", "Backup and Reset Room Usage page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	toDB := staff.ComposeDbName(roomUsageEntity, "c")
	if err := cleanupUsage(toDB); err != nil {
		log.Println("BackupRoomUsage:ERROR: Failed to cleanup db=", toDB, " : err=", err)
	}
	fromDB := staff.ComposeDbName(roomUsageEntity, "b")
	if err := copyUsage(fromDB, toDB); err != nil {
		log.Println("BackupRoomUsage:ERROR: Failed to copy usage from db=", fromDB, " to=", toDB, " : err=", err)
	}

	bkupTime, err := database.DbwRead(fromDB, "BackupTime")
	if err == nil {
		// write it to the toDB
		database.DbwUpdate(toDB, "BackupTime", bkupTime)
	} else {
		log.Println("BackupRoomUsage:ERROR: Failed to copy backup time from=", fromDB, " to=", toDB, " : err=", err)
	}

	toDB = fromDB
	if err := cleanupUsage(toDB); err != nil {
		log.Println("BackupRoomUsage:ERROR: Failed to cleanup db=", toDB, " : err=", err)
	}
	if err := copyUsage(roomUsageEntity, toDB); err != nil {
		log.Println("BackupRoomUsage:ERROR: Failed to copy usage from db=", roomUsageEntity, " to=", toDB, " : err=", err)
	}
	bkupTime, err = database.DbwRead(roomUsageEntity, "BackupTime")
	if err == nil {
		// write it to the toDB
		database.DbwUpdate(toDB, "BackupTime", bkupTime)
	} else {
		log.Println("BackupRoomUsage:ERROR: Failed to copy backup time from=", fromDB, " to=", toDB, " : err=", err)
	}

	// lastly reset the current room usage
	// 0 the TotHours and Guest count
	resArray, err := database.DbwReadAll(roomUsageEntity)
	if err != nil {
		log.Println(`BackupStaffHours:ERROR: db readall: err=`, err)
		return
	}

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		_, exists := vm[roomUsageRN]
		if !exists {
			continue
		}

		(vm)[roomUsageTH] = float64(0)
		(vm)[roomUsageTNG] = int(0)
		(vm)[roomUsageNTO] = int(0)
		if err := database.DbwUpdate(roomUsageEntity, "", &vm); err != nil {
			log.Println(`BackupStaffHours:ERROR: db update: err=`, err)
		}
	}

	nowStr, _ := misc.TimeNow()

	bkupTime, err = database.DbwRead(roomUsageEntity, "BackupTime")
	if err == nil {
		// write it to the toDB
		(*bkupTime)["BackupTime"] = nowStr
		if err := database.DbwUpdate(roomUsageEntity, "", bkupTime); err != nil {
			log.Println("BackupRoomUsage:ERROR: Failed to update backup time for=", roomUsageEntity, " : err=", err)
		}
	} else {
		tstamp := map[string]interface{}{"BackupTime": nowStr}
		if err := database.DbwUpdate(roomUsageEntity, "BackupTime", &tstamp); err != nil {
			log.Println("BackupRoomUsage:ERROR: Failed to create backup time for=", roomUsageEntity, " : err=", err)
		}
	}

	http.Redirect(w, r, "/manager/report_room_usage", http.StatusFound)
}

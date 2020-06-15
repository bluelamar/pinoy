package room

import (
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bluelamar/pinoy/config"
	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
	"github.com/bluelamar/pinoy/staff"
)

const (
	hopperChoice = "Choose User ID"

	roomHistoryRN = "Room"
	roomHistoryDE = "Desk"
	roomHistoryBH = "Bellhop"
	roomHistoryAC = "Activity"
	roomHistoryTS = "Timestamp"
)

type BellHop struct {
	UserID     string
	Room       string
	TimeStamp  string
	LoginValid bool // true=valid or false
	// key= UserID + ":" + Room + ":" + TimeStamp
}

type BellHopsTable struct {
	*psession.SessionDetails
	BellHops []BellHop
}

type HopperTable struct {
	*psession.SessionDetails
	RoomNum        string
	CheckinTime    string
	Hoppers        []string
	Repeat         string
	Total          string
	OldCost        string
	MonetarySymbol string
	CheckoutTime   string
	OverCost       string
	TotGuests      string
	ExtraGuests    string
}

// currently clocked in
type HopShift struct {
	UserID  string
	ClockIn string
}

type HopShiftTable struct {
	*psession.SessionDetails
	Shift []HopShift
}

type RoomHistory struct {
	Room      string
	Desk      string
	Bellhop   string
	Activity  string // describes checkin, checkout, update, etc
	Timestamp string
}

type RoomHistoryTable struct {
	*psession.SessionDetails
	Title        string
	RoomHistList []RoomHistory
	BackupTime   string
}

// make list of hop names
func getHoppers() ([]string, error) {
	// do a Find of staff that are Role == ROLE_HOP
	hoppers := make([]string, 0)
	hlist, err := database.DbwFind(staff.StaffEntity, "Role", psession.ROLE_HOP)
	if err != nil {
		return hoppers, err
	}

	hoppers = append(hoppers, hopperChoice)

	for _, v := range hlist {
		vm := v.(map[string]interface{})
		id := ""
		if name, exists := vm["name"]; exists {
			id = name.(string)
		} else {
			continue
		}
		hoppers = append(hoppers, id)
	}
	return hoppers, nil
}

func RoomHop(w http.ResponseWriter, r *http.Request) {
	sessDetails := psession.GetSessDetails(r, "Room Bell Hop", "Bell Hop page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	misc.IncrRequestCnt()
	if r.Method == "GET" {

		room := ""
		if rooms, ok := r.URL.Query()["room"]; ok {
			room = rooms[0]
		} else {
			log.Println("room_hop: missing Url param: room")
		}

		citime := ""
		if citimes, ok := r.URL.Query()["citime"]; ok {
			citime = citimes[0]
		} else {
			log.Println("room_hop: missing check in time param: citime")
		}

		repeat := ""
		if repeats, ok := r.URL.Query()["repeat"]; ok {
			repeat = repeats[0]
		} else {
			log.Println("room_hop: missing repeat param: repeat")
		}

		total := ""
		if totals, ok := r.URL.Query()["total"]; ok {
			total = totals[0]
		} else {
			log.Println("room_hop: missing room total param: total")
		}

		oldCost := ""
		if oldcosts, ok := r.URL.Query()["oldcost"]; ok {
			oldCost = oldcosts[0]
		} else {
			log.Println("room_hop: missing room old cost param: oldcost")
		}

		totGuests := ""
		if totGuestCnt, ok := r.URL.Query()["totguests"]; ok {
			totGuests = totGuestCnt[0]
		}

		extraGuests := ""
		if extraGuestCnt, ok := r.URL.Query()["extguests"]; ok {
			extraGuests = extraGuestCnt[0]
		}

		hoppers, err := getHoppers()
		if err != nil {
			log.Println("room_hop:ERROR: Failed to retrieve bell hops list: err=", err)
			sessDetails.Sess.Message = "No bell hops"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/room_hop.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("room_hop:ERROR: Failed to parse template: err=", err)
			sessDetails.Sess.Message = "Error with bell hops"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		monSymbol := config.GetConfig().MonetarySymbol
		regData := HopperTable{
			sessDetails,
			room,
			citime,
			hoppers,
			repeat,
			total,
			oldCost,
			monSymbol,
			"",
			"",
			totGuests,
			extraGuests,
		}
		err = t.Execute(w, regData)
		if err != nil {
			log.Println("room_hop:ERROR: Failed to execute template : err=", err)
			sessDetails.Sess.Message = "Error with bell hops"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}
	} else {
		r.ParseForm()

		hopper := r.Form["hopper"]
		hopperID := r.Form["user_id"]
		bellHopPin := r.Form["bell_hop_pin"]
		room_num := r.Form["room_num"]
		repeat := r.Form["repeat"] // if true then can repeat upon failure
		oldCost := r.Form["oldcost"]
		total := r.Form["total"]
		totGuests := r.Form["totguests"]
		extraGuests := r.Form["extguests"]

		if strings.Compare(hopper[0], hopperChoice) == 0 {
			hopper[0] = hopperID[0]
		}
		umap, err := database.DbwRead(staff.StaffEntity, hopper[0])
		if err != nil {
			log.Println("room_hop:ERROR: Failed to read room attendent=", hopper[0], " : err=", err)
			sessDetails.Sess.Message = "Error with bell hop " + hopper[0]
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}
		passwd, ok := (*umap)["Pwd"]
		if !ok {
			log.Println("room_hop:ERROR: Failed to check passwd for room attendent=", hopper[0], " : room=", room_num[0], " : err=", err)
			sessDetails.Sess.Message = "PIN check Error with bell hop " + hopper[0]
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		nowStr, ctime := misc.TimeNow()
		activity := "Customer checkin: "
		if totGuests[0] != "" {
			activity = activity + "Total-num-guests=" + totGuests[0]
		}
		if extraGuests[0] != "" {
			activity = activity + " : Extra-num-guests=" + extraGuests[0]
		}
		hist := RoomHistory{
			Room:      room_num[0],
			Desk:      sessDetails.Sess.User,
			Bellhop:   hopper[0],
			Activity:  activity,
			Timestamp: nowStr,
		}
		rs := xlateHistToMap(&hist, ctime)

		pwd := config.HashIt(bellHopPin[0]) // use hash for user password
		if strings.Compare(pwd, passwd.(string)) != 0 {
			// invalid match
			log.Println("room_hop: room attendent pin not a match=", hopper[0], " : room=", room_num[0])
			misc.IncrFailedLoginCnt()
			if strings.Compare(repeat[0], "true") == 0 {
				activity = activity + " : Bellhop Failed login"
				(*rs)["Activity"] = activity
				err = database.DbwUpdate(roomHistoryEntity, (*rs)["key"].(string), rs)
				if err != nil {
					log.Println("room_hop:ERROR: Failed to update room history=", activity, " : err=", err)
				}
				http.Redirect(w, r, "/desk/room_hop?room="+room_num[0]+"&citime="+nowStr+"&repeat=false&total="+total[0]+"&oldcost="+oldCost[0], http.StatusFound)
			} else {

				sessDetails.Sess.Message = "PIN error again for bell hop " + hopper[0]
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			}
			return
		}

		// check if hop is clocked in, if not, clock them in and show warning
		clockedIn, _, _ := staff.IsUserLoggedIn(hopper[0])
		if clockedIn == false {
			misc.IncrLoginCnt()
			sessAttrs := psession.PinoySession{
				User:      hopper[0],
				Role:      psession.ROLE_HOP,
				Auth:      false,
				SessID:    "",
				CsrfToken: "",
				CsrfParam: "",
				Message:   "",
			}
			go staff.UpdateEmployeeHours(hopper[0], true, 8, &sessAttrs)
		}

		// add to room history
		activity = activity + " : Bellhop is clocked-in=" + strconv.FormatBool(clockedIn)
		(*rs)["Activity"] = activity
		err = database.DbwUpdate(roomHistoryEntity, (*rs)["key"].(string), rs)
		if err != nil {
			log.Println("room_hop:ERROR: Failed to update room history=", activity, " : err=", err)
			sessDetails.Sess.Message = "Failed to update room history: room=" + room_num[0]
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

func RoomCheckout(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()

	// check session expiration and authorization
	sessDetails := psession.GetSessDetails(r, "RoomCheckout", "RoomCheckout page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {

		room := ""
		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("RoomCheckout:ERROR: Missing required room param")
			sessDetails.Sess.Message = "Failed to check out - missing room number"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}
		room = rooms[0]

		charge := ""
		if charges, ok := r.URL.Query()["charge"]; ok {
			charge = charges[0]
		}
		over := ""
		if overs, ok := r.URL.Query()["over"]; ok {
			over = overs[0]
		}
		origCost := ""
		if origCosts, ok := r.URL.Query()["origCost"]; ok {
			origCost = origCosts[0]
		}
		checkin := ""
		if checkins, ok := r.URL.Query()["checkin"]; ok {
			checkin = checkins[0]
		}
		checkout := ""
		if checkouts, ok := r.URL.Query()["checkout"]; ok {
			checkout = checkouts[0]
		}
		totGuests := ""
		if totGuestCnt, ok := r.URL.Query()["totguests"]; ok {
			totGuests = totGuestCnt[0]
		}

		extraGuests := ""
		if extraGuestCnt, ok := r.URL.Query()["extguests"]; ok {
			extraGuests = extraGuestCnt[0]
		}

		hoppers, err := getHoppers()
		if err != nil {
			log.Println("RoomCheckout:ERROR: Failed to retrieve bell hops list: err=", err)
		}

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/room_co.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("RoomCheckout:ERROR: Failed to parse template: err=", err)
			sessDetails.Sess.Message = "Error checkout with bell hops"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}
		monSymbol := config.GetConfig().MonetarySymbol
		hopTbl := HopperTable{
			sessDetails,
			room,
			checkin,
			hoppers,
			"",
			charge,
			origCost,
			monSymbol,
			checkout,
			over,
			totGuests,
			extraGuests,
		}
		err = t.Execute(w, hopTbl)
		if err != nil {
			log.Println("RoomCheckout:ERROR: Failed to execute template : err=", err)
			sessDetails.Sess.Message = "Error with bell hops"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}
	} else {
		r.ParseForm()

		hopper := r.Form["hopper"]
		hopperID := r.Form["user_id"]
		room_num := r.Form["room_num"]
		ciTime := r.Form["citime"]
		coTime := r.Form["cotime"]
		oldCost := r.Form["oldcost"]
		overCost := r.Form["overcost"]
		total := r.Form["total"]

		if strings.Compare(hopper[0], hopperChoice) == 0 {
			hopper[0] = hopperID[0]
		}

		// add to room history
		nowStr, ctime := misc.TimeNow()
		activity := "Customer checkout: checkout=" + coTime[0] + " : checkin=" + ciTime[0] + " : Total=" + total[0]
		if len(oldCost) > 0 {
			activity = activity + " : Original-cost=" + oldCost[0]
		}
		if len(overCost) > 0 {
			activity = activity + " : Overage-cost=" + overCost[0]
		}
		hist := RoomHistory{
			Room:      room_num[0],
			Desk:      sessDetails.Sess.User,
			Bellhop:   hopper[0],
			Activity:  activity,
			Timestamp: nowStr,
		}
		rs := xlateHistToMap(&hist, ctime)

		err := database.DbwUpdate(roomHistoryEntity, (*rs)["key"].(string), rs)
		if err != nil {
			log.Println("RoomCheckout:ERROR: Failed to update room history=", activity, " : err=", err)
			sessDetails.Sess.Message = "Failed to update room history upon checkout: room=" + room_num[0]
			err = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

func xlateHistToMap(hist *RoomHistory, curTime time.Time) *map[string]interface{} {
	rs := make(map[string]interface{})
	rs["Room"] = hist.Room
	rs["Desk"] = hist.Desk
	rs["Bellhop"] = hist.Bellhop
	rs["Activity"] = hist.Activity
	rs[roomHistoryTS] = hist.Timestamp
	var n int64 = int64(curTime.Nanosecond() / 1000000) // set to millis
	rs["key"] = hist.Room + ":" + hist.Timestamp + ":" + strconv.FormatInt(n, 10)
	return &rs
}

func ReportRoomHistory(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Room History", "Room History page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	if r.Method != "GET" {
		log.Println("ReportRoomHistory:ERROR: bad http method: should only be a GET")
		sessDetails.Sess.Message = "Failed to get room usage"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/room_history.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("ReportRoomHistory:ERROR: Failed to parse templates: err=", err)
		sessDetails.Sess.Message = "Failed to get all room history"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		return
	}

	title := `Current Room History`
	dbName := roomHistoryEntity
	if bkups, ok := r.URL.Query()["bkup"]; ok {
		dbName = staff.ComposeDbName(roomHistoryEntity, bkups[0])
		log.Println("ReportRoomHistory: use backup db=", dbName)
		if bkups[0] == "b" {
			title = `Previous Room History`
		} else {
			title = `Oldest Room History`
		}
	}

	resArray, err := database.DbwReadAll(dbName)
	if err != nil {
		log.Println(`ReportRoomHistory:ERROR: db readall error`, err)
		sessDetails.Sess.Message = "Failed to get all room history"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		return
	}

	timeStamp := ""
	usageList := make([]RoomHistory, 0)

	sort.SliceStable(resArray, func(i, j int) bool {
		vmi := resArray[i].(map[string]interface{})
		if _, exists := vmi[roomHistoryRN]; exists {
			// ok record we want, but sort on timestamp
			if tsi, exists := vmi[roomHistoryTS]; exists {
				timei, _ := tsi.(string)
				var timej string
				vmj := resArray[j].(map[string]interface{})
				if tsj, exists := vmj[roomHistoryTS]; exists {
					timej, _ = tsj.(string)
					return strings.Compare(timei, timej) > 0
				}
			}
		}

		return true
	})

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		id := ""
		name, exists := vm[roomHistoryRN]
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

		desk, _ := vm[roomHistoryDE].(string)
		bellHop, _ := vm[roomHistoryBH].(string)
		activity, _ := vm[roomHistoryAC].(string)
		ts, _ := vm[roomHistoryTS].(string)

		rusage := RoomHistory{
			Room:      id,
			Desk:      desk,
			Bellhop:   bellHop,
			Activity:  activity,
			Timestamp: ts,
		}
		usageList = append(usageList, rusage)
	}

	tblData := RoomHistoryTable{
		sessDetails,
		title,
		usageList,
		timeStamp,
	}

	if err = t.Execute(w, &tblData); err != nil {
		log.Println("ReportRoomHistory:ERROR: could not execute template: err=", err)
		sessDetails.Sess.Message = "Failed to report room history"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
	}
}

func BackupRoomHistory(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	// check session expiration and authorization
	sessDetails := psession.GetSessDetails(r, "Backup and Reset Room History", "Backup and Reset Room History page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	toDB := staff.ComposeDbName(roomHistoryEntity, "c")
	if err := misc.CleanupDbUsage(toDB, roomHistoryRN); err != nil {
		log.Println("BackupRoomHistory:ERROR: Failed to cleanup db=", toDB, " : err=", err)
	}
	fromDB := staff.ComposeDbName(roomHistoryEntity, "b")
	if err := misc.CopyDbUsage(fromDB, toDB, roomHistoryRN); err != nil {
		log.Println("BackupRoomHistory:ERROR: Failed to copy history from db=", fromDB, " to=", toDB, " : err=", err)
	}

	bkupTime, err := database.DbwRead(fromDB, "BackupTime")
	if err == nil {
		// write it to the toDB
		database.DbwUpdate(toDB, "BackupTime", bkupTime)
	} else {
		log.Println("BackupRoomHistory:ERROR: Failed to copy backup time from=", fromDB, " to=", toDB, " : err=", err)
	}

	toDB = fromDB
	if err := misc.CleanupDbUsage(toDB, roomHistoryRN); err != nil {
		log.Println("BackupRoomHistory:ERROR: Failed to cleanup db=", toDB, " : err=", err)
	}
	if err := misc.CopyDbUsage(roomHistoryEntity, toDB, roomHistoryRN); err != nil {
		log.Println("BackupRoomHistory:ERROR: Failed to copy history from db=", roomHistoryEntity, " to=", toDB, " : err=", err)
	}
	bkupTime, err = database.DbwRead(roomHistoryEntity, "BackupTime")
	if err == nil {
		// write it to the toDB
		database.DbwUpdate(toDB, "BackupTime", bkupTime)
	} else {
		log.Println("BackupRoomHistory:ERROR: Failed to copy backup time from=", fromDB, " to=", toDB, " : err=", err)
	}

	// lastly reset the current room usage
	resArray, err := database.DbwReadAll(roomHistoryEntity)
	if err != nil {
		log.Println(`BackupRoomHistory:ERROR: db readall: err=`, err)
		return
	}

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		_, exists := vm[roomHistoryRN]
		if !exists {
			continue
		}

		if err := database.DbwDelete(roomHistoryEntity, &vm); err != nil {
			log.Println(`BackupRoomHistory:ERROR: db update: err=`, err)
		}
	}

	nowStr, _ := misc.TimeNow()

	bkupTime, err = database.DbwRead(roomHistoryEntity, "BackupTime")
	if err == nil {
		// write it to the toDB
		(*bkupTime)["BackupTime"] = nowStr
		if err := database.DbwUpdate(roomHistoryEntity, "", bkupTime); err != nil {
			log.Println("BackupRoomHistory:ERROR: Failed to update backup time for=", roomHistoryEntity, " : err=", err)
		}
	} else {
		tstamp := map[string]interface{}{"BackupTime": nowStr}
		if err := database.DbwUpdate(roomHistoryEntity, "BackupTime", &tstamp); err != nil {
			log.Println("BackupRoomHistory:ERROR: Failed to create backup time for=", roomHistoryEntity, " : err=", err)
		}
	}

	http.Redirect(w, r, "/manager/report_room_history", http.StatusFound)
}

package shift

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/bluelamar/pinoy/config"
	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
)

const (
	shiftInfoEntity = "shiftinfo"
	ShiftItemEntity = "shiftitem"
)

type ShiftItem struct {
	Shift     int
	StartTime int
	EndTime   int
}

type ShiftData struct {
	*psession.SessionDetails
	Shifts []ShiftItem
}

type ShiftDailyItem struct {
	Shift    int    // shift number
	Day      int    // day of the year
	Shiftday string // indexed for easy lookup "<shift>-<day>"
	Time     string // date-time
	Type     string // ie. "room" or "food"
	Subtype  string // for rooms use "Rate Class"; for food use item name ie. "coke"
	Subtype2 string // for rooms use "Duration" or "Overtime"; for food  use "size"
	Volume   int    // for food this is the number of items, for room this is number times Duration/Overtime chosen
	Total    float64
}

type ShiftDailyData struct {
	*psession.SessionDetails
	Shift        int
	Month        string
	Day          int
	CurShift     int
	LastDay      int
	DayOfYear    int
	CurDayOfYear int
	Rooms        []ShiftDailyItem
	Food         []ShiftDailyItem
}

// Cleanup is used to cleanup the database of old shift items - used with the misc.CleanerInterface
type Cleanup struct {
}

var shiftList []ShiftItem

var monthList []string

func init() {
	monthList = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
}

// NewCleaner used in main to to setup the cleaner impl slice
func NewCleaner() *Cleanup {
	return &Cleanup{}
}

// Cleanup is called by main timer
func (c *Cleanup) Cleanup(cfg *config.PinoyConfig, now time.Time) {
	// FIX TODO need to cleanup the DB
	// if day==7, in January so delete all days >= (365-7)
	// else clean up everything up to the previous 7 days
	log.Println("shiftinfo: Start cleanup now=", now, " : day-of-year=", now.YearDay())
	// FIX gather entities where the field Day == dayOfYear, find using index for Day, should get room and food entities for each shift
	// delete each one
}

// BuildShiftList will set up the shift info used for the shift reporting
func BuildShiftList(refresh bool) []ShiftItem {
	if refresh == false && shiftList != nil {
		return shiftList
	}
	sis, err := database.DbwReadAll(shiftInfoEntity)
	if err != nil {
		log.Println("shiftinfo:buildShiftList: failed to read and load shifts: err=", err)
		return nil
	}
	shifts := make([]ShiftItem, len(sis))
	for k, v := range sis {
		val := v.(map[string]interface{})
		shiftID := misc.XtractIntField("Shift", &val)
		startTime := misc.XtractIntField("StartTime", &val)
		endTime := misc.XtractIntField("EndTime", &val)
		shift := ShiftItem{
			Shift:     shiftID,
			StartTime: startTime,
			EndTime:   endTime,
		}
		shifts[k] = shift
	}

	sort.SliceStable(shifts, func(i, j int) bool {
		si := shifts[i]
		sj := shifts[j]
		return si.Shift < sj.Shift
	})

	shiftList = shifts
	return shiftList
}

// AdjustDayForXOverShift returns day-of-year adjusted for shift that crosses over midnight
func AdjustDayForXOverShift(year, dayOfYear, hourOfDay, shiftNum int) int {
	lastShift := shiftList[len(shiftList)-1]
	if hourOfDay < lastShift.StartTime {
		return dayOfYear // day-of-year is already on right day
	}
	// if leap year then 366 days rather than 365
	leapYear := year%4 == 0
	divisor := 366
	if leapYear {
		divisor = 367
	}
	day := (dayOfYear + 1) % divisor // adjust to next day
	if day == 0 {
		return 1
	}
	return day
}

// CalcShift returns the day-of-the-year, hour-of-the-day, shift-number
func CalcShift() (int, int, int, time.Time) {
	// calculate day of the year
	_, t := misc.TimeNow()
	return calcShift(t)
}

func calcShift(t time.Time) (int, int, int, time.Time) {
	dayOfTheYear := t.YearDay()

	// calculate shift number acording to the current time
	hour := t.Hour()

	// look thru list of shifts to determine which shift this hour lies in
	shift := -1
	if shiftList == nil {
		BuildShiftList(true)
	}
	// loop thru shift list to find shift the hour lies within
	if shiftList != nil {
		for _, v := range shiftList {
			if hour >= v.StartTime && hour < v.EndTime {
				shift = v.Shift
				break
			} else if v.StartTime > v.EndTime { // crossover midnight
				if hour >= 0 && hour < v.EndTime || hour >= v.StartTime && hour < 24 {
					shift = v.Shift
					break
				}
			}
		}
	}

	return dayOfTheYear, hour, shift, t
}

// MakeShiftMap is useful for features to update shift data entries
func MakeShiftMap(shiftNum, dayOfYear int, shiftID, itemType, subType, subType2, nowStr string) *map[string]interface{} {
	rs := make(map[string]interface{})
	(rs)["Shift"] = shiftNum   // shift number
	(rs)["Day"] = dayOfYear    // day of the year
	(rs)["Shiftday"] = shiftID // indexed for easy lookup "<shift>-<day>"
	(rs)["Type"] = itemType
	(rs)["Subtype"] = subType
	(rs)["Subtype2"] = subType2
	(rs)["Time"] = nowStr
	(rs)["Volume"] = 1
	return &rs
}

// Info is the REST API to return the shift configuration
func Info(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "ShiftInfo", "Shift Info page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method != "GET" {
		log.Println("shiftinfo: bad http method: should only be a GET")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/shiftinfo.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("shiftinfo:ERROR: Failed to parse template: err=", err)
		sessDetails.Sess.Message = "Internal error"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		return
	}

	shifts := BuildShiftList(false)
	if shifts == nil {
		log.Println("shiftinfo:ERROR: Failed to read shift info: err=", err)
		sessDetails.Sess.Message = `Failed to read shift info`
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		return
	}

	shiftData := ShiftData{
		sessDetails,
		shifts,
	}
	err = t.Execute(w, &shiftData)
	if err != nil {
		log.Println("shiftinfo:ERROR: Failed to exec template: err=", err)
		sessDetails.Sess.Message = `Internal error for shift info`
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
	}
}

// UpdateShiftInfo Info is the REST API to update the shift info
func UpdateShiftInfo(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "UpdateShiftInfo", "Update Shift Info page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {

		// update values: add, delete, update
		update := ""
		if updates, ok := r.URL.Query()["update"]; ok {
			update = updates[0]
		}
		if update == "" {
			log.Println("upd_shiftinfo: required update parameter is missing")
			sessDetails.Sess.Message = `Internal error for shift info`
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		shiftID := ""
		if shiftIDs, ok := r.URL.Query()["shift"]; ok {
			shiftID = shiftIDs[0]
		}

		var sMap *map[string]interface{}
		if update == "delete" || update == "update" {
			// read in shift from db
			var err error
			sMap, err = database.DbwRead(shiftInfoEntity, shiftID)
			if err != nil {
				log.Println("upd_shiftinfo: Invalid shift-id=", shiftID, " : err=", err)
				sessDetails.Sess.Message = `Internal error for shift info`
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
				return
			}
		}

		if update == "delete" {
			if sMap == nil {
				log.Println("upd_shiftinfo:delete: Shift info missing")
				sessDetails.Sess.Message = `Shift info missing`
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
				return
			}

			if err := database.DbwDelete(shiftInfoEntity, sMap); err != nil {
				sessDetails.Sess.Message = "Failed to delete shift: " + shiftID
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
				return
			}

			BuildShiftList(true) // rebuild list

			http.Redirect(w, r, "/manager/shiftinfo", http.StatusFound)
			return
		}

		// user wants to add or update existing shift
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_shiftinfo.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("upd_shiftinfo:ERROR: Failed to parse template: err=", err)
			sessDetails.Sess.Message = "Failed to Update shift: " + shiftID
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		var shiftItem ShiftItem
		if sMap != nil {
			shiftid := misc.XtractIntField("Shift", sMap)
			stime := misc.XtractIntField("StartTime", sMap)
			etime := misc.XtractIntField("EndTime", sMap)
			shiftItem = ShiftItem{
				Shift:     shiftid,
				StartTime: stime,
				EndTime:   etime,
			}
		} else {
			shiftItem = ShiftItem{
				Shift:     1,
				StartTime: 0,
				EndTime:   0,
			}
		}

		shifts := make([]ShiftItem, 1)
		shifts[0] = shiftItem
		updData := ShiftData{
			sessDetails,
			shifts,
		}

		err = t.Execute(w, updData)
		if err != nil {
			log.Println("upd_shiftinfo:ERROR: Failed to exec template: err=", err)
			sessDetails.Sess.Message = `Internal error in Update shift info`
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		}
	} else {
		r.ParseForm()

		shiftIDs, _ := r.Form["shiftid"]
		startHours, _ := r.Form["starthour"]
		endHours, _ := r.Form["endhour"]

		// validate incoming form fields
		if len(shiftIDs[0]) == 0 || len(startHours[0]) == 0 || len(endHours[0]) == 0 {
			log.Println("upd_shiftinfo:POST: Missing form data")
			sessDetails.Sess.Message = `Missing required fields in Update Shift Info`
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		key := ""
		var sMap *map[string]interface{}
		var err error
		sMap, err = database.DbwRead(shiftInfoEntity, shiftIDs[0])
		if err != nil {
			log.Println("upd_shiftinfo:POST: must be new shift record: num=", shiftIDs[0], " : err=", err)
			sm := make(map[string]interface{})
			sMap = &sm

			(*sMap)["Shift"] = shiftIDs[0]
			(*sMap)["StartTime"] = startHours[0]
			(*sMap)["EndTime"] = endHours[0]
			key = shiftIDs[0] // create rather than update
		} else {
			(*sMap)["StartTime"] = startHours[0]
			(*sMap)["EndTime"] = endHours[0]
		}

		err = database.DbwUpdate(shiftInfoEntity, key, sMap)
		if err != nil {
			log.Println("upd_shiftinfo:POST:ERROR: Failed to create or update shift=", shiftIDs[0], " :err=", err)
			sessDetails.Sess.Message = "Failed to create or update shift=" + shiftIDs[0]
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		BuildShiftList(true) // rebuild list

		http.Redirect(w, r, "/manager/shiftinfo", http.StatusFound)
	}
}

// Info is the REST API to return the shift configuration
func DailyInfo(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "DailyInfo", "Daily Shift Info page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method != "GET" {
		log.Println("shift:DailyInfo: bad http method: should only be a GET")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	beforeShift := "" // if is set in query, then calc the day before it
	if bshifts, ok := r.URL.Query()["beforeShift"]; ok {
		beforeShift = bshifts[0]
	}
	beforeDay := "" // if is set in query, then calc the day before it
	if days, ok := r.URL.Query()["beforeDay"]; ok {
		beforeDay = days[0]
	}

	befShift := -1
	if beforeShift != "" {
		// get the int for the shift and decrement
		if num, err := strconv.Atoi(beforeShift); err == nil {
			befShift = num
		}
	}
	befDay := -1
	if beforeDay != "" {
		// get the int for the day and decrement
		if num, err := strconv.Atoi(beforeDay); err == nil {
			befDay = num
		}
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/shiftdailyinfo.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("shift:DailyInfo:ERROR: Failed to parse template: err=", err)
		sessDetails.Sess.Message = "Internal error"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		return
	}

	// get the daily shift items
	shiftRoomItems := make([]ShiftDailyItem, 0)
	shiftFoodItems := make([]ShiftDailyItem, 0)

	dayOfYear, hourOfDay, shiftNum, timeNow := CalcShift()
	curShift := shiftNum
	curDayOfYear := dayOfYear
	//log.Println("FIX shift: doy=", dayOfYear, " shiftnum=", shiftNum, " hour=", hourOfDay, " befshift=", befShift, " befday=", befDay)
	if befDay != -1 {
		dayOfYear = befDay
	}

	lastShiftStartDay := false
	if befShift != -1 {
		lastShift := shiftList[len(shiftList)-1]
		if befShift > 1 {
			shiftNum = befShift - 1
		} else {
			// 1st shift: previous shift will be the last shift of the previous day
			shiftNum = lastShift.Shift
			dayOfYear--
		}
		if shiftNum == lastShift.Shift {
			lastShiftStartDay = true
		}
	} else {
		dayOfYear = AdjustDayForXOverShift(timeNow.Year(), dayOfYear, hourOfDay, shiftNum)
	}

	t2 := time.Date(timeNow.Year(), timeNow.Month(), dayOfYear, hourOfDay, 0, 0, 0, misc.GetLocale())

	//log.Println("FIX shift: doy=", dayOfYear, " shiftnum=", shiftNum)

	shiftDay := fmt.Sprintf("%d-%d", dayOfYear, shiftNum)
	silist, err := database.DbwFind(ShiftItemEntity, "Shiftday", shiftDay)
	if err != nil {
		log.Println("shift:DailyInfo:ERROR: Failed to find shift daily items for ", shiftDay)
	} else {
		// fill up the shifts
		for _, v := range silist {
			vm := v.(map[string]interface{})
			sdi := ShiftDailyItem{}
			shift := misc.XtractIntField("Shift", &vm)
			sdi.Shift = shift
			day := misc.XtractIntField("Day", &vm)
			sdi.Day = day
			sday, _ := vm["ShiftDay"].(string)
			sdi.Shiftday = sday
			t, _ := vm["Time"].(string)
			sdi.Time = t
			sitype, _ := vm["Type"].(string)
			sdi.Type = sitype
			subtype, _ := vm["Subtype"].(string)
			sdi.Subtype = subtype
			subtype2, _ := vm["Subtype2"].(string)
			sdi.Subtype2 = subtype2
			vol := misc.XtractIntField("Volume", &vm)
			sdi.Volume = vol
			tot := misc.XtractFloatField("Total", &vm)
			sdi.Total = tot

			if sitype == "room" {
				shiftRoomItems = append(shiftRoomItems, sdi)
			} else {
				shiftFoodItems = append(shiftFoodItems, sdi)
			}
		}
	}

	// FIX TODO sort the lists

	month, dom := calcMonthDayOfMonth(t2)
	prevDay := dom
	if lastShiftStartDay {
		prevDay--
	}

	shiftData := ShiftDailyData{
		sessDetails,
		shiftNum,
		month,
		dom,
		curShift,
		prevDay,
		dayOfYear,
		curDayOfYear,
		shiftRoomItems,
		shiftFoodItems,
	}
	err = t.Execute(w, &shiftData)
	if err != nil {
		log.Println("shift:DailyInfo:ERROR: Failed to exec template: err=", err)
		sessDetails.Sess.Message = `Internal error for shift info`
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
	}
}

func calcMonthDayOfMonth(t time.Time) (string, int) {
	m := t.Month()
	month := monthList[m-1]
	dom := t.Day()

	return month, dom
}

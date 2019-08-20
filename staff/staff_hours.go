package staff

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
)

const (
	StaffHoursEntity = "staff_hours"
	DateTimeLongForm = "2006-01-02 15:04"
)

type EmpHours struct {
	UserID           string
	LastClockinTime  string
	LastClockoutTime string
	ExpectedHours    int // hours expected to work on the shift when clocks in
	// -- default per Role? desk=12, etc
	TotalExpectedHours float64 // sum the total hours using expected hours
	TotalHours         float64 // gets updated when employee clocks out
	ClockInCnt         int
	ClockOutCnt        int
}
type EmpHoursTable struct {
	*psession.SessionDetails
	Title      string
	StaffHours []EmpHours
	BackupTime string
}

type UpdateEmpHours struct {
	*psession.SessionDetails
	Emp EmpHours
}

func ComposeDbName(prefix, suffix string) string {
	dbName := prefix + "_bkup_" + suffix
	return dbName
}

func ReportStaffHours(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.Get_sess_details(r, "Staff Hours", "Staff Hours page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	if r.Method != "GET" {
		log.Println("ReportStaffHours:ERROR: bad http method: should only be a GET")
		sessDetails.Sess.Message = "Failed to get staff hours"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/staff_hours.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("ReportStaffHours:ERROR: Failed to parse templates: err=", err)
		sessDetails.Sess.Message = "Failed to get all staff hours"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	title := `Current Staff Hours`
	dbName := StaffHoursEntity
	if bkups, ok := r.URL.Query()["bkup"]; ok {
		dbName = ComposeDbName(StaffHoursEntity, bkups[0])
		log.Println("ReportStaffHours: use backup db=", dbName)
		if bkups[0] == "b" {
			title = `Previous Staff Hours`
		} else {
			title = `Oldest Staff Hours`
		}
	}

	resArray, err := database.DbwReadAll(dbName)
	if err != nil {
		log.Println(`ReportStaffHours:ERROR: db readall error`, err)
		sessDetails.Sess.Message = "Failed to get all staff hours"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	timeStamp := ""
	emps := make([]EmpHours, 0)
	deskRole := sessDetails.Sess.Role == psession.ROLE_DSK
	for _, v := range resArray {
		vm := v.(map[string]interface{})
		id := ""
		name, exists := vm["UserID"]
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

		if deskRole {
			// only show the hoppers
			entity, err := database.DbwRead(StaffEntity, id)
			if err != nil {
				log.Println("report_staff_hours:ERROR: desk role: Failed to read user=", id, " : err=", err)
				continue
			}

			if role, exists := (*entity)["Role"]; exists {
				if role.(string) != psession.ROLE_HOP {
					continue // Desk can only update hours for bell hops
				}
			}
		}

		clockin := ""
		if name, exists = vm["LastClockinTime"]; exists {
			clockin = name.(string)
		}

		clockout := ""
		if name, exists = vm["LastClockoutTime"]; exists {
			clockout = name.(string)
		}

		expHours := 0
		if num, exists := vm["ExpectedHours"]; exists {
			expHours = int(num.(float64))
		}

		totExpHours := float64(0)
		if num, exists := vm["TotalExpectedHours"]; exists {
			totExpHours = num.(float64)
		}

		totHours := float64(0)
		if num, exists := vm["TotalHours"]; exists {
			totHours = num.(float64)
		}

		ciCnt := int(0)
		if num, exists := vm["ClockInCnt"]; exists {
			ciCnt = int(num.(float64))
		}
		coCnt := int(0)
		if num, exists := vm["ClockOutCnt"]; exists {
			coCnt = int(num.(float64))
		}

		emp := EmpHours{
			UserID:             id,
			LastClockinTime:    clockin,
			LastClockoutTime:   clockout,
			ExpectedHours:      expHours,
			TotalExpectedHours: totExpHours,
			TotalHours:         totHours,
			ClockInCnt:         ciCnt,
			ClockOutCnt:        coCnt,
		}
		emps = append(emps, emp)
	}

	tblData := EmpHoursTable{
		sessDetails,
		title,
		emps,
		timeStamp,
	}
	if err = t.Execute(w, &tblData); err != nil {
		log.Println("report_staff_hours:ERROR: could not execute template: err=", err)
		sessDetails.Sess.Message = "Failed to report staff hours"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
	}

}

func getEmpFromMap(empMap map[string]interface{}) *map[string]interface{} {

	id := ""
	name, exists := empMap["UserID"]
	if !exists {
		return nil
	}
	id = name.(string)
	if id == "" {
		// ignore this record
		return nil
	}

	clockin := ""
	if name, exists = empMap["LastClockinTime"]; exists {
		clockin = name.(string)
	}

	clockout := ""
	if name, exists = empMap["LastClockoutTime"]; exists {
		clockout = name.(string)
	}

	expHours := 0
	if num, exists := empMap["ExpectedHours"]; exists {
		expHours = int(num.(float64))
	}

	totExpHours := float64(0)
	if num, exists := empMap["TotalExpectedHours"]; exists {
		totExpHours = num.(float64)
	}

	totHours := float64(0)
	if num, exists := empMap["TotalHours"]; exists {
		totHours = num.(float64)
	}

	ciCnt := int(0)
	if num, exists := empMap["ClockInCnt"]; exists {
		ciCnt = int(num.(float64))
	}
	coCnt := int(0)
	if num, exists := empMap["ClockOutCnt"]; exists {
		coCnt = int(num.(float64))
	}

	emp := map[string]interface{}{
		//emp := EmpHours{
		"UserID":             id,
		"LastClockinTime":    clockin,
		"LastClockoutTime":   clockout,
		"ExpectedHours":      expHours,
		"TotalExpectedHours": totExpHours,
		"TotalHours":         totHours,
		"ClockInCnt":         ciCnt,
		"ClockOutCnt":        coCnt,
	}
	return &emp
}

func cleanupHours(dbName string) error {
	// remove all entities from specified db
	resArray, err := database.DbwReadAll(dbName)
	if err != nil {
		log.Println(`cleanupHours:ERROR: db readall: err=`, err)
		return err
	}

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		emp := getEmpFromMap(vm)
		if emp == nil {
			continue
		}
		if err := database.DbwDelete(dbName, emp); err != nil {
			log.Println(`cleanupHours:ERROR: db delete: err=`, err)
		}
	}
	return nil
}
func copyHours(fromDB, toDB string) error {
	// copy each entity from fromDB to the toDB
	resArray, err := database.DbwReadAll(fromDB)
	if err != nil {
		log.Println(`copyHours:ERROR: db readall: err=`, err)
		return err
	}
	for _, v := range resArray {
		vm := v.(map[string]interface{})
		emp := getEmpFromMap(vm)
		if emp == nil {
			continue
		}
		err = database.DbwUpdate(toDB, (*emp)["UserID"].(string), emp)
		if err != nil {
			log.Println("copyHours:ERROR: Failed to update db for staff hours for userid=", (*emp)["UserID"].(string), " : err=", err)
			return err
		}
	}

	return nil
}
func BackupStaffHours(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	// check session expiration and authorization
	sessDetails := psession.Get_sess_details(r, "Backup and Reset Employee Hours", "Backup and Reset Employee Hours page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	toDB := ComposeDbName(StaffHoursEntity, "c")
	if err := cleanupHours(toDB); err != nil {
		log.Println("BackupStaffHours:ERROR: Failed to cleanup db=", toDB, " : err=", err)
	}
	fromDB := ComposeDbName(StaffHoursEntity, "b")
	if err := copyHours(fromDB, toDB); err != nil {
		log.Println("BackupStaffHours:ERROR: Failed to copy hours from db=", fromDB, " to=", toDB, " : err=", err)
	}

	bkupTime, err := database.DbwRead(fromDB, "BackupTime")
	if err == nil {
		// write it to the toDB
		database.DbwUpdate(toDB, "BackupTime", bkupTime)
	} else {
		log.Println("BackupStaffHours:ERROR: Failed to copy backup time from=", fromDB, " to=", toDB, " : err=", err)
	}

	toDB = fromDB
	if err := cleanupHours(toDB); err != nil {
		log.Println("BackupStaffHours:ERROR: Failed to cleanup db=", toDB, " : err=", err)
	}
	if err := copyHours(StaffHoursEntity, toDB); err != nil {
		log.Println("BackupStaffHours:ERROR: Failed to copy hours from db=", StaffHoursEntity, " to=", toDB, " : err=", err)
	}
	bkupTime, err = database.DbwRead(StaffHoursEntity, "BackupTime")
	if err == nil {
		// write it to the toDB
		database.DbwUpdate(toDB, "BackupTime", bkupTime)
	} else {
		log.Println("BackupStaffHours:ERROR: Failed to copy backup time from=", fromDB, " to=", toDB, " : err=", err)
	}

	// lastly reset the current staff hours
	// 0 the TotalHours and TotalExpectedHours, ClockIn and ClockOut
	resArray, err := database.DbwReadAll(StaffHoursEntity)
	if err != nil {
		log.Println(`BackupStaffHours:ERROR: db readall: err=`, err)
		return
	}

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		_, exists := vm["UserID"]
		if !exists {
			continue
		}

		(vm)["TotalHours"] = 0
		(vm)["TotalExpectedHours"] = 0
		(vm)["ClockInCnt"] = int(0)
		(vm)["ClockOutCnt"] = int(0)
		if err := database.DbwUpdate(StaffHoursEntity, "", &vm); err != nil {
			log.Println(`BackupStaffHours:ERROR: db update: err=`, err)
		}
	}

	nowStr, _ := misc.TimeNow()
	tstamp := map[string]interface{}{"BackupTime": nowStr}
	if err := database.DbwUpdate(StaffHoursEntity, "BackupTime", &tstamp); err != nil {
		log.Println(`BackupStaffHours:ERROR: db update timestamp: err=`, err)
	}

	http.Redirect(w, r, "/desk/report_staff_hours", http.StatusFound)
}

func DeleteStaffHoursEntity(userId string) error {
	misc.IncrRequestCnt()
	var err error
	var rMap *map[string]interface{}
	rMap, err = database.DbwRead(StaffHoursEntity, userId)
	if err != nil {
		log.Println("DeleteStaffHoursEntity:ERROR: Failed to read staff hours with name=", userId)
		return err
	}

	if err := database.DbwDelete(StaffHoursEntity, rMap); err != nil {
		log.Println("DeleteStaffHoursEntity:ERROR: Failed to delete staff hours with name=", userId)
		return err
	}
	return nil
}

func UpdateStaffHours(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	// check session expiration and authorization
	sessDetails := psession.Get_sess_details(r, "Update Employee Hours", "Update Employee Hours page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {

		userid := ""
		if ids, ok := r.URL.Query()["user"]; !ok || len(ids[0]) < 1 {
			log.Println("upd_staff: Url Param 'user' is missing")
		} else {
			userid = ids[0]
		}

		if userid == "" {
			log.Println("update_staff_hours: Missing required usersid=", userid)
			sessDetails.Sess.Message = "Failed to update hours for user: " + userid
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotFound)
			return
		}

		// handle clockin or clockout requests
		// if the clockin is requested, and TotalHours == 0 and LastClockoutTime != ''
		// then set TotalHours to ExpectedHours

		update := ""
		updates, ok := r.URL.Query()["update"]
		if !ok || len(updates[0]) < 1 {
			log.Println("update_staff_hours: Url Param 'update' is missing")
			sessDetails.Sess.Message = "Failed to update hours for user: " + userid
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotModified)
			return
		}
		update = updates[0]

		if update == "clockout" {
			if err := UpdateEmployeeHours(userid, false, 12, sessDetails.Sess); err != nil {
				log.Println("update_staff_hours:ERROR: Failed to update hours for clockout of userid=", userid, " : err=", err)
				sessDetails.Sess.Message = "Failed to update hours for clockout of user: " + userid
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotModified)
			} else {
				http.Redirect(w, r, "/desk/report_staff_hours", http.StatusFound)
			}
			return
		} else if update == "clockin" {
			// give user a clockin page to specify expected hours
			t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/staff_hours_clockin.gtpl", "static/header.gtpl")
			if err != nil {
				log.Println("update_staff_hours:ERROR: clockin template failure: err=", err)
				sessDetails.Sess.Message = "Failed to setup clockin for user: " + userid
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotFound)
				return
			}

			emp := EmpHours{
				userid,
				"",
				"",
				0,
				0,
				0,
				0,
				0,
			}
			tblData := UpdateEmpHours{
				sessDetails,
				emp,
			}
			err = t.Execute(w, &tblData)
			if err != nil {
				log.Println("update_staff_hours:ERROR: Failed to execute clockin template: err=", err)
				sessDetails.Sess.Message = "Failed clockin for user: " + userid
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotFound)
			}
			return
		}

		if update == "delete" {
			if sessDetails.Sess.Role != psession.ROLE_MGR { // only manager can delete the record
				sessDetails.Sess.Message = "No Permissions"
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
				return
			}

			if err := DeleteStaffHoursEntity(userid); err != nil {
				sessDetails.Sess.Message = "Failed to delete user hours data for user: " + userid
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
			}
		}
	} else {
		// clockin request with expected hours
		r.ParseForm()
		// get the form parameters - staffid, expected hours=hours
		userid := r.Form["staffid"][0]
		hours := r.Form["hours"][0]

		expHours, err := strconv.Atoi(hours)
		if err != nil {
			log.Println("update_staff_hours:ERROR: bad hours param=", hours, " : use default hours : err=", err)
			expHours = 12
		}

		if err := UpdateEmployeeHours(userid, true, expHours, sessDetails.Sess); err != nil {
			log.Println("update_staff_hours:ERROR: Failed to clockin hours for userid=", userid, " : err=", err)
			sessDetails.Sess.Message = "Failed to clockin hours for user: " + userid
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotModified)
			return
		}
	}

	http.Redirect(w, r, "/desk/report_staff_hours", http.StatusFound)
}

func IsUserLoggedIn(userID string) (bool, *map[string]interface{}, error) {
	rMap, err := database.DbwRead(StaffHoursEntity, userID)
	if err != nil {
		log.Println("IsUserLoggedIn:ERROR: Failed to read user=", userID, " : err=", err)
		return false, nil, err
	}

	num, ciCntOk := (*rMap)["ClockInCnt"].(float64)
	if !ciCntOk {
		(*rMap)["ClockInCnt"] = int(0)
	} else {
		(*rMap)["ClockInCnt"] = int(num)
	}

	num, coCntOk := (*rMap)["ClockOutCnt"].(float64)
	if !coCntOk {
		(*rMap)["ClockOutCnt"] = int(0)
	} else {
		(*rMap)["ClockOutCnt"] = int(num)
	}

	ci, ok := (*rMap)["LastClockinTime"].(string)
	if !ok || strings.Compare(ci, "") == 0 {
		return false, rMap, nil
	}

	co, ok := (*rMap)["LastClockoutTime"].(string)
	if !ok {
		return false, rMap, nil
	}

	clockoutTime, err := time.ParseInLocation(DateTimeLongForm, co, misc.GetLocale())
	if err != nil {
		return false, rMap, err
	}
	(*rMap)["clockoutTime"] = clockoutTime

	clockinTime, err := time.ParseInLocation(DateTimeLongForm, ci, misc.GetLocale())
	if err != nil {
		return false, rMap, err
	}
	(*rMap)["clockinTime"] = clockinTime

	dur := clockoutTime.Sub(clockinTime)
	(*rMap)["duration"] = dur
	mins := dur.Minutes()
	if mins >= 0 {
		return true, rMap, nil
	}
	return false, rMap, nil
}

func UpdateEmployeeHours(userid string, clockin bool, expHours int, sess *psession.PinoySession) error {

	if userid == "" {
		log.Println("UpdateEmployeeHours: Missing staff name")
		return errors.New("Missing staff id")
	}
	var err error
	var rMap *map[string]interface{}
	// check if userid is Role == ROLE_DSK - can only update themself and Hoppers
	// they cannot update other desk or manager role users
	if sess.Role == psession.ROLE_DSK && sess.User != userid {
		rMap, err = database.DbwRead(StaffEntity, userid)
		if err != nil {
			log.Println("UpdateEmployeeHours:ERROR: No staff with name=", userid, " : err=", err)
			return err
		}
		role := ""
		name, exists := (*rMap)["Role"]
		if exists && name != nil {
			role = name.(string)
		}
		if role != psession.ROLE_HOP {
			msg := "Not allowed to update hours for user: " + userid
			log.Println("UpdateEmployeeHours: ", msg, userid)
			return errors.New(msg)
		}
	}

	// get hours record
	key := ""
	_, rMap, err = IsUserLoggedIn(userid)
	if err != nil {
		if strings.Compare(err.Error(), "not_found") == 0 {
			log.Println("UpdateEmployeeHours:ERROR: No staff hours for user=", userid, " : err=", err)
			// create a new record
			if rMap == nil {
				rm := map[string]interface{}{"UserID": userid,
					"LastClockinTime":  "",
					"LastClockoutTime": "",
					"ExpectedHours":    expHours, // hours expected to work on the shift when clocks in
					// -- default per Role? desk=12, etc
					"TotalHours":         0,
					"TotalExpectedHours": 0,
					"ClockInCnt":         0,
					"ClockOutCnt":        0,
				}
				rMap = &rm
			} else {
				(*rMap)["ExpectedHours"] = expHours
			}
			key = userid
		} else {
			log.Println("UpdateEmployeeHours:ERROR: Failed to read staff hours for user=", userid, " : err=", err)
		}
	}

	nowStr, nowTime := misc.TimeNow()
	var lastClockinTime time.Time
	var lastClockoutTime time.Time
	var duration time.Duration
	var okIn, okOut, okDur bool

	var nowMinusLastClockin float64

	lastClockinTime, okIn = (*rMap)["clockinTime"].(time.Time)
	lastClockoutTime, okOut = (*rMap)["clockoutTime"].(time.Time)
	duration, okDur = (*rMap)["duration"].(time.Duration)
	if okIn {
		nowMinusLastClockin = nowTime.Sub(lastClockinTime).Hours()
	}

	lastExpHrs, _ := (*rMap)["ExpectedHours"].(float64)
	if clockin {
		cnt, _ := (*rMap)["ClockInCnt"].(int)
		(*rMap)["ClockInCnt"] = cnt + 1
		// how to handle if desk didnt clockout bell hop
		// - so if old clockin is > clockout means the hop was not clocked out
		if okDur && duration.Hours() < 0 { // dur = clockout - clockin
			if okOut {
				log.Println("UpdateEmployeeHours: detected missing clock-out of user=", userid,
					" : last-clockin=", lastClockinTime, " : last clock-out=", lastClockoutTime, " : last expected-hours=", (*rMap)["ExpectedHours"])

			} else {
				log.Println("UpdateEmployeeHours: detected missing clock-out of user=", userid,
					" : last-clockin=", lastClockinTime, " : last clock-out=empty : last expected-hours=", (*rMap)["ExpectedHours"])
			}
			// did staff member not get clocked out?
			total, _ := (*rMap)["TotalHours"].(float64)
			total += lastExpHrs
			(*rMap)["TotalHours"] = total
		} else {
			total, _ := (*rMap)["TotalHours"].(float64)
			total += duration.Hours()
			(*rMap)["TotalHours"] = total
		}

		total, _ := (*rMap)["TotalExpectedHours"].(float64)
		total += float64(expHours)
		(*rMap)["TotalExpectedHours"] = total

		// reset the LastClockinTime
		(*rMap)["LastClockinTime"] = nowStr
		(*rMap)["ExpectedHours"] = expHours

	} else {
		cnt, _ := (*rMap)["ClockOutCnt"].(int)
		(*rMap)["ClockOutCnt"] = cnt + 1
		(*rMap)["LastClockoutTime"] = nowStr
		// clocked out so recalc the total hours
		if okIn {
			total, _ := (*rMap)["TotalHours"].(float64)
			total += nowMinusLastClockin
			(*rMap)["TotalHours"] = total
		} else {
			// something is missing so try to guess
			total, _ := (*rMap)["TotalHours"].(float64)
			total += lastExpHrs
			(*rMap)["TotalHours"] = total
		}
	}

	err = database.DbwUpdate(StaffHoursEntity, key, rMap)
	if err != nil {
		log.Println("UpdateEmployeeHours:ERROR: Failed to update db for total hours for userid=", userid, " : err=", err)
		return err
	}
	signInOut := " SIGNED-IN"
	if clockin {
		signInOut = " SIGNED-OUT"
	}
	log.Println("UpdateEmployeeHours: updated hours: user=", userid, signInOut, " : record=", (*rMap))

	return nil
}

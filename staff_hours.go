package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	StaffHoursEntity = "staff_hours"
)

type EmpHours struct {
	UserID           string
	LastClockinTime  string
	LastClockoutTime string
	ExpectedHours    int // hours expected to work on the shift when clocks in
	// -- default per Role? desk=12, etc
	TotalHours float64 // gets updated when employee clocks out
}
type EmpHoursTable struct {
	*SessionDetails
	StaffHours []EmpHours
}

type UpdateEmpHours struct {
	*SessionDetails
	Emp EmpHours
}

func report_staff_hours(w http.ResponseWriter, r *http.Request) {
	fmt.Println("report_staff_hours:FIX method:", r.Method)
	sessDetails := get_sess_details(r, "Staff Hours", "Staff Hours page to Pinoy Lodge")
	if sessDetails.Sess.Role != ROLE_MGR && sessDetails.Sess.Role != ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	if r.Method != "GET" {
		log.Println("report_staff_hours: bad http method: should only be a GET")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/staff_hours.gtpl", "static/header.gtpl")
	if err != nil {
		fmt.Printf("report_staff_hours: err: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {

		resArray, err := PDb.ReadAll(StaffHoursEntity)
		if err != nil {
			log.Println(`report_staff_hours: db readall error`, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Println("FIX report_staff_hours: res=", resArray)

		emps := make([]EmpHours, 0) // FIX len(resArray))
		deskRole := sessDetails.Sess.Role == ROLE_DSK
		for _, v := range resArray {
			vm := v.(map[string]interface{})
			fmt.Println("FIX report_staff_hours: emp=", vm)
			id := ""
			name, exists := vm["UserID"]
			if !exists {
				continue
			}
			id = name.(string)
			if id == "" {
				// ignore this record
				continue
			}

			if deskRole {
				// only show the hoppers
				entity, err := PDb.Read(StaffEntity, id)
				if err != nil {
					log.Println("report_staff_hours: desk role: failed to read user=", id, " :err=", err)
					continue
				}

				if role, exists := (*entity)["Role"]; exists {
					if role.(string) != ROLE_HOP {
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

			totHours := float64(0)
			if num, exists := vm["TotalHours"]; exists {
				totHours = num.(float64)
			}

			emp := EmpHours{
				UserID:           id,
				LastClockinTime:  clockin,
				LastClockoutTime: clockout,
				ExpectedHours:    expHours,
				TotalHours:       totHours,
			}
			emps = append(emps, emp)
		}

		tblData := EmpHoursTable{
			sessDetails,
			emps,
		}
		err = t.Execute(w, &tblData)
		if err != nil {
			fmt.Println("report_staff_hours: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func update_staff_hours(w http.ResponseWriter, r *http.Request) {
	fmt.Println("update_staff_hours:FIX:method:", r.Method)
	// check session expiration and authorization
	sessDetails := get_sess_details(r, "Update Employee Hours", "Update Employee Hours page of Pinoy Lodge")
	if sessDetails.Sess.Role != ROLE_MGR && sessDetails.Sess.Role != ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
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
			//sessDetails := get_sess_details(r, "Update Employee Hours", "Update Employee Hours page of Pinoy Lodge")
			sessDetails.Sess.Message = "Failed to update hours for user: " + userid
			_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotFound)
			return
		}

		// handle clockin or clockout requests
		// if the clockin is requested, and TotalHours == 0 and LastClockoutTime != ''
		// then set TotalHours to ExpectedHours

		update := ""
		if updates, ok := r.URL.Query()["update"]; !ok || len(updates[0]) < 1 {
			log.Println("update_staff_hours: Url Param 'update' is missing")
		} else {
			update = updates[0]
		}

		clockIn := true
		deleteStaffHours := false
		if update == "delete" {
			deleteStaffHours = true
		} else if update == "clockout" {
			clockIn = false
		}

		fmt.Printf("update_staff_hours:FIX: user=%s update=%s\n", userid, update)

		if !deleteStaffHours {
			if err := UpdateEmployeeHours(userid, clockIn, sessDetails.Sess); err != nil {
				log.Println("update_staff_hours: Failed to update hours for userid=", userid, " :err=", err)
				sessDetails.Sess.Message = "Failed to update hours for user: " + userid
				_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotModified)
			}
			return
		}

		var err error
		var rMap *map[string]interface{}
		rMap, err = PDb.Read(StaffHoursEntity, userid)
		if err != nil {
			log.Println("update_staff_hours: No staff with name=", userid)
			sessDetails.Sess.Message = "Failed to update hours for user: " + userid
			_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotFound)
			return
		}

		if deleteStaffHours {
			fmt.Printf("update_staff_hours: delete user=%s\n", userid)

			if sessDetails.Sess.Role != ROLE_MGR { // only manager can delete the record
				sessDetails.Sess.Message = "No Permissions"
				_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
				return
			}

			if err := PDb.DbwDelete(StaffHoursEntity, rMap); err != nil {
				//sessDetails := get_sess_details(r, "UpdateEmployee Hours", "Update Employee Hours page of Pinoy Lodge")
				sessDetails.Sess.Message = "Failed to delete user hours data: " + userid
				_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
			}
		}
	} else {
		fmt.Println("update_staff_hours:FIX should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}
		// FIX TODO get the form parameters

		//err = UpdateEmployeeHours(userid string, clockin bool, session *PinoySession)
	}
	fmt.Printf("update_staff_hours:FIX post about to redirect to rooms\n")
	http.Redirect(w, r, "/static/frontpage", http.StatusOK)
}

func UpdateEmployeeHours(userid string, clockin bool, session *PinoySession) error {

	var err error
	var rMap *map[string]interface{}
	rMap, err = PDb.Read(StaffEntity, userid)
	// check if userid is Role == ROLE_DSK - can only update themself and Hoppers
	// they cannot update other desk or manager role users
	if session.Role == ROLE_DSK && session.User != userid {
		fmt.Println("update_staff_hours:FIX role is desk: check that userid role is bellhop")
		rMap, err = PDb.Read(StaffEntity, userid)
		if err != nil {
			log.Println("UpdateEmployeeHours: No staff with name=", userid, " :err=", err)
			return err
		}
		role := ""
		name, exists := (*rMap)["Role"]
		if exists && name != nil {
			role = name.(string)
		}
		if role != ROLE_HOP {
			msg := "Not allowed to update hours for user: " + userid
			log.Println("UpdateEmployeeHours: ", msg, userid)
			return errors.New(msg)
		}
	}

	// get hours record
	key := ""
	rMap, err = PDb.Read(StaffHoursEntity, userid)
	if err != nil {
		log.Println("UpdateEmployeeHours: No staff hours for user=", userid, " :err=", err)
		// create a new record

		rm := map[string]interface{}{"UserID": userid,
			"LastClockinTime":  "",
			"LastClockoutTime": "",
			"ExpectedHours":    12, // hours expected to work on the shift when clocks in
			// -- default per Role? desk=12, etc
			"TotalHours": 0,
		}
		rMap = &rm
		key = userid
	}

	nowStr, nowTime := TimeNow(Locale)
	if clockin {
		// reset the LastClockinTime
		(*rMap)["LastClockinTime"] = nowStr
		// if the clockin is requested, and TotalHours == 0 and LastClockoutTime != ''
		// then set TotalHours to ExpectedHours

	} else {
		(*rMap)["LastClockoutTime"] = nowStr
		// clocked out so recalc the hours
		// create Time from the LastClockinTime
		const longForm = "2006-01-02 15:04"
		// ex clockinTime: 2019-06-11 12:49
		clockinTime, err := time.ParseInLocation(longForm, (*rMap)["LastClockinTime"].(string), Locale)
		if err != nil {
			log.Println("UpdateEmployeeHours: Failed to calc clockin time for userid=", userid, " :err=", err)
			return err
		}
		dur := nowTime.Sub(clockinTime) // subtract clockouttime from LastClockinTime
		hours := dur.Hours()            // float64: add this to TotalHours
		total := hours + (*rMap)["TotalHours"].(float64)
		(*rMap)["TotalHours"] = total
	}

	err = PDb.DbwUpdate(StaffHoursEntity, key, rMap)
	if err != nil {
		log.Println("UpdateEmployeeHours: Failed to update db for total hours for userid=", userid, " :err=", err)
		return err
	}

	return nil
}

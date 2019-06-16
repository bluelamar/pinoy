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
	StaffHoursEntity = "staff_hours"
)

type EmpHours struct {
	UserID           string
	LastClockinTime  string
	LastClockoutTime string
	ExpectedHours    int // hours expected to work on the shift when clocks in
	// -- default per Role? desk=12, etc
	TotalHours int // gets updated when employee clocks out
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
	if sessDetails.Sess.Role != ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	if r.Method != "GET" {
		log.Println("report_staff_hours: bad http method: should only be a GET")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/staff_hours.gtpl", "static/header.gtpl")
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

		emps := make([]EmpHours, len(resArray))

		for k, v := range resArray {
			vm := v.(map[string]interface{})
			fmt.Println("FIX report_staff_hours: emp=", vm)
			id := ""
			name, exists := vm["name"]
			if !exists {
				continue
			}
			id = name.(string)

			clockin := ""
			name, exists = vm["LastClockinTime"]
			if exists {
				clockin = name.(string)
			}

			clockout := ""
			name, exists = vm["LastClockoutTime"]
			if exists {
				clockout = name.(string)
			}

			expHours := 0
			name, exists = vm["ExpectedHours"]
			if exists {
				if num, err := strconv.Atoi(name.(string)); err == nil {
					expHours = num
				}
			}

			totHours := 0
			name, exists = vm["TotalHours"]
			if exists {
				if num, err := strconv.Atoi(name.(string)); err == nil {
					totHours = num
				}
			}

			if id == "" {
				// ignore this record
				continue
			}
			emps[k] = EmpHours{
				UserID:           id,
				LastClockinTime:  clockin,
				LastClockoutTime: clockout,
				ExpectedHours:    expHours,
				TotalHours:       totHours,
			}
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
			sessDetails := get_sess_details(r, "Update Employee Hours", "Update Employee Hours page of Pinoy Lodge")
			sessDetails.Sess.Message = "Failed to update hours for user: " + userid
			_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusNotFound)
			return
		}

		citime := ""
		if cis, ok := r.URL.Query()["checkin"]; !ok || len(cis[0]) < 1 {
			log.Println("update_staff_hours: Url Param 'checkin' is missing")
		} else {
			citime = cis[0]
		}

		cotime := ""
		if cos, ok := r.URL.Query()["checkout"]; !ok || len(cos[0]) < 1 {
			log.Println("update_staff_hours: Url Param 'checkout' is missing")
		} else {
			cotime = cos[0]
		}

		update := ""
		if updates, ok := r.URL.Query()["update"]; !ok || len(updates[0]) < 1 {
			log.Println("update_staff_hours: Url Param 'update' is missing")
		} else {
			update = updates[0]
		}

		deleteStaffHours := false
		if update == "delete" {
			deleteStaffHours = true
		}

		fmt.Printf("update_staff_hours:FIX: user=%s update=%s\n", userid, update)

		var err error
		var rMap *map[string]interface{}
		rMap, err = PDb.Read(StaffHoursEntity, userid)
		if err != nil {
			log.Println("update_staff_hours: No staff with name=", userid)
			sessDetails := get_sess_details(r, "Update Employee Hours", "Update Employee Hours page of Pinoy Lodge")
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

			err := PDb.DbwDelete(StaffHoursEntity, rMap)
			if err != nil {
				sessDetails := get_sess_details(r, "UpdateEmployee Hours", "Update Employee Hours page of Pinoy Lodge")
				sessDetails.Sess.Message = "Failed to delete user hours data: " + userid
				_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
			} else {
				http.Redirect(w, r, "/desk/update_staff_hours", http.StatusOK)
			}
			return
		}

		// FIX TODO check if userid is Role == ROLE_DSK - can only update themself and BellHops
		// they cannot update other desk or manager role users
		if sessDetails.Sess.Role == ROLE_DSK && sessDetails.Sess.User != userid {
			// FIX read user staff record to get their role
			fmt.Println("update_staff_hours:FIX role is desk: check that userid role is bellhop")
		}

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/upd_emp_hours.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("update_staff_hours: err=", err.Error())
			sessDetails.Sess.Message = "Failed to Update employee hours: " + userid
			err = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		var empData EmpHours
		fmt.Printf("update_staff_hours: r-map=%v\n", (*rMap))

		expHours, err := strconv.Atoi((*rMap)["ExpectedHours"].(string))
		if err != nil {
			log.Println("update_staff_hours: Failed to convert expected hours: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError) // FIX
			return
		}
		totHours, err := strconv.Atoi((*rMap)["TotalHours"].(string))
		if err != nil {
			log.Println("update_staff_hours: Failed to convert total hours: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError) // FIX
			return
		}

		if citime == "" {
			citime = (*rMap)["LastClockinTime"].(string)
		}
		if cotime == "" {
			cotime = (*rMap)["LastClockoutTime"].(string)
		} else {
			// user clocked out so update total hours
			// FIX subtract cotime from citime and add that diff to totHours
			// ex checkinTime: 2019-06-11 12:49
			TODO
			ci := strings.Split(citime, " ")
			date, hourMin := ci[0], ci[1]
			dateSlice := strings.Split(date, "-")
			hm := strings.Split(hourMin, ":")
			hourStr := hm[0]
			min := hm[1]
			hourNum, err := strconv.Atoi(hourStr)
			if err != nil {
				return "", err
			}
			start := time.Date(2009, 1, 1, 12, 0, 0, 0, time.UTC)
		}
		empData = EmpHours{
			UserID:           (*rMap)["UserID"].(string),
			LastClockinTime:  citime,
			LastClockoutTime: (*rMap)["LastClockoutTime"].(string),
			ExpectedHours:    expHours, // FIX (*rMap)["ExpectedHours"].(int), // hours expected to work on the shift when clocks in
			// -- default per Role? desk=12, etc
			TotalHours: totHours, // FIX (*rMap)["TotalHours"].(int),
		}

		updData := UpdateEmpHours{
			sessDetails,
			empData,
		}
		err = t.Execute(w, updData)
		if err != nil {
			log.Println("update_staff_hours: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError) // FIX
		}
		return
	} else {
		fmt.Println("update_staff_hours:FIX should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}
		// FIX TODO get the form parameters
	}
	fmt.Printf("update_staff_hours:FIX post about to redirect to rooms\n")
	http.Redirect(w, r, "/frontpage", http.StatusOK)
}

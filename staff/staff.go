package staff

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/bluelamar/pinoy/config"
	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
)

const (
	StaffEntity = "staff"
)

type Employee struct {
	Last   string
	First  string
	Middle string
	Salary string
	Role   string
	Name   string // the unique user id for this employee
	Pwd    string // users password
}
type EmpTable struct {
	*psession.SessionDetails
	Staff []Employee
}

type UpdateEmployee struct {
	*psession.SessionDetails
	Employee
}

func Staff(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Staff", "Staff page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	if r.Method != "GET" {
		log.Println("staff: bad http method: should only be a GET")
		sessDetails.Sess.Message = "Bad request"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/staff.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("staff:ERROR: Failed to parse template: err=", err)
		sessDetails.Sess.Message = "Internal error"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	resArray, err := database.DbwReadAll(StaffEntity)
	if err != nil {
		log.Println("staff:ERROR: Failed to readall staff from db: err=", err)
		sessDetails.Sess.Message = "Failed to read staff"
		psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	emps := make([]Employee, 0)

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		id := ""
		name, exists := vm["name"]
		if !exists {
			continue
		}
		id = name.(string)
		last := ""
		name, exists = vm["Last"]
		if exists {
			last = name.(string)
		}
		first := ""
		name, exists = vm["First"]
		if exists {
			first = name.(string)
		}
		middle := ""
		name, exists = vm["Middle"]
		if exists {
			middle = name.(string)
		}
		salary := ""
		name, exists = vm["Salary"]
		if exists {
			salary = name.(string)
		}
		role := "Staff"
		name, exists = vm["Role"]
		if exists && name != nil {
			role = name.(string)
		}
		if last == "" || id == "" {
			// ignore this record
			continue
		}

		emp := Employee{
			Last:   last,
			First:  first,
			Middle: middle,
			Salary: salary,
			Role:   role,
			Name:   id,
		}
		emps = append(emps, emp)
	}

	tblData := EmpTable{
		sessDetails,
		emps,
	}
	err = t.Execute(w, &tblData)
	if err != nil {
		log.Println("staff:ERROR: Failed to execute template: err=", err)
		sessDetails.Sess.Message = "Failed to read staff"
		psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
	}
}

func AddStaff(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Add Employee", "Add Employee page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_empl.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("add_staff: err=", err)
			sessDetails.Sess.Message = "Internal error"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		empData := Employee{
			"",
			"",
			"",
			"",
			"Staff",
			"",
			"",
		}
		updData := UpdateEmployee{
			sessDetails,
			empData,
		}
		err = t.Execute(w, updData)
		if err != nil {
			log.Println("add_staff:ERROR: Failed to add staff: err=", err)
			sessDetails.Sess.Message = "Failed to add staff"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		}
	}
}

func UpdStaff(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	// check session expiration and authorization
	sessDetails := psession.GetSessDetails(r, "Update Employee", "Update Employee page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {

		id := ""
		ids, ok := r.URL.Query()["id"]
		if !ok || len(ids[0]) < 1 {
			log.Println("upd_staff: Url Param 'id' is missing")
		} else {
			id = ids[0]
		}

		lname := ""
		lnames, ok := r.URL.Query()["last"]
		if !ok || len(lnames[0]) < 1 {
			log.Println("upd_staff: Url Param 'last' is missing")
		} else {
			lname = lnames[0]
		}

		fname := ""
		fnames, ok := r.URL.Query()["first"]
		if !ok || len(fnames[0]) < 1 {
			log.Println("upd_staff: Url Param 'first' is missing")
		} else {
			fname = fnames[0]
		}

		mname := ""
		mnames, ok := r.URL.Query()["middle"]
		if !ok || len(mnames[0]) < 1 {
			log.Println("upd_staff: Url Param 'middle' is missing")
		} else {
			mname = mnames[0]
		}

		salary := ""
		salaries, ok := r.URL.Query()["salary"]
		if !ok || len(salaries[0]) < 1 {
			log.Println("upd_staff: Url Param 'salary' is missing")
		} else {
			salary = salaries[0]
		}

		role := "Staff"
		roles, ok := r.URL.Query()["role"]
		if !ok || len(roles[0]) < 1 {
			log.Println("upd_staff: Url Param 'role' is missing")
		} else {
			role = roles[0]
		}

		passwd := ""
		/* passwds, ok := r.URL.Query()["pwd"]
		if !ok || len(passwds[0]) < 1 {
			log.Println("upd_staff: Url Param 'pwd' is missing")
		} else {
			passwd = passwds[0]
		} */

		update := ""
		updates, ok := r.URL.Query()["update"]
		if !ok || len(updates[0]) < 1 {
			log.Println("upd_staff: Url Param 'update' is missing")
		} else {
			update = updates[0]
		}

		deleteEmp := false
		if update == "delete" {
			deleteEmp = true
		}

		if id == "" && lname == "" {
			sessDetails.Sess.Message = "Missing user name to make staff update page"
			psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
			return
		}

		var entry *map[string]interface{}
		if id == "" {
			elist, err := database.Find(database.GetDB(), StaffEntity, "Last", lname)
			if err != nil {
				log.Println("upd_staff:ERROR: No staff with last name=", lname, " : err=", err)
				sessDetails.Sess.Message = "No such employee: " + lname
				psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
				return
			}

			for _, v := range elist {
				// if more than one object ensure matched all parts of the name
				vm := v.(map[string]interface{})
				name, exists := vm["First"]
				if !exists {
					continue
				}
				if strings.Compare(fname, name.(string)) != 0 {
					continue
				}
				name, exists = vm["Middle"]
				if !exists {
					continue
				}
				if strings.Compare(mname, name.(string)) != 0 {
					continue
				}
				name, exists = vm["Salary"]
				if exists {
					salary = name.(string)
				}
				name, exists = vm["Role"]
				if exists {
					role = name.(string)
				}
				/* name, exists = vm["Pwd"]
				if exists && name != nil {
					passwd = name.(string)
				} */

				name, exists = vm["_id"]
				if !exists {
					break
				}
				id = name.(string)
			}
		} else {
			// read the entry to get the revision
			var err error
			entry, err = database.DbwRead(StaffEntity, id)
			if err != nil {
				log.Println("upd_staff:ERROR: No staff with name=", id, " : err=", err)
				sessDetails.Sess.Message = "Cant get user name to make staff update"
				psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
				return
			}

			name, exists := (*entry)["Salary"]
			if exists {
				salary = name.(string)
			}
			name, exists = (*entry)["Role"]
			if exists && name != nil {
				role = name.(string)
			}
		}

		if deleteEmp {
			err := database.DbwDelete(StaffEntity, entry)
			if err != nil {
				log.Println("upd_staff:ERROR: Failed to delete staff=", id, " : err=", err)
				sessDetails.Sess.Message = "Failed to delete staff: " + id
				psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			} else {
				// delete the employee from staff hours too
				if err := DeleteStaffHoursEntity(id); err != nil {
					sessDetails.Sess.Message = "Failed to delete staff hours data for user: " + id
					_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
					return
				}
				http.Redirect(w, r, "/manager/staff", http.StatusFound)
			}
			return
		}

		// user wants to update existing employee
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_empl.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("upd_staff:ERROR: Failed to parse template for update employee: err=", err)
			sessDetails.Sess.Message = "Failed to update staff: " + id
			psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		empData := Employee{
			lname,
			fname,
			mname,
			salary,
			role,
			id,
			passwd,
		}
		updData := UpdateEmployee{
			sessDetails,
			empData,
		}
		err = t.Execute(w, updData)
		if err != nil {
			log.Println("upd_staff:ERROR: Failed to execute template to update staff: err=", err)
			sessDetails.Sess.Message = "Failed to update staff: " + id
			psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}
	} else {

		r.ParseForm()

		lname := r.Form["last"][0]
		fname := r.Form["first"][0]
		mname := r.Form["middle"][0]
		salary := r.Form["salary"][0]
		name := r.Form["name"][0]
		role := r.Form["role"][0]
		passwd := r.Form["pwd"][0]

		// determine if new user or existing to be updated
		var emap *map[string]interface{}
		var err error

		key := ""
		emap, err = database.DbwRead(StaffEntity, name)
		if err != nil {
			log.Println("upd_staff:ERROR: Failed to read db:staff for name=", name, " : err=", err)
			emp := make(map[string]interface{})
			emap = &emp
			key = name
		}

		(*emap)["id"] = name
		(*emap)["name"] = name
		(*emap)["Last"] = lname
		(*emap)["Middle"] = mname
		(*emap)["First"] = fname
		(*emap)["Salary"] = salary
		(*emap)["Role"] = role
		(*emap)["Pwd"] = config.HashIt(passwd)

		err = database.DbwUpdate(StaffEntity, key, emap)
		if err != nil {
			log.Println("upd_staff:ERROR: Failed to execute template to update staff=", name, " : err=", err)
			sessDetails.Sess.Message = "Failed to update staff: " + name
			psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			return
		}

		// make staff hours entry if this is a new user
		if key != "" {
			UpdateEmployeeHours(name, false, 12, sessDetails.Sess)
		}

		http.Redirect(w, r, "/manager/staff", http.StatusFound)
	}
}

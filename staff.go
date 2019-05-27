package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
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
	*SessionDetails
	Staff []Employee
}

type UpdateEmployee struct {
	*SessionDetails
	Employee
}

func staff(w http.ResponseWriter, r *http.Request) {
	fmt.Println("staff:method:", r.Method)

	if r.Method != "GET" {
		fmt.Printf("staff: bad http method: should only be a GET\n")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/staff.gtpl", "static/header.gtpl")
	if err != nil {
		fmt.Printf("staff: err: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		sessDetails := get_sess_details(r, "Staff", "Staff page to Pinoy Lodge")

		resArray, err := PDb.ReadAll("staff")
		if err != nil {
			log.Println(`staff: db readall error`, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Println("FIX staff: res=", resArray)

		emps := make([]Employee, len(resArray))

		for k, v := range resArray {
			vm := v.(map[string]interface{})
			log.Println("FIX staff: emp=", vm)
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
			emps[k] = Employee{
				Last:   last,
				First:  first,
				Middle: middle,
				Salary: salary,
				Role:   role,
				Name:   id,
			}
		}

		//emps := make([]Employee, 2)
		//emps[0] = a1
		//emps[1] = a2

		tblData := EmpTable{
			sessDetails,
			emps,
		}
		err = t.Execute(w, &tblData)
		if err != nil {
			fmt.Println("staff: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func add_staff(w http.ResponseWriter, r *http.Request) {
	fmt.Println("add_staff:method:", r.Method)
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_empl.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("add_staff:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sessDetails := get_sess_details(r, "Add Employee", "Add Employee page of Pinoy Lodge")
			empData := Employee{
				"",
				"",
				"",
				"",
				"Staff",
				"id",
				"",
			}
			updData := UpdateEmployee{
				sessDetails,
				empData,
			}
			err = t.Execute(w, updData)
			if err != nil {
				fmt.Println("add_staff err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func upd_staff(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upd_staff:method:", r.Method)
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

		delete_emp := false
		if update == "delete" {
			delete_emp = true
		}

		fmt.Printf("upd_staff: last=%s first=%s middle=%s salary=%s update=%s\n", lname, fname, mname, salary, update)
		if id == "" && lname == "" {
			http.Error(w, "Last name not specified", http.StatusBadRequest)
			return
		}

		rev := ""
		if id == "" {
			elist, err := PDb.Find("staff", "Last", lname)
			if err != nil {
				log.Println("upd_staff: No staff with last name=", lname)
				http.Error(w, "No such employee", http.StatusBadRequest)
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

				name, exists = vm["_rev"]
				if !exists {
					break
				}
				rev = name.(string)
			}
		} else {
			// read the entry to get the revision
			entry, err := PDb.Read("staff", id)
			if err != nil {
				log.Println("upd_staff: No staff with name=", id)
				http.Error(w, "No such employee", http.StatusBadRequest)
			}
			rev = (*entry)["_rev"].(string)

			name, exists := (*entry)["Salary"]
			if exists {
				salary = name.(string)
			}
			name, exists = (*entry)["Role"]
			if exists && name != nil {
				role = name.(string)
			}
			/* name, exists = (*entry)["Pwd"]
			if exists && name != nil {
				passwd = name.(string)
			} */
		}

		if id == "" || rev == "" {
			http.Error(w, "Failed to process user: "+lname, http.StatusBadRequest)
			return
		}

		if delete_emp {
			fmt.Printf("upd_staff: delete employee=%s, %s %s\n", lname, fname, mname)

			err := PDb.Delete("staff", id, rev)
			if err != nil {
				http.Error(w, "Failed to delete user: "+lname, http.StatusInternalServerError)
			} else {
				http.Redirect(w, r, "/manager/staff", http.StatusFound)
			}
			return
		}

		// user wants to update existing employee
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_empl.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("upd_staff:err: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sessDetails := get_sess_details(r, "Update Employee", "Update Employee page of Pinoy Lodge")
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
				fmt.Println("upd_staff err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	} else {
		fmt.Println("upd_staff:FIX should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}

		lname := r.Form["last"][0]
		fname := r.Form["first"][0]
		mname := r.Form["middle"][0]
		salary := r.Form["salary"][0]
		name := r.Form["name"][0]
		role := r.Form["role"][0]
		log.Println("upd_staff:FIX post got role=", role)
		passwd := r.Form["pwd"][0]

		// determine if new user or existing to be updated
		rev := ""
		entry, err := PDb.Read("staff", name)
		if err != nil {
			log.Println("upd_staff: Failed to read db:staff for name=", name)
			http.Error(w, "Failed to apply employee="+name, http.StatusInternalServerError)
		} else {
			errMsg, exists := (*entry)["error"]
			if exists {
				log.Printf("upd_staff:FIX: create entity=staff id=%s: error=%v\n", name, errMsg)
			} else {
				rev = (*entry)["_rev"].(string) // update existing employee
			}
		}

		fmt.Printf("upd_staff:FIX: last=%s first=%s middle=%s salary=%s\n", lname, fname, mname, salary)

		emap := make(map[string]interface{})
		emap["id"] = name
		emap["name"] = name
		emap["Last"] = lname
		emap["Middle"] = mname
		emap["First"] = fname
		emap["Salary"] = salary
		emap["Role"] = role
		emap["Pwd"] = HashIt(passwd)

		if rev == "" {
			_, err = PDb.Create("staff", name, emap)
		} else {
			_, err = PDb.Update("staff", name, rev, emap)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		fmt.Printf("upd_staff: post about to redirect to staff\n")
		http.Redirect(w, r, "/manager/staff", http.StatusFound)
	}
}

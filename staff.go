package pinoy

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
		a1 := Employee{
			Last:   "Fuentes",
			First:  "Mario",
			Middle: "T",
			Salary: "$2.50",
		}
		a2 := Employee{
			Last:   "Johnson",
			First:  "Jay",
			Middle: "R",
			Salary: "$2.75",
		}

		emps := make([]Employee, 2)
		emps[0] = a1
		emps[1] = a2

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

func upd_staff(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upd_staff:method:", r.Method)
	// item size room
	// for get - prefill fields based on query parameters
	if r.Method == "GET" {

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

		if delete_emp {
			// TODO delete specified room - error if name is not set
			if lname == "" {
				http.Error(w, "Name not specified", http.StatusBadRequest)
			}
			fmt.Printf("upd_staff: delete employee=%s, %s %s\n", lname, fname, mname)
			http.Redirect(w, r, "/manager/staff", http.StatusFound)
		}

		// TODO get the room details from the db

		// user wants to add or update existing room
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/staff.gtpl", "static/header.gtpl")
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
		fmt.Println("upd_staff: should be post")
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}

		lname := r.Form["last"]
		fname := r.Form["first"]
		mname := r.Form["middle"]
		salary := r.Form["salary"]

		// TODO set in db
		fmt.Printf("upd_staff: last=%s first=%s middle=%s salary=%s\n", lname, fname, mname, salary)

		fmt.Printf("upd_staff: post about to redirect to staff\n")
		http.Redirect(w, r, "/manager/staff", http.StatusFound)
	}
}

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
		/* FIX
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
		} */

		resArray, err := PDb.ReadAll("staff")
		if err != nil {
			log.Println(`staff: db readall error`, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		emps := make([]Employee, len(resArray))

		for k, v := range resArray {
			vm := v.(map[string]interface{})
			doc, exists := vm["doc"]
			if !exists {
				continue
			}
			docm := doc.(map[string]interface{})
			emps[k] = Employee{
				Last:   docm["Last"].(string),
				First:  docm["First"].(string),
				Middle: docm["Middle"].(string),
				Salary: docm["Salary"].(string),
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
		/* FIX
		update_emp := false
		if update == "true" {
			update_emp = true
		} */

		fmt.Printf("upd_staff: last=%s first=%s middle=%s salary=%s update=%s\n", lname, fname, mname, salary, update)
		if lname == "" {
			http.Error(w, "Last name not specified", http.StatusBadRequest)
			return
		}
		elist, err := PDb.Find("staff", "Last", lname)
		if err != nil {
			log.Println("upd_staff: No staff with last name=", lname)
			http.Error(w, "No such employee", http.StatusBadRequest)
		}

		id := ""
		rev := ""
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
		if id == "" || rev == "" {
			http.Error(w, "Failed to process user: "+lname, http.StatusBadRequest)
			return
		}

		if delete_emp {
			// delete specified room
			fmt.Printf("upd_staff: delete employee=%s, %s %s\n", lname, fname, mname)

			err = PDb.Delete("staff", id, rev)
			if err != nil {
				http.Error(w, "Failed to delete user: "+lname, http.StatusInternalServerError)
			} else {
				http.Redirect(w, r, "/manager/staff", http.StatusFound)
			}
			return
		} /* FIX else if update_emp {
			val := Employee{
				lname,
				fname,
				mname,
				salary,
			}
			_, err := PDb.Update("staff", id, rev, val)
		} */

		// user wants to update existing employee
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

		lname := r.Form["last"][0]
		fname := r.Form["first"][0]
		mname := r.Form["middle"][0]
		salary := r.Form["salary"][0]

		// set in db
		fmt.Printf("upd_staff:FIX: last=%s first=%s middle=%s salary=%s\n", lname, fname, mname, salary)
		staffVal := Employee{
			Last:   lname,
			First:  fname,
			Middle: mname,
			Salary: salary,
		}
		_, err := PDb.Create("staff", staffVal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		fmt.Printf("upd_staff: post about to redirect to staff\n")
		http.Redirect(w, r, "/manager/staff", http.StatusFound)
	}
}

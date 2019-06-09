package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

const (
	RoomRatesEntity = "room_rates"
)

type RoomRate struct {
	TUnit string // time unit, ex: "3 Hours"
	Cost  string // cost of the TUnit, ex: "$10"
}
type RoomRateData struct {
	RateClass string // ex: "Small Room"
	Rates     []RoomRate
}

type RateDataTable struct {
	*SessionDetails
	RateData []RoomRateData
}

type RateDataEntry struct {
	*SessionDetails
	RateData RoomRateData
}

type ByTUnit []RoomRate

func (a ByTUnit) Len() int           { return len(a) }
func (a ByTUnit) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTUnit) Less(i, j int) bool { return a[i].TUnit < a[j].TUnit }

func room_rates(w http.ResponseWriter, r *http.Request) {
	fmt.Println("room_rates:method:", r.Method)
	if r.Method != "GET" {
		fmt.Printf("room_rates: bad http method: should only be a GET\n")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/room_rates.gtpl", "static/header.gtpl")
	if err != nil {
		fmt.Printf("room_rates: err: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		sessDetails := get_sess_details(r, "Room Rates", "Room Rates page to Pinoy Lodge")
		// []interface{}, error
		rrs, err := PDb.ReadAll(RoomRatesEntity)
		if err != nil {
			log.Println("room_rates: Failed to read room rates: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		rrds := make([]RoomRateData, len(rrs))
		for k, v := range rrs {
			log.Println("FIX got k=", k, " v=", v)
			val := v.(map[string]interface{})
			rs := val["Rates"].([]interface{})
			rates := make([]RoomRate, len(rs))
			for k2, v2 := range rs {
				val2 := v2.(map[string]interface{})
				rr := RoomRate{
					val2["TUnit"].(string),
					val2["Cost"].(string),
				}
				rates[k2] = rr
			}
			sort.Sort(ByTUnit(rates))
			rrd := RoomRateData{
				val["RateClass"].(string),
				rates,
			}
			rrds[k] = rrd
		}

		tblData := RateDataTable{
			sessDetails,
			rrds,
		}
		err = t.Execute(w, &tblData)
		if err != nil {
			fmt.Println("room_rates: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func upd_room_rate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upd_room_rate:method:", r.Method)
	if r.Method == "GET" {

		rate_class := ""
		rate_classes, ok := r.URL.Query()["rate_class"]
		if !ok || len(rate_classes[0]) < 1 {
			log.Println("upd_room_rate: Url Param 'rate_class' is missing")
		} else {
			rate_class = rate_classes[0]
		}

		update := ""
		updates, ok := r.URL.Query()["update"]
		if !ok || len(updates[0]) < 1 {
			log.Println("upd_room_rate: Url Param 'update' is missing")
		} else {
			update = updates[0]
		}

		deleteRate := false
		if update == "delete" {
			deleteRate = true
		}

		fmt.Printf("upd_room_rate: rate-class=%s update=%s\n", rate_class, update)
		var rateMap *map[string]interface{}
		if len(rate_class) > 1 {
			var err error
			rateMap, err = PDb.Read(RoomRatesEntity, rate_class)
			if err != nil {
				http.Error(w, "Invalid Rate class specified", http.StatusBadRequest)
				return
			}
		}

		if deleteRate {
			// delete specified room rate - error if room rate is not set
			if rate_class == "" {
				http.Error(w, "Rate class not specified", http.StatusBadRequest)
				return
			}
			fmt.Printf("upd_room_rate: delete room-rate=%s\n", rate_class)
			/* FIX
			id := (*rateMap)["_id"].(string)
			rev := (*rateMap)["_rev"].(string)
			err := PDb.Delete(RoomRatesEntity, id, rev) */
			err := PDb.DbwDelete(RoomRatesEntity, rateMap)
			if err != nil {
				sessDetails := get_sess_details(r, "Update Room Rate", "Update Room Rate page of Pinoy Lodge")
				sessDetails.Sess.Message = "Failed to delete room rate: " + rate_class
				_ = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusConflict)
			} else {
				http.Redirect(w, r, "/manager/room_rates", http.StatusFound)
			}
			return
		}

		// user wants to add or update existing room rate
		sessDetails := get_sess_details(r, "Update Room Rate", "Update Room Rate page of Pinoy Lodge")
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_room_rate.gtpl", "static/header.gtpl")
		if err != nil {
			fmt.Printf("upd_room_rate:err: %s", err.Error())
			sessDetails.Sess.Message = "Failed to update room rate: " + rate_class
			err = SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
			if err != nil {
				return
			}
		} else {

			var rrs []RoomRate
			if rateMap != nil {
				fmt.Printf("upd_room_rate: rate-map=%v\n", (*rateMap))
				rrs2, ok := (*rateMap)["Rates"]
				if !ok {
					fmt.Printf("upd_room_rate: failed to get rates\n")
					http.Error(w, "No rates", http.StatusInternalServerError)
					return
				}
				rrs3 := rrs2.([]interface{})
				fmt.Printf("upd_room_rate: rates=%v\n", rrs2)
				rrs = make([]RoomRate, len(rrs3))
				for k, v := range rrs3 {
					fmt.Printf("upd_room_rate: k=%d v=%v\n", k, v)

					v2 := v.(map[string]interface{})
					rrs[k] = RoomRate{
						TUnit: v2["TUnit"].(string),
						Cost:  v2["Cost"].(string),
					}
				}
			} else {
				rrs = nil
			}
			roomData := RoomRateData{
				rate_class,
				rrs,
			}
			updData := RateDataEntry{
				sessDetails,
				roomData,
			}
			err = t.Execute(w, updData)
			if err != nil {
				fmt.Println("upd_room_rate err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	} else {
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}

		rate_class := r.Form["rate_class"]
		newNumUnits := r.Form["num_units"]
		newTimeUnit := r.Form["new_rate_time_unit"]
		newCost := r.Form["new_rate_cost"]
		fmt.Printf("upd_room_rate: rate_class=%s num-units=%s time-unit=%s cost=%s\n", rate_class, newNumUnits, newTimeUnit, newCost)

		// validate incoming form fields
		if len(rate_class[0]) == 0 || len(newTimeUnit[0]) == 0 || len(newCost[0]) == 0 {
			log.Println("upd_room_rate:POST: Missing form data")
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		newTimeUnitStr := newNumUnits[0] + " " + newTimeUnit[0]
		fmt.Println("upd_room_rate:FIX new tu=", newTimeUnitStr)

		key := ""
		var rateMap *map[string]interface{}
		var err error
		rateMap, err = PDb.Read(RoomRatesEntity, rate_class[0])
		if err != nil {
			log.Println("upd_room_rate:POST: err=", err)

			// no such entry so the rate-class must be new
			rm := make(map[string]interface{})
			rateMap = &rm
			(*rateMap)["RateClass"] = rate_class[0]
			(*rateMap)["Rates"] = make([]interface{}, 0)
			key = rate_class[0]
		}

		rates := (*rateMap)["Rates"]
		fmt.Printf("upd_room_rate: rate_class=%s rates=%v\n", rate_class, rates)
		// if rates has TUnit entry matching newTimeUnit, remove it since it will be replaced
		newRates := make([]map[string]interface{}, 0)
		rts := rates.([]interface{})
		for _, v := range rts {
			v2 := v.(map[string]interface{})
			tu := v2["TUnit"].(string)
			if newTimeUnitStr == tu {
				continue
			}
			newRates = append(newRates, v2)
		}

		newRate := make(map[string]interface{})
		newRate["TUnit"] = newTimeUnitStr // FIX newTimeUnit[0]
		newRate["Cost"] = newCost[0]

		newRates = append(newRates, newRate)
		(*rateMap)["Rates"] = newRates

		// set in db
		fmt.Printf("upd_room_rate:FIX rate_class=%s newrates=%v\n", rate_class, newRates)
		err = PDb.DbwUpdate(RoomRatesEntity, key, rateMap)
		if err != nil {
			log.Println("upd_room_rate:POST: Failed to create or updated rate=", rate_class[0], " :err=", err)
			http.Error(w, "Failed to create or update rate="+rate_class[0], http.StatusInternalServerError)
			return
		}

		fmt.Printf("upd_room_rate: post about to redirect to room rates\n")
		http.Redirect(w, r, "/manager/room_rates", http.StatusFound)
	}
}

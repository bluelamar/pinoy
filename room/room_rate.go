package room

import (
	"html/template"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
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
	*psession.SessionDetails
	RateData []RoomRateData
}

type RateDataEntry struct {
	*psession.SessionDetails
	RateData RoomRateData
}

type ByTUnit []RoomRate

// global rate class map
// keyed by rate class string
// value for a rate class will be a map, key=num-hours(int), value=cost(float64)
var rateClassMap map[string]map[int]float64

// used to protect updates to rateClassMap
var rateClassMutex sync.Mutex

func getRoomRates() ([]RoomRateData, error) {
	// []interface{}, error
	rrs, err := database.DbwReadAll(RoomRatesEntity)
	if err != nil {
		log.Println("room_rates:ERROR: Failed to read room rates: err=", err)
		return nil, err
	}

	rrds := make([]RoomRateData, len(rrs))
	for k, v := range rrs {
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
	return rrds, nil
}

// InitRoomStatus sets up global data to be read in memory rather than frequently from the DB
func InitRoomRates() {
	// load the rateClassMap
	rrds, err := getRoomRates() // []RoomRateData, error
	if err != nil {
		log.Println("room_rates:ERROR: InitRoomRates failed to read room rates: err=", err)
		return
	}

	newRateClassMap := make(map[string]map[int]float64)
	vmap := make(map[string]interface{})
	for _, rrd := range rrds {
		rrm := make(map[int]float64)
		newRateClassMap[rrd.RateClass] = rrm
		for _, rr := range rrd.Rates { //[]RoomRate
			// translate cost into float64 and hours into int
			//vmap := make([string]interface{})
			vmap["cost"] = misc.StripMonPrefix(rr.Cost)
			cost := misc.XtractFloatField("cost", &vmap)
			dur := ParseDuration(rr.TUnit)
			rrm[dur] = cost
		}
	}
	rateClassMutex.Lock()
	rateClassMap = newRateClassMap
	rateClassMutex.Unlock()
}

func (a ByTUnit) Len() int           { return len(a) }
func (a ByTUnit) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTUnit) Less(i, j int) bool { return a[i].TUnit < a[j].TUnit }

func RoomRates(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Room Rates", "Room Rates page to Pinoy Lodge")
	if r.Method != "GET" {
		log.Println("room_rates: bad http method: should only be a GET")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/room_rates.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("room_rates:ERROR: Failed to parse template: err=", err)
		sessDetails.Sess.Message = "Internal error"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		return
	}

	// []RoomRateData, error
	rrds, err := getRoomRates()
	if err != nil {
		log.Println("room_rates:ERROR: Failed to read room rates: err=", err)
		sessDetails.Sess.Message = "Internal error getting room rates"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		return
	}

	tblData := RateDataTable{
		sessDetails,
		rrds,
	}
	err = t.Execute(w, &tblData)
	if err != nil {
		log.Println("room_rates:ERROR: Failed to execute template: err=", err)
		sessDetails.Sess.Message = "Internal error getting room rates"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
	}
}

func UpdRoomRate(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sessDetails := psession.GetSessDetails(r, "Update Room Rate", "Update Room Rate page of Pinoy Lodge")
	if r.Method == "GET" {

		rateClass := ""
		rateClasses, ok := r.URL.Query()["rate_class"]
		if ok && len(rateClasses[0]) > 0 {
			rateClass = rateClasses[0]
		}

		update := ""
		updates, ok := r.URL.Query()["update"]
		if !ok || len(updates[0]) < 1 {
			log.Println("upd_room_rate: missing url param: update")
		} else {
			update = updates[0]
		}

		deleteRate := false
		if update == "delete" {
			deleteRate = true
		}

		var rateMap *map[string]interface{}
		if len(rateClass) > 1 {
			var err error
			rateMap, err = database.DbwRead(RoomRatesEntity, rateClass)
			if err != nil {
				log.Println("upd_room_rate: Invalid Rate class specified: err=", err)
				sessDetails.Sess.Message = "Invalid rate class error"
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
				return
			}
		}

		if deleteRate {
			// delete specified room rate - error if room rate is not set
			if rateClass == "" {
				log.Println("upd_room_rate: Rate class not specified")
				sessDetails.Sess.Message = "Invalid rate class error"
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
				return
			}

			err := database.DbwDelete(RoomRatesEntity, rateMap)
			if err != nil {
				sessDetails.Sess.Message = "Failed to delete room rate: " + rateClass
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			} else {
				http.Redirect(w, r, "/manager/room_rates", http.StatusFound)
			}
			return
		}

		// user wants to add or update existing room rate
		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/upd_room_rate.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("upd_room_rate:ERROR: Failed to parse template: err=", err)
			sessDetails.Sess.Message = "Failed to update room rate: " + rateClass
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		var rrs []RoomRate
		if rateMap != nil {
			rrs2, ok := (*rateMap)["Rates"]
			if !ok {
				log.Println("upd_room_rate:ERROR: Failed to get rates")
				sessDetails.Sess.Message = "No rates"
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
				return
			}
			rrs3 := rrs2.([]interface{})
			rrs = make([]RoomRate, len(rrs3))
			for k, v := range rrs3 {

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
			rateClass,
			rrs,
		}
		updData := RateDataEntry{
			sessDetails,
			roomData,
		}
		err = t.Execute(w, updData)
		if err != nil {
			log.Println("upd_room_rate:ERROR: Failed to exec template: err=", err)
			sessDetails.Sess.Message = "Failed to update room rate: " + rateClass
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
		}
	} else {
		r.ParseForm()

		rateClass, _ := r.Form["rate_class"]
		newNumUnits, _ := r.Form["num_units"]
		newTimeUnit, _ := r.Form["new_rate_time_unit"]
		newCost, _ := r.Form["new_rate_cost"]

		// validate incoming form fields
		if len(rateClass[0]) == 0 || len(newTimeUnit[0]) == 0 || len(newCost[0]) == 0 || len(newNumUnits[0]) == 0 {
			log.Println("upd_room_rate:POST: Missing form data")
			sessDetails.Sess.Message = "Missing required rate class fields"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusBadRequest)
			return
		}

		newTimeUnitStr := newNumUnits[0] + " " + newTimeUnit[0]

		key := ""
		var rateMap *map[string]interface{}
		var err error
		rateMap, err = database.DbwRead(RoomRatesEntity, rateClass[0])
		if err != nil {
			log.Println("upd_room_rate:POST:WARN:: Perhaps new room: err=", err)

			// no such entry so the rate-class must be new
			rm := make(map[string]interface{})
			rateMap = &rm
			(*rateMap)["RateClass"] = rateClass[0]
			(*rateMap)["Rates"] = make([]interface{}, 0)
			key = rateClass[0]
		}

		// translate rateMap to form rateClassMap needs
		newRateMap := make(map[int]float64)
		vmap := make(map[string]interface{})

		// if rates has TUnit entry matching newTimeUnit, remove it since it will be replaced
		rates := (*rateMap)["Rates"]
		newRates := make([]map[string]interface{}, 0)
		rts := rates.([]interface{})
		for _, v := range rts {
			v2 := v.(map[string]interface{})
			tu := v2["TUnit"].(string)
			if newTimeUnitStr == tu {
				continue
			}
			newRates = append(newRates, v2)

			vmap["cost"] = misc.StripMonPrefix(v2["Cost"].(string))
			cost := misc.XtractFloatField("cost", &vmap)
			dur := ParseDuration(tu)
			newRateMap[dur] = cost
		}

		vmap["cost"] = misc.StripMonPrefix(newCost[0])
		cost := misc.XtractFloatField("cost", &vmap)
		dur := ParseDuration(newTimeUnitStr)
		newRateMap[dur] = cost

		newRate := make(map[string]interface{})
		newRate["TUnit"] = newTimeUnitStr
		newRate["Cost"] = newCost[0]

		newRates = append(newRates, newRate)
		(*rateMap)["Rates"] = newRates

		// set in db
		err = database.DbwUpdate(RoomRatesEntity, key, rateMap)
		if err != nil {
			log.Println("upd_room_rate:POST:ERROR: Failed to create or updated rate=", rateClass[0], " :err=", err)
			sessDetails.Sess.Message = "Failed to create or update rate=" + rateClass[0]
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		// update the in memory rateClassMap
		rateClassMutex.Lock()
		rateClassMap[rateClass[0]] = newRateMap
		rateClassMutex.Unlock()

		http.Redirect(w, r, "/manager/room_rates", http.StatusFound)
	}
}

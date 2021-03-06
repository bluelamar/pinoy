package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bluelamar/pinoy/config"
	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/food"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
	"github.com/bluelamar/pinoy/room"
	"github.com/bluelamar/pinoy/shift"
	"github.com/bluelamar/pinoy/staff"
	"github.com/client9/reopen"
	"github.com/gorilla/context"
)

var logger reopen.WriteCloser
var curRoomStati []room.RoomState
var doCleanup bool = true // used by the runCleanup function
var cleaners []misc.CleanerInterface

// signout revokes authentication for a user
func signout(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	sess, err := psession.GetUserSession(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uVal := sess.Values["user"]
	if uVal == nil {
		// no user logged in for the session
		log.Println("signout:WARN: No user session to log user out=", sess)
		http.Redirect(w, r, "/", http.StatusFound)
	}

	// update the employee report record
	staff.UpdateEmployeeHours(sess.Values["user"].(string), false, 12, psession.SessAttrs(r))

	sess.Options.MaxAge = -1
	sess.Values["authenticated"] = false
	sess.Values["user"] = nil
	sess.Values["role"] = nil
	sess.Values[food.SessFItems] = nil
	sess.Values[food.SessTCost] = nil

	err = sess.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func signin(w http.ResponseWriter, r *http.Request) {
	misc.IncrRequestCnt()
	if r.Method == "GET" {

		t, err := template.ParseFiles("static/login.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("signin:ERROR: Failed parse template: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sess := psession.SessAttrs(r)
		pageContent := &psession.PageContent{
			PageTitle: "Login",
			PageDescr: "Login for Pinoy Lodge",
		}

		loginPage := psession.SessionDetails{
			Sess:   sess,
			PgCont: pageContent,
		}

		err = t.Execute(w, loginPage)
		if err != nil {
			log.Println("signin:ERROR: Failed to exec template: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// user signing in
		r.ParseForm()
		username := r.Form["user_id"]
		password := r.Form["user_password"]

		// verify user in db and set cookie et al]
		entity := staff.StaffEntity
		umap, err := database.DbwRead(entity, username[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		passwd, ok := (*umap)["Pwd"]
		if !ok {
			http.Error(w, "Not authorized", http.StatusInternalServerError)
			return
		}
		// use hash only for user password
		pwd := config.HashIt(password[0])
		if passwd != pwd {
			misc.IncrFailedLoginCnt()
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		// get the role from the map
		role, ok := (*umap)["Role"]
		if !ok {
			role = "Bypasser"
		}

		sess, err := psession.GetUserSession(w, r)
		if err != nil {
			log.Println("signin: Failed to get user session: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sess.Values["authenticated"] = true
		sess.Values["user"] = username[0]
		sess.Values["role"] = role // "Manager" "desk" "staff"
		sess.Save(r, w)

		// update the employee report record - dont wait for it
		go staff.UpdateEmployeeHours(username[0], true, 12, psession.SessAttrs(r))

		misc.IncrLoginCnt()
		log.Println("signin: logged in=", username[0], " : auth=", sess.Values["authenticated"].(bool))
		http.Redirect(w, r, "/frontpage", http.StatusFound)
	}
}

func frontpage(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/frontpage.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("frontpage:ERROR: Failed to parse template: err=", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		sessDetails := psession.GetSessDetails(r, "Front page", "Front page to Pinoy Lodge")
		if err = t.Execute(w, sessDetails); err != nil {
			log.Println("frontpage:ERROR: Failed to execute template: err=", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	misc.IncrRequestCnt()
}

func initDB(cfg *config.PinoyConfig) error {
	// which db implementation are we using?
	var pDb database.DBInterface
	var dbName string
	if strings.Compare("", cfg.DbType) == 0 || strings.Compare("couchdb", cfg.DbType) == 0 {
		dbName = "couchdb"
		pDb = &database.CDBInterface{}
	} else if strings.Compare("mongodb", cfg.DbType) == 0 {
		dbName = "mongodb"
		pDb = &database.MDBInterface{}
	}

	dbiOld := database.SetDB(pDb)
	err := database.DbwInit(cfg)
	if err != nil {
		log.Println("main:ERROR: db init error: ", dbName, " error=", err)
		if dbiOld != nil {
			database.SetDB(dbiOld) // reset to the old db if the new one failed
		}
		return err
	}
	if dbiOld != nil {
		dbiOld.Close(cfg)
	}

	//log.Println("pinoy:main: db init success")
	return nil
}

func runDiags(cfg *config.PinoyConfig) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	log.Println("diags: mem-stats: total-bytes-virtual-memory(sys)=", memStats.Sys,
		" : alloc-heap=", memStats.HeapAlloc, " : num-GCs=", memStats.NumGC)
	// TODO report scoreboard stats - ex: number of room registrations, et al?

}

func runWeeklyCleanup(cfg *config.PinoyConfig) {
	// if it is day 7 then call cleanup impls
	// be sure not to repeat every hour if it was already done
	_, t := misc.TimeNow()
	dayOfTheYear := t.YearDay()
	cleanupTime := dayOfTheYear%7 == 0
	if doCleanup == false {
		if cleanupTime == false {
			doCleanup = true
		}
		return
	}

	doCleanup = false
	// run the cleanup impls
	if cleaners == nil {
		cleaners = make([]misc.CleanerInterface, 0)
		var cleaner misc.CleanerInterface
		cleaner = shift.NewCleaner()
		cleaners = append(cleaners, cleaner)
	}

	for _, c := range cleaners {
		go c.Cleanup(cfg, t)
	}
}

func roomStati(w http.ResponseWriter, r *http.Request) {
	bytesRepresentation, err := json.Marshal(curRoomStati)
	if err != nil {
		log.Println("roomStati:ERROR: Failed to convert to json: :err=", err)
		return
	}
	_, err = w.Write(bytesRepresentation) // []byte) // int, error
	if err != nil {
		log.Println("room_stati:ERROR: Failed to write response: err=", err)
	}
}

func runRoomCheck(cfg *config.PinoyConfig) {
	// check if any rooms should be checked out within configurable
	// time period : cfg.RoomStatusMonitorInterval
	// set this up in memory so that the browser javascript will make call to
	// load it quickly
	dur := time.Duration(time.Minute * time.Duration(cfg.CheckoutGracePeriod))
	stati, err := room.GetRoomStati(room.BookedStatus, dur) // ([]RoomState, error)
	if err != nil {
		log.Println("main:runRoomCheck:ERROR: err=", err)
		return
	}
	curRoomStati = stati
}

// perform inits on modules that are dependent on the database to be ready
func doOneTimeInits() {
	room.InitRoomStatus()
	shift.BuildShiftList(true)
}

func main() {
	/* FIX TODO   setup templates first
	 */

	curRoomStati = make([]room.RoomState, 0)

	cfg, err := config.LoadConfig("/etc/pinoy/config.json")
	if err != nil {
		log.Println("main:ERROR: config load error=", err)
		return
	}
	config.SetConfig(cfg)

	switch cfg.LogOutput {
	case "stdout":
		logger = reopen.Stdout
	case "stderr":
		logger = reopen.Stderr
	case "file":
		logger, err = reopen.NewFileWriter(cfg.LogFile)
		if err != nil {
			log.Println("main:ERROR: Failed to open file=", cfg.LogFile, " : err=", err)
		} else {
			sighup := make(chan os.Signal, 1)
			signal.Notify(sighup, syscall.SIGHUP)
			go func() {
				for {
					<-sighup
					fmt.Println("main: handle sighup to reopen logger")
					logger.Reopen()
				}
			}()
		}
	default:
		logger = reopen.Stdout
	}
	log.SetOutput(logger)

	{
		// setup serialization for types stored in sessions
		var items []food.FoodItem
		gob.Register(items)
	}
	psession.InitStore(cfg)

	misc.InitTime("Singapore", 8)
	misc.InitStats()

	// initialize DB then the "about to checkout rooms"
	doOneTimeInit := true
	initDbErr := initDB(cfg)
	if initDbErr != nil {
		log.Println("main:ERROR: Failed to init db - retry in a few minutes")
	} else {
		log.Println("main: DB initialized")
		runRoomCheck(cfg)
		doOneTimeInits()
		doOneTimeInit = false // FIX TODO this should be set according to success of doOneTimeInits
	}

	// setup background tasks
	minutes := cfg.StatsMonitorInterval
	if minutes == 0 {
		minutes = 60
	}
	statsTicker := time.NewTicker(time.Duration(minutes) * time.Minute)

	minutes = cfg.RoomStatusMonitorInterval
	if minutes == 0 {
		minutes = 5
	}
	roomTicker := time.NewTicker(time.Duration(minutes) * time.Minute)

	quit := make(chan string)
	go func() {
		for {
			select {
			case <-statsTicker.C:
				runDiags(cfg)
				initDB(cfg)
				runWeeklyCleanup(cfg)
			case <-roomTicker.C:
				if initDbErr != nil {
					// upon startup db can take a minute or so to start so we catch
					// that situation here
					initDbErr = initDB(cfg)
				}
				if initDbErr == nil {
					runRoomCheck(cfg)
					if doOneTimeInit {
						doOneTimeInits()
						doOneTimeInit = false
						log.Println("main: room init complete")
					}
				}
			case <-quit:
				statsTicker.Stop()
				roomTicker.Stop()
				return
			}
		}
	}()

	// setup routes
	http.HandleFunc("/", frontpage)
	http.HandleFunc("/frontpage", frontpage)
	http.HandleFunc("/signin", signin)
	http.HandleFunc("/signout", signout)
	http.HandleFunc("/desk/register", room.Register)
	http.HandleFunc("/desk/checkout", room.RoomCheckout)
	http.HandleFunc("/desk/room_status", room.RoomStatus)
	http.HandleFunc("/desk/room_stati", roomStati) // AJAX api return JSON
	http.HandleFunc("/desk/room_hop", room.RoomHop)
	http.HandleFunc("/desk/report_staff_hours", staff.ReportStaffHours)
	http.HandleFunc("/desk/upd_staff_hours", staff.UpdateStaffHours)
	http.HandleFunc("/desk/food", food.Food)
	http.HandleFunc("/desk/purchase", food.Purchase)
	http.HandleFunc("/desk/purchase_summary", food.PurchaseSummary)
	http.HandleFunc("/desk/shiftdailyinfo", shift.DailyInfo)
	http.HandleFunc("/manager/staff", staff.Staff)
	http.HandleFunc("/manager/upd_staff", staff.UpdStaff)
	http.HandleFunc("/manager/add_staff", staff.AddStaff)
	http.HandleFunc("/manager/backup_staff_hours", staff.BackupStaffHours)
	http.HandleFunc("/manager/rooms", room.Rooms)
	http.HandleFunc("/manager/upd_room", room.UpdRoom)
	http.HandleFunc("/manager/room_rates", room.RoomRates)
	http.HandleFunc("/manager/upd_room_rate", room.UpdRoomRate)
	http.HandleFunc("/manager/report_room_usage", room.ReportRoomUsage)
	http.HandleFunc("/manager/backup_room_usage", room.BackupRoomUsage)
	http.HandleFunc("/manager/report_room_history", room.ReportRoomHistory)
	http.HandleFunc("/manager/backup_room_history", room.BackupRoomHistory)
	http.HandleFunc("/manager/upd_food", food.UpdFood)
	http.HandleFunc("/manager/report_food_usage", food.ReportFoodUsage)
	http.HandleFunc("/manager/backup_food_usage", food.BackupFoodUsage)
	http.HandleFunc("/manager/shiftinfo", shift.Info)
	http.HandleFunc("/manager/upd_shiftinfo", shift.UpdateShiftInfo)
	http.HandleFunc("/manager/svc_stats", misc.SvcStats)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	err = http.ListenAndServe("127.0.0.1:8080", context.ClearHandler(http.DefaultServeMux)) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	quit <- "quit"
}

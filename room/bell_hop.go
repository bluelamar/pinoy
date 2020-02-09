package room

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/bluelamar/pinoy/config"
	"github.com/bluelamar/pinoy/database"
	"github.com/bluelamar/pinoy/misc"
	"github.com/bluelamar/pinoy/psession"
	"github.com/bluelamar/pinoy/staff"
)

const (
	hopperChoice = "Choose User ID"
)

type BellHop struct {
	UserID     string
	Room       string
	TimeStamp  string
	LoginValid bool // true=valid or false
	// key= UserID + ":" + Room + ":" + TimeStamp
}

type BellHopsTable struct {
	*psession.SessionDetails
	BellHops []BellHop
}

type HopperTable struct {
	*psession.SessionDetails
	RoomNum     string
	CheckinTime string
	Hoppers     []string
	Repeat      string
}

// currently clocked in
type HopShift struct {
	UserID  string
	ClockIn string
}

type HopShiftTable struct {
	*psession.SessionDetails
	Shift []HopShift
}

func RoomHop(w http.ResponseWriter, r *http.Request) {
	sessDetails := psession.GetSessDetails(r, "Room Bell Hop", "Bell Hop page of Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR && sessDetails.Sess.Role != psession.ROLE_DSK {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	misc.IncrRequestCnt()
	if r.Method == "GET" {

		room := ""
		rooms, ok := r.URL.Query()["room"]
		if !ok || len(rooms[0]) < 1 {
			log.Println("room_hop: missing Url param: room")
		} else {
			room = rooms[0]
		}

		citime := ""
		citimes, ok := r.URL.Query()["citime"]
		if !ok || len(citimes[0]) < 1 {
			log.Println("room_hop: missing Url param: citime")
		} else {
			citime = citimes[0]
		}

		repeat := ""
		repeats, ok := r.URL.Query()["repeat"]
		if !ok || len(repeats[0]) < 1 {
			log.Println("room_hop: missing Url param: repeat")
		} else {
			repeat = repeats[0]
		}

		// make list of hop names
		// do a Find of staff that are Role == ROLE_HOP
		hlist, err := database.DbwFind(staff.StaffEntity, "Role", psession.ROLE_HOP)
		if err != nil {
			sessDetails.Sess.Message = "No bell hops"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		hoppers := make([]string, 0)
		hoppers = append(hoppers, hopperChoice)

		for _, v := range hlist {
			vm := v.(map[string]interface{})
			id := ""
			if name, exists := vm["name"]; exists {
				id = name.(string)
			} else {
				continue
			}

			hoppers = append(hoppers, id)
		}

		t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/desk/room_hop.gtpl", "static/header.gtpl")
		if err != nil {
			log.Println("room_hop:ERROR: Failed to parse template: err=", err)
			sessDetails.Sess.Message = "Error with bell hops"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		regData := HopperTable{
			sessDetails,
			room,
			citime,
			hoppers,
			repeat,
		}
		err = t.Execute(w, regData)
		if err != nil {
			log.Println("room_hop:ERROR: Failed to execute template : err=", err)
			sessDetails.Sess.Message = "Error with bell hops"
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}
	} else {
		r.ParseForm()

		hopper := r.Form["hopper"]
		hopperID := r.Form["user_id"]
		bellHopPin := r.Form["bell_hop_pin"]
		room_num := r.Form["room_num"]
		repeat := r.Form["repeat"] // if true then can repeat upon failure

		if strings.Compare(hopper[0], hopperChoice) == 0 {
			hopper[0] = hopperID[0]
		}
		umap, err := database.DbwRead(staff.StaffEntity, hopper[0])
		if err != nil {
			log.Println("room_hop:ERROR: Failed to read room attendent=", hopper[0], " : err=", err)
			sessDetails.Sess.Message = "Error with bell hop " + hopper[0]
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}
		passwd, ok := (*umap)["Pwd"]
		if !ok {
			log.Println("room_hop:ERROR: Failed to check passwd for room attendent=", hopper[0], " : room=", room_num[0], " : err=", err)
			sessDetails.Sess.Message = "PIN check Error with bell hop " + hopper[0]
			_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			return
		}

		/* TODO set in db: BellHopEntity
			type BellHop struct {
			- userid : UserID
			- room : Room
			- time stamp : TimeStamp
			- success login : LoginValid = true or false
			- key= UserID + ":" + Room + ":" + TimeStamp
		UserID     string
		Room       string
		TimeStamp  string
		LoginValid bool // true=valid or false
		// key= UserID + ":" + Room + ":" + TimeStamp
		*/
		// use hash for user password
		pwd := config.HashIt(bellHopPin[0])
		if strings.Compare(pwd, passwd.(string)) != 0 {
			// invalid match
			log.Println("room_hop: room attendent pin not a match=", hopper[0], " : room=", room_num[0])
			misc.IncrFailedLoginCnt()
			if strings.Compare(repeat[0], "true") == 0 {
				nowStr, _ := misc.TimeNow()
				http.Redirect(w, r, "/desk/room_hop?room="+room_num[0]+"&citime="+nowStr+"&repeat=false", http.StatusFound)
			} else {
				sessDetails.Sess.Message = "PIN check Error again with bell hop " + hopper[0]
				_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusAccepted)
			}
			return
		}

		// check if hop is clocked in, if not, clock them in and show warning
		clockedIn, _, _ := staff.IsUserLoggedIn(hopper[0])
		if clockedIn == false {
			misc.IncrLoginCnt()
			sessAttrs := psession.PinoySession{
				User:      hopper[0],
				Role:      psession.ROLE_HOP,
				Auth:      false,
				SessID:    "",
				CsrfToken: "",
				CsrfParam: "",
				Message:   "",
			}
			go staff.UpdateEmployeeHours(hopper[0], true, 8, &sessAttrs)
		}

		http.Redirect(w, r, "/desk/room_status", http.StatusFound)
	}
}

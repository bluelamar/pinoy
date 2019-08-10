package misc

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/bluelamar/pinoy/psession"
)

var (
	startupTime     time.Time // time when the service started
	numRequests     int64     // number requests received since the server was started
	numLogins       int64     // number times users have logged in
	numFailedLogins int64     // number times users failed to login
	mutex           sync.Mutex
)

type SvcStatsData struct {
	Name  string
	Value string
}
type SvcStatsTable struct {
	*psession.SessionDetails
	Title     string
	StatsData []SvcStatsData
}

func InitStats() {
	// called by main
	_, startupTime = TimeNow()
}

func IncrRequestCnt() {
	mutex.Lock()
	numRequests++
	mutex.Unlock()
}
func IncrLoginCnt() {
	mutex.Lock()
	numLogins++
	mutex.Unlock()
}
func IncrFailedLoginCnt() {
	mutex.Lock()
	numFailedLogins++
	mutex.Unlock()
}

func SvcStats(w http.ResponseWriter, r *http.Request) {

	sessDetails := psession.Get_sess_details(r, "Service Stats", "Service Stats page to Pinoy Lodge")
	if sessDetails.Sess.Role != psession.ROLE_MGR {
		sessDetails.Sess.Message = "No Permissions"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusUnauthorized)
		return
	}

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", "static/manager/svc_stats.gtpl", "static/header.gtpl")
	if err != nil {
		log.Println("svcStats:ERROR: Failed to parse template: err=", err)
		sessDetails.Sess.Message = "Failed to get service stats"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
		return
	}

	// gather the stats
	title := `Service Statistics Snapshot`
	stats := make([]SvcStatsData, 0)

	// calculate uptime
	_, now := TimeNow()
	dur := now.Sub(startupTime)
	hours := dur.Hours()
	stat := SvcStatsData{
		Name:  "Up Time Hours",
		Value: fmt.Sprint(hours),
	}
	stats = append(stats, stat)

	stat = SvcStatsData{
		Name:  "Number Requests",
		Value: fmt.Sprint(numRequests),
	}
	stats = append(stats, stat)

	stat = SvcStatsData{
		Name:  "Number Logins",
		Value: fmt.Sprint(numLogins),
	}
	stats = append(stats, stat)

	stat = SvcStatsData{
		Name:  "Number Failed Logins",
		Value: fmt.Sprint(numFailedLogins),
	}
	stats = append(stats, stat)

	tblData := SvcStatsTable{
		sessDetails,
		title,
		stats,
	}
	if err = t.Execute(w, &tblData); err != nil {
		log.Println("svc_stats:ERROR: could not execute template: err=", err)
		sessDetails.Sess.Message = "Failed to report service statistics"
		_ = psession.SendErrorPage(sessDetails, w, "static/frontpage.gtpl", http.StatusInternalServerError)
	}

}

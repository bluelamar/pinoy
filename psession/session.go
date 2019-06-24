package psession

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/bluelamar/pinoy/config"
	"github.com/bluelamar/pinoy/database"
	"github.com/gorilla/sessions"
)

const (
	ROLE_MGR        = "Manager"
	ROLE_DSK        = "Desk"
	ROLE_HOP        = "BellHop"
	CookieNameSID   = "PinoySID"
	CookieSecretKey = "cookieID"
)

// TODO create secret using random values - should be from db? so all servers use
// the same secret
// Note: Don't store your key in your source code. Pass it via an
// environmental variable, or flag (or both), and don't accidentally commit it
// alongside your code. Ensure your key is sufficiently random - i.e. use Go's
// crypto/rand or securecookie.GenerateRandomKey(32) and persist the result
// TODO get the cookie store secret from the config file
// FIX var store = sessions.NewCookieStore([]byte("something-very-secret"))
var store *sessions.CookieStore

type PinoySession struct {
	User      string
	Role      string
	Auth      bool
	SessID    string
	CsrfToken string
	CsrfParam string
	Message   string
}

type PageContent struct {
	PageTitle string
	PageDescr string
}

type SessionDetails struct {
	Sess   *PinoySession
	PgCont *PageContent
}

type SessionMgr interface {
	getSession(db *database.DBInterface, user string) (*PinoySession, error)
	putSession(db *database.DBInterface, sess *PinoySession) error
}

func InitStore(cfg *config.PinoyConfig) {
	// FIX var store = sessions.NewCookieStore([]byte("something-very-secret"))
	cookieSecret := cfg.CookieSecret
	cookieSecretDbEntity := cfg.CookieSecretDbEntity
	if cookieSecret == "" {
		// retrieve coookiessecret from the db
		entry, err := database.DbwRead(cookieSecretDbEntity, CookieSecretKey)
		if err != nil {
			log.Println("initStore: DB error: No cookie entry=", cookieSecretDbEntity, " : key=", CookieSecretKey, " : err=", err)
		} else {
			if sec, ok := (*entry)[CookieSecretKey].(string); ok {
				cookieSecret = sec
			} else {
				log.Println("initStore: Missing cookie key=", CookieSecretKey)
			}
		}
		if cookieSecret == "" {
			cookieSecret = "something-very-secret" // use a default for the CookieSecret
		}
	}
	store = sessions.NewCookieStore([]byte(cookieSecret))
	storeOptions := store.Options
	fmt.Printf("store options:FIX path=%s domain=%s httponly=%t maxage=%d secure=%t\n",
		storeOptions.Path, storeOptions.Domain, storeOptions.HttpOnly, storeOptions.MaxAge, storeOptions.Secure)

	// FIX TODO add session expiry to config
	// set session expiry to default of 12 hours
	storeOptions.MaxAge = 12 * 60 * 60
	store.Options = storeOptions
}

func GetUserSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	sess, err := store.Get(r, CookieNameSID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	return sess, nil
}

func Sess_attrs(r *http.Request) *PinoySession {

	session, err := store.Get(r, CookieNameSID)
	if err != nil {
		log.Printf("sess_attrs: err=%v\n", err)
	} else {
		fmt.Printf("sess_attrs:FIX: sess=%v\n", session)
	}
	for k, v := range session.Values {
		fmt.Printf("s-key:FIX: %v", k)
		fmt.Printf(" : s-val:FIX: %v\n", v)
	}
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		//http.Error(w, "Forbidden", http.StatusForbidden)
		session.Values["authenticated"] = false
		fmt.Printf("sess_attrs:FIX: set auth=%t\n", session.Values["authenticated"].(bool))
	}
	fmt.Printf("sess_attrs:FIX: auth=%t\n", session.Values["authenticated"].(bool))

	user := ""
	if sess_user, ok := session.Values["user"].(string); ok {
		user = sess_user
	}
	role := ""
	if sess_role, ok := session.Values["role"].(string); ok {
		role = sess_role
	}
	sess := &PinoySession{
		Auth:      session.Values["authenticated"].(bool),
		User:      user,
		Role:      role,
		SessID:    session.ID,
		CsrfToken: "",
		CsrfParam: "",
	}
	return sess
}

func Get_sess_details(r *http.Request, title, desc string) *SessionDetails {

	sess := Sess_attrs(r)

	pageContent := &PageContent{
		title,
		desc,
	}

	sessDetails := SessionDetails{
		sess,
		pageContent,
	}
	return &sessDetails
}

func SendErrorPage(sess *SessionDetails, w http.ResponseWriter, webPageTmplt string, httpCode int) error {

	t, err := template.ParseFiles("static/layout.gtpl", "static/body_prefix.gtpl", webPageTmplt, "static/header.gtpl")
	if err != nil {
		log.Printf("SendErrorPage: %s: Parse template err: %s", webPageTmplt, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	err = t.Execute(w, sess)
	if err != nil {
		log.Printf("SendErrorPage: %s: Execute err: %s", webPageTmplt, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

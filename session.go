// session
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const (
	ROLE_MGR = "Manager"
	ROLE_DSK = "Desk"
)

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
	getSession(db *DBInterface, user string) (*PinoySession, error)
	putSession(db *DBInterface, sess *PinoySession) error
}

func sess_attrs(r *http.Request) *PinoySession {

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

func get_sess_details(r *http.Request, title, desc string) *SessionDetails {

	sess := sess_attrs(r)

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

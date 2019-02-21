// session
package main

//"fmt"

type PinoySession struct {
	User      string
	Role      string
	Auth      bool
	SessID    string
	CsrfToken string
	CsrfParam string
}

type PageContent struct {
	PageTitle string
	PageDescr string
}

type DBInterface interface {
	open(dbName, dbUser, dbPwd string) error
	write(key string, val interface{}) error
	read(key string) (interface{}, error)
}
type SessionMgr interface {
	getSession(db *DBInterface, user string) (*PinoySession, error)
	putSession(db *DBInterface, sess *PinoySession) error
}

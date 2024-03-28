package session

import "github.com/gorilla/sessions"

var Store *sessions.CookieStore

func InitStore(key []byte) {
	Store = sessions.NewCookieStore(key)
}

package main

import (
	"encoding/binary"
	"time"

	"github.com/MJKWoolnough/sessions"
)

var loginStore, basketStore *sessions.CookieStore

func initStores(sessionKey, basketKey []byte) {
	loginStore, _ = sessions.NewCookieStore(sessionKey, sessions.HTTPOnly(), sessions.Name("session"), sessions.Expiry(time.Hour*24*30))
	basketStore, _ = sessions.NewCookieStore(basketKey, sessions.HTTPOnly(), sessions.Name("basket"))
}

func GetLogin(r *http.Request) uint64 {
	buf := loginStore.Get(r)
	if len(buf) != 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(buf)
}

func SetLogin(w http.ResonseWriter, userID uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], userID)
	loginStore.Set(w, buf[:])
}

// TODO: do basket store/retrieve

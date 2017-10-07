package main

import (
	"encoding/binary"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/MJKWoolnough/memio"
	"github.com/MJKWoolnough/sessions"
)

var loginStore, basketStore *sessions.CookieStore

func initStores(sessionKey, basketKey []byte, basketTypes ...interface{}) {
	loginStore, _ = sessions.NewCookieStore(sessionKey, sessions.HTTPOnly(), sessions.Name("session"), sessions.Expiry(time.Hour*24*30))
	basketStore, _ = sessions.NewCookieStore(basketKey, sessions.HTTPOnly(), sessions.Name("basket"))
	for _, typ := range basketTypes {
		gob.Register(typ)
	}
}

func GetLogin(r *http.Request) uint64 {
	buf := loginStore.Get(r)
	if len(buf) != 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(buf)
}

func SetLogin(w http.ResponseWriter, userID uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], userID)
	loginStore.Set(w, buf[:])
}

func ClearLogin(w http.ResponseWriter) {
	loginStore.Set(w, nil)
}

func LoadBasket(r *http.Request) *Basket {
	buf := memio.Buffer(basketStore.Get(r))
	basket := new(Basket)
	if len(buf) == 0 {
		return basket
	}
	gob.NewDecoder(&buf).Decode(basket)
	basket.Validate()
	return basket
}

func SaveBasket(w http.ResponseWriter, basket *Basket) {
	var buf memio.Buffer
	gob.NewEncoder(&buf).Encode(basket)
	basketStore.Set(w, buf)
}

func ClearBasket(w http.ResponseWriter) {
	basketStore.Set(w, nil)
}

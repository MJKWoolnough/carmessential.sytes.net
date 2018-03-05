package main

import (
	"encoding/binary"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/MJKWoolnough/memio"
	"github.com/MJKWoolnough/sessions"
)

var Session sess

type sess struct {
	loginStore, basketStore *sessions.CookieStore
}

func (s *sess) init(sessionKey, basketKey string, basketTypes ...interface{}) {
	s.loginStore, _ = sessions.NewCookieStore([]byte(sessionKey), sessions.HTTPOnly(), sessions.Name("session"), sessions.Expiry(time.Hour*24*30))
	s.basketStore, _ = sessions.NewCookieStore([]byte(basketKey), sessions.HTTPOnly(), sessions.Name("basket"))
	for _, typ := range basketTypes {
		gob.Register(typ)
	}
}

func (s *sess) GetLogin(r *http.Request) uint64 {
	buf := s.loginStore.Get(r)
	if len(buf) != 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(buf)
}

func (s *sess) SetLogin(w http.ResponseWriter, userID uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], userID)
	s.loginStore.Set(w, buf[:])
}

func (s *sess) ClearLogin(w http.ResponseWriter) {
	s.loginStore.Set(w, nil)
}

func (s *sess) LoadBasket(r *http.Request) *Basket {
	buf := memio.Buffer(s.basketStore.Get(r))
	basket := new(Basket)
	if len(buf) == 0 {
		return basket
	}
	gob.NewDecoder(&buf).Decode(basket)
	//basket.Validate()
	return basket
}

func (s *sess) SaveBasket(w http.ResponseWriter, basket *Basket) {
	var buf memio.Buffer
	gob.NewEncoder(&buf).Encode(basket)
	s.basketStore.Set(w, buf)
}

func (s *sess) ClearBasket(w http.ResponseWriter) {
	s.basketStore.Set(w, nil)
}

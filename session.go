package main

import (
	"encoding/binary"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/MJKWoolnough/errors"
	"github.com/MJKWoolnough/memio"
	"github.com/MJKWoolnough/sessions"
)

var Session sess

type sess struct {
	loginStore, basketStore *sessions.CookieStore
}

func (s *sess) Init() error {
	var err error
	s.loginStore, err = sessions.NewCookieStore([]byte(Config.Get("sessionKey")), sessions.HTTPOnly(), sessions.Name("session"), sessions.Expiry(time.Hour*24*30))
	if err != nil {
		return errors.WithContext("error initialising Login Store: ", err)
	}
	s.basketStore, err = sessions.NewCookieStore([]byte(Config.Get("basketKey")), sessions.HTTPOnly(), sessions.Name("basket"))
	if err != nil {
		return errors.WithContext("error initialising Basket Store: ", err)
	}
	return nil
}

func (s *sess) RegisterBasketType(basketTypes ...interface{}) {
	for _, typ := range basketTypes {
		gob.Register(typ)
	}
}

func (s *sess) GetLogin(r *http.Request) int64 {
	buf := s.loginStore.Get(r)
	if len(buf) != 8 {
		return 0
	}
	return int64(binary.LittleEndian.Uint64(buf))
}

func (s *sess) SetLogin(w http.ResponseWriter, userID int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(userID))
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

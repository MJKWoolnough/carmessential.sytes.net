package main

import "net/http"

var Admin admin

type admin struct {
}

func (a *admin) init() {

}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if uid := Session.GetLogin(r); !Session.IsAdmin(uid) {
		if uid == 0 {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/user/", http.StatusFound)
		return
	}
	Pages.WriteHeader(w, r, PageHeader{
		Title:       []byte("CARMEssential - Admin"),
		Style:       []byte("user"),
		WriteBasket: true,
	})
	w.Write([]byte("ADMIN"))
	Pages.WriteFooter(w)
}

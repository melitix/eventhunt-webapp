package main

import (
	"net/http"
	"net/url"

	"github.com/eventhunt-org/webapp/webapp/db"

	"github.com/gopherlibs/gpic/gpic"
	log "github.com/sirupsen/logrus"
)

// Deprecated. Will be removed.
// loads a user object if a user is logged in
// otherwise its nil
func initUser(app *app, r *http.Request) (*db.User, *url.URL) {

	session, err := store.Get(r, "login")
	if err != nil {
		log.Error(err)
		return nil, nil
	}

	var userInt uint64
	if session.Values["uid"] == nil || session.Values["uid"] == 0 {
		userInt = uint64(0)
	} else {
		userInt = session.Values["uid"].(uint64)
	}

	user, err := db.GetUserByID(app.DB, userInt)
	if err != nil {
		log.Error(err)
		return nil, nil
	} else if user == nil {
		return nil, nil
	}

	avatar, err := gpic.NewAvatar(user.Email())
	if err != nil {
		log.Error(err)
		return nil, nil
	}

	avatar.SetSize(100)
	picURL, err := avatar.URL()
	if err != nil {
		log.Error(err)
		return nil, nil
	}

	return user, picURL
}

func (this *app) util404Get(w http.ResponseWriter, r *http.Request) {

	user, picURL := initUser(this, r)

	w.WriteHeader(http.StatusNotFound)

	renderPage(this, "util/404.html", w, r, map[string]interface{}{
		"User":        user,
		"GravatarURL": picURL,
	})
}

func (this *app) utilVerifyEmail(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "login")
	if err != nil {
		log.Error(err)
		return
	}

	r.ParseForm()
	code := r.Form.Get("code")

	defer r.Body.Close()

	tok, e, err := db.GetEmailToken(this.DB, code)
	if err != nil {
		log.Error("Error: Failed to get token.")
		session.Values["message"] = "Error: Invalid token."
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if db.IsTokenExpired(tok) {

		session.Values["message"] = "Error: Expired token."
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if db.IsTokenUsed(tok) {

		session.Values["message"] = "Error: This token has already been used."
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	tok.Save() // mark the token as used
	e.Verified = true
	e.Save()

	session.Values["message"] = "Email has been verified."
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

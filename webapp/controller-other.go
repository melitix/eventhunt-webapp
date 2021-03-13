package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

/*
 * Process /invite GET
 */
func (this *app) otherInviteGet(w http.ResponseWriter, r *http.Request) {

	user, picURL := initUser(this, r)

	renderPage(this, "other/invite.html", w, r, map[string]interface{}{
		"User":        user,
		"GravatarURL": picURL,
	})
}

/*
 * Process /invite POST
 */
func (this *app) otherInvitePost(w http.ResponseWriter, r *http.Request) {

	var email string

	user, _ := initUser(this, r)

	if user.ID == 0 {
		log.Fatal("Error, not a user")
	}

	r.ParseForm()
	email = r.Form.Get("email")

	defer r.Body.Close()

	if err := sendEmailInvite(email); err != nil {
		log.Error("Failed to send invite email to: " + email)
		log.Error(err)
	}

	http.Redirect(w, r, "/invite", http.StatusFound)
}

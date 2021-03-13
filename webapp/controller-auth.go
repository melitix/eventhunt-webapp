package main

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/eventhunt-org/webapp/framework"
	"github.com/eventhunt-org/webapp/webapp/db"

	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

/*
 * Display /login GET
 */
func (a *app) authLogin(w http.ResponseWriter, r *http.Request) {

	var messages []string

	// we may or may not have a logged in user
	u, _ := r.Context().Value("user").(*db.User)

	session, _ := store.Get(r, "login")

	if session.Values["message"] != nil && session.Values["message"].(string) != "" {

		messages = append(messages, session.Values["message"].(string))
		session.Values["message"] = ""
		session.Save(r, w)
	}

	renderPage(a, "auth/login.html", w, r, map[string]interface{}{
		"User":     u,
		"Messages": messages,
	})
}

/*
 * Process /login POST
 */
func (this *app) authLoginPost(w http.ResponseWriter, r *http.Request) {

	var verified bool
	var userID uint64

	session, _ := store.Get(r, "login")

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	defer r.Body.Close()

	errs := framework.Validator.Var(username, "required")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: username is required"
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(username, "gt=0,lte=25")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: username must be between 1 and 25 characters"
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(username, "alphanum")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: username cannot contain special characters"
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(password, "required")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: password is required"
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(password, "gt=0,lte=100")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: password must be between 1 and 100 characters"
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	userID = db.VerifyPassword(this.DB, username, password)

	if userID == 0 {
		verified = false
	} else {
		verified = true
	}

	if verified {

		session.Values["authenticated"] = true
		session.Values["uid"] = userID
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	log.Error(errs)

	session.Values["message"] = "Error: username or password is incorrect"
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusFound)
	return
}

/*
 * Process /logout GET
 */
func (this *app) authLogout(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")
	session.Values["authenticated"] = false
	session.Values["uid"] = uint64(0)
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusFound)
	return
}

/*
 * Display /signup GET
 */
func (this *app) authSignup(w http.ResponseWriter, r *http.Request) {

	var messages []string

	user, picURL := initUser(this, r)

	session, _ := store.Get(r, "login")

	if session.Values["message"] != nil && session.Values["message"].(string) != "" {

		messages = append(messages, session.Values["message"].(string))
		session.Values["message"] = ""
		session.Save(r, w)
	}

	renderPage(this, "auth/signup.html", w, r, map[string]interface{}{
		"User":        user,
		"GravatarURL": picURL,
		"Messages":    messages,
	})
}

/*
 * Process /signup POST
 */
func (this *app) authSignupPost(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	email := r.Form.Get("email")
	defer r.Body.Close()

	errs := framework.Validator.Var(username, "required")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: username is required"
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(username, "gt=0,lte=25")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: username must be between 1 and 25 characters"
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(username, "alphanum")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: username cannot contain special characters"
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(email, "required")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: email address is required"
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(email, "gt=0,lte=100")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: email address must be between 1 and 100 characters"
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(password, "required")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: password is required"
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(password, "gt=0,lte=100")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: password must be between 1 and 100 characters"
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	if db.IsEmailTaken(this.DB, email) {

		session.Values["message"] = "Error: this email address is already in use."
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	u, err := db.CreateUser(this.DB, username, password, email, "", "")
	if err != nil {

		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {

			if strings.Contains(err.Error(), "'username'") {
				session.Values["message"] = "Error: username already taken."
				log.Errorf("Failed signup. The username %s is already taken.", username)
			}
		} else {

			session.Values["message"] = "Error: there was an error signing up."
			log.Errorf("There was an error signing up. Message: %s", err)
		}

		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	e, err := db.AddEmailAddress(u, email, true, false)
	if err != nil {
		slog.Error("There was an error signing up. Message: %s", err)
		session.Values["message"] = "Error: there was an error saving the email address."
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	// send verification email
	tok, err := db.NewUserToken(u, "email-verify")
	if err != nil {
		slog.Error("Failed to create user token. Message: %s", err)
		session.Values["message"] = "Error: there was an error saving the email address."
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	if err := sendEmailVerification(e.Value, tok.Token); err != nil {
		slog.Error("There was an error signing up. Message: %s", err)
		session.Values["message"] = "Error: there was an error saving the email address."
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	// send an admin email to Ricardo alerting of new user
	if err := sendEmailGeneric("Ricardo@Feliciano.Tech", "New User Signed Up", "The user "+username+" just signed up to "+AppName+"."); err != nil {
		log.Error("Failed to send admin signup email for user: " + username)
		log.Error(err)
	}

	session.Values["authenticated"] = true
	session.Values["uid"] = u.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
	return
}

/*
 * Display /forgot-password GET
 */
func (this *app) authForgotPasswordGet(w http.ResponseWriter, r *http.Request) {

	var messages []string

	user, picURL := initUser(this, r)

	session, _ := store.Get(r, "login")

	if session.Values["message"] != nil && session.Values["message"].(string) != "" {

		messages = append(messages, session.Values["message"].(string))
		session.Values["message"] = ""
		session.Save(r, w)
	}

	renderPage(this, "auth/forgot.html", w, r, map[string]interface{}{
		"User":        user,
		"GravatarURL": picURL,
		"Messages":    messages,
	})
}

/*
 * Process /forgot-password POST
 */
func (this *app) authForgotPasswordPost(w http.ResponseWriter, r *http.Request) {

	var username string

	session, _ := store.Get(r, "login")

	r.ParseForm()
	username = r.Form.Get("username")
	defer r.Body.Close()

	errs := framework.Validator.Var(username, "required")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: username is required"
		session.Save(r, w)

		http.Redirect(w, r, "/forgot-password", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(username, "gt=0,lte=25")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: username must be between 1 and 25 characters"
		session.Save(r, w)

		http.Redirect(w, r, "/forgot-password", http.StatusFound)
		return
	}

	errs = framework.Validator.Var(username, "alphanum")
	if errs != nil {

		log.Error(errs)

		session.Values["message"] = "Error: username cannot contain special characters"
		session.Save(r, w)

		http.Redirect(w, r, "/forgot-password", http.StatusFound)
		return
	}

	u, err := db.GetUserByUsername(this.DB, username)
	if err != nil {
		log.Error("Error: Failed to get user by username.")
	}

	// if a user is found
	if u != nil {

		t, err := db.NewUserToken(u, "pw-reset")
		if err == nil {
			sendEmailPasswordReset(u, t.Token)
		}
	}

	session.Values["message"] = "If the user exists, an email was sent to reset the password."
	session.Save(r, w)

	http.Redirect(w, r, "/forgot-password", http.StatusFound)
	return
}

/*
 * Display /reset-password GET
 */
func (this *app) resetPasswordGet(w http.ResponseWriter, r *http.Request) {

	var messages []string

	username := r.URL.Query().Get("username")
	t, _ := url.QueryUnescape(r.URL.Query().Get("token"))

	user, picURL := initUser(this, r)

	if username == "" || t == "" {
		messages = append(messages, "Error: Invalid request to reset a password.")
	}

	session, _ := store.Get(r, "login")

	if session.Values["message"] != nil && session.Values["message"].(string) != "" {

		messages = append(messages, session.Values["message"].(string))
		session.Values["message"] = ""
		session.Save(r, w)
	}

	renderPage(this, "auth/reset.html", w, r, map[string]interface{}{
		"User":        user,
		"GravatarURL": picURL,
		"Messages":    messages,
		"Username":    username,
		"Token":       t,
	})
}

/*
 * Process /forgot-password POST
 */
func (this *app) resetPasswordPost(w http.ResponseWriter, r *http.Request) {

	var password string
	var password2 string
	var username string
	var theToken string

	session, _ := store.Get(r, "login")

	r.ParseForm()
	password = r.Form.Get("password")
	password2 = r.Form.Get("password2")
	username = r.Form.Get("username")
	theToken = r.Form.Get("token")
	defer r.Body.Close()

	u, err := db.GetUserByUsername(this.DB, username)
	if err != nil {
		log.Error("Error: Failed to get user.")
		session.Values["message"] = "Error: Failed to get user to reset password."
		session.Save(r, w)

		http.Redirect(w, r, "/forgot-password", http.StatusFound)
		return
	}

	t, err := db.GetTokenByValue(u, theToken)
	if err != nil {
		log.Error("Error: Failed to get token.")
		session.Values["message"] = "Error: Invalid token."
		session.Save(r, w)

		http.Redirect(w, r, "/forgot-password", http.StatusFound)
		return
	}

	if db.IsTokenExpired(t) {

		session.Values["message"] = "Error: Expired token."
		session.Save(r, w)

		http.Redirect(w, r, "/forgot-password", http.StatusFound)
		return
	}

	if db.IsTokenUsed(t) {

		session.Values["message"] = "Error: This token has already been used."
		session.Save(r, w)

		http.Redirect(w, r, "/forgot-password", http.StatusFound)
		return
	}

	err = u.UpdatePassword(password, password2)
	if err != nil {

		log.Error(err)
		session.Values["message"] = "Error: Failed to reset password."
		session.Save(r, w)

		http.Redirect(w, r, "/forgot-password", http.StatusFound)
		return
	}

	// mark token as used
	t.Save()

	session.Values["message"] = "Your password has been reset."
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusFound)
	return
}

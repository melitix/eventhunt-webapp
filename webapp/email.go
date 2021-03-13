package main

import (
	"net/smtp"
	"os"

	"github.com/eventhunt-org/webapp/webapp/db"
	log "github.com/sirupsen/logrus"
)

// Send a generic email where to, subject, and body are passed in
func sendEmailGeneric(email, subject, body string) error {

	from := "no-reply@" + HostnameEmail
	username := os.Getenv("RAG_EMAIL_USER")
	password := os.Getenv("RAG_EMAIL_PWD")
	host := os.Getenv("RAG_EMAIL_HOST")
	port := "587"

	message := []byte("To: " + email + "\r\n" +
		"From: " + AppName + " <" + from + ">\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	if environment == "development" {
		log.Info("We're not in production so outputing generic email here:")
		log.Info(string(message))

		return nil
	}

	auth := smtp.PlainAuth("", username, password, host)

	return smtp.SendMail(host+":"+port, auth, from, []string{email}, message)
}

func sendEmailInvite(email string) error {

	from := "notifications@" + HostnameEmail
	username := os.Getenv("RAG_EMAIL_USER")
	password := os.Getenv("RAG_EMAIL_PWD")
	host := os.Getenv("RAG_EMAIL_HOST")
	port := "587"

	message := []byte("To: " + email + "\r\n" +
		"From: " + AppName + " <" + from + ">\r\n" +
		"Subject: Checkout " + AppName + "\r\n" +
		"\r\n" +
		"Hey, you should try " + AppName + " (https://" + hostname + ")!\r\n")

	auth := smtp.PlainAuth("", username, password, host)

	return smtp.SendMail(host+":"+port, auth, from, []string{email}, message)
}

func sendEmailPasswordReset(u *db.User, token string) error {

	from := "notifications@" + HostnameEmail
	username := os.Getenv("RAG_EMAIL_USER")
	password := os.Getenv("RAG_EMAIL_PWD")
	host := os.Getenv("RAG_EMAIL_HOST")
	port := "587"

	message := []byte("To: " + u.Email() + "\r\n" +
		"From: " + AppName + " <" + from + ">\r\n" +
		"Subject: " + AppName + " - Reset your Password\r\n" +
		"\r\n" +
		"You can reset your password by clicking the following link. If you " + "\r\n" +
		"didn't request this email, please ignore it." + "\r\n" +
		"\r\n" +
		"Reset your password: https://" + hostname + "/reset-password?username=" + u.Username + "&token=" + token + "\r\n")

	if environment == "development" {
		log.Info("We're not in production so outputing a password reset email here:")
		log.Info(string(message))

		return nil
	}

	auth := smtp.PlainAuth("", username, password, host)

	return smtp.SendMail(host+":"+port, auth, from, []string{u.Email()}, message)
}

func sendEmailVerification(email, code string) error {

	from := "notifications@" + HostnameEmail
	username := os.Getenv("RAG_EMAIL_USER")
	password := os.Getenv("RAG_EMAIL_PWD")
	host := os.Getenv("RAG_EMAIL_HOST")
	port := "587"

	message := []byte("To: " + email + "\r\n" +
		"From: " + AppName + " <" + from + ">\r\n" +
		"Subject: " + AppName + " - Verify Your Email\r\n" +
		"\r\n" +
		"To start using " + AppName + ", please verify your email address by " + "\r\n" +
		"clicking this link: https://" + hostname + "/verify-email?code=" + code + "\r\n")

	if environment == "development" || environment == "testing" {
		log.Info("We're not in production so outputing an email verification email here:")
		log.Info(string(message))

		return nil
	}

	auth := smtp.PlainAuth("", username, password, host)

	return smtp.SendMail(host+":"+port, auth, from, []string{email}, message)
}

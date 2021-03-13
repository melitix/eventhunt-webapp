package sso

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/url"

	"github.com/spf13/viper"
)

func PayloadExtract(encPayload string, outbound bool) (url.Values, error) {

	payload, err := base64.URLEncoding.DecodeString(encPayload)
	if err != nil {
		slog.Error("Failed to decode payload.", "msg", err, "encPayload", encPayload)
		return nil, err
	}

	values, err := url.ParseQuery(string(payload))
	if err != nil {
		slog.Error("Failed to parse payload.", "msg", err, "payload", payload)
		return nil, err
	}

	//
	// check for required values
	//
	if !values.Has("nonce") {
		err := errors.New("Required key nonce is missing.")
		slog.Error("Payload is missing a key.", "msg", err)
		return nil, err
	}

	if outbound && !values.Has("return_sso_url") {
		err := errors.New("Required key return_sso_url is missing.")
		slog.Error("Payload is missing a key.", "msg", err)
		return nil, err
	}

	if !outbound && !values.Has("external_id") {
		err := errors.New("Required key external_id is missing.")
		slog.Error("Payload is missing a key.", "msg", err)
		return nil, err
	}

	return values, nil
}

func PayloadPack(urlStr string, values url.Values) *url.URL {

	encPayload := base64.URLEncoding.EncodeToString([]byte(values.Encode()))

	hash := hmac.New(sha256.New, []byte(viper.GetString("AUTH_SECRET")))
	hash.Write([]byte(encPayload))
	sig := hex.EncodeToString(hash.Sum(nil))

	payloadURL, err := url.Parse(urlStr)
	if err != nil {
		slog.Error("Failed to build URL.", "msg", err)
		return nil
	}
	q := payloadURL.Query()
	q.Set("sso", encPayload)
	q.Set("sig", sig)
	payloadURL.RawQuery = q.Encode()

	return payloadURL
}

func PayloadValidate(payload, signature string) bool {

	hash := hmac.New(sha256.New, []byte(viper.GetString("AUTH_SECRET")))
	hash.Write([]byte(payload))
	sig := hex.EncodeToString(hash.Sum(nil))

	return signature == sig
}

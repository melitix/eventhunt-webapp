package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type token struct {
	framework.BaseModel
	userID     uint      `json:"user_id"`
	Token      string    `json:"token"`
	tokenHash  string    `json:"tokenHash"`
	expiration time.Time `json:"expiration"`
	purpose    string    `json:"purpose"`
}

/*
 * Delete token from database.
 */
func (t *token) delete() error {

	_, err := t.DB.Exec(context.Background(), "DELETE FROM user_tokens WHERE id=$1", t.ID)
	if err != nil {
		log.Error(err)
	}

	return err
}

/*
 * Load from the db. Essentially updating the struct with potentially newer info.
 */
func (this *token) load() error {

	err := this.DB.QueryRow(context.Background(), "SELECT * FROM user_tokens WHERE id=$1", this.ID).Scan(
		&this.ID, &this.userID, &this.tokenHash, &this.expiration, &this.purpose, &this.CreatedTime, &this.UpdatedTime)
	if err != nil {

		if err == sql.ErrNoRows {
			// this is an okay error
			return err
		} else {
			log.Error(err)
		}
	}

	return err
}

/*
 * Save by simply updating the 'UpdatedAt' field.
 */
func (this *token) Save() error {

	_, err := this.DB.Exec(context.Background(), "UPDATE user_tokens SET updated_at = NOW() WHERE id=$1",
		this.ID)
	if err != nil {
		log.Error(err)
	}

	return err
}

//=============================================================================
// End of methods, start of functions
//=============================================================================

/*
 * Create a new user token. This can be used for password resets or the API.
 */
func NewUserToken(u *User, purpose string) (*token, error) {

	var lastInsertID int

	if purpose != "pw-reset" && purpose != "email-verify" {

		return nil, errors.New("Error: An invalid tokek type was attempted to be created.")
	}

	rBytes := make([]byte, 15)
	_, err := rand.Read(rBytes)
	if err != nil {
		return nil, errors.New("Error: Reading random failed.")
	}

	rToken := base64.RawURLEncoding.EncodeToString(rBytes)

	// hash token
	hToken, err := bcrypt.GenerateFromPassword([]byte(rToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Error: Hashing token failed.")
	}

	err = u.DB.QueryRow(context.Background(), "INSERT INTO user_tokens (user_id, the_value, expiration, purpose) VALUES ($1, $2, $3, $4) RETURNING id",
		u.ID, hToken, time.Now().UTC().Add(time.Hour*1), purpose).Scan(&lastInsertID)
	if err != nil {
		return nil, errors.New("Error: Failed to store new user token.")
	}

	var t token
	t.ID = uint64(lastInsertID)
	t.DB = u.DB
	t.Token = string(rToken)
	t.load()

	return &t, nil
}

/*
 * Checks if a token is expired. It does this by confirming that the token is in
 * the DB and it is not expired.
 * Will also return false on error.
 */
func IsTokenExpired(t *token) bool {

	if time.Now().After(t.expiration) || time.Now().Equal(t.expiration) {
		return true
	}

	return false
}

/*
 * Checks if a token has been used before. It does this by comparing the
 * creation date to the updated date.
 * Will also return false on error.
 */
func IsTokenUsed(t *token) bool {

	if t.UpdatedTime.Equal(t.CreatedTime) {
		return false
	}

	return true
}

/*
 * This is the main function that retrieves tokens from the DB.
 * Several helper functions may exists to specific by what field/clause to
 * retrieve the address.
 */
func GetUserTokenBy(db *pgxpool.Pool, clause string) (*token, error) {

	var t token

	q := fmt.Sprintf("SELECT id FROM user_tokens WHERE %s", clause)

	err := db.QueryRow(context.Background(), q).Scan(&t.ID)
	if err != nil {

		if err == sql.ErrNoRows {
			// this is an okay error
			return nil, err
		} else {
			log.Error(err)
			return nil, err
		}
	}

	t.DB = db
	t.load()

	return &t, err
}

/*
 * Get token by its value.
 */
func GetEmailToken(db *pgxpool.Pool, tValue string) (*token, *emailAddress, error) {

	tokens, err := GetActiveTokens(db, 0, 100)
	if err != nil {
		return nil, nil, errors.New("Failed to get active token list.")
	}

	for _, t := range tokens {

		err = bcrypt.CompareHashAndPassword([]byte(t.tokenHash), []byte(tValue))
		if err == nil {
			e, err := GetEmailAddressByID(db, int(t.userID))
			if err != nil {
				return nil, nil, err
			}
			return t, e, nil
		}
	}

	return nil, nil, errors.New("Error: Token not found.")
}

/*
 * Get token by its value.
 */
func GetTokenByValue(u *User, tValue string) (*token, error) {

	tokens, err := GetActiveTokens(u.DB, 0, 100)
	if err != nil {
		return nil, errors.New("Error: Failed to get token list.")
	}

	for _, t := range tokens {

		err = bcrypt.CompareHashAndPassword([]byte(t.tokenHash), []byte(tValue))
		if err == nil {
			return t, nil
		}
	}

	return nil, errors.New("Error: Token not found.")
}

/*
 * Get every token for a user.
 */
func GetTokensByUser(u *User, start int, count int) ([]*token, error) {

	rows, err := u.DB.Query(context.Background(), "SELECT id FROM user_tokens WHERE user_id = $1 LIMIT $2 OFFSET $3", u.ID, count, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := []*token{}

	for rows.Next() {

		var t token

		if err := rows.Scan(&t.ID); err != nil {
			return nil, errors.New("Error: Failed to receive next token.")
		}

		t.DB = u.DB
		t.load()
		tokens = append(tokens, &t)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return tokens, nil
}

/*
 * Get active tokens.
 */
func GetActiveTokens(db *pgxpool.Pool, start int, count int) ([]*token, error) {

	rows, err := db.Query(context.Background(), "SELECT id FROM user_tokens WHERE created_time = updated_time LIMIT $1 OFFSET $2", count, start)
	if err != nil {
		return nil, errors.New("Active tokens query failed. Err: " + err.Error())
	}
	defer rows.Close()

	tokens := []*token{}

	for rows.Next() {

		var t token

		if err := rows.Scan(&t.ID); err != nil {
			return nil, errors.New("Error: Failed to receive next token.")
		}

		t.DB = db
		t.load()
		tokens = append(tokens, &t)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return tokens, nil
}

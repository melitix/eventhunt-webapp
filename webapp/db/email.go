package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	log "github.com/sirupsen/logrus"
)

type emailAddress struct {
	framework.BaseModel
	userID    uint   `json:"user_id"`
	Value     string `json:"value"`
	preferred bool   `json:"preferred"`
	Verified  bool   `json:"verified"`
}

/*
 * Delete this item.
 */
func (ea *emailAddress) delete() error {

	_, err := ea.DB.Exec(context.Background(), "DELETE FROM email_addresses WHERE `id`=$1", ea.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

/*
 * Load from the db. Essentially updating the struct with potentially newer info.
 */
func (this *emailAddress) load() error {

	err := this.DB.QueryRow(context.Background(), "SELECT * FROM email_addresses WHERE id=$1", this.ID).Scan(
		&this.ID, &this.userID, &this.Value, &this.preferred, &this.Verified, &this.CreatedTime, &this.UpdatedTime)
	if err != nil {

		if err == sql.ErrNoRows {
			// this is an okay error
			return err
		} else {
			log.Fatal(err)
		}
	}

	return err
}

/*
 * Save struct to database.
 */
func (this *emailAddress) Save() error {

	_, err := this.DB.Exec(context.Background(), "UPDATE email_addresses SET preferred = $1, verified = $2 WHERE id=$3",
		this.preferred, this.Verified, this.ID)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

//=============================================================================
// End of methods, start of functions
//=============================================================================

/*
 * Add new email. If verified equals false, will send an email verification
 * email.
 */
func AddEmailAddress(u *User, value string, preferred, verified bool) (*emailAddress, error) {

	var lastInsertID int

	if value == "" {
		return nil, errors.New("Error: Email address cannot be blank.")
	}

	err := u.DB.QueryRow(context.Background(), "INSERT INTO email_addresses (user_id, the_value, preferred, verified) VALUES ($1, $2, $3, $4) RETURNING id",
		u.ID, value, preferred, verified).Scan(&lastInsertID)
	if err != nil {
		slog.Error("Failed to add email.", "address", value, "userID", u.ID)
		return nil, err
	}

	e := emailAddress{}
	e.ID = uint64(lastInsertID)
	e.DB = u.DB
	e.load()

	return &e, nil
}

/*
 * This is the main function that retrieves an email address from the DB.
 * Several helper functions may exists to specific by what field/clause to
 * retrieve the address.
 */
func GetEmailAddressBy(db *pgxpool.Pool, clause string) (*emailAddress, error) {

	e := emailAddress{}

	q := fmt.Sprintf("SELECT id FROM email_addresses WHERE %s", clause)

	rows, _ := db.Query(context.Background(), q)
	eaID, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[uint64])
	if err != nil {
		return nil, err
	}

	e.ID = eaID
	e.DB = db
	e.load()

	return &e, err
}

/*
 * Get struct by its database ID.
 */
func GetEmailAddressByID(db *pgxpool.Pool, id int) (*emailAddress, error) {

	return GetEmailAddressBy(db, "id="+strconv.Itoa(id))
}

/*
 * Get preferred email of a user.
 */
func GetPreferredEmailByUser(u *User) (*emailAddress, error) {
	return GetEmailAddressBy(u.DB, fmt.Sprintf("user_id=%d AND preferred=true", u.ID))
}

/*
 * Checks if an email address is already in-use in the database.
 */
func IsEmailTaken(db *pgxpool.Pool, email string) bool {

	if e, _ := GetEmailAddressBy(db, "the_value='"+email+"'"); e != nil {
		return true
	}

	return false
}

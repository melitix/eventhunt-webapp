package db

import (
	"context"
	"fmt"
	"time"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const DB_TABLE_RSVP = "rsvps"

type RSVPStatus string

const (
	RSVPYes      RSVPStatus = "yes"
	RSVPInPerson RSVPStatus = "in-person"
	RSVPOnline   RSVPStatus = "online"
	RSVPMaybe    RSVPStatus = "maybe"
	RSVPNo       RSVPStatus = "no"
)

type RSVPRole string

const (
	RSVPAttendee RSVPRole = "attendee"
	RSVPHost     RSVPRole = "host"
	RSVPCrew     RSVPRole = "crew"
)

/*
 * RSVP represents a user's intention (and result) of attending an event.
 */
type RSVP struct {
	framework.BaseModel
	EventID      uint64      `db:"event_id" validate:"required"`
	UserID       uint64      `db:"user_id" validate:"required"`
	Intent       RSVPStatus  `db:"intent" validate:"required"`
	Actual       *RSVPStatus `db:"actual"`
	Role         RSVPRole    `db:"role"`
	RemindedTime *time.Time  `db:"reminded_time"`
}

/*
 * primaryKey returns the primary key name of the table
 */
func (r *RSVP) primaryKey() string { return "" }

/*
 * save serializes the struct to the database. The update is done via primary
 * key.
 */
func (r *RSVP) Save() error {

	q := `UPDATE ` + r.table() + ` 
		SET intent=@intent,
			actual=@actual,
			role=@role,
			reminded_time=@remindedTime,
			updated_time=@updatedTime
		WHERE event_id=@eventID AND user_id=@userID`
	_, err := r.DB.Exec(context.Background(), q, pgx.NamedArgs{
		"intent":       r.Intent,
		"actual":       r.Actual,
		"role":         r.Role,
		"remindedTime": r.RemindedTime,
		"updatedTime":  r.UpdatedTime,
		"eventID":      r.EventID,
		"userID":       r.UserID,
	})

	return err
}

/*
 * table returns the table name used in the database.
 */
func (r *RSVP) table() string { return DB_TABLE_RSVP }

//==============================================================================
// End of methods, start of functions
//==============================================================================

/*
 * Internal init function.
 */
func initRSVP(db *pgxpool.Pool) *RSVP {

	r := new(RSVP)
	r.DB = db

	return r
}

/*
 * NewRSVP creates a new RSVP struct, validates it, and if good, saves it to
 * the database.
 */
func NewRSVP(eventID uint64, u *User, intent RSVPStatus, role RSVPRole) (*RSVP, error) {

	r := initRSVP(u.DB)
	r.EventID = eventID
	r.UserID = u.ID
	r.Intent = intent
	r.Role = role

	err := validate.Struct(r)
	if err != nil {
		return nil, err
	}

	q := `INSERT INTO ` + r.table() + ` 
		(event_id, user_id, intent, role) 
		VALUES (@eventID, @userID, @intent, @role) RETURNING *`
	rows, _ := u.DB.Query(context.Background(), q, pgx.NamedArgs{
		"eventID": r.EventID,
		"userID":  r.UserID,
		"intent":  r.Intent,
		"role":    r.Role,
	})

	r, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[RSVP])
	if err != nil {
		return nil, fmt.Errorf("Failed to create RSVP. Err: %s", err)
	}

	return r, nil
}

/*
 * GetRSVP returns a user's RSVP based on the event provided.
 *
 */
func GetRSVP(db *pgxpool.Pool, eventID, userID uint64) (*RSVP, error) {

	q := `SELECT * FROM ` + DB_TABLE_RSVP + ` WHERE event_id=@eventID AND user_id=@userID`
	rows, _ := db.Query(context.Background(), q, pgx.NamedArgs{
		"eventID": eventID,
		"userID":  userID,
	})

	return pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[RSVP])
}

/*
 * GetRSVPsByEvent returns an event's RSVPs.
 *
 */
func GetRSVPsByEvent(db *pgxpool.Pool, eventID uint64) ([]*RSVP, error) {

	q := `SELECT * FROM ` + DB_TABLE_RSVP + ` WHERE event_id=@eventID`
	rows, _ := db.Query(context.Background(), q, pgx.NamedArgs{
		"eventID": eventID,
	})

	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RSVP])
}

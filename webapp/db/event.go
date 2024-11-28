package db

import (
	"context"
	"strconv"
	"time"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const DB_TABLE_EVENT = "events"

/*
 * event represents a physical event in the real-world.
 */
type Event struct {
	framework.BaseModel
	GroupID       uint64    `db:"group_id" validate:"required"`
	TheGroup      *Group    `db:"-"`
	Name          string    `db:"name" validate:"required,min=3,max=80"`
	StartTime     time.Time `db:"start_time"`
	EndTime       time.Time `db:"end_time"`
	Summary       string    `db:"summary"`
	Description   string    `db:"description"`
	WebURL        string    `db:"web_url"`
	AnnounceURL   string    `db:"announce_url"`
	AttendeeLimit int       `db:"attendee_limit"`
	VenueID       *uint64   `db:"venue_id"`
	Venue         *venue    `db:"-"`
	LocationURL   string    `db:"location_url"`
}

/*
 * primaryKey returns the primary key name of the table
 */
func (e *Event) primaryKey() string { return "id" }

/*
 * RSVPs returns a slice of RSVP for this event.
 */
func (e *Event) RSVPs() ([]*RSVP, error) {
	return GetRSVPsByEvent(e.DB, e.ID)
}

/*
 * save serializes the struct to the database. The update is done via primary
 * key.
 */
func (e *Event) Save() error {

	q := `UPDATE ` + e.table() + `
	SET name=@name,
		start_time=@startTime,
		end_time=@endTime,
		summary=@summary,
		description=@description,
		web_url=@webURL,
		announce_url=@announceURL,
		attendee_limit=@attendeeLimit,
		venue_id=@venueID,
		location_url=@locationURL,
		updated_time=@updatedTime
	WHERE ` + e.primaryKey() + `=@id`

	_, err := e.DB.Exec(context.Background(), q,
		pgx.NamedArgs{
			"name":          e.Name,
			"startTime":     e.StartTime,
			"endTime":       e.EndTime,
			"summary":       e.Summary,
			"description":   e.Description,
			"webURL":        e.WebURL,
			"announceURL":   e.AnnounceURL,
			"attendeeLimit": e.AttendeeLimit,
			"venueID":       e.VenueID,
			"locationURL":   e.LocationURL,
			"updatedTime":   e.UpdatedTime,
			"id":            e.ID,
		})

	return err
}

/*
 * table returns the table name used in the database.
 */
func (e *Event) table() string { return "events" }

//==============================================================================
// End of methods, start of functions
//==============================================================================

/*
 * Internal init function.
 */
func initEvent(db *pgxpool.Pool) *Event {

	e := new(Event)
	e.DB = db

	return e
}

/*
 * Create a new event in the system. Includes the base event itself, not
 * information such as the venue or groups.
 */
func NewEvent(u *User, groupID uint64, name string, startTime, endTime time.Time, summary string) (*Event, error) {

	e := initEvent(u.DB)
	e.Name = name
	e.StartTime = startTime.UTC()
	e.EndTime = endTime.UTC()
	e.GroupID = groupID
	e.Summary = summary

	// validate inputs
	err := validate.Struct(e)
	if err != nil {
		return nil, err
	}

	q := `INSERT INTO ` + e.table() + ` (group_id, name, start_time, end_time, summary) VALUES (@groupID, @name, @startTime, @endTime, @summary) RETURNING *`
	rows, _ := e.DB.Query(context.Background(), q, pgx.NamedArgs{
		"groupID":   groupID,
		"name":      e.Name,
		"startTime": e.StartTime,
		"endTime":   e.EndTime,
		"summary":   e.Summary,
	})

	e, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[Event])
	if err != nil {
		return nil, err
	}

	// When creating a new event, the user who created the event is automatically
	// added as the host.
	_, err = NewRSVP(e.ID, u, RSVPYes, RSVPHost)

	return e, err
}

/*
 * The generic helper to retrieve the model by a column.
 */
func getEventBy(db *pgxpool.Pool, column, value string) (*Event, error) {

	e := initEvent(db)

	q := `SELECT * FROM ` + e.table() + ` WHERE ` + column + `=$1`
	rows, _ := db.Query(context.Background(), q, value)
	e, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[Event])
	if err != nil {
		return nil, err
	}
	e.DB = db
	g, err := GetGroupByID(db, e.GroupID)
	if err != nil {
		return nil, err
	}
	e.TheGroup = g

	// Try loading venue only if we have an ID
	if e.VenueID != nil && *e.VenueID != uint64(0) {

		e.Venue, err = GetVenueByID(db, *e.VenueID)
		if err != nil {
			return nil, err
		}
		e.Venue.DB = e.DB
	}

	return e, nil
}

/*
 * Get the model by ID. This is better to use than the generic helper to ensure
 * type safety.
 */
func GetEventByID(db *pgxpool.Pool, id uint64) (*Event, error) {
	return getEventBy(db, "id", strconv.FormatUint(id, 10))
}

/*
 * GetEventsByQuery returns a slice of Event from the DB based on the provided query.
 */
func GetEventsByQuery(db *pgxpool.Pool, q string, args any) ([]*Event, error) {

	var rows pgx.Rows

	if args == nil {

		rows, _ = db.Query(context.Background(), q)
	} else {
		rows, _ = db.Query(context.Background(), q, args)
	}
	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Event])
	if err != nil {
		return nil, err
	}

	for _, e := range events {

		e.DB = db

		e.TheGroup, err = GetGroupByID(db, e.GroupID)
		if err != nil {
			return nil, err
		}
	}

	return events, nil
}

/*
 * The generic helper to retrieve the model by a column.
 */
func GetEvents(db *pgxpool.Pool, limit int) ([]*Event, error) {

	q := `SELECT * FROM ` + DB_TABLE_EVENT + ` LIMIT @limit`
	args := pgx.NamedArgs{
		"limit": limit,
	}

	return GetEventsByQuery(db, q, args)
}

/*
 * GetEventsByGroup returns a slice of Event.
 */
func GetEventsByGroup(db *pgxpool.Pool, groupID uint64, pastEvents bool, limit uint8) ([]*Event, error) {

	var op string

	if pastEvents {
		op = "<"
	} else {
		op = ">="
	}

	q := `SELECT * FROM ` + DB_TABLE_EVENT + ` WHERE start_time ` + op + ` CURRENT_TIMESTAMP AND group_id=@id`
	args := pgx.NamedArgs{
		"id": groupID,
	}

	return GetEventsByQuery(db, q, args)
}

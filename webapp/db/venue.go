package db

import (
	"context"
	"fmt"
	"net/url"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
 * venue represents a place where an event takes place.
 */
type venue struct {
	framework.BaseModel
	Name    string  `db:"name"`
	Address string  `db:"address"`
	CityID  int     `db:"city_id"`
	WebURL  url.URL `db:"web_url"`
}

/*
 * primaryKey returns the primary key name of the table
 */
func (g *venue) primaryKey() string { return "id" }

/*
 * save serializes the struct to the database. The update is done via primary
 * key.
 */
func (v *venue) save() error {

	_, err := v.DB.Exec(context.Background(), "UPDATE "+v.table()+" SET name = $1, address = $2, city_id = $3, web_url = $4, updated_time = $5 WHERE "+v.primaryKey()+" = $6",
		v.Name,
		v.Address,
		v.CityID,
		v.WebURL,
		v.UpdatedTime,
		v.ID,
	)

	return err
}

/*
 * table returns the table name used in the database.
 */
func (e *venue) table() string { return "venues" }

//==============================================================================
// End of methods, start of functions
//==============================================================================

/*
 * getVenueByID returns a Venue from the DB by its ID.
 *
 */
func getVenueByID(db *pgxpool.Pool, id string) (*venue, error) {

	v := initVenue(db)

	q := `SELECT ` + v.primaryKey() + ` FROM ` + v.table() + ` WHERE ` + v.primaryKey() + ` = $1`
	rows, _ := db.Query(context.Background(), q, id)
	v, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[venue])
	if err != nil {
		return nil, err
	}

	return v, nil
}

/*
 * Internal init function.
 */
func initVenue(db *pgxpool.Pool) *venue {

	v := new(venue)
	v.DB = db

	return v
}

/*
 * NewVenue creates a new Venue in the DB.
 */
func NewVenue(db *pgxpool.Pool, name, address string, cityID uint64) (*venue, error) {

	// validate inputs
	errs := validate.Var(name, "required,min=3,max=20")
	if errs != nil {
		return nil, errs
	}
	errs = validate.Var(address, "required,min=3,max=20")
	if errs != nil {
		return nil, errs
	}

	v := initVenue(db)

	_, err := v.DB.Exec(context.Background(), "INSERT INTO "+v.table()+" (name, address, city_id, web_url, capacity) VALUES ($1, $2, $3, '', 0)",
		name,
		address,
		cityID,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed save venue to DB. Message: %s", err)
	}

	return v, nil
}

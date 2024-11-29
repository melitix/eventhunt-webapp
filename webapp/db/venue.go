package db

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
 * venue represents a place where an event takes place.
 */
type venue struct {
	framework.BaseModel
	Name     string `db:"name"`
	Address  string `db:"address"`
	CityID   uint64 `db:"city_id"`
	WebURL   string `db:"web_url"`
	Capacity uint   `db:"capacity"`
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
 * String returns a one-line representation of the venue address.
 */
func (v *venue) String() string {

	city, err := GetCityByID(v.DB, v.CityID)
	if err != nil {
		slog.Error("GetCityByID failed.", "err", err)
		return "error"
	}

	return v.Address + ", " + city.Name + ", " + city.Admin1
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
func GetVenueByID(db *pgxpool.Pool, id uint64) (*venue, error) {

	v := initVenue(db)

	q := `SELECT * FROM ` + v.table() + ` WHERE ` + v.primaryKey() + ` = $1`
	rows, _ := db.Query(context.Background(), q, strconv.FormatUint(id, 10))
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

	v := initVenue(db)
	v.Name = name
	v.Address = address
	v.CityID = cityID

	// validate inputs
	errs := validate.Var(name, "required,min=3,max=26")
	if errs != nil {
		return nil, errs
	}
	errs = validate.Var(address, "required,min=3,max=40")
	if errs != nil {
		return nil, errs
	}

	q := `INSERT INTO ` + v.table() + ` (name, address, city_id, web_url, capacity) VALUES (@name, @address, @cityID, '', 0) RETURNING *`
	rows, _ := v.DB.Query(context.Background(), q, pgx.NamedArgs{
		"name":    v.Name,
		"address": v.Address,
		"cityID":  v.CityID,
	})

	v, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[venue])
	if err != nil {
		return nil, fmt.Errorf("Failed create venue in DB. Err: %s", err)
	}

	return v, nil
}

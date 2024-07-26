package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const DB_TABLE_CITY = "cities"

/*
 * event represents a physical event in the real-world.
 * City represents a city in the world as represented by GeoNames.org.
 */
type City struct {
	framework.BaseModel
	// In the U.S., this is the state
	Admin1 string `db:"admin1" validate:"required"`
	Name   string `db:"name" validate:"required,min=2,max=200"`
}

/*
 * primaryKey returns the primary key name of the table.
 */
func (city *City) primaryKey() string { return "id" }

//==============================================================================
// End of methods, start of functions
//==============================================================================

/*
 * GetCitiesBy returns a slice of city as determined by the provided SQL query.
 */
func GetCitiesBy(db *pgxpool.Pool, q string, args ...any) ([]City, error) {

	var rows pgx.Rows

	if args == nil {
		rows, _ = db.Query(context.Background(), q)
	} else {
		rows, _ = db.Query(context.Background(), q, args)
	}

	cities, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[City])
	if err != nil {
		return nil, fmt.Errorf("Failed to get city list from DB. Err: %s", err)
	}

	return cities, nil
}

func GetCityByID(db *pgxpool.Pool, id uint64) (*City, error) {

	q := `SELECT id,admin1,name FROM ` + DB_TABLE_CITY + ` WHERE id=` + strconv.FormatUint(id, 10)

	cities, err := GetCitiesBy(db, q)
	if err != nil {
		return nil, fmt.Errorf("Failed to get city list from DB. Err: %s", err)
	}

	city := cities[0]

	return &city, nil
}

func GetCitiesByAll(db *pgxpool.Pool) ([]City, error) {

	q := `SELECT id,admin1,name FROM ` + DB_TABLE_CITY

	return GetCitiesBy(db, q)
}

/*
 * Internal init function.
 */
func initCity(db *pgxpool.Pool) *City {

	city := new(City)
	city.DB = db

	return city
}

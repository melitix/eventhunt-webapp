package framework

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
 * A model represents structured data that can undergo standard CRUD operations
 * with a MariaDB database.
 *
 * As the current plan is not to create an ORM, this interface is around to
 * simply help reduce boilerplate and for programmers to remember what their
 * models should have in terms of methods.
 *
 * Models implementing this interface should use the suffix `Model` and should
 * embed [BaseModel].
 */
type Model interface {

	// just a timesaver
	IDString()

	// Loads fresh data from the database into the struct
	load() error

	// Returns the primary key used in this database table. This is a method
	// instead of a field because interfaces don't include fields in Go.
	primaryKey() string

	// Saves data from the struct to the database. Ideally should only include
	// fields that could change in value.
	save() error

	// Returns the table name used in the database. This is a method instead of
	// a field because there can be occassions where a model can be saved in
	// one of multiple database tables. In that case, the table name needs to
	// be dynamic. Also, interfaces don't include fields in Go.
	table() string
}

/*
 * BaseModel defines the minimum that all models should contain in terms of
 * fields. It's intended to be used with the [Model] interface.
 */
type BaseModel struct {
	DB          *pgxpool.Pool `db:"-"`
	ID          uint64        `db:"id"`
	CreatedTime time.Time     `db:"created_time",json:"created_time"`
	UpdatedTime time.Time     `db:"updated_time",json:"updated_time"`
}

func (m *BaseModel) IDString() string {
	return fmt.Sprintf("%d", m.ID)
}

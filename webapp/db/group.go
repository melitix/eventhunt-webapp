package db

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const DB_TABLE_GROUP = "groups"

/*
 * group represents an owner of events. Sometimes called a team.
 */
type Group struct {
	framework.BaseModel
	UserID      uint64   `db:"user_id"`
	Name        string   `db:"name" validate:"required,min=3,max=20"`
	Summary     string   `db:"summary" validate:"required,min=3,max=200"`
	Description string   `db:"description"`
	Slug        string   `db:"slug"`
	WebURL      *url.URL `db:"web_url"`
	CityID      uint64   `db:"city_id" validate:"required"`
	IsPrivate   bool     `db:"is_private"`
}

/*
 * IsMember returns true if the provided ID (User) is a member of this Group.
 */
func (g *Group) IsMember(id uint64) bool {

	memberships := g.Memberships()

	for _, ms := range memberships {
		if ms.TheUser.ID == id {
			return true
		}
	}

	return false
}

/*
 * Members returns a slice of User who belong to the Group..
 */
func (g *Group) Memberships() []*Membership {

	memberships, err := GetMembershipsByGroup(g.DB, g.ID)
	if err != nil {
		slog.Error("Failed to get members for group.", "groupID", g.ID, "err", err)
	}

	return memberships
}

/*
 * PastEvents returns n number of past events.
 */
func (g *Group) PastEvents(count uint8) []*Event {

	events, err := GetEventsByGroup(g.DB, g.ID, true, 10)
	if err != nil {
		slog.Error("Failed to pull past events.", "groupID", g.ID)
	}

	return events
}

/*
 * primaryKey returns the primary key name of the table
 */
func (g *Group) primaryKey() string { return "id" }

/*
 * save serializes the struct to the database. The update is done via primary
 * key.
 */
func (g *Group) Save() error {

	q := `UPDATE ` + g.table() + ` 
		SET user_id=@userID,
			name=@name,
			city_id=@cityID,
			is_private=@isPrivate,
			updated_time=@updatedTime
		WHERE ` + g.primaryKey() + ` = @id`
	_, err := g.DB.Exec(context.Background(), q, pgx.NamedArgs{
		"userID":      g.UserID,
		"name":        g.Name,
		"cityID":      g.CityID,
		"isPrivate":   g.IsPrivate,
		"updatedTime": g.UpdatedTime,
		"id":          g.ID,
	})

	return err
}

/*
 * table returns the table name used in the database.
 */
func (g *Group) table() string { return "groups" }

/*
 * UpcomingEvents returns n number of future events.
 */
func (g *Group) UpcomingEvents(count uint8) []*Event {

	events, err := GetEventsByGroup(g.DB, g.ID, false, 10)
	if err != nil {
		slog.Error("Failed to pull upcoming events.", "groupID", g.ID)
	}

	return events
}

//==============================================================================
// End of methods, start of functions
//==============================================================================

/*
 * getGroupByID tries to get a Group from the database by its ID.
 */
func getGroupByID(db *pgxpool.Pool, id string) (*Group, error) {

	g := initGroup(db)

	q := `SELECT ` + g.primaryKey() + ` FROM ` + g.table() + ` WHERE ` + g.primaryKey() + ` = $1`
	rows, _ := db.Query(context.Background(), q, id)
	g, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[Group])
	if err != nil {
		return nil, err
	}

	return g, nil
}

/*
 * Internal init function.
 */
func initGroup(db *pgxpool.Pool) *Group {

	g := new(Group)
	g.DB = db

	return g
}

/*
 * NewGroup creates a new Group struct, validates it, and if good, saves itr to
 * the database.
 */
func NewGroup(u *User, name string, cityID uint64, summary string, isPrivate bool) (*Group, error) {

	g := initGroup(u.DB)
	g.Name = name
	g.CityID = cityID
	g.Summary = summary
	g.IsPrivate = isPrivate

	err := validate.Struct(g)
	if err != nil {
		return nil, err
	}

	q := `INSERT INTO ` + g.table() + ` 
		(user_id, name, summary, description, slug, city_id, is_private) 
		VALUES (@userID, @name, @summary, '', '', @cityID, @isPrivate) RETURNING *`
	rows, _ := u.DB.Query(context.Background(), q, pgx.NamedArgs{
		"userID":    u.ID,
		"name":      g.Name,
		"cityID":    g.CityID,
		"summary":   g.Summary,
		"isPrivate": g.IsPrivate,
	})
	g, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[Group])
	if err != nil {
		return nil, fmt.Errorf("Failed to create group. Err: %s", err)
	}

	// Now that we have a Group, let's add the User who created it as its first
	// member. The User's role will be 'owner'.
	_, err = NewMembership(g.ID, u, MemberOwner)
	if err != nil {
		return nil, fmt.Errorf("Failed to create group membership. Err: %s", err)
	}

	return g, nil
}

/*
 * GetGroupsByQuery returns a slice of group based on the SQL query provided.
 */
func GetGroupsByQuery(db *pgxpool.Pool, q string, args any) ([]*Group, error) {

	var rows pgx.Rows

	if args == nil {

		rows, _ = db.Query(context.Background(), q)
	} else {
		rows, _ = db.Query(context.Background(), q, args)
	}
	groups, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Group])
	if err != nil {
		return nil, err
	}

	for _, g := range groups {
		g.DB = db
	}

	return groups, nil
}

/*
 * GetGroupsByID returns a group with the provided group ID.
 */
func GetGroupByID(db *pgxpool.Pool, id uint64) (*Group, error) {

	q := `SELECT * FROM ` + DB_TABLE_GROUP + ` WHERE id = @id`
	args := pgx.NamedArgs{
		"id": id,
	}

	groups, err := GetGroupsByQuery(db, q, args)
	if err != nil {
		return nil, err
	}

	return groups[0], nil
}

/*
 * GetGroupsByLimit returns a slice of Group with a max count of 'limit'.
 */
func GetGroupsByLimit(db *pgxpool.Pool, limit uint64) ([]*Group, error) {

	q := `SELECT * FROM ` + DB_TABLE_GROUP + ` LIMIT ` + strconv.FormatUint(limit, 10)

	return GetGroupsByQuery(db, q, nil)
}

/*
 * GetGroupsByUser returns a slice of group containing groups owned by the user.
 */
func GetGroupsByUser(u *User) ([]*Group, error) {

	q := `SELECT * FROM ` + DB_TABLE_GROUP + ` WHERE user_id = @userID LIMIT 25`
	args := pgx.NamedArgs{
		"userID": u.ID,
	}

	return GetGroupsByQuery(u.DB, q, args)
}

package db

import (
	"context"
	"fmt"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const DB_TABLE_MEMBERSHIPS = "memberships"

type MemberRole string

const (
	MemberMember MemberRole = "member"
	MemberCohost MemberRole = "cohost"
	MemberHost   MemberRole = "host"
	MemberOwner  MemberRole = "owner"
)

/*
 * Membership represents a relationship between a Group and a User. It forms a
 * many-to-many relationship and also provdes the user's role in the Group.
 */
type Membership struct {
	framework.BaseModel
	GroupID uint64     `db:"group_id" validate:"required"`
	UserID  uint64     `db:"user_id" validate:"required"`
	TheUser *User      `db:"-"`
	Role    MemberRole `db:"role"`
}

/*
 * primaryKey returns the primary key name of the table
 */
func (ms *Membership) primaryKey() string { return "" }

/*
 * save serializes the struct to the database. The update is done via primary
 * key.
 */
func (ms *Membership) Save() error {

	q := `UPDATE ` + ms.table() + ` 
		SET role=@role,
			updated_time=@updatedTime
		WHERE group_id=@groupID AND user_id=@userID`
	_, err := ms.DB.Exec(context.Background(), q, pgx.NamedArgs{
		"role":        ms.Role,
		"updatedTime": ms.UpdatedTime,
		"groupID":     ms.GroupID,
		"userID":      ms.UserID,
	})

	return err
}

/*
 * table returns the table name used in the database.
 */
func (ms *Membership) table() string { return DB_TABLE_MEMBERSHIPS }

//==============================================================================
// End of methods, start of functions
//==============================================================================

/*
 * Internal init function.
 */
func initMembership(db *pgxpool.Pool) *Membership {

	ms := new(Membership)
	ms.DB = db

	return ms
}

/*
 * NewMembership creates a new Membership struct, validates it, and if good,
 * saves it to the database.
 */
func NewMembership(groupID uint64, u *User, role MemberRole) (*Membership, error) {

	ms := initMembership(u.DB)
	ms.GroupID = groupID
	ms.UserID = u.ID
	ms.Role = role

	err := validate.Struct(ms)
	if err != nil {
		return nil, err
	}

	q := `INSERT INTO ` + ms.table() + ` 
		(group_id, user_id, role) 
		VALUES (@groupID, @userID, @role) RETURNING *`
	rows, _ := u.DB.Query(context.Background(), q, pgx.NamedArgs{
		"groupID": ms.GroupID,
		"userID":  ms.UserID,
		"role":    ms.Role,
	})

	ms, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[Membership])
	if err != nil {
		return nil, fmt.Errorf("Failed to create Membership. Err: %s", err)
	}

	return ms, nil
}

/*
 * GetMembershipsByQuery returns a slice of Membership from the DB based on the
 * provided query.
 *
 * Note: this function can likely be reused for all DBs using an interface or
 * generic sometime in the future.
 */
func GetMembershipsByQuery(db *pgxpool.Pool, q string, args any) ([]*Membership, error) {

	var rows pgx.Rows

	if args == nil {

		rows, _ = db.Query(context.Background(), q)
	} else {
		rows, _ = db.Query(context.Background(), q, args)
	}
	memberships, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Membership])
	if err != nil {
		return nil, err
	}

	for _, ms := range memberships {

		ms.DB = db

		ms.TheUser, err = GetUserByID(db, ms.UserID)
		if err != nil {
			return nil, err
		}
	}

	return memberships, nil
}

/*
 * GetMembershipsByGroup
 *
 */
func GetMembershipsByGroup(db *pgxpool.Pool, groupID uint64) ([]*Membership, error) {

	q := `SELECT * FROM ` + DB_TABLE_MEMBERSHIPS + ` WHERE group_id=@groupID`
	args := pgx.NamedArgs{
		"groupID": groupID,
	}

	return GetMembershipsByQuery(db, q, args)
}

/*
 * GetMembershipsByUser
 *
 */
func GetMembershipsByUser(db *pgxpool.Pool, userID uint64) ([]*Membership, error) {

	q := `SELECT * FROM ` + DB_TABLE_MEMBERSHIPS + ` WHERE user_id=@userID`
	args := pgx.NamedArgs{
		"userID": userID,
	}

	return GetMembershipsByQuery(db, q, args)
}

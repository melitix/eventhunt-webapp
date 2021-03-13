package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/gopherlibs/gpic/gpic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// lastActive is a timestamp showing when the user last did something. This is
// currently handled by the initUser function running active()
type User struct {
	framework.BaseModel
	Username   string    `db:"username"`
	Password   string    `db:"password"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	LastActive time.Time `db:"last_active"`
}

/*
 * Update lastActive timestamp to use the user did something.
 */
func (u *User) active() {

	u.LastActive = time.Now()
	u.save()
}

/*
 *
 */
func (u *User) AvatarURL() string {

	avatar, err := gpic.NewAvatar(u.Email())
	if err != nil {
		slog.Error("Failed to create avatar struct for user.", "id", u.ID, "err", err)
		return ""
	}

	avatar.SetSize(100)
	picURL, err := avatar.URL()
	if err != nil {
		slog.Error("Failed to get avatar URL.", "userID", u.ID, "err", err)
		return ""
	}

	return picURL.String()
}

/*
 * Delete user from database.
 */
func (u *User) delete() error {

	_, err := u.DB.Exec(context.Background(), "DELETE FROM users WHERE id=$1", u.ID)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

/*
 * Return the preferred email address as a string.
 */
func (u *User) Email() string {

	email, err := GetPreferredEmailByUser(u)
	if err != nil {
		log.Error("Error: Failed to get preferred email for username: " + u.Username)
		return ""
	}

	return email.Value
}

/*
 * Save user to database.
 */
func (u *User) save() error {

	q := `UPDATE ` + u.table() + ` SET username = @username, first_name = @firstName, last_name = @lastName, last_active = @lastActive WHERE id=$5`
	_, err := u.DB.Exec(context.Background(), q, pgx.NamedArgs{
		"username":   u.Username,
		"firstName":  u.FirstName,
		"lastName":   u.LastName,
		"lastActive": u.LastActive,
		"id":         u.ID,
	})

	if err != nil {
		return fmt.Errorf("Failed to save user (id:%d) to db. Msg: %s", u.ID, err)
	}

	return nil
}

func (u *User) table() string { return "users" }

/*
 * Update the user's password.
 */
func (u *User) UpdatePassword(password, password2 string) error {

	if password != password2 {
		return errors.New("Error: Passwords don't match.")
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Warn("Password hashing failed.")
	}

	_, err = u.DB.Exec(context.Background(), "UPDATE users SET password = $1 WHERE id=$2",
		hashedPassword, u.ID)
	if err != nil {
		return errors.New("Error: Failed to save new password.")
	}

	return nil
}

//=============================================================================
// End of methods, start of functions
//=============================================================================

/*
 * Adds a user to the DB. This is similar to createUser but not the same. createUser should be used when user auth is handled by this app specifically. addUser should be used when auth is handled via SSO and we simply need some local user data.
 */
func AddUser(db *pgxpool.Pool, userID uint64, username, email string) (*User, error) {

	if userID == 0 {
		return nil, errors.New("User ID cannot be 0.")
	}

	if username == "" {
		return nil, errors.New("Username cannot be blank.")
	}

	if email == "" {
		return nil, errors.New("Email address cannot be blank.")
	}

	var u User

	u.DB = db
	u.ID = userID

	_, err := db.Exec(context.Background(), "INSERT INTO users (id, username, password, first_name, last_name) VALUES ($1, $2, 'n/a', '', '')",
		userID,
		username,
	)
	if err != nil {
		slog.Error("Failed to addUser.", "userID", userID, "username", username, "msg", err)
		return nil, err
	}

	_, err = AddEmailAddress(&u, email, true, true)
	if err != nil {
		return nil, errors.New("addEmailAddress failed. Message: " + err.Error())
	}

	return &u, nil
}

/*
 * Creates a new user.
 */
func CreateUser(db *pgxpool.Pool, username, password, email, firstName, lastName string) (*User, error) {

	var u User
	var lastInsertID uint64

	u.DB = db

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Warn("Password hashing failed.")
	}

	err = db.QueryRow(context.Background(), "INSERT INTO users (username, password, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id",
		username, string(hashedPassword), firstName, lastName).Scan(&lastInsertID)
	if err != nil {
		return nil, err
	}

	u.ID = lastInsertID

	return &u, nil
}

/*
 * This is the main function that retrieves users from the DB. Several helper
 * functions may exists to specific by what field/clause to retrieve the
 * user.
 */
func GetUserBy(db *pgxpool.Pool, clause string) (*User, error) {

	q := `SELECT * FROM users WHERE ` + clause
	rows, _ := db.Query(context.Background(), q)
	u, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[User])
	if err != nil {

		return nil, fmt.Errorf("Failed to get user from the DB. Msg: %s", err)
	}

	u.DB = db

	return u, nil
}

/*
 * Get a user by email.
 */
func GetUserByEmail(db *pgxpool.Pool, email string) (*User, error) {

	return GetUserBy(db, "email='"+email+"'")
}

/*
 * Get a website by its database ID.
 */
func GetUserByID(db *pgxpool.Pool, id uint64) (*User, error) {

	return GetUserBy(db, fmt.Sprintf("id = %d", id))
}

/*
 * Get a user by username.
 */
func GetUserByUsername(db *pgxpool.Pool, username string) (*User, error) {

	return GetUserBy(db, "username='"+username+"'")
}

func GetUsers(db *pgxpool.Pool, start int, count int) ([]*User, error) {

	q := `SELECT * FROM users WHERE LIMIT @limit OFFSET @offset`
	args := pgx.NamedArgs{
		"limit":  count,
		"offset": start,
	}

	return GetUsersByQuery(db, q, args)
}

/*
 * Get group of users.
 */
func GetUsersByQuery(db *pgxpool.Pool, q string, args ...any) ([]*User, error) {

	rows, _ := db.Query(context.Background(), q, args)
	users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[User])
	if err != nil {

		return nil, fmt.Errorf("Failed to get user list from the DB. Msg: %s", err)
	}

	for _, u := range users {
		u.DB = db
	}

	return users, nil
}

/*
 * Get list of users active in the last 'days'.
 */
func getActiveUsers(db *pgxpool.Pool, days int) ([]*User, error) {

	// hard DB ratelimit just in case
	if days < 1 || days > 90 {
		return nil, fmt.Errorf("Days count for active users is outside of range 0 < days < 91.")
	}

	q := `SELECT id FROM users WHERE last_active >= DATE(NOW() - INTERVAL ` + strconv.Itoa(days) + ` DAY)`

	return GetUsersByQuery(db, q)
}

/*
 * Get 'number' of newest users.
 */
func getNewestUsers(db *pgxpool.Pool, number int) ([]*User, error) {

	// hard DB ratelimit just in case
	if number < 1 || number > 100 {
		return nil, fmt.Errorf("Number count for newest users is outside of range 0 < number < 101.")
	}

	q := `SELECT id FROM users ORDER BY created_at DESC LIMIT ` + strconv.Itoa(number)

	return GetUsersByQuery(db, q)
}

/*
 * Verify username & password.
 *
 * Returns user ID if verified, 0 otherwise.
 */
func VerifyPassword(db *pgxpool.Pool, username, password string) uint64 {

	var userID uint64
	var hashedPassword string

	err := db.QueryRow(context.Background(), "SELECT id, password FROM users WHERE username = $1", username).Scan(
		&userID, &hashedPassword)
	if err != nil {

		if err == sql.ErrNoRows {
			// Username not found
			return 0
		} else {
			log.Fatal(err)
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// Password incorrect
		return 0
	}

	return userID
}

package storage

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/pbkdf2"
	
)

// PBKDF2 hash generation parameters.
// Salt size
var SALT_SIZE = 16
// Iterations
var ITERATIONS = 1024*1024

// Any upper, lower, digit characters, at least 6 characters.
var usernameValidator = regexp.MustCompile(`^[[:alnum:]]+$`)

// Any printable characters, at least 6 characters.
var passwordValidator = regexp.MustCompile(`^.{6,}$`)

// User represents a user to external code.
type User struct {
	ID           int64 `db:"id"`
	Username     string `db:"username"`
	IsAdmin	     bool `db:"is_admin"`
	SessionTTL   int32 `db:"session_ttl"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
var externalUserFields = "id, username, is_admin, session_ttl, created_at, updated_at" 

// UserUpdate represents the user fields that can be "updated" in the simple 
// case.  This excludes id, which identifies the record to be updated.
type UserUpdate struct {
	ID           int64 `db:"id"`
	Username     string `db:"username"`
	IsAdmin	     bool `db:"is_admin"`
	SessionTTL   int32 `db:"session_ttl"`
}

type UserApplicationAccess struct {
	Id int64			`db:"application_id"`
	Name string			`db:"application_name"`
	IsAdmin bool		`db:"is_admin"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// userInternal represents a user as known by the database.
type userInternal struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	IsAdmin	     bool `db:"is_admin"`
	IsActive     bool `db:"is_active"`
	SessionTTL   int32 `db:"session_ttl"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
var internalUserFields = "*"

var jwtsecret []byte

func SetUserSecret( s string ) {
	jwtsecret = []byte(s)
}

// Validate validates the data of the Application.
func ValidateUsername( username string ) error {
	if !usernameValidator.MatchString( username ) {
		return errors.New("user name may only be composed of upper and lower case characters and digits")
	}
	return nil
}

func ValidatePassword( password string ) error {
	if !passwordValidator.MatchString( password ) {
		return errors.New("passwords must be at least 6 characters long")
	}
	return nil
}

// CreateApplication creates the given Application.
func CreateUser(db *sqlx.DB, user *User, password string) ( int64, error ) {
	if err := ValidateUsername( user.Username ); err != nil {
		return 0, fmt.Errorf("validate username error: %s", err)
	}

	if err := ValidatePassword( password ); err != nil {
		return 0, fmt.Errorf("validate password error: %s", err)
	}
	
	now := time.Now()
	
	// Add the new user.
	rows, err := db.Queryx( "insert into \"user\" (username, password_hash, is_admin, is_active, session_ttl, created_at, updated_at) values( $1, $2, $3, 'true', $4, $5, $6 ) returning id", 
							user.Username, hash( password, SALT_SIZE, ITERATIONS ), user.IsAdmin, user.SessionTTL, now, now )
    if err != nil {
    	// Unexpected errorerinarian
    	return 0, err
    }
    rows.Next();
    rows.Scan( &user.ID )
    rows.Close()

	log.WithFields(log.Fields{
		"username":   user.Username,
		"session_ttl": user.SessionTTL,
		"is_admin": user.IsAdmin,
	}).Info("user created")
	return user.ID, nil
}

// Generate the hash of a password for storage in the database.
// NOTE: We store the details of the hashing algorithm with the hash itself, 
// making it easy to recreate the hash for password checking, even if we change
// the default criteria here. 
func hash( password string, saltSize int, iterations int) string {
	// Generate a random salt value, 128 bits.
	salt := make( []byte, saltSize )
	rand.Read( salt )
	
	return hashWithSalt( password, salt, iterations )
}

func hashWithSalt( password string, salt []byte, iterations int ) string {
	// Generate the hash.  This should be a little painful, adjust ITERATIONS
	// if it needs performance tweeking.  Greatly depends on the hardware.  
	// NOTE: We store these details with the returned hash, so changes will not
	// affect our ability to do password compares.
	hash := pbkdf2.Key( []byte( password ), salt, iterations, sha512.Size, sha512.New );
	
	// Build up the parameters and hash into a single string so we can compare 
	// other string to the same hash.  Note that the hash algorithm is hard-
	// coded here, as it is above.  Introducing alternate encodings must support
	// old encodings as well, and build this string appropriately.
	var buffer bytes.Buffer

	buffer.WriteString( "PBKDF2$" )
	buffer.WriteString( "sha512$" )
	buffer.WriteString( strconv.Itoa( iterations ) )
	buffer.WriteString( "$" ) 
	buffer.WriteString( base64.StdEncoding.EncodeToString( salt ) )
	buffer.WriteString( "$" )
	buffer.WriteString( base64.StdEncoding.EncodeToString( hash ) )
	
	return buffer.String()
}

// HashCompare verifies that passed password hashes to the same value as the 
// passed passwordHash.
func hashCompare( password string, passwordHash string ) bool {
	// SPlit the hash string into its parts.
	hashSplit := strings.Split( passwordHash, "$" )
	
	// Get the iterations and the salt and use them to encode the password
	// being compared.cre
	iterations, _ := strconv.Atoi( hashSplit[ 2 ] )
	salt, _ := base64.StdEncoding.DecodeString( hashSplit[ 3 ] )
	newHash := hashWithSalt( password, salt, iterations )
	return newHash == passwordHash
}

// GetUser returns the User for the given id.
func GetUser(db *sqlx.DB, id int64) (User, error) {
	var user User
	err := db.Get( &user, "select " + externalUserFields + " from \"user\" where id = $1", id)
	if err != nil {
		return user, fmt.Errorf("get user error: %s", err)
	}
	
	return user, nil
}

// GetUSerByUsername returns the User for the given username.
func GetUserByUsername(db *sqlx.DB, username string) (User, error) {
	var user User
	err := db.Get( &user, "select " + externalUserFields + " from \"user\" where username = $1", username)
	if err != nil {
		return user, fmt.Errorf("get user error: %s", err)
	}
	
	return user, nil
}

// GetUserCount returns the total number of users.
func GetUserCount(db *sqlx.DB) (int32, error) {
	var count int32
	err := db.Get(&count, "select count(*) from \"user\"")
	if err != nil {
		return 0, fmt.Errorf("get user count error: %s", err)
	}
	return count, nil
}

// GetUsers returns a slice of users, respecting the given limit and offset.
func GetUsers(db *sqlx.DB, limit, offset int32) ([]User, error) {
	var users []User
	err := db.Select(&users, "select " + externalUserFields + " from \"user\" order by username limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get users error: %s", err)
	}
	return users, nil
}

// UpdateUser updates the given User.
func UpdateUser(db *sqlx.DB, item UserUpdate) error {
	if err := ValidateUsername( item.Username ); err != nil {
		return fmt.Errorf("validate username error: %s", err)
	}

	res, err := db.Exec(`
		update "user"
		set
			username = $2,
			is_admin = $3,
			session_ttl = $4,
			updated_at = now()
		where id = $1`,
		item.ID,
		item.Username,
		item.IsAdmin,
		item.SessionTTL,
	)
	if err != nil {
		return fmt.Errorf("update user error: %s", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("user %d does not exist", item.ID)
	}
	log.WithFields(log.Fields{
		"id":   item.ID,
		"username": item.Username,
		"is_admin": item.IsAdmin,
		"session_ttl": item.SessionTTL,
	}).Info("user updated")

	return nil
}

// DeleteUSer deletes the User record matching the given ID.
func DeleteUser(db *sqlx.DB, id int64) error {
	res, err := db.Exec("delete from \"user\" where id = $1", id)
	if err != nil {
		return fmt.Errorf("delete user error: %s", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("user with id %d does not exist", id)
	}
	log.WithFields(log.Fields{
		"id": id,
	}).Info("user deleted")

	return nil
}


// Login the user.
func LoginUser(db *sqlx.DB, username string, password string ) ( string, error ) {
	// Find the user by username
	var user userInternal
	err := db.Get( &user, "select " + internalUserFields + " from \"user\" where username = $1", username)
	if err != nil {
		return "", fmt.Errorf("Invalid username or password (unknown username)")
	}
	
	// Compare the passed in password with the hash in the database.
	if !hashCompare( password, user.PasswordHash ) {
		return "", fmt.Errorf("Invalid username or password")
	}
	
	// Generate the token.
	now := time.Now()
	nowSecondsSinceEpoch := now.Unix()
	expSecondsSinceEpoch := nowSecondsSinceEpoch + ( 60 * int64(user.SessionTTL) )
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			    "iss": "lora-app-server",
			    "aud": "lora-app-server",
			    "nbf": nowSecondsSinceEpoch,
			    "exp": expSecondsSinceEpoch,
			    "sub": "user",
			    "username": user.Username,
			})
	
	jwt, err := token.SignedString( jwtsecret )
	if nil != err {
		return jwt, fmt.Errorf( "Failed to generate jwt: %s", err )
	}
	return jwt, err
}

// Update password.
func UpdatePassword( db *sqlx.DB, id int64, newpassword string ) error {
	if err := ValidatePassword( newpassword ); err != nil {
		return fmt.Errorf("validate password error: %s", err)
	}
	
	// Add the new user.
	rows, err := db.Queryx( "update \"user\" set password_hash = $1, updated_at = now() where id = $2", 
							hash( newpassword, SALT_SIZE, ITERATIONS ), id )
    if err != nil {
    	// Unexpected error
    	return err
    }
    rows.Close()

	log.WithFields(log.Fields{
		"id":   id,
	}).Info("user password updated")
	return nil

}

// Gets the User Profile, which is the set of applications the user can interact
// with, including permissions.
func GetProfile( db *sqlx.DB, id int64 ) ([]UserApplicationAccess, error) {
	// Get the user applications.
	var userProfile []UserApplicationAccess
	err := db.Select( &userProfile,
					  `select au.application_id as application_id,
                              a.name as application_name,
                              au.is_admin as is_admin,
                              au.created_at as created_at,
                              au.updated_at as updated_at
                       from application_user au, application a
                       where au.user_id = $1 and au.application_id = a.id`,
					  id )
   return userProfile, err
}
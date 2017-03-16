package storage

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

// saltSize defines the salt size
const saltSize = 16

// HashIterations defines the number of hash iterations.
var HashIterations = 100000

// defaultSessionTTL defines the default session TTL
const defaultSessionTTL = time.Hour * 24

// Any upper, lower, digit characters, at least 6 characters.
var usernameValidator = regexp.MustCompile(`^[[:alnum:]]+$`)

// Any printable characters, at least 6 characters.
var passwordValidator = regexp.MustCompile(`^.{6,}$`)

// User represents a user to external code.
type User struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	IsAdmin      bool      `db:"is_admin"`
	IsActive     bool      `db:"is_active"`
	SessionTTL   int32     `db:"session_ttl"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	PasswordHash string    `db:"password_hash"`
}

const externalUserFields = "id, username, is_admin, is_active, session_ttl, created_at, updated_at"
const internalUserFields = "*"

// UserUpdate represents the user fields that can be "updated" in the simple
// case.  This excludes id, which identifies the record to be updated.
type UserUpdate struct {
	ID         int64  `db:"id"`
	Username   string `db:"username"`
	IsAdmin    bool   `db:"is_admin"`
	IsActive   bool   `db:"is_active"`
	SessionTTL int32  `db:"session_ttl"`
}

type UserApplicationAccess struct {
	ID        int64     `db:"application_id"`
	Name      string    `db:"application_name"`
	IsAdmin   bool      `db:"is_admin"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// userInternal represents a user as known by the database.
type userInternal struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	IsAdmin      bool      `db:"is_admin"`
	IsActive     bool      `db:"is_active"`
	SessionTTL   int32     `db:"session_ttl"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

var jwtsecret []byte

func SetUserSecret(s string) {
	jwtsecret = []byte(s)
}

// Validate validates the data of the Application.
func ValidateUsername(username string) error {
	if !usernameValidator.MatchString(username) {
		return ErrUserInvalidUsername
	}
	return nil
}

func ValidatePassword(password string) error {
	if !passwordValidator.MatchString(password) {
		return ErrUserPasswordLength
	}
	return nil
}

// CreateApplication creates the given Application.
func CreateUser(db *sqlx.DB, user *User, password string) (int64, error) {
	if err := ValidateUsername(user.Username); err != nil {
		return 0, errors.Wrap(err, "validation error")
	}

	if err := ValidatePassword(password); err != nil {
		return 0, errors.Wrap(err, "validation error")
	}

	pwHash, err := hash(password, saltSize, HashIterations)
	if err != nil {
		return 0, err
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Add the new user.
	err = db.Get(&user.ID, `
		insert into "user" (
			username,
			password_hash,
			is_admin,
			is_active,
			session_ttl,
			created_at,
			updated_at)
		values (
			$1, $2, $3, $4, $5, $6, $7) returning id`,
		user.Username,
		pwHash,
		user.IsAdmin,
		user.IsActive,
		user.SessionTTL,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return 0, ErrAlreadyExists
			default:
				return 0, errors.Wrap(err, "insert error")
			}
		default:
			return 0, errors.Wrap(err, "insert error")
		}
	}

	log.WithFields(log.Fields{
		"username":    user.Username,
		"session_ttl": user.SessionTTL,
		"is_admin":    user.IsAdmin,
	}).Info("user created")
	return user.ID, nil
}

// Generate the hash of a password for storage in the database.
// NOTE: We store the details of the hashing algorithm with the hash itself,
// making it easy to recreate the hash for password checking, even if we change
// the default criteria here.
func hash(password string, saltSize int, iterations int) (string, error) {
	// Generate a random salt value, 128 bits.
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return "", errors.Wrap(err, "read random bytes error")
	}

	return hashWithSalt(password, salt, iterations), nil
}

func hashWithSalt(password string, salt []byte, iterations int) string {
	// Generate the hash.  This should be a little painful, adjust ITERATIONS
	// if it needs performance tweeking.  Greatly depends on the hardware.
	// NOTE: We store these details with the returned hash, so changes will not
	// affect our ability to do password compares.
	hash := pbkdf2.Key([]byte(password), salt, iterations, sha512.Size, sha512.New)

	// Build up the parameters and hash into a single string so we can compare
	// other string to the same hash.  Note that the hash algorithm is hard-
	// coded here, as it is above.  Introducing alternate encodings must support
	// old encodings as well, and build this string appropriately.
	var buffer bytes.Buffer

	buffer.WriteString("PBKDF2$")
	buffer.WriteString("sha512$")
	buffer.WriteString(strconv.Itoa(iterations))
	buffer.WriteString("$")
	buffer.WriteString(base64.StdEncoding.EncodeToString(salt))
	buffer.WriteString("$")
	buffer.WriteString(base64.StdEncoding.EncodeToString(hash))

	return buffer.String()
}

// HashCompare verifies that passed password hashes to the same value as the
// passed passwordHash.
func hashCompare(password string, passwordHash string) bool {
	// SPlit the hash string into its parts.
	hashSplit := strings.Split(passwordHash, "$")

	// Get the iterations and the salt and use them to encode the password
	// being compared.cre
	iterations, _ := strconv.Atoi(hashSplit[2])
	salt, _ := base64.StdEncoding.DecodeString(hashSplit[3])
	newHash := hashWithSalt(password, salt, iterations)
	return newHash == passwordHash
}

// GetUser returns the User for the given id.
func GetUser(db *sqlx.DB, id int64) (User, error) {
	var user User
	err := db.Get(&user, "select "+externalUserFields+" from \"user\" where id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrDoesNotExist
		}
		return user, errors.Wrap(err, "select error")
	}

	return user, nil
}

// GetUSerByUsername returns the User for the given username.
func GetUserByUsername(db *sqlx.DB, username string) (User, error) {
	var user User
	err := db.Get(&user, "select "+externalUserFields+" from \"user\" where username = $1", username)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrDoesNotExist
		}
		return user, errors.Wrap(err, "select error")
	}

	return user, nil
}

// GetUserCount returns the total number of users.
func GetUserCount(db *sqlx.DB, search string) (int32, error) {
	var count int32
	if search != "" {
		search = "%" + search + "%"
	}
	err := db.Get(&count, `
		select
			count(*)
		from "user"
		where
			($1 != '' and username like $1)
			or ($1 = '')
		`, search)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetUsers returns a slice of users, respecting the given limit and offset.
func GetUsers(db *sqlx.DB, limit, offset int32, search string) ([]User, error) {
	var users []User
	if search != "" {
		search = search + "%"
	}
	err := db.Select(&users, "select "+externalUserFields+` from "user" where ($3 != '' and username like $3) or ($3 = '') order by username limit $1 offset $2`, limit, offset, search)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return users, nil
}

// UpdateUser updates the given User.
func UpdateUser(db *sqlx.DB, item UserUpdate) error {
	if err := ValidateUsername(item.Username); err != nil {
		return fmt.Errorf("validate username error: %s", err)
	}

	res, err := db.Exec(`
		update "user"
		set
			username = $2,
			is_admin = $3,
			is_active = $4,
			session_ttl = $5,
			updated_at = now()
		where id = $1`,
		item.ID,
		item.Username,
		item.IsAdmin,
		item.IsActive,
		item.SessionTTL,
	)
	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return ErrAlreadyExists
			default:
				return errors.Wrap(err, "update error")
			}
		default:
			return errors.Wrap(err, "update error")
		}
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"id":          item.ID,
		"username":    item.Username,
		"is_admin":    item.IsAdmin,
		"session_ttl": item.SessionTTL,
	}).Info("user updated")

	return nil
}

// DeleteUSer deletes the User record matching the given ID.
func DeleteUser(db *sqlx.DB, id int64) error {
	res, err := db.Exec("delete from \"user\" where id = $1", id)
	if err != nil {
		return errors.Wrap(err, "delete error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"id": id,
	}).Info("user deleted")
	return nil
}

// Login the user.
func LoginUser(db *sqlx.DB, username string, password string) (string, error) {
	// Find the user by username
	var user userInternal
	err := db.Get(&user, "select "+internalUserFields+" from \"user\" where username = $1", username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidUsernameOrPassword
		}
		return "", errors.Wrap(err, "select error")
	}

	// Compare the passed in password with the hash in the database.
	if !hashCompare(password, user.PasswordHash) {
		return "", ErrInvalidUsernameOrPassword
	}

	// Generate the token.
	now := time.Now()
	nowSecondsSinceEpoch := now.Unix()
	var expSecondsSinceEpoch int64
	if user.SessionTTL > 0 {
		expSecondsSinceEpoch = nowSecondsSinceEpoch + (60 * int64(user.SessionTTL))
	} else {
		expSecondsSinceEpoch = nowSecondsSinceEpoch + int64(defaultSessionTTL/time.Second)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "lora-app-server",
		"aud":      "lora-app-server",
		"nbf":      nowSecondsSinceEpoch,
		"exp":      expSecondsSinceEpoch,
		"sub":      "user",
		"username": user.Username,
	})

	jwt, err := token.SignedString(jwtsecret)
	if nil != err {
		return jwt, errors.Wrap(err, "get jwt signed string error")
	}
	return jwt, err
}

// Update password.
func UpdatePassword(db *sqlx.DB, id int64, newpassword string) error {
	if err := ValidatePassword(newpassword); err != nil {
		return errors.Wrap(err, "validation error")
	}

	pwHash, err := hash(newpassword, saltSize, HashIterations)
	if err != nil {
		return err
	}

	// Add the new user.
	_, err = db.Exec("update \"user\" set password_hash = $1, updated_at = now() where id = $2",
		pwHash, id)
	if err != nil {
		return errors.Wrap(err, "update error")
	}

	log.WithFields(log.Fields{
		"id": id,
	}).Info("user password updated")
	return nil

}

// Gets the User Profile, which is the set of applications the user can interact
// with, including permissions.
func GetProfile(db *sqlx.DB, id int64) ([]UserApplicationAccess, error) {
	// Get the user applications.
	var userProfile []UserApplicationAccess
	err := db.Select(&userProfile,
		`select au.application_id as application_id,
                              a.name as application_name,
                              au.is_admin as is_admin,
                              au.created_at as created_at,
                              au.updated_at as updated_at
                       from application_user au, application a
                       where au.user_id = $1 and au.application_id = a.id`,
		id)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return userProfile, nil
}

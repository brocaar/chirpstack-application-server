package storage

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
)

// saltSize defines the salt size
const saltSize = 16

// defaultSessionTTL defines the default session TTL
const defaultSessionTTL = time.Hour * 24

// Any upper, lower, digit characters, at least 6 characters.
var usernameValidator = regexp.MustCompile(`^[[:alnum:]]+$`)

// Any printable characters, at least 6 characters.
var passwordValidator = regexp.MustCompile(`^.{6,}$`)

// Must contain @ (this is far from perfect)
var emailValidator = regexp.MustCompile(`.+@.+`)

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
	Email        string    `db:"email"`
	Note         string    `db:"note"`
}

const externalUserFields = "id, username, is_admin, is_active, session_ttl, created_at, updated_at, email, note"
const internalUserFields = "*"

// UserUpdate represents the user fields that can be "updated" in the simple
// case.  This excludes id, which identifies the record to be updated.
type UserUpdate struct {
	ID         int64  `db:"id"`
	Username   string `db:"username"`
	IsAdmin    bool   `db:"is_admin"`
	IsActive   bool   `db:"is_active"`
	SessionTTL int32  `db:"session_ttl"`
	Email      string `db:"email"`
	Note       string `db:"note"`
}

// UserProfile contains the profile of the user.
type UserProfile struct {
	User          UserProfileUser
	Organizations []UserProfileOrganization
}

// UserProfileUser contains the user information of the profile.
type UserProfileUser struct {
	ID         int64     `db:"id"`
	Username   string    `db:"username"`
	IsAdmin    bool      `db:"is_admin"`
	IsActive   bool      `db:"is_active"`
	SessionTTL int32     `db:"session_ttl"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// UserProfileApplication contains the applications to which the user
// is linked.
type UserProfileApplication struct {
	ID        int64     `db:"application_id"`
	Name      string    `db:"application_name"`
	IsAdmin   bool      `db:"is_admin"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// UserProfileOrganization contains the organizations to which the user
// is linked.
type UserProfileOrganization struct {
	ID        int64     `db:"organization_id"`
	Name      string    `db:"organization_name"`
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
	Email        string    `db:"email"`
	Note         string    `db:"note"`
}

// ValidateUsername validates the given username.
func ValidateUsername(username string) error {
	if !usernameValidator.MatchString(username) {
		return ErrUserInvalidUsername
	}
	return nil
}

// ValidatePassword validates the given password.
func ValidatePassword(password string) error {
	if !passwordValidator.MatchString(password) {
		return ErrUserPasswordLength
	}
	return nil
}

// ValidateEmail validates the given e-mail.
func ValidateEmail(email string) error {
	if !emailValidator.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// CreateUser creates the given user.
func CreateUser(db sqlx.Queryer, user *User, password string) (int64, error) {
	if err := ValidateUsername(user.Username); err != nil {
		return 0, errors.Wrap(err, "validation error")
	}

	if err := ValidatePassword(password); err != nil {
		return 0, errors.Wrap(err, "validation error")
	}

	if err := ValidateEmail(user.Email); err != nil {
		return 0, errors.Wrap(err, "validation error")
	}

	pwHash, err := hash(password, saltSize, HashIterations)
	if err != nil {
		return 0, err
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Add the new user.
	err = sqlx.Get(db, &user.ID, `
		insert into "user" (
			username,
			password_hash,
			is_admin,
			is_active,
			session_ttl,
			created_at,
			updated_at,
			email,
			note
		)
		values (
			$1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`,
		user.Username,
		pwHash,
		user.IsAdmin,
		user.IsActive,
		user.SessionTTL,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
		user.Note,
	)
	if err != nil {
		return 0, handlePSQLError(Insert, err, "insert error")
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
func GetUser(db sqlx.Queryer, id int64) (User, error) {
	var user User
	err := sqlx.Get(db, &user, "select "+externalUserFields+" from \"user\" where id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrDoesNotExist
		}
		return user, errors.Wrap(err, "select error")
	}

	return user, nil
}

// GetUserByUsername returns the User for the given username.
func GetUserByUsername(db sqlx.Queryer, username string) (User, error) {
	var user User
	err := sqlx.Get(db, &user, "select "+externalUserFields+" from \"user\" where username = $1", username)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrDoesNotExist
		}
		return user, errors.Wrap(err, "select error")
	}

	return user, nil
}

// GetUserCount returns the total number of users.
func GetUserCount(db sqlx.Queryer, search string) (int32, error) {
	var count int32
	if search != "" {
		search = "%" + search + "%"
	}
	err := sqlx.Get(db, &count, `
		select
			count(*)
		from "user"
		where
			($1 != '' and username ilike $1)
			or ($1 = '')
		`, search)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetUsers returns a slice of users, respecting the given limit and offset.
func GetUsers(db sqlx.Queryer, limit, offset int, search string) ([]User, error) {
	var users []User
	if search != "" {
		search = "%" + search + "%"
	}
	err := sqlx.Select(db, &users, "select "+externalUserFields+` from "user" where ($3 != '' and username ilike $3) or ($3 = '') order by username limit $1 offset $2`, limit, offset, search)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return users, nil
}

// UpdateUser updates the given User.
func UpdateUser(db sqlx.Execer, item UserUpdate) error {
	if err := ValidateUsername(item.Username); err != nil {
		return errors.Wrap(err, "validation error")
	}

	if err := ValidateEmail(item.Email); err != nil {
		return errors.Wrap(err, "validation error")
	}

	res, err := db.Exec(`
		update "user"
		set
			username = $2,
			is_admin = $3,
			is_active = $4,
			session_ttl = $5,
			updated_at = now(),
			email = $6,
			note = $7
		where id = $1`,
		item.ID,
		item.Username,
		item.IsAdmin,
		item.IsActive,
		item.SessionTTL,
		item.Email,
		item.Note,
	)
	if err != nil {
		return handlePSQLError(Update, err, "update error")
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

// DeleteUser deletes the User record matching the given ID.
func DeleteUser(db sqlx.Execer, id int64) error {
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

// LoginUser returns a JWT token for the user matching the given username
// and password.
func LoginUser(db sqlx.Queryer, username string, password string) (string, error) {
	// Find the user by username
	var user userInternal
	err := sqlx.Get(db, &user, "select "+internalUserFields+" from \"user\" where username = $1", username)
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

// UpdatePassword updates the user with the new password.
func UpdatePassword(db sqlx.Execer, id int64, newpassword string) error {
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

// GetProfile returns the user profile (user, applications and organizations
// to which the user is linked).
func GetProfile(db sqlx.Queryer, id int64) (UserProfile, error) {
	var prof UserProfile

	user, err := GetUser(db, id)
	if err != nil {
		return prof, errors.Wrap(err, "get user error")
	}
	prof.User = UserProfileUser{
		ID:         user.ID,
		Username:   user.Username,
		SessionTTL: user.SessionTTL,
		IsAdmin:    user.IsAdmin,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

	err = sqlx.Select(db, &prof.Organizations, `
		select
			ou.organization_id as organization_id,
			o.name as organization_name,
			ou.is_admin as is_admin,
			ou.created_at as created_at,
			ou.updated_at as updated_at
		from
			organization_user ou,
			organization o
		where
			ou.user_id = $1
			and ou.organization_id = o.id`,
		id,
	)
	if err != nil {
		return prof, errors.Wrap(err, "select error")
	}

	return prof, nil
}

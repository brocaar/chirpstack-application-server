package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"

	"github.com/brocaar/chirpstack-application-server/internal/logging"
)

// saltSize defines the salt size
const saltSize = 16

// defaultSessionTTL defines the default session TTL
const defaultSessionTTL = time.Hour * 24

// Any printable characters, at least 6 characters.
var passwordValidator = regexp.MustCompile(`^.{6,}$`)

// Email validation regexp taken from:
// https://html.spec.whatwg.org/multipage/input.html#e-mail-state-(type%3Demail)
var emailValidator = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// User defines the user structure.
type User struct {
	ID            int64     `db:"id"`
	IsAdmin       bool      `db:"is_admin"`
	IsActive      bool      `db:"is_active"`
	SessionTTL    int32     `db:"session_ttl"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
	PasswordHash  string    `db:"password_hash"`
	Email         string    `db:"email"`
	EmailVerified bool      `db:"email_verified"`
	EmailOld      string    `db:"email_old"`
	Note          string    `db:"note"`
	ExternalID    *string   `db:"external_id"` // must be pointer for unique index
}

// Validate validates the user data.
func (u User) Validate() error {
	if u.Email != "admin" && !emailValidator.MatchString(u.Email) {
		return ErrInvalidEmail
	}

	return nil
}

// SetPasswordHash hashes the given password and sets it.
func (u *User) SetPasswordHash(pw string) error {
	if !passwordValidator.MatchString(pw) {
		return ErrUserPasswordLength
	}

	pwHash, err := hash(pw, saltSize, HashIterations)
	if err != nil {
		return err
	}

	u.PasswordHash = pwHash

	return nil
}

// UserProfile contains the profile of the user.
type UserProfile struct {
	User          UserProfileUser
	Organizations []UserProfileOrganization
}

// UserProfileUser contains the user information of the profile.
type UserProfileUser struct {
	ID         int64     `db:"id"`
	Email      string    `db:"email"`
	IsAdmin    bool      `db:"is_admin"`
	IsActive   bool      `db:"is_active"`
	SessionTTL int32     `db:"session_ttl"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// UserProfileOrganization contains the organizations to which the user
// is linked.
type UserProfileOrganization struct {
	ID             int64     `db:"organization_id"`
	Name           string    `db:"organization_name"`
	IsAdmin        bool      `db:"is_admin"`
	IsDeviceAdmin  bool      `db:"is_device_admin"`
	IsGatewayAdmin bool      `db:"is_gateway_admin"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// CreateUser creates the given user.
func CreateUser(ctx context.Context, db sqlx.Queryer, user *User) error {
	if err := user.Validate(); err != nil {
		return errors.Wrap(err, "validation error")
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err := sqlx.Get(db, &user.ID, `
		insert into "user" (
			is_admin,
			is_active,
			session_ttl,
			created_at,
			updated_at,
			password_hash,
			email,
			email_verified,
			note,
			external_id
		)
		values (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		returning
			id`,
		user.IsAdmin,
		user.IsActive,
		user.SessionTTL,
		user.CreatedAt,
		user.UpdatedAt,
		user.PasswordHash,
		user.Email,
		user.EmailVerified,
		user.Note,
		user.ExternalID,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	var externalID string
	if user.ExternalID != nil {
		externalID = *user.ExternalID
	}

	log.WithFields(log.Fields{
		"id":          user.ID,
		"external_id": externalID,
		"email":       user.Email,
		"ctx_id":      ctx.Value(logging.ContextIDKey),
	}).Info("storage: user created")

	return nil
}

// GetUser returns the User for the given id.
func GetUser(ctx context.Context, db sqlx.Queryer, id int64) (User, error) {
	var user User

	err := sqlx.Get(db, &user, `
		select
			*
		from
			"user"
		where
			id = $1
	`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrDoesNotExist
		}
		return user, errors.Wrap(err, "select error")
	}

	return user, nil
}

// GetUserByExternalID returns the User for the given ext. ID.
func GetUserByExternalID(ctx context.Context, db sqlx.Queryer, externalID string) (User, error) {
	var user User

	err := sqlx.Get(db, &user, `
		select
			*
		from
			"user"
		where
			external_id = $1
	`, externalID)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrDoesNotExist
		}
		return user, errors.Wrap(err, "select error")
	}

	return user, nil
}

// GetUserByEmail returns the User for the given email.
func GetUserByEmail(ctx context.Context, db sqlx.Queryer, email string) (User, error) {
	var user User

	err := sqlx.Get(db, &user, `
		select
			*
		from
			"user"
		where
			email = $1
	`, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrDoesNotExist
		}
		return user, errors.Wrap(err, "select error")
	}

	return user, nil
}

// GetUserCount returns the total number of users.
func GetUserCount(ctx context.Context, db sqlx.Queryer) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select
			count(*)
		from "user"
	`)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetUsers returns a slice of users, respecting the given limit and offset.
func GetUsers(ctx context.Context, db sqlx.Queryer, limit, offset int) ([]User, error) {
	var users []User

	err := sqlx.Select(db, &users, `
		select
			*
		from
			"user"
		order by
			email
		limit $1
		offset $2
	`, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return users, nil
}

// UpdateUser updates the given User.
func UpdateUser(ctx context.Context, db sqlx.Execer, u *User) error {
	if err := u.Validate(); err != nil {
		return errors.Wrap(err, "validate user error")
	}

	u.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update "user"
		set
			updated_at = $2,
			is_admin = $3,
			is_active = $4,
			session_ttl = $5,
			email = $6,
			email_verified = $7,
			note = $8,
			external_id = $9,
			password_hash = $10
		where
			id = $1`,
		u.ID,
		u.UpdatedAt,
		u.IsAdmin,
		u.IsActive,
		u.SessionTTL,
		u.Email,
		u.EmailVerified,
		u.Note,
		u.ExternalID,
		u.PasswordHash,
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

	var extUser string
	if u.ExternalID != nil {
		extUser = *u.ExternalID
	}

	log.WithFields(log.Fields{
		"id":          u.ID,
		"external_id": extUser,
		"ctx_id":      ctx.Value(logging.ContextIDKey),
	}).Info("storage: user updated")

	return nil
}

// DeleteUser deletes the User record matching the given ID.
func DeleteUser(ctx context.Context, db sqlx.Execer, id int64) error {
	res, err := db.Exec(`
		delete from
			"user"
		where
			id = $1
	`, id)
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
		"id":     id,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("storage: user deleted")
	return nil
}

// LoginUserByPassword returns a JWT token for the user matching the given email
// and password combination.
func LoginUserByPassword(ctx context.Context, db sqlx.Queryer, email string, password string) (string, error) {
	// get the user by email
	var user User
	err := sqlx.Get(db, &user, `
		select
			*
		from
			"user"
		where
			email = $1
	`, email)
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

	return GetUserToken(user)
}

// GetProfile returns the user profile (user, applications and organizations
// to which the user is linked).
func GetProfile(ctx context.Context, db sqlx.Queryer, id int64) (UserProfile, error) {
	var prof UserProfile

	user, err := GetUser(ctx, db, id)
	if err != nil {
		return prof, errors.Wrap(err, "get user error")
	}
	prof.User = UserProfileUser{
		ID:         user.ID,
		Email:      user.Email,
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
			ou.is_device_admin as is_device_admin,
			ou.is_gateway_admin as is_gateway_admin,
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

// hashCompare verifies that passed password hashes to the same value as the
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

// GetUserToken returns a JWT token for the given user.
func GetUserToken(u User) (string, error) {
	// Generate the token.
	now := time.Now()
	nowSecondsSinceEpoch := now.Unix()
	var expSecondsSinceEpoch int64
	if u.SessionTTL > 0 {
		expSecondsSinceEpoch = nowSecondsSinceEpoch + (60 * int64(u.SessionTTL))
	} else {
		expSecondsSinceEpoch = nowSecondsSinceEpoch + int64(defaultSessionTTL/time.Second)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "as",
		"aud":      "as",
		"nbf":      nowSecondsSinceEpoch,
		"exp":      expSecondsSinceEpoch,
		"sub":      "user",
		"id":       u.ID,
		"username": u.Email, // backwards compatibility
	})

	jwt, err := token.SignedString(jwtsecret)
	if err != nil {
		return jwt, errors.Wrap(err, "get jwt signed string error")
	}
	return jwt, err
}

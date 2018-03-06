package auth

import (
	"fmt"
	"regexp"

	"github.com/gusseleet/lora-app-server/internal/storage"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

var validAuthorizationRegexp = regexp.MustCompile(`(?i)^bearer (.*)$`)

// Claims defines the struct containing the token claims.
type Claims struct {
	jwt.StandardClaims

	// Username defines the identity of the user.
	Username string `json:"username"`
}

// Validator defines the interface a validator needs to implement.
type Validator interface {
	// Validate validates the given set of validators against the given context.
	// Must return after the first validator function either returns true or
	// and error. The way how the validation must be seens is:
	//   if validatorFunc1 || validatorFunc2 || validatorFunc3 ...
	// In case multiple validators must validate to true, then a validator
	// func needs to be implemented which validates a given set of funcs as:
	//   if validatorFunc1 && validatorFunc2 && ValidatorFunc3 ...
	Validate(context.Context, ...ValidatorFunc) error

	// GetUsername returns the name of the authenticated user.
	GetUsername(context.Context) (string, error)

	// GetIsAdmin returns if the authenticated user is a global admin.
	GetIsAdmin(context.Context) (bool, error)
}

// ValidatorFunc defines the signature of a claim validator function.
// It returns a bool indicating if the validation passed or failed and an
// error in case an error occured (e.g. db connectivity).
type ValidatorFunc func(sqlx.Queryer, *Claims) (bool, error)

// JWTValidator validates JWT tokens.
type JWTValidator struct {
	db        sqlx.Ext
	secret    string
	algorithm string
}

// NewJWTValidator creates a new JWTValidator.
func NewJWTValidator(db sqlx.Ext, algorithm, secret string) *JWTValidator {
	return &JWTValidator{
		db:        db,
		secret:    secret,
		algorithm: algorithm,
	}
}

// Validate validates the token from the given context against the given
// validator funcs.
func (v JWTValidator) Validate(ctx context.Context, funcs ...ValidatorFunc) error {
	claims, err := v.getClaims(ctx)
	if err != nil {
		return err
	}

	for _, f := range funcs {
		ok, err := f(v.db, claims)
		if err != nil {
			return errors.Wrap(err, "validator func error")
		}
		if ok {
			return nil
		}
	}

	return ErrNotAuthorized
}

// GetUsername returns the username of the authenticated user.
func (v JWTValidator) GetUsername(ctx context.Context) (string, error) {
	claims, err := v.getClaims(ctx)
	if err != nil {
		return "", err
	}

	return claims.Username, nil
}

// GetIsAdmin returns if the authenticated user is a global amin.
func (v JWTValidator) GetIsAdmin(ctx context.Context) (bool, error) {
	claims, err := v.getClaims(ctx)
	if err != nil {
		return false, err
	}

	user, err := storage.GetUserByUsername(v.db, claims.Username)
	if err != nil {
		return false, errors.Wrap(err, "get user by username error")
	}

	return user.IsAdmin, nil
}

func (v JWTValidator) getClaims(ctx context.Context) (*Claims, error) {
	tokenStr, err := getTokenFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get token from context error")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Header["alg"] != v.algorithm {
			return nil, ErrInvalidAlgorithm
		}
		return []byte(v.secret), nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "jwt parse error")
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		// no need to use a static error, this should never happen
		return nil, fmt.Errorf("api/auth: expected *Claims, got %T", token.Claims)
	}

	return claims, nil
}

func getTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrNoMetadataInContext
	}

	token, ok := md["authorization"]
	if !ok || len(token) == 0 {
		return "", ErrNoAuthorizationInMetadata
	}

	match := validAuthorizationRegexp.FindStringSubmatch(token[0])

	// authorization header should respect RFC1945
	if len(match) == 0 {
		log.Warning("Deprecated Authorization header : RFC1945 format expected : Authorization: <type> <credentials>")
		return token[0], nil
	}

	return match[1], nil
}

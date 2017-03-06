package auth

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

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
}

// ValidatorFunc defines the signature of a claim validator function.
// It returns a bool indicating if the validation passed or failed and an
// error in case an error occured (e.g. db connectivity).
type ValidatorFunc func(*sqlx.DB, *Claims) (bool, error)

// NopValidator doesn't perform any validation and returns alway true.
type NopValidator struct{}

// Validate validates the given token against the given validator funcs.
// In the case of the NopValidator, it returns always nil.
func (v NopValidator) Validate(db *sqlx.DB, ctx context.Context, funcs ...ValidatorFunc) error {
	return nil
}

// JWTValidator validates JWT tokens.
type JWTValidator struct {
	db        *sqlx.DB
	secret    string
	algorithm string
}

// NewJWTValidator creates a new JWTValidator.
func NewJWTValidator(db *sqlx.DB, algorithm, secret string) *JWTValidator {
	return &JWTValidator{
		db:        db,
		secret:    secret,
		algorithm: algorithm,
	}
}

// Validate validates the token from the given context against the given
// validator funcs.
func (v JWTValidator) Validate(ctx context.Context, funcs ...ValidatorFunc) error {
	tokenStr, err := getTokenFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get token from context error")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Header["alg"] != v.algorithm {
			return nil, ErrInvalidAlgorithm
		}
		return []byte(v.secret), nil
	})
	if err != nil {
		return errors.Wrap(err, "jwt parse error")
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		// no need to use a static error, this should never happen
		return fmt.Errorf("api/auth: expected *Claims, got %T", token.Claims)
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

func getTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return "", ErrNoMetadataInContext
	}

	token, ok := md["authorization"]
	if !ok || len(token) == 0 {
		return "", ErrNoAuthorizationInMetadata
	}

	return token[0], nil
}

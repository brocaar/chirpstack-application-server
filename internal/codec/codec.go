package codec

import (
	"fmt"

	"github.com/brocaar/chirpstack-application-server/internal/codec/cayennelpp"
	"github.com/brocaar/chirpstack-application-server/internal/codec/js"
	"github.com/lib/pq/hstore"
)

// Type defines the codec type.
type Type string

// Available codec types.
const (
	None                = ""
	CayenneLPPType Type = "CAYENNE_LPP"
	CustomJSType   Type = "CUSTOM_JS"
)

// BinaryToJSON encodes the given binary payload to JSON.
func BinaryToJSON(t Type, fPort uint8, variables hstore.Hstore, decodeScript string, b []byte) ([]byte, error) {
	vars := make(map[string]string)
	for k, v := range variables.Map {
		if v.Valid {
			vars[k] = v.String
		}
	}

	switch t {
	case CayenneLPPType:
		return cayennelpp.BinaryToJSON(b)
	case CustomJSType:
		return js.BinaryToJSON(fPort, vars, decodeScript, b)
	default:
		return nil, fmt.Errorf("unknown codec type: %s", t)
	}
}

// JSONToBinary encodes the given JSON to binary.
func JSONToBinary(t Type, fPort uint8, variables hstore.Hstore, encodeScript string, jsonB []byte) ([]byte, error) {
	vars := make(map[string]string)
	for k, v := range variables.Map {
		if v.Valid {
			vars[k] = v.String
		}
	}

	switch t {
	case CayenneLPPType:
		return cayennelpp.JSONToBinary(jsonB)
	case CustomJSType:
		return js.JSONToBinary(fPort, vars, encodeScript, jsonB)
	default:
		return nil, fmt.Errorf("unknown codec type: %s", t)
	}
}

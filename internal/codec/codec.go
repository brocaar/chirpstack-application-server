package codec

// Type defines the codec type.
type Type string

// Available codec types.
const (
	CayenneLPPType Type = "CAYENNE_LPP"
	CustomJSType   Type = "CUSTOM_JS"
)

// Payload defines a codec payload.
type Payload interface {
	DecodeBytes(data []byte) error
	EncodeToBytes() ([]byte, error)
	Object() interface{}
}

// NewPayload returns a new codec payload. In case of an unknown Type, nil is
// returned.
func NewPayload(t Type, fPort uint8, encodeScript, decodeScript string) Payload {
	switch t {
	case CayenneLPPType:
		return &CayenneLPP{}
	case CustomJSType:
		return NewCustomJS(fPort, encodeScript, decodeScript)
	default:
		return nil
	}
}

package helpers

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
)

// HEXBytes defines a type which represents bytes as HEX when marshaled to
// text.
type HEXBytes []byte

// String implements fmt.Stringer.
func (hb HEXBytes) String() string {
	return hex.EncodeToString(hb[:])
}

// MarshalText implements encoding.TextMarshaler.
func (hb HEXBytes) MarshalText() ([]byte, error) {
	return []byte(hb.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (hb *HEXBytes) UnmarshalText(text []byte) error {
	b, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}
	*hb = HEXBytes(b)
	return nil
}

// GetTimestamp returns the RX timestamp.
func GetTimestamp(rxInfo []*gw.UplinkRXInfo) time.Time {
	for i := range rxInfo {
		if rxInfo[i].Time != nil {
			t, err := ptypes.Timestamp(rxInfo[i].Time)
			if err == nil {
				return t
			}
		}
	}

	return time.Now()
}

// EUI64 defines the EUI64.
type EUI64 [8]byte

// MarshalText implements encoding.TextMarshaler.
func (b EUI64) MarshalText() ([]byte, error) {
	var str []string
	for i := range b[:] {
		str = append(str, hex.EncodeToString([]byte{b[i]}))
	}

	return []byte(strings.Join(str, "-")), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *EUI64) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), "-")
	if len(parts) != 8 {
		return errors.New("eui64 must be 8 bytes")
	}

	for i := range parts {
		bb, err := hex.DecodeString(parts[i])
		if err != nil {
			return errors.Wrap(err, "decode hex error")
		}
		if len(bb) != 1 {
			return errors.New("exactly 1 byte expected")
		}
		b[i] = bb[0]
	}

	return nil
}

// GetTimeSinceGPSEpoch returns the time since GPS epoch if it is available
// in the uplink payload.
func GetTimeSinceGPSEpoch(rxInfo []*gw.UplinkRXInfo) *time.Duration {
	for i := range rxInfo {
		if rxInfo[i].TimeSinceGpsEpoch != nil {
			d, err := ptypes.Duration(rxInfo[i].TimeSinceGpsEpoch)
			if err == nil {
				return &d
			}
		}
	}

	return nil
}

// GetStartLocation returns the location of the gateway closest to the device
// in terms of SNR.
func GetStartLocation(rxInfo []*gw.UplinkRXInfo) *common.Location {
	var snr *float64
	var loc *common.Location

	for i := range rxInfo {
		if rxInfo[i].Location == nil {
			continue
		}

		if snr == nil || *snr < rxInfo[i].LoraSnr {
			snr = &rxInfo[i].LoraSnr
			loc = rxInfo[i].Location
		}
	}

	return loc
}

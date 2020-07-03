package geolocation

import (
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
)

// getTimeSinceGPSEpoch returns the time since GPS epoch if it is available
// in the uplink payload.
func getTimeSinceGPSEpoch(rxInfo []*gw.UplinkRXInfo) *time.Duration {
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

// getStartLocation returns the location of the gateway closest to the device
// in terms of SNR.
func getStartLocation(rxInfo []*gw.UplinkRXInfo) *common.Location {
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

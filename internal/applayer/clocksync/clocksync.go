package clocksync

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/downlink"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/applayer/clocksync"
)

// HandleClockSyncCommand handles an uplink clock synchronization command.
func HandleClockSyncCommand(db sqlx.Ext, devEUI lorawan.EUI64, timeSinceGPSEpoch time.Duration, b []byte) error {
	var cmd clocksync.Command

	if err := cmd.UnmarshalBinary(true, b); err != nil {
		return errors.Wrap(err, "unmarshal command error")
	}

	switch cmd.CID {
	case clocksync.AppTimeReq:
		pl, ok := cmd.Payload.(*clocksync.AppTimeReqPayload)
		if !ok {
			return fmt.Errorf("expected *clocksync.AppTimeReqPayload, got: %T", cmd.Payload)
		}
		if err := handleAppTimeReq(db, devEUI, timeSinceGPSEpoch, pl); err != nil {
			return errors.Wrap(err, "handle AppTimeReq error")
		}
	default:
		return fmt.Errorf("CID not implemented: %s", cmd.CID)
	}

	return nil
}

func handleAppTimeReq(db sqlx.Ext, devEUI lorawan.EUI64, timeSinceGPSEpoch time.Duration, pl *clocksync.AppTimeReqPayload) error {
	deviceGPSTime := int64(pl.DeviceTime)
	networkGPSTime := int64((timeSinceGPSEpoch / time.Second) % (1 << 32))

	log.WithFields(log.Fields{
		"dev_eui":      devEUI,
		"device_time":  pl.DeviceTime,
		"ans_required": pl.Param.AnsRequired,
		"token_req":    pl.Param.TokenReq,
	}).Info("AppTimeReq received")

	ans := clocksync.Command{
		CID: clocksync.AppTimeAns,
		Payload: &clocksync.AppTimeAnsPayload{
			TimeCorrection: int32(networkGPSTime - deviceGPSTime),
			Param: clocksync.AppTimeAnsPayloadParam{
				TokenAns: pl.Param.TokenReq,
			},
		},
	}
	b, err := ans.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "marshal command error")
	}

	_, err = downlink.EnqueueDownlinkPayload(db, devEUI, false, uint8(clocksync.DefaultFPort), b)
	if err != nil {
		return errors.Wrap(err, "enqueue downlink payload error")
	}

	log.WithFields(log.Fields{
		"dev_eui":         devEUI,
		"time_correction": int32(networkGPSTime - deviceGPSTime),
		"token_ans":       pl.Param.TokenReq,
	}).Info("AppTimeAns enqueued")

	return nil
}

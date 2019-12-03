package marshaler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-api/go/as/integration"
	models "github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

// Type defines the marshaler type.
type Type int

// Marshaler types.
const (
	JSONV3 Type = iota
	Protobuf
	ProtobufJSON
)

// Marshal marshals the given payload.
func Marshal(t Type, msg proto.Message) ([]byte, error) {
	switch t {
	case Protobuf:
		return marshalProtobuf(msg)
	case ProtobufJSON:
		return marshalProtobufJSON(msg)
	case JSONV3:
		return marshalJSONV3(msg)
	default:
		return nil, fmt.Errorf("unknown marshaler type: %v", t)
	}
}

func marshalProtobuf(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

func marshalProtobufJSON(msg proto.Message) ([]byte, error) {
	m := &jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: true,
	}
	str, err := m.MarshalToString(msg)
	return []byte(str), err
}

func unmarshalProtobufJSON(b []byte, msg proto.Message) error {
	um := &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
	return um.Unmarshal(bytes.NewReader(b), msg)
}

func marshalJSONV3(msg proto.Message) ([]byte, error) {
	switch v := msg.(type) {
	case *integration.UplinkEvent:
		return jsonv3MarshalUplinkEvent(v)
	case *integration.JoinEvent:
		return jsonv3MarshalJoinEvent(v)
	case *integration.AckEvent:
		return jsonv3MarshalAckEvent(v)
	case *integration.ErrorEvent:
		return jsonv3MarshalErrorEvent(v)
	case *integration.StatusEvent:
		return jsonv3MarshalStatusEvent(v)
	case *integration.LocationEvent:
		return jsonv3MarshalLocationEvent(v)
	default:
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}
}

func jsonv3MarshalUplinkEvent(msg *integration.UplinkEvent) ([]byte, error) {
	//obj := make(map[string]interface{})
	var obj interface{}
	if msg.ObjectJson != "" {
		if err := json.Unmarshal([]byte(msg.ObjectJson), &obj); err != nil {
			fmt.Println(msg.ObjectJson)
			log.WithError(err).Error("integration/marshaler: unmarshal json error")
		}
	}

	m := models.DataUpPayload{
		ApplicationID:   int64(msg.ApplicationId),
		ApplicationName: msg.ApplicationName,
		DeviceName:      msg.DeviceName,
		TXInfo: models.TXInfo{
			Frequency: int(msg.TxInfo.Frequency),
			DR:        int(msg.Dr),
		},
		ADR:    msg.Adr,
		FCnt:   msg.FCnt,
		FPort:  uint8(msg.FPort),
		Data:   msg.Data,
		Tags:   msg.Tags,
		Object: obj,
	}

	copy(m.DevEUI[:], msg.DevEui)

	var gatewayIDs []lorawan.EUI64
	for i := range msg.RxInfo {
		rxInfo := models.RXInfo{
			RSSI:    int(msg.RxInfo[i].Rssi),
			LoRaSNR: float64(msg.RxInfo[i].LoraSnr),
		}

		copy(rxInfo.GatewayID[:], msg.RxInfo[i].GatewayId)
		copy(rxInfo.UplinkID[:], msg.RxInfo[i].UplinkId)

		if msg.RxInfo[i].Time != nil {
			t, err := ptypes.Timestamp(msg.RxInfo[i].Time)
			if err == nil {
				rxInfo.Time = &t
			}
		}

		if msg.RxInfo[i].Location != nil {
			rxInfo.Location = &models.Location{
				Latitude:  msg.RxInfo[i].Location.Latitude,
				Longitude: msg.RxInfo[i].Location.Longitude,
				Altitude:  msg.RxInfo[i].Location.Altitude,
			}
		}

		m.RXInfo = append(m.RXInfo, rxInfo)
		gatewayIDs = append(gatewayIDs, rxInfo.GatewayID)
	}

	gws, err := storage.GetGatewaysForMACs(context.Background(), storage.DB(), gatewayIDs)
	if err != nil {
		return nil, errors.Wrap(err, "get gateways for ids error")
	}
	for i := range m.RXInfo {
		if gw, ok := gws[m.RXInfo[i].GatewayID]; ok {
			m.RXInfo[i].Name = gw.Name
		}
	}

	return json.Marshal(m)
}

func jsonv3MarshalJoinEvent(msg *integration.JoinEvent) ([]byte, error) {
	m := models.JoinNotification{
		ApplicationID:   int64(msg.ApplicationId),
		ApplicationName: msg.ApplicationName,
		DeviceName:      msg.DeviceName,
		TXInfo: models.TXInfo{
			Frequency: int(msg.TxInfo.Frequency),
			DR:        int(msg.Dr),
		},
		Tags: msg.Tags,
	}

	copy(m.DevEUI[:], msg.DevEui)
	copy(m.DevAddr[:], msg.DevAddr)

	var gatewayIDs []lorawan.EUI64
	for i := range msg.RxInfo {
		rxInfo := models.RXInfo{
			RSSI:    int(msg.RxInfo[i].Rssi),
			LoRaSNR: float64(msg.RxInfo[i].LoraSnr),
		}

		copy(rxInfo.GatewayID[:], msg.RxInfo[i].GatewayId)
		copy(rxInfo.UplinkID[:], msg.RxInfo[i].UplinkId)

		if msg.RxInfo[i].Time != nil {
			t, err := ptypes.Timestamp(msg.RxInfo[i].Time)
			if err == nil {
				rxInfo.Time = &t
			}
		}

		if msg.RxInfo[i].Location != nil {
			rxInfo.Location = &models.Location{
				Latitude:  msg.RxInfo[i].Location.Latitude,
				Longitude: msg.RxInfo[i].Location.Longitude,
				Altitude:  msg.RxInfo[i].Location.Altitude,
			}
		}

		m.RXInfo = append(m.RXInfo, rxInfo)
		gatewayIDs = append(gatewayIDs, rxInfo.GatewayID)
	}

	gws, err := storage.GetGatewaysForMACs(context.Background(), storage.DB(), gatewayIDs)
	if err != nil {
		return nil, errors.Wrap(err, "get gateways for ids error")
	}
	for i := range m.RXInfo {
		if gw, ok := gws[m.RXInfo[i].GatewayID]; ok {
			m.RXInfo[i].Name = gw.Name
		}
	}

	return json.Marshal(m)
}

func jsonv3MarshalAckEvent(msg *integration.AckEvent) ([]byte, error) {
	m := models.ACKNotification{
		ApplicationID:   int64(msg.ApplicationId),
		ApplicationName: msg.ApplicationName,
		DeviceName:      msg.DeviceName,
		Acknowledged:    msg.Acknowledged,
		FCnt:            msg.FCnt,
		Tags:            msg.Tags,
	}

	copy(m.DevEUI[:], msg.DevEui)

	return json.Marshal(m)
}

func jsonv3MarshalErrorEvent(msg *integration.ErrorEvent) ([]byte, error) {
	m := models.ErrorNotification{
		ApplicationID:   int64(msg.ApplicationId),
		ApplicationName: msg.ApplicationName,
		DeviceName:      msg.DeviceName,
		FCnt:            msg.FCnt,
		Type:            msg.Type.String(),
		Error:           msg.Error,
		Tags:            msg.Tags,
	}

	copy(m.DevEUI[:], msg.DevEui)

	return json.Marshal(m)
}

func jsonv3MarshalStatusEvent(msg *integration.StatusEvent) ([]byte, error) {
	m := models.StatusNotification{
		ApplicationID:           int64(msg.ApplicationId),
		ApplicationName:         msg.ApplicationName,
		DeviceName:              msg.DeviceName,
		Margin:                  int(msg.Margin),
		ExternalPowerSource:     msg.ExternalPowerSource,
		BatteryLevelUnavailable: msg.BatteryLevelUnavailable,
		BatteryLevel:            msg.BatteryLevel,
		Tags:                    msg.Tags,
	}

	copy(m.DevEUI[:], msg.DevEui)

	return json.Marshal(m)
}

func jsonv3MarshalLocationEvent(msg *integration.LocationEvent) ([]byte, error) {
	m := models.LocationNotification{
		ApplicationID:   int64(msg.ApplicationId),
		ApplicationName: msg.ApplicationName,
		DeviceName:      msg.DeviceName,
		Tags:            msg.Tags,
	}

	if msg.Location != nil {
		m.Location = models.Location{
			Latitude:  msg.Location.Latitude,
			Longitude: msg.Location.Longitude,
			Altitude:  msg.Location.Altitude,
		}
	}

	copy(m.DevEUI[:], msg.DevEui)

	return json.Marshal(m)
}

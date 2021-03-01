package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lib/pq/hstore"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

type deviceUp struct {
	ID              uuid.UUID       `db:"id"`
	ReceivedAt      time.Time       `db:"received_at"`
	DevEUI          lorawan.EUI64   `db:"dev_eui"`
	DeviceName      string          `db:"device_name"`
	ApplicationID   int64           `db:"application_id"`
	ApplicationName string          `db:"application_name"`
	Frequency       int             `db:"frequency"`
	DR              int             `db:"dr"`
	ADR             bool            `db:"adr"`
	FCnt            int             `db:"f_cnt"`
	FPort           int             `db:"f_port"`
	Data            []byte          `db:"data"`
	RXInfo          json.RawMessage `db:"rx_info"`
	Object          json.RawMessage `db:"object"`
	Tags            hstore.Hstore   `db:"tags"`
}

type deviceStatus struct {
	ID                      uuid.UUID     `db:"id"`
	ReceivedAt              time.Time     `db:"received_at"`
	DevEUI                  lorawan.EUI64 `db:"dev_eui"`
	DeviceName              string        `db:"device_name"`
	ApplicationID           int64         `db:"application_id"`
	ApplicationName         string        `db:"application_name"`
	Margin                  int           `db:"margin"`
	ExternalPowerSource     bool          `db:"external_power_source"`
	BatteryLevelUnavailable bool          `db:"battery_level_unavailable"`
	BatteryLevel            float32       `db:"battery_level"`
	Tags                    hstore.Hstore `db:"tags"`
}

type deviceJoin struct {
	ID              uuid.UUID       `db:"id"`
	ReceivedAt      time.Time       `db:"received_at"`
	DevEUI          lorawan.EUI64   `db:"dev_eui"`
	DeviceName      string          `db:"device_name"`
	ApplicationID   int64           `db:"application_id"`
	ApplicationName string          `db:"application_name"`
	DevAddr         lorawan.DevAddr `db:"dev_addr"`
	Tags            hstore.Hstore   `db:"tags"`
}

type deviceAck struct {
	ID              uuid.UUID     `db:"id"`
	ReceivedAt      time.Time     `db:"received_at"`
	DevEUI          lorawan.EUI64 `db:"dev_eui"`
	DeviceName      string        `db:"device_name"`
	ApplicationID   int64         `db:"application_id"`
	ApplicationName string        `db:"application_name"`
	Acknowledged    bool          `db:"acknowledged"`
	FCnt            int           `db:"f_cnt"`
	Tags            hstore.Hstore `db:"tags"`
}

type deviceError struct {
	ID              uuid.UUID     `db:"id"`
	ReceivedAt      time.Time     `db:"received_at"`
	DevEUI          lorawan.EUI64 `db:"dev_eui"`
	DeviceName      string        `db:"device_name"`
	ApplicationID   int64         `db:"application_id"`
	ApplicationName string        `db:"application_name"`
	Type            string        `db:"type"`
	Error           string        `db:"error"`
	FCnt            int           `db:"f_cnt"`
	Tags            hstore.Hstore `db:"tags"`
}

type deviceLocation struct {
	ID              uuid.UUID     `db:"id"`
	ReceivedAt      time.Time     `db:"received_at"`
	DevEUI          lorawan.EUI64 `db:"dev_eui"`
	DeviceName      string        `db:"device_name"`
	ApplicationID   int64         `db:"application_id"`
	ApplicationName string        `db:"application_name"`
	Altitude        float64       `db:"altitude"`
	Latitude        float64       `db:"latitude"`
	Longitude       float64       `db:"longitude"`
	Geohash         string        `db:"geohash"`
	Accuracy        int           `db:"accuracy"`
	Tags            hstore.Hstore `db:"tags"`
}

func init() {
	log.SetLevel(log.ErrorLevel)
}

type PostgreSQLTestSuite struct {
	suite.Suite

	db          *sqlx.DB
	integration *Integration
}

func (ts *PostgreSQLTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))
	test.MustResetDB(storage.DB().DB)

	dsn := "postgres://localhost/chirpstack_as_test?sslmode=disable"
	if v := os.Getenv("TEST_POSTGRES_DSN"); v != "" {
		dsn = v
	}

	var err error
	ts.db, err = sqlx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	ts.integration, err = New(config.IntegrationPostgreSQLConfig{
		DSN: dsn,
	})
	if err != nil {
		panic(err)
	}

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	ns := storage.NetworkServer{
		Name: "test-ns",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &ns))

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	gw := storage.Gateway{
		MAC:             lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
		Name:            "test-gw",
		OrganizationID:  org.ID,
		NetworkServerID: ns.ID,
	}
	assert.NoError(storage.CreateGateway(context.Background(), storage.DB(), &gw))
}

func (ts *PostgreSQLTestSuite) TearDownSuite() {
	if err := ts.integration.Close(); err != nil {
		panic(err)
	}
}

func (ts *PostgreSQLTestSuite) SetupTest() {
	if err := MigrateDown(ts.db); err != nil {
		panic(err)
	}
	if err := MigrateUp(ts.db); err != nil {
		panic(err)
	}
}

func (ts *PostgreSQLTestSuite) TestHandleUplinkEvent() {
	assert := require.New(ts.T())

	timestamp := time.Now().Round(time.Second).UTC()
	tsProto, err := ptypes.TimestampProto(timestamp)
	assert.NoError(err)

	pl := pb.UplinkEvent{
		ApplicationId:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
		RxInfo: []*gw.UplinkRXInfo{
			{
				GatewayId: []byte{8, 7, 6, 5, 4, 3, 2, 1},
				Time:      tsProto,
				Rssi:      20,
				LoraSnr:   10,
			},
		},
		TxInfo: &gw.UplinkTXInfo{
			Frequency: 868100000,
		},
		Dr:    4,
		Adr:   true,
		FCnt:  2,
		FPort: 3,
		Data:  []byte{1, 2, 3, 4},
		ObjectJson: `{
			"temp": 21.5,
			"hum":  44.3
		}`,
		Tags: map[string]string{
			"foo": "bar",
		},
	}

	assert.NoError(ts.integration.HandleUplinkEvent(context.Background(), nil, nil, pl))

	var up deviceUp
	assert.NoError(ts.db.Get(&up, "select * from device_up"))

	up.ReceivedAt = up.ReceivedAt.UTC()

	assert.NotEqual(json.RawMessage("null"), up.RXInfo)
	assert.NotEqual(json.RawMessage("null"), up.Object)
	up.RXInfo = nil
	up.Object = nil

	assert.NotEqual(uuid.Nil, up.ID)
	up.ID = uuid.Nil

	assert.Equal(deviceUp{
		ReceivedAt:      timestamp,
		ApplicationID:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Frequency:       868100000,
		DR:              4,
		ADR:             true,
		FCnt:            2,
		FPort:           3,
		Data:            []byte{1, 2, 3, 4},
		Tags: hstore.Hstore{
			Map: map[string]sql.NullString{
				"foo": sql.NullString{String: "bar", Valid: true},
			},
		},
	}, up)
}

func (ts *PostgreSQLTestSuite) TestHandleUplinkEventNoObject() {
	assert := require.New(ts.T())

	timestamp := time.Now().Round(time.Second).UTC()
	tsProto, err := ptypes.TimestampProto(timestamp)
	assert.NoError(err)

	pl := pb.UplinkEvent{
		ApplicationId:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
		RxInfo: []*gw.UplinkRXInfo{
			{
				GatewayId: []byte{8, 7, 6, 5, 4, 3, 2, 1},
				Time:      tsProto,
				Rssi:      20,
				LoraSnr:   10,
			},
		},
		TxInfo: &gw.UplinkTXInfo{
			Frequency: 868100000,
		},
		Dr:    4,
		Adr:   true,
		FCnt:  2,
		FPort: 3,
		Data:  []byte{1, 2, 3, 4},
		Tags: map[string]string{
			"foo": "bar",
		},
	}

	assert.NoError(ts.integration.HandleUplinkEvent(context.Background(), nil, nil, pl))

	var up deviceUp
	assert.NoError(ts.db.Get(&up, "select * from device_up"))

	up.ReceivedAt = up.ReceivedAt.UTC()

	assert.NotEqual(json.RawMessage("null"), up.RXInfo)
	up.RXInfo = nil

	assert.NotEqual(uuid.Nil, up.ID)
	up.ID = uuid.Nil

	assert.Equal(deviceUp{
		ReceivedAt:      timestamp,
		ApplicationID:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Frequency:       868100000,
		DR:              4,
		ADR:             true,
		FCnt:            2,
		FPort:           3,
		Data:            []byte{1, 2, 3, 4},
		Object:          json.RawMessage("null"),
		Tags: hstore.Hstore{
			Map: map[string]sql.NullString{
				"foo": sql.NullString{String: "bar", Valid: true},
			},
		},
	}, up)
}

func (ts *PostgreSQLTestSuite) TestUplinkEventNoData() {
	assert := require.New(ts.T())

	timestamp := time.Now().Round(time.Second).UTC()
	tsPB, err := ptypes.TimestampProto(timestamp)
	assert.NoError(err)

	pl := pb.UplinkEvent{
		ApplicationId:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
		RxInfo: []*gw.UplinkRXInfo{
			{
				GatewayId: []byte{8, 7, 6, 5, 4, 3, 2, 1},
				Time:      tsPB,
				Rssi:      20,
				LoraSnr:   10,
			},
		},
		TxInfo: &gw.UplinkTXInfo{
			Frequency: 868100000,
		},
		Dr:    4,
		Adr:   true,
		FCnt:  2,
		FPort: 3,
		Tags: map[string]string{
			"foo": "bar",
		},
	}

	assert.NoError(ts.integration.HandleUplinkEvent(context.Background(), nil, nil, pl))

	var up deviceUp
	assert.NoError(ts.db.Get(&up, "select * from device_up"))

	up.ReceivedAt = up.ReceivedAt.UTC()

	assert.NotEqual(json.RawMessage("null"), up.RXInfo)
	up.RXInfo = nil

	assert.NotEqual(uuid.Nil, up.ID)
	up.ID = uuid.Nil

	assert.Equal(deviceUp{
		ReceivedAt:      timestamp,
		ApplicationID:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Frequency:       868100000,
		DR:              4,
		ADR:             true,
		FCnt:            2,
		FPort:           3,
		Data:            []byte{},
		Object:          json.RawMessage("null"),
		Tags: hstore.Hstore{
			Map: map[string]sql.NullString{
				"foo": sql.NullString{String: "bar", Valid: true},
			},
		},
	}, up)
}

func (ts *PostgreSQLTestSuite) TestHandleStatusEvent() {
	assert := require.New(ts.T())

	timestamp := time.Now()

	pl := pb.StatusEvent{
		ApplicationId:           1,
		ApplicationName:         "test-app",
		DeviceName:              "test-device",
		DevEui:                  []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Margin:                  10,
		ExternalPowerSource:     true,
		BatteryLevelUnavailable: true,
		BatteryLevel:            75.5,
		Tags: map[string]string{
			"foo": "bar",
		},
	}

	assert.NoError(ts.integration.HandleStatusEvent(context.Background(), nil, nil, pl))

	var status deviceStatus
	assert.NoError(ts.db.Get(&status, "select * from device_status"))

	assert.True(status.ReceivedAt.After(timestamp))
	status.ReceivedAt = timestamp

	assert.NotEqual(uuid.Nil, status.ID)
	status.ID = uuid.Nil

	assert.Equal(deviceStatus{
		ReceivedAt:              timestamp,
		ApplicationID:           1,
		ApplicationName:         "test-app",
		DeviceName:              "test-device",
		DevEUI:                  lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Margin:                  10,
		ExternalPowerSource:     true,
		BatteryLevelUnavailable: true,
		BatteryLevel:            75.5,
		Tags: hstore.Hstore{
			Map: map[string]sql.NullString{
				"foo": sql.NullString{String: "bar", Valid: true},
			},
		},
	}, status)

}

func (ts *PostgreSQLTestSuite) TestHandleJoinEvent() {
	assert := require.New(ts.T())

	timestamp := time.Now()

	pl := pb.JoinEvent{
		ApplicationId:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
		DevAddr:         []byte{1, 2, 3, 4},
		Tags: map[string]string{
			"foo": "bar",
		},
	}

	assert.NoError(ts.integration.HandleJoinEvent(context.Background(), nil, nil, pl))

	var join deviceJoin
	assert.NoError(ts.db.Get(&join, "select * from device_join"))

	assert.True(join.ReceivedAt.After(timestamp))
	join.ReceivedAt = timestamp

	assert.NotEqual(uuid.Nil, join.ID)
	join.ID = uuid.Nil

	assert.Equal(deviceJoin{
		ReceivedAt:      timestamp,
		ApplicationID:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		DevAddr:         lorawan.DevAddr{1, 2, 3, 4},
		Tags: hstore.Hstore{
			Map: map[string]sql.NullString{
				"foo": sql.NullString{String: "bar", Valid: true},
			},
		},
	}, join)
}

func (ts *PostgreSQLTestSuite) TestHandleAckEvent() {
	assert := require.New(ts.T())

	timestamp := time.Now()

	pl := pb.AckEvent{
		ApplicationId:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Acknowledged:    true,
		FCnt:            10,
		Tags: map[string]string{
			"foo": "bar",
		},
	}

	assert.NoError(ts.integration.HandleAckEvent(context.Background(), nil, nil, pl))

	var ack deviceAck
	assert.NoError(ts.db.Get(&ack, "select * from device_ack"))

	assert.True(ack.ReceivedAt.After(timestamp))
	ack.ReceivedAt = timestamp

	assert.NotEqual(uuid.Nil, ack.ID)
	ack.ID = uuid.Nil

	assert.Equal(deviceAck{
		ReceivedAt:      timestamp,
		ApplicationID:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Acknowledged:    true,
		FCnt:            10,
		Tags: hstore.Hstore{
			Map: map[string]sql.NullString{
				"foo": sql.NullString{String: "bar", Valid: true},
			},
		},
	}, ack)
}

func (ts *PostgreSQLTestSuite) TestHandleErrorEvent() {
	assert := require.New(ts.T())

	timestamp := time.Now()

	pl := pb.ErrorEvent{
		ApplicationId:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Type:            pb.ErrorType_DOWNLINK_PAYLOAD_SIZE,
		Error:           "Everything blew up!",
		FCnt:            10,
		Tags: map[string]string{
			"foo": "bar",
		},
	}

	assert.NoError(ts.integration.HandleErrorEvent(context.Background(), nil, nil, pl))

	var e deviceError
	assert.NoError(ts.db.Get(&e, "select * from device_error"))

	assert.True(e.ReceivedAt.After(timestamp))
	e.ReceivedAt = timestamp

	assert.NotEqual(uuid.Nil, e.ID)
	e.ID = uuid.Nil

	assert.Equal(deviceError{
		ReceivedAt:      timestamp,
		ApplicationID:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Type:            "DOWNLINK_PAYLOAD_SIZE",
		Error:           "Everything blew up!",
		FCnt:            10,
		Tags: hstore.Hstore{
			Map: map[string]sql.NullString{
				"foo": sql.NullString{String: "bar", Valid: true},
			},
		},
	}, e)
}

func (ts *PostgreSQLTestSuite) TestLocationEvent() {
	assert := require.New(ts.T())

	timestamp := time.Now()

	pl := pb.LocationEvent{
		ApplicationId:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Location: &common.Location{
			Altitude:  1.123,
			Latitude:  2.123,
			Longitude: 3.123,
		},
		Tags: map[string]string{
			"foo": "bar",
		},
	}

	assert.NoError(ts.integration.HandleLocationEvent(context.Background(), nil, nil, pl))

	var loc deviceLocation
	assert.NoError(ts.db.Get(&loc, "select * from device_location"))

	assert.True(loc.ReceivedAt.After(timestamp))
	loc.ReceivedAt = timestamp

	assert.NotEqual(uuid.Nil, loc.ID)
	loc.ID = uuid.Nil

	assert.Equal(deviceLocation{
		ReceivedAt:      timestamp,
		ApplicationID:   1,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Altitude:        1.123,
		Latitude:        2.123,
		Longitude:       3.123,
		Geohash:         "s06hp46p75vs",
		Tags: hstore.Hstore{
			Map: map[string]sql.NullString{
				"foo": sql.NullString{String: "bar", Valid: true},
			},
		},
	}, loc)
}

func TestPostgreSQL(t *testing.T) {
	suite.Run(t, new(PostgreSQLTestSuite))
}

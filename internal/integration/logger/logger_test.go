package logger

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/eventlog"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

type LoggerTestSuite struct {
	suite.Suite

	logChannel  chan eventlog.EventLog
	ctx         context.Context
	cancelFunc  context.CancelFunc
	devEUI      lorawan.EUI64
	integration *Integration
}

func (ts *LoggerTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()
	conf.Monitoring.PerDeviceEventLogMaxHistory = 10
	config.Set(conf)
	assert.NoError(storage.Setup(conf))

	ts.logChannel = make(chan eventlog.EventLog, 1)
	ts.ctx, ts.cancelFunc = context.WithCancel(context.Background())

	ts.integration, _ = New(Config{})

	go func() {
		if err := eventlog.GetEventLogForDevice(ts.ctx, ts.devEUI, ts.logChannel); err != nil {
			panic(err)
		}
	}()

	// some time to subscribe
	time.Sleep(100 * time.Millisecond)
}

func (ts *LoggerTestSuite) TearDownSuite() {
	ts.cancelFunc()
}

func (ts *LoggerTestSuite) TestHandleUplinkEvent() {
	assert := require.New(ts.T())
	pl := pb.UplinkEvent{
		DevEui: ts.devEUI[:],
	}
	assert.NoError(ts.integration.HandleUplinkEvent(context.Background(), nil, nil, pl))
	el := <-ts.logChannel
	assert.NotEqual("", el.StreamID)
	el.StreamID = ""
	assert.Equal(toEventLog("up", &pl), el)
}

func (ts *LoggerTestSuite) TestJoinEvent() {
	assert := require.New(ts.T())
	pl := pb.JoinEvent{
		DevEui: ts.devEUI[:],
	}
	assert.NoError(ts.integration.HandleJoinEvent(context.Background(), nil, nil, pl))
	el := <-ts.logChannel
	assert.NotEqual("", el.StreamID)
	el.StreamID = ""
	assert.Equal(toEventLog("join", &pl), el)
}

func (ts *LoggerTestSuite) TestAckEvent() {
	assert := require.New(ts.T())
	pl := pb.AckEvent{
		DevEui: ts.devEUI[:],
	}
	assert.NoError(ts.integration.HandleAckEvent(context.Background(), nil, nil, pl))
	el := <-ts.logChannel
	assert.NotEqual("", el.StreamID)
	el.StreamID = ""
	assert.Equal(toEventLog("ack", &pl), el)
}

func (ts *LoggerTestSuite) TestErrorEvent() {
	assert := require.New(ts.T())
	pl := pb.ErrorEvent{
		DevEui: ts.devEUI[:],
	}
	assert.NoError(ts.integration.HandleErrorEvent(context.Background(), nil, nil, pl))
	el := <-ts.logChannel
	assert.NotEqual("", el.StreamID)
	el.StreamID = ""
	assert.Equal(toEventLog("error", &pl), el)
}

func (ts *LoggerTestSuite) TestStatusEvent() {
	assert := require.New(ts.T())
	pl := pb.StatusEvent{
		DevEui: ts.devEUI[:],
	}
	assert.NoError(ts.integration.HandleStatusEvent(context.Background(), nil, nil, pl))
	el := <-ts.logChannel
	assert.NotEqual("", el.StreamID)
	el.StreamID = ""
	assert.Equal(toEventLog("status", &pl), el)
}

func (ts *LoggerTestSuite) TestLocationEvent() {
	assert := require.New(ts.T())
	pl := pb.LocationEvent{
		DevEui: ts.devEUI[:],
	}
	assert.NoError(ts.integration.HandleLocationEvent(context.Background(), nil, nil, pl))
	el := <-ts.logChannel
	assert.NotEqual("", el.StreamID)
	el.StreamID = ""
	assert.Equal(toEventLog("location", &pl), el)
}

func (ts *LoggerTestSuite) TestTxAckEvent() {
	assert := require.New(ts.T())
	pl := pb.TxAckEvent{
		DevEui: ts.devEUI[:],
	}
	assert.NoError(ts.integration.HandleTxAckEvent(context.Background(), nil, nil, pl))
	el := <-ts.logChannel
	assert.NotEqual("", el.StreamID)
	el.StreamID = ""
	assert.Equal(toEventLog("txack", &pl), el)
}

func (ts *LoggerTestSuite) TestIntegrationEvent() {
	assert := require.New(ts.T())
	pl := pb.IntegrationEvent{
		DevEui: ts.devEUI[:],
	}
	assert.NoError(ts.integration.HandleIntegrationEvent(context.Background(), nil, nil, pl))
	el := <-ts.logChannel
	assert.NotEqual("", el.StreamID)
	el.StreamID = ""
	assert.Equal(toEventLog("integration", &pl), el)
}

func toEventLog(t string, msg proto.Message) eventlog.EventLog {
	b, err := marshaler.Marshal(marshaler.ProtobufJSON, msg)
	if err != nil {
		panic(err)
	}

	return eventlog.EventLog{
		Type:    t,
		Payload: b,
	}
}

func TestLogger(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

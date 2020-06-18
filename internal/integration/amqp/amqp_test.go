package amqp

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

type IntegrationTestSuite struct {
	suite.Suite

	integration *Integration

	applicationID uint64
	devEUI        lorawan.EUI64

	amqpConn      *amqp.Connection
	amqpChan      *amqp.Channel
	amqpEventChan <-chan amqp.Delivery
}

func (ts *IntegrationTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()

	ts.applicationID = 10
	ts.devEUI = lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	var err error
	ts.amqpConn, err = amqp.Dial(conf.ApplicationServer.Integration.AMQP.URL)
	assert.NoError(err)

	ts.amqpChan, err = ts.amqpConn.Channel()
	assert.NoError(err)

	_, err = ts.amqpChan.QueueDeclare(
		"test-event-queue",
		true,
		false,
		false,
		false,
		nil,
	)
	assert.NoError(err)

	err = ts.amqpChan.QueueBind(
		"test-event-queue",
		"application.*.device.*.event.*",
		"amq.topic",
		false,
		nil,
	)
	assert.NoError(err)

	ts.amqpEventChan, err = ts.amqpChan.Consume(
		"test-event-queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	assert.NoError(err)

	ts.integration, err = New(marshaler.Protobuf, conf.ApplicationServer.Integration.AMQP)
	assert.NoError(err)
}

func (ts *IntegrationTestSuite) TearDownSuite() {
	assert := require.New(ts.T())

	assert.NoError(ts.amqpChan.Close())
	assert.NoError(ts.amqpConn.Close())
}

func (ts *IntegrationTestSuite) TestUplinkEvent() {
	assert := require.New(ts.T())

	pl := pb.UplinkEvent{
		ApplicationId: ts.applicationID,
		DevEui:        ts.devEUI[:],
	}

	assert.NoError(ts.integration.HandleUplinkEvent(context.Background(), nil, nil, pl))

	msg := <-ts.amqpEventChan

	assert.Equal("application.10.device.0102030405060708.event.up", msg.RoutingKey)
	assert.Equal("application/octet-stream", msg.ContentType)

	var plReceived pb.UplinkEvent

	assert.NoError(proto.Unmarshal(msg.Body, &plReceived))
	assert.True(proto.Equal(&pl, &plReceived))
}

func (ts *IntegrationTestSuite) TestJoinEvent() {
	assert := require.New(ts.T())

	pl := pb.JoinEvent{
		ApplicationId: ts.applicationID,
		DevEui:        ts.devEUI[:],
	}

	assert.NoError(ts.integration.HandleJoinEvent(context.Background(), nil, nil, pl))

	msg := <-ts.amqpEventChan

	assert.Equal("application.10.device.0102030405060708.event.join", msg.RoutingKey)
	assert.Equal("application/octet-stream", msg.ContentType)

	var plReceived pb.JoinEvent

	assert.NoError(proto.Unmarshal(msg.Body, &plReceived))
	assert.True(proto.Equal(&pl, &plReceived))
}

func (ts *IntegrationTestSuite) TestAckEvent() {
	assert := require.New(ts.T())

	pl := pb.AckEvent{
		ApplicationId: ts.applicationID,
		DevEui:        ts.devEUI[:],
	}

	assert.NoError(ts.integration.HandleAckEvent(context.Background(), nil, nil, pl))

	msg := <-ts.amqpEventChan

	assert.Equal("application.10.device.0102030405060708.event.ack", msg.RoutingKey)
	assert.Equal("application/octet-stream", msg.ContentType)

	var plReceived pb.AckEvent

	assert.NoError(proto.Unmarshal(msg.Body, &plReceived))
	assert.True(proto.Equal(&pl, &plReceived))
}

func (ts *IntegrationTestSuite) TestErrorEvent() {
	assert := require.New(ts.T())

	pl := pb.ErrorEvent{
		ApplicationId: ts.applicationID,
		DevEui:        ts.devEUI[:],
	}

	assert.NoError(ts.integration.HandleErrorEvent(context.Background(), nil, nil, pl))

	msg := <-ts.amqpEventChan

	assert.Equal("application.10.device.0102030405060708.event.error", msg.RoutingKey)
	assert.Equal("application/octet-stream", msg.ContentType)

	var plReceived pb.ErrorEvent

	assert.NoError(proto.Unmarshal(msg.Body, &plReceived))
	assert.True(proto.Equal(&pl, &plReceived))
}

func (ts *IntegrationTestSuite) TestStatusEvent() {
	assert := require.New(ts.T())

	pl := pb.StatusEvent{
		ApplicationId: ts.applicationID,
		DevEui:        ts.devEUI[:],
	}

	assert.NoError(ts.integration.HandleStatusEvent(context.Background(), nil, nil, pl))

	msg := <-ts.amqpEventChan

	assert.Equal("application.10.device.0102030405060708.event.status", msg.RoutingKey)
	assert.Equal("application/octet-stream", msg.ContentType)

	var plReceived pb.StatusEvent

	assert.NoError(proto.Unmarshal(msg.Body, &plReceived))
	assert.True(proto.Equal(&pl, &plReceived))
}

func (ts *IntegrationTestSuite) TestLocationEvent() {
	assert := require.New(ts.T())

	pl := pb.LocationEvent{
		ApplicationId: ts.applicationID,
		DevEui:        ts.devEUI[:],
	}

	assert.NoError(ts.integration.HandleLocationEvent(context.Background(), nil, nil, pl))

	msg := <-ts.amqpEventChan

	assert.Equal("application.10.device.0102030405060708.event.location", msg.RoutingKey)
	assert.Equal("application/octet-stream", msg.ContentType)

	var plReceived pb.LocationEvent

	assert.NoError(proto.Unmarshal(msg.Body, &plReceived))
	assert.True(proto.Equal(&pl, &plReceived))
}

func (ts *IntegrationTestSuite) TestTxAckEvent() {
	assert := require.New(ts.T())

	pl := pb.TxAckEvent{
		ApplicationId: ts.applicationID,
		DevEui:        ts.devEUI[:],
	}

	assert.NoError(ts.integration.HandleTxAckEvent(context.Background(), nil, nil, pl))

	msg := <-ts.amqpEventChan

	assert.Equal("application.10.device.0102030405060708.event.txack", msg.RoutingKey)
	assert.Equal("application/octet-stream", msg.ContentType)

	var plReceived pb.TxAckEvent

	assert.NoError(proto.Unmarshal(msg.Body, &plReceived))
	assert.True(proto.Equal(&pl, &plReceived))
}

func (ts *IntegrationTestSuite) TestIntegrationEvent() {
	assert := require.New(ts.T())

	pl := pb.IntegrationEvent{
		ApplicationId: ts.applicationID,
		DevEui:        ts.devEUI[:],
	}

	assert.NoError(ts.integration.HandleIntegrationEvent(context.Background(), nil, nil, pl))

	msg := <-ts.amqpEventChan

	assert.Equal("application.10.device.0102030405060708.event.integration", msg.RoutingKey)
	assert.Equal("application/octet-stream", msg.ContentType)

	var plReceived pb.IntegrationEvent

	assert.NoError(proto.Unmarshal(msg.Body, &plReceived))
	assert.True(proto.Equal(&pl, &plReceived))
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

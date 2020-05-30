package kafka

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"

	"github.com/segmentio/kafka-go"
)

type IntegrationTestSuite struct {
	suite.Suite
	integration *Integration

	applicationID uint64
	devEUI        lorawan.EUI64

	conn *kafka.Conn
}

func (ts *IntegrationTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()

	ts.applicationID = 10
	ts.devEUI = lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	kafkaConf := conf.ApplicationServer.Integration.Kafka

	// Note: We instantiate a reader before we instantiate the integration. An
	// integration immediately creates a writer. If kafka was just started for the
	// first time for our tests, that first writer connection and kafka are in a state
	// where the first message write fails: "Leader Not Available: the cluster is in
	// the middle of a leadership election and there is currently no leader for this
	// partition and hence it is unavailable for writes". Creating a reader first seems
	// to prevent that.

	var err error
	ts.conn, err = kafka.DialLeader(context.Background(), "tcp", kafkaConf.Brokers[0], kafkaConf.Topic, 0)
	assert.NoError(err)

	_, err = ts.conn.Seek(0, kafka.SeekEnd)
	assert.NoError(err)

	ts.integration, err = New(marshaler.Protobuf, kafkaConf)
	assert.NoError(err)
}

func (ts *IntegrationTestSuite) TearDownSuite() {
	assert := require.New(ts.T())
	assert.NoError(ts.conn.Close())
}

func (ts *IntegrationTestSuite) TestUplink() {
	tx := pb.UplinkEvent{
		ApplicationId: ts.applicationID,
		DevEui:        ts.devEUI[:],
	}
	var rx pb.UplinkEvent
	err := ts.integration.HandleUplinkEvent(context.Background(), nil, nil, tx)
	ts.checkMessage(err, &tx, &rx)
}

func (ts *IntegrationTestSuite) checkMessage(err error, tx, rx proto.Message) {
	assert := require.New(ts.T())

	assert.NoError(err)

	msg, err := ts.conn.ReadMessage(1 << 10)
	assert.NoError(err)

	assert.Equal("application.10.device.0102030405060708.event.up", string(msg.Key))

	assert.NoError(proto.Unmarshal(msg.Value, rx))
	assert.True(proto.Equal(tx, rx))
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

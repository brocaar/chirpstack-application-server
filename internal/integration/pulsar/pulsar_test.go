package pulsar

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
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

	client   pulsar.Client
	consumer pulsar.Consumer
}

func (ts *IntegrationTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()

	ts.applicationID = 10
	ts.devEUI = lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	pulsarConf := conf.ApplicationServer.Integration.Pulsar

	clientOpts := pulsar.ClientOptions{
		URL:                        pulsarConf.Brokers,
		OperationTimeout:           30 * time.Second,
		ConnectionTimeout:          30 * time.Second,
		TLSTrustCertsFilePath:      pulsarConf.TLSTrustCertsFilePath,
		TLSAllowInsecureConnection: pulsarConf.TLSAllowInsecureConnection,
		MaxConnectionsPerBroker:    pulsarConf.MaxConnectionsPerBroker,
	}

	if pulsarConf.AuthType == config.PulsarAuthTypeOAuth2 {
		oauth2 := pulsar.NewAuthenticationOAuth2(map[string]string{
			"type":       "client_credentials",
			"issuerUrl":  pulsarConf.OAuth2.IssuerURL,
			"audience":   pulsarConf.OAuth2.Audience,
			"clientId":   pulsarConf.OAuth2.ClientID,
			"privateKey": pulsarConf.OAuth2.PrivateKey,
		})
		clientOpts.Authentication = oauth2
	}

	var err error
	ts.client, err = pulsar.NewClient(clientOpts)
	assert.NoError(err)
	assert.NotNil(ts.client)

	ts.consumer, err = ts.client.Subscribe(pulsar.ConsumerOptions{
		Topic:            pulsarConf.Topic,
		SubscriptionName: "chirpstack_as",
	})
	assert.NoError(err)

	err = ts.consumer.Seek(pulsar.LatestMessageID())
	assert.NoError(err)

	ts.integration, err = New(marshaler.Protobuf, pulsarConf)
	assert.NoError(err)
}

func (ts *IntegrationTestSuite) TearDownSuite() {
	ts.consumer.Unsubscribe()
	ts.consumer.Close()
	ts.client.Close()
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

	msg, err := ts.consumer.Receive(context.Background())
	assert.NoError(err)

	assert.Equal("application.10.device.0102030405060708.event.up", msg.Key())

	assert.Equal("up", msg.Properties()["event"])

	assert.NoError(proto.Unmarshal(msg.Payload(), rx))
	assert.True(proto.Equal(tx, rx))

	err = ts.consumer.Ack(msg)
	assert.NoError(err)
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

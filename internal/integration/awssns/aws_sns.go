// Package awssns implements an AWS NSN integration.
package awssns

import (
	"context"
	"encoding/base64"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Integration implements the AWS SNS integration.
type Integration struct {
	marshaler marshaler.Type
	sns       *sns.SNS
	topicARN  string
}

// New creates a new AWS SNS integration.
func New(m marshaler.Type, conf config.IntegrationAWSSNSConfig) (*Integration, error) {
	if conf.Marshaler != "" {
		switch conf.Marshaler {
		case "PROTOBUF":
			m = marshaler.Protobuf
		case "JSON":
			m = marshaler.ProtobufJSON
		case "JSON_V3":
			m = marshaler.JSONV3
		}
	}

	i := Integration{
		marshaler: m,
		topicARN:  conf.TopicARN,
	}

	log.Info("integration/awssns: setting up session")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(conf.AWSRegion),
		Credentials: credentials.NewStaticCredentials(conf.AWSAccessKeyID, conf.AWSSecretAccessKey, ""),
	})
	if err != nil {
		return nil, errors.Wrap(err, "new session error")
	}
	i.sns = sns.New(sess)

	return &i, nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	return i.publish(ctx, "up", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	return i.publish(ctx, "join", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	return i.publish(ctx, "ack", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	return i.publish(ctx, "error", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	return i.publish(ctx, "status", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	return i.publish(ctx, "location", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleTxAckEvent sends a TxAckEevent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	return i.publish(ctx, "txack", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	return i.publish(ctx, "integration", pl.ApplicationId, pl.DevEui, &pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	return nil
}

func (i *Integration) publish(ctx context.Context, event string, applicationID uint64, devEUIB []byte, msg proto.Message) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIB)

	b, err := marshaler.Marshal(i.marshaler, msg)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	// base64 encode the Protobuf payload as the message must be a UTF-8 string.
	if i.marshaler == marshaler.Protobuf {
		b = []byte(base64.StdEncoding.EncodeToString(b))
	}

	_, err = i.sns.Publish(&sns.PublishInput{
		Message: aws.String(string(b)),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"event":          &sns.MessageAttributeValue{DataType: aws.String("String"), StringValue: aws.String(event)},
			"dev_eui":        &sns.MessageAttributeValue{DataType: aws.String("String"), StringValue: aws.String(devEUI.String())},
			"application_id": &sns.MessageAttributeValue{DataType: aws.String("String"), StringValue: aws.String(strconv.FormatInt(int64(applicationID), 10))},
		},
		TopicArn: aws.String(i.topicARN),
	})
	if err != nil {
		return errors.Wrap(err, "sns publish")
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"event":   event,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/awssns: event published")

	return nil
}

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

	pb "github.com/brocaar/chirpstack-api/go/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
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

	log.WithField("topic_arn", i.topicARN).Info("integration/awssns: testing if topic exists")
	_, err = i.sns.GetTopicAttributes(&sns.GetTopicAttributesInput{
		TopicArn: aws.String(i.topicARN),
	})
	if err != nil {
		return nil, errors.Wrap(err, "get topic error")
	}

	return &i, nil
}

// SendDataUp sends an uplink data payload.
func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	return i.publish(ctx, "up", pl.ApplicationId, pl.DevEui, &pl)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	return i.publish(ctx, "join", pl.ApplicationId, pl.DevEui, &pl)
}

// SendACKNotification sends an ack notification.
func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	return i.publish(ctx, "ack", pl.ApplicationId, pl.DevEui, &pl)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {
	return i.publish(ctx, "error", pl.ApplicationId, pl.DevEui, &pl)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	return i.publish(ctx, "status", pl.ApplicationId, pl.DevEui, &pl)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	return i.publish(ctx, "location", pl.ApplicationId, pl.DevEui, &pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
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

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"event":   event,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/awssns: event published")

	return nil
}

// Package awssns implements an AWS NSN integration.
package awssns

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lora-app-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Config holds the AWS SNS integration configuration.
type Config struct {
	AWSRegion          string `mapstructure:"aws_region"`
	AWSAccessKeyID     string `mapstructure:"aws_access_key_id"`
	AWSSecretAccessKey string `mapstructure:"aws_secret_access_key"`
	TopicARN           string `mapstructure:"topic_arn"`
}

// Integration implements the AWS SNS integration.
type Integration struct {
	sns      *sns.SNS
	topicARN string
}

// New creates a new AWS SNS integration.
func New(conf Config) (*Integration, error) {
	i := Integration{
		topicARN: conf.TopicARN,
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
func (i *Integration) SendDataUp(ctx context.Context, pl integration.DataUpPayload) error {
	return i.publish(ctx, "up", pl.ApplicationID, pl.DevEUI, pl)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(ctx context.Context, pl integration.JoinNotification) error {
	return i.publish(ctx, "join", pl.ApplicationID, pl.DevEUI, pl)
}

// SendACKNotification sends an ack notification.
func (i *Integration) SendACKNotification(ctx context.Context, pl integration.ACKNotification) error {
	return i.publish(ctx, "ack", pl.ApplicationID, pl.DevEUI, pl)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(ctx context.Context, pl integration.ErrorNotification) error {
	return i.publish(ctx, "error", pl.ApplicationID, pl.DevEUI, pl)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(ctx context.Context, pl integration.StatusNotification) error {
	return i.publish(ctx, "status", pl.ApplicationID, pl.DevEUI, pl)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(ctx context.Context, pl integration.LocationNotification) error {
	return i.publish(ctx, "location", pl.ApplicationID, pl.DevEUI, pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	return nil
}

func (i *Integration) publish(ctx context.Context, event string, applicationID int64, devEUI lorawan.EUI64, v interface{}) error {
	jsonB, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	_, err = i.sns.Publish(&sns.PublishInput{
		Message: aws.String(string(jsonB)),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"event":          &sns.MessageAttributeValue{DataType: aws.String("String"), StringValue: aws.String(event)},
			"dev_eui":        &sns.MessageAttributeValue{DataType: aws.String("String"), StringValue: aws.String(devEUI.String())},
			"application_id": &sns.MessageAttributeValue{DataType: aws.String("String"), StringValue: aws.String(strconv.FormatInt(applicationID, 10))},
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

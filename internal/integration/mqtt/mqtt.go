package mqtt

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"sync"
	"text/template"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

const downlinkLockTTL = time.Millisecond * 100

var (
	clientCACert       string
	clientCAKey        string
	clientCertLifetime time.Duration
)

// Integration implements a MQTT integration.
type Integration struct {
	marshaler            marshaler.Type
	conn                 mqtt.Client
	dataDownChan         chan models.DataDownPayload
	wg                   sync.WaitGroup
	config               config.IntegrationMQTTConfig
	eventTopicTemplate   *template.Template
	commandTopicTemplate *template.Template
	downlinkTopic        string
	downlinkRegexp       *regexp.Regexp
	retainEvents         bool

	// For backwards compatibility.
	uplinkTemplate      *template.Template
	downlinkTemplate    *template.Template
	joinTemplate        *template.Template
	ackTemplate         *template.Template
	errorTemplate       *template.Template
	statusTemplate      *template.Template
	locationTemplate    *template.Template
	txAckTemplate       *template.Template
	integrationTemplate *template.Template
	uplinkRetained      bool
	joinRetained        bool
	ackRetained         bool
	errorRetained       bool
	statusRetained      bool
	locationRetained    bool
	txAckRetained       bool
	integrationRetained bool
}

// Setup configures the MQTT package.
func Setup(c config.Config) error {
	clientCACert = c.ApplicationServer.Integration.MQTT.Client.CACert
	clientCAKey = c.ApplicationServer.Integration.MQTT.Client.CAKey
	clientCertLifetime = c.ApplicationServer.Integration.MQTT.Client.ClientCertLifetime
	return nil
}

// New creates a new MQTT integration.
func New(m marshaler.Type, conf config.IntegrationMQTTConfig) (*Integration, error) {
	var err error
	i := Integration{
		marshaler:    m,
		dataDownChan: make(chan models.DataDownPayload),
		config:       conf,
	}

	i.retainEvents = i.config.RetainEvents
	i.eventTopicTemplate, err = template.New("event").Parse(i.config.EventTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse event template error")
	}
	i.commandTopicTemplate, err = template.New("command").Parse(i.config.CommandTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse command template error")
	}

	// For backwards compatibility.
	if i.config.UplinkTopicTemplate != "" {
		i.uplinkTemplate, err = template.New("uplink").Parse(i.config.UplinkTopicTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parse uplink template error")
		}
	}
	if i.config.DownlinkTopicTemplate != "" {
		i.downlinkTemplate, err = template.New("downlink").Parse(i.config.DownlinkTopicTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parse downlink template error")
		}
	}
	if i.config.JoinTopicTemplate != "" {
		i.joinTemplate, err = template.New("join").Parse(i.config.JoinTopicTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parse join template error")
		}
	}
	if i.config.AckTopicTemplate != "" {
		i.ackTemplate, err = template.New("ack").Parse(i.config.AckTopicTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parse ack template error")
		}
	}
	if i.config.ErrorTopicTemplate != "" {
		i.errorTemplate, err = template.New("error").Parse(i.config.ErrorTopicTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parse error template error")
		}
	}
	if i.config.StatusTopicTemplate != "" {
		i.statusTemplate, err = template.New("status").Parse(i.config.StatusTopicTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parse status template error")
		}
	}
	if i.config.LocationTopicTemplate != "" {
		i.locationTemplate, err = template.New("location").Parse(i.config.LocationTopicTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parse location template error")
		}
	}
	if i.config.TxAckTopicTemplate != "" {
		i.txAckTemplate, err = template.New("txack").Parse(i.config.TxAckTopicTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parse tx ack template error")
		}
	}
	if i.config.IntegrationTopicTemplate != "" {
		i.integrationTemplate, err = template.New("integration").Parse(i.config.IntegrationTopicTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parse integration template error")
		}
	}
	i.uplinkRetained = i.config.UplinkRetainedMessage
	i.joinRetained = i.config.JoinRetainedMessage
	i.ackRetained = i.config.AckRetainedMessage
	i.errorRetained = i.config.ErrorRetainedMessage
	i.statusRetained = i.config.StatusRetainedMessage
	i.locationRetained = i.config.LocationRetainedMessage
	i.txAckRetained = i.config.TxAckRetainedMessage
	i.integrationRetained = i.config.IntegrationRetainedMessage

	// generate downlink topic matching all applications and devices
	i.downlinkTopic, err = i.getDownlinkTopic()
	if err != nil {
		return nil, errors.Wrap(err, "get downlink topic error")
	}

	// generate downlink topic regexp
	i.downlinkRegexp, err = i.getDownlinkTopicRegexp()
	if err != nil {
		return nil, errors.Wrap(err, "get downlink topic regexp error")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(i.config.Server)
	opts.SetUsername(i.config.Username)
	opts.SetPassword(i.config.Password)
	opts.SetCleanSession(i.config.CleanSession)
	opts.SetClientID(i.config.ClientID)
	opts.SetOnConnectHandler(i.onConnected)
	opts.SetConnectionLostHandler(i.onConnectionLost)
	opts.SetMaxReconnectInterval(i.config.MaxReconnectInterval)

	tlsconfig, err := newTLSConfig(i.config.CACert, i.config.TLSCert, i.config.TLSKey)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"ca_cert":  i.config.CACert,
			"tls_cert": i.config.TLSCert,
			"tls_key":  i.config.TLSKey,
		}).Fatalf("error loading mqtt certificate files")
	}
	if tlsconfig != nil {
		opts.SetTLSConfig(tlsconfig)
	}

	log.WithField("server", i.config.Server).Info("integration/mqtt: connecting to mqtt broker")
	i.conn = mqtt.NewClient(opts)
	for {
		if token := i.conn.Connect(); token.Wait() && token.Error() != nil {
			log.Errorf("integration/mqtt: connecting to broker error, will retry in 2s: %s", token.Error())
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	return &i, nil
}

func newTLSConfig(cafile, certFile, certKeyFile string) (*tls.Config, error) {
	// Here are three valid options:
	//   - Only CA
	//   - TLS cert + key
	//   - CA, TLS cert + key

	if cafile == "" && certFile == "" && certKeyFile == "" {
		log.Info("integration/mqtt: TLS config is empty")
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	// Import trusted certificates from CAfile.pem.
	if cafile != "" {
		cacert, err := ioutil.ReadFile(cafile)
		if err != nil {
			log.Errorf("integration/mqtt: couldn't load cafile: %s", err)
			return nil, err
		}
		certpool := x509.NewCertPool()
		certpool.AppendCertsFromPEM(cacert)

		tlsConfig.RootCAs = certpool // RootCAs = certs used to verify server cert.
	}

	// Import certificate and the key
	if certFile != "" || certKeyFile != "" {
		kp, err := tls.LoadX509KeyPair(certFile, certKeyFile) // here raises error when the pair of cert and key are invalid (e.g. either one is empty)
		if err != nil {
			log.Errorf("integration/mqtt: couldn't load MQTT TLS key pair: %s", err)
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{kp}
	}

	return tlsConfig, nil
}

// Close stops the handler.
func (i *Integration) Close() error {
	log.Info("integration/mqtt: closing handler")
	log.WithField("topic", i.downlinkTopic).Info("integration/mqtt: unsubscribing from tx topic")
	if token := i.conn.Unsubscribe(i.downlinkTopic); token.Wait() && token.Error() != nil {
		return fmt.Errorf("integration/mqtt: unsubscribe from %s error: %s", i.downlinkTopic, token.Error())
	}
	log.Info("integration/mqtt: handling last items in queue")
	i.wg.Wait()
	close(i.dataDownChan)
	return nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, payload pb.UplinkEvent) error {
	return i.publish(ctx, payload.ApplicationId, payload.DevEui, "up", &payload)
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, payload pb.JoinEvent) error {
	return i.publish(ctx, payload.ApplicationId, payload.DevEui, "join", &payload)
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, payload pb.AckEvent) error {
	return i.publish(ctx, payload.ApplicationId, payload.DevEui, "ack", &payload)
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, payload pb.ErrorEvent) error {
	return i.publish(ctx, payload.ApplicationId, payload.DevEui, "error", &payload)
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, payload pb.StatusEvent) error {
	return i.publish(ctx, payload.ApplicationId, payload.DevEui, "status", &payload)
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, payload pb.LocationEvent) error {
	return i.publish(ctx, payload.ApplicationId, payload.DevEui, "location", &payload)
}

// HandleTxAckEvent sends a TxAckEvent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, payload pb.TxAckEvent) error {
	return i.publish(ctx, payload.ApplicationId, payload.DevEui, "txack", &payload)
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, payload pb.IntegrationEvent) error {
	return i.publish(ctx, payload.ApplicationId, payload.DevEui, "integration", &payload)
}

func (i *Integration) publish(ctx context.Context, applicationID uint64, devEUIB []byte, eventType string, msg proto.Message) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIB)

	topic, err := i.getTopic(applicationID, devEUI, eventType)
	if err != nil {
		return errors.Wrap(err, "get topic error")
	}

	retain := i.getRetainEvents(eventType)

	b, err := marshaler.Marshal(i.marshaler, msg)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"retain":  retain,
		"topic":   topic,
		"qos":     i.config.QOS,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/mqtt: publishing event")
	if token := i.conn.Publish(topic, i.config.QOS, retain, b); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	mqttEventCounter(eventType).Inc()

	return nil
}

// DataDownChan returns the channel containing the received DataDownPayload.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return i.dataDownChan
}

func (i *Integration) getTXTopicVariables(topic string) (int64, lorawan.EUI64, error) {
	var applicationID int64
	var devEUI lorawan.EUI64
	var err error

	match := i.downlinkRegexp.FindStringSubmatch(topic)
	if len(match) != len(i.downlinkRegexp.SubexpNames()) {
		return applicationID, devEUI, errors.New("topic regex match error")
	}

	result := make(map[string]string)
	for i, name := range i.downlinkRegexp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	if idStr, ok := result["application_id"]; ok {
		applicationID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return applicationID, devEUI, errors.Wrap(err, "parse application id error")
		}
	} else {
		return applicationID, devEUI, errors.New("topic regexp does not contain application id")
	}

	if devEUIStr, ok := result["dev_eui"]; ok {
		if err = devEUI.UnmarshalText([]byte(devEUIStr)); err != nil {
			return applicationID, devEUI, errors.Wrap(err, "parse deveui error")
		}
	}

	return applicationID, devEUI, nil
}

func (i *Integration) txPayloadHandler(mqttc mqtt.Client, msg mqtt.Message) {
	i.wg.Add(1)
	defer i.wg.Done()

	log.WithField("topic", msg.Topic()).Info("integration/mqtt: downlink event received")
	topicApplicationID, topicDevEUI, err := i.getTXTopicVariables(msg.Topic())
	if err != nil {
		log.WithError(err).Warning("integration/mqtt: get variables from topic error")
		return
	}

	var pl models.DataDownPayload
	dec := json.NewDecoder(bytes.NewReader(msg.Payload()))
	if err := dec.Decode(&pl); err != nil {
		log.WithFields(log.Fields{
			"data_base64": base64.StdEncoding.EncodeToString(msg.Payload()),
		}).Errorf("integration/mqtt: tx payload unmarshal error: %s", err)
		return
	}

	pl.ApplicationID = topicApplicationID
	pl.DevEUI = topicDevEUI

	if pl.FPort == 0 || pl.FPort > 224 {
		log.WithFields(log.Fields{
			"topic":   msg.Topic(),
			"dev_eui": pl.DevEUI,
			"f_port":  pl.FPort,
		}).Error("integration/mqtt: fPort must be between 1 - 224")
		return
	}

	// Since with MQTT all subscribers will receive the downlink messages sent
	// by the application, the first instance receiving the message must lock it,
	// so that other instances can ignore the message.
	key := storage.GetRedisKey("lora:as:downlink:lock:%d:%s:%x", pl.ApplicationID, pl.DevEUI, sha256.Sum256(msg.Payload()))
	set, err := storage.RedisClient().SetNX(context.Background(), key, "lock", downlinkLockTTL).Result()
	if err != nil {
		log.WithError(err).Error("integration/mqtt: acquire lock error")
		return
	}

	// If we could not set, it means it is already locked by an other process.
	if !set {
		return
	}

	mqttCommandCounter("down").Inc()

	i.dataDownChan <- pl
}

func (i *Integration) onConnected(mqttc mqtt.Client) {
	log.Info("integration/mqtt: connected to mqtt broker")
	for {
		log.WithFields(log.Fields{
			"topic": i.downlinkTopic,
			"qos":   i.config.QOS,
		}).Info("integration/mqtt: subscribing to tx topic")
		if token := i.conn.Subscribe(i.downlinkTopic, i.config.QOS, i.txPayloadHandler); token.Wait() && token.Error() != nil {
			log.WithField("topic", i.downlinkTopic).Errorf("integration/mqtt: subscribe error: %s", token.Error())
			time.Sleep(time.Second)
			continue
		}
		return
	}
}

func (i *Integration) onConnectionLost(mqttc mqtt.Client, reason error) {
	log.Errorf("integration/mqtt: mqtt connection error: %s", reason)
}

func (i *Integration) getDownlinkTopic() (string, error) {
	topic := bytes.NewBuffer(nil)
	topicTemplate := i.commandTopicTemplate
	if i.downlinkTemplate != nil {
		topicTemplate = i.downlinkTemplate
	}

	err := topicTemplate.Execute(topic, struct {
		ApplicationID string
		DevEUI        string
		CommandType   string
	}{"+", "+", "down"})
	if err != nil {
		return "", errors.Wrap(err, "execute template error")
	}
	return topic.String(), nil
}

func (i *Integration) getDownlinkTopicRegexp() (*regexp.Regexp, error) {
	topic := bytes.NewBuffer(nil)
	topicTemplate := i.commandTopicTemplate
	if i.downlinkTemplate != nil {
		topicTemplate = i.downlinkTemplate
	}

	err := topicTemplate.Execute(topic, struct {
		ApplicationID string
		DevEUI        string
		CommandType   string
	}{`(?P<application_id>\w+)`, `(?P<dev_eui>\w+)`, `(?P<command_type>\w)`})
	if err != nil {
		return nil, errors.Wrap(err, "execute template error")
	}

	r, err := regexp.Compile(topic.String())
	if err != nil {
		return nil, errors.Wrap(err, "compile regexp error")
	}

	return r, nil
}

func (i *Integration) getTopic(applicationID uint64, devEUI lorawan.EUI64, eventType string) (string, error) {
	var topicTemplate *template.Template

	// For backwards compatibility.
	switch eventType {
	case "up":
		topicTemplate = i.uplinkTemplate
	case "join":
		topicTemplate = i.joinTemplate
	case "ack":
		topicTemplate = i.ackTemplate
	case "error":
		topicTemplate = i.errorTemplate
	case "status":
		topicTemplate = i.statusTemplate
	case "location":
		topicTemplate = i.locationTemplate
	case "txack":
		topicTemplate = i.txAckTemplate
	case "integration":
		topicTemplate = i.integrationTemplate
	}

	if topicTemplate == nil {
		topicTemplate = i.eventTopicTemplate
	}

	topic := bytes.NewBuffer(nil)
	err := topicTemplate.Execute(topic, struct {
		ApplicationID uint64
		DevEUI        lorawan.EUI64
		EventType     string
	}{applicationID, devEUI, eventType})
	if err != nil {
		return "", errors.Wrap(err, "execute template error")
	}

	return topic.String(), nil
}

func (i *Integration) getRetainEvents(eventType string) bool {
	if i.retainEvents {
		return true
	}

	// For backwards compatibility
	switch eventType {
	case "up":
		return i.uplinkRetained
	case "join":
		return i.joinRetained
	case "ack":
		return i.ackRetained
	case "error":
		return i.errorRetained
	case "status":
		return i.statusRetained
	case "location":
		return i.locationRetained
	case "txack":
		return i.txAckRetained
	case "integration":
		return i.integrationRetained
	default:
		return false
	}
}

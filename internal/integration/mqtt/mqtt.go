package mqtt

import (
	"bytes"
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
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lorawan"
)

const downlinkLockTTL = time.Millisecond * 100

// Config holds the configuration for the MQTT integration.
type Config struct {
	Server                  string
	Username                string
	Password                string
	QOS                     uint8  `mapstructure:"qos"`
	CleanSession            bool   `mapstructure:"clean_session"`
	ClientID                string `mapstructure:"client_id"`
	CACert                  string `mapstructure:"ca_cert"`
	TLSCert                 string `mapstructure:"tls_cert"`
	TLSKey                  string `mapstructure:"tls_key"`
	UplinkTopicTemplate     string `mapstructure:"uplink_topic_template"`
	DownlinkTopicTemplate   string `mapstructure:"downlink_topic_template"`
	JoinTopicTemplate       string `mapstructure:"join_topic_template"`
	AckTopicTemplate        string `mapstructure:"ack_topic_template"`
	ErrorTopicTemplate      string `mapstructure:"error_topic_template"`
	StatusTopicTemplate     string `mapstructure:"status_topic_template"`
	LocationTopicTemplate   string `mapstructure:"location_topic_template"`
	UplinkRetainedMessage   bool   `mapstructure:"uplink_retained_message"`
	JoinRetainedMessage     bool   `mapstructure:"join_retained_message"`
	AckRetainedMessage      bool   `mapstructure:"ack_retained_message"`
	ErrorRetainedMessage    bool   `mapstructure:"error_retained_message"`
	StatusRetainedMessage   bool   `mapstructure:"status_retained_message"`
	LocationRetainedMessage bool   `mapstructure:"location_retained_message"`
}

// Integration implements a MQTT integration.
type Integration struct {
	conn             mqtt.Client
	dataDownChan     chan integration.DataDownPayload
	wg               sync.WaitGroup
	redisPool        *redis.Pool
	config           Config
	uplinkTemplate   *template.Template
	downlinkTemplate *template.Template
	joinTemplate     *template.Template
	ackTemplate      *template.Template
	errorTemplate    *template.Template
	statusTemplate   *template.Template
	locationTemplate *template.Template
	downlinkTopic    string
	downlinkRegexp   *regexp.Regexp
	uplinkRetained   bool
	joinRetained     bool
	ackRetained      bool
	errorRetained    bool
	statusRetained   bool
	locationRetained bool
}

// New creates a new MQTT integration.
func New(p *redis.Pool, conf Config) (*Integration, error) {
	var err error
	i := Integration{
		dataDownChan: make(chan integration.DataDownPayload),
		redisPool:    p,
		config:       conf,
	}

	i.uplinkTemplate, err = template.New("uplink").Parse(i.config.UplinkTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse uplink template error")
	}
	i.downlinkTemplate, err = template.New("downlink").Parse(i.config.DownlinkTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse downlink template error")
	}
	i.joinTemplate, err = template.New("join").Parse(i.config.JoinTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse join template error")
	}
	i.ackTemplate, err = template.New("ack").Parse(i.config.AckTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse ack template error")
	}
	i.errorTemplate, err = template.New("error").Parse(i.config.ErrorTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse error template error")
	}
	i.statusTemplate, err = template.New("status").Parse(i.config.StatusTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse status template error")
	}
	i.locationTemplate, err = template.New("location").Parse(i.config.LocationTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse location template error")
	}
	i.uplinkRetained = i.config.UplinkRetainedMessage
	i.joinRetained = i.config.JoinRetainedMessage
	i.ackRetained = i.config.AckRetainedMessage
	i.errorRetained = i.config.ErrorRetainedMessage
	i.statusRetained = i.config.StatusRetainedMessage
	i.locationRetained = i.config.LocationRetainedMessage

	// generate downlink topic matching all applications and devices
	topic := bytes.NewBuffer(nil)
	err = i.downlinkTemplate.Execute(topic, struct {
		ApplicationID string
		DevEUI        string
	}{"+", "+"})
	if err != nil {
		return nil, errors.Wrap(err, "execute template error")
	}
	i.downlinkTopic = topic.String()

	// generate downlink topic regexp
	topic.Reset()
	err = i.downlinkTemplate.Execute(topic, struct {
		ApplicationID string
		DevEUI        string
	}{`(?P<application_id>\w+)`, `(?P<dev_eui>\w+)`})
	if err != nil {
		return nil, errors.Wrap(err, "execute template error")
	}
	i.downlinkRegexp, err = regexp.Compile(topic.String())
	if err != nil {
		return nil, errors.Wrap(err, "compile regexp error")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(i.config.Server)
	opts.SetUsername(i.config.Username)
	opts.SetPassword(i.config.Password)
	opts.SetCleanSession(i.config.CleanSession)
	opts.SetClientID(i.config.ClientID)
	opts.SetOnConnectHandler(i.onConnected)
	opts.SetConnectionLostHandler(i.onConnectionLost)

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

// SendDataUp sends a DataUpPayload.
func (i *Integration) SendDataUp(payload integration.DataUpPayload) error {
	return i.publish(payload.ApplicationID, payload.DevEUI, i.uplinkTemplate, i.uplinkRetained, payload)
}

// SendJoinNotification sends a JoinNotification.
func (i *Integration) SendJoinNotification(payload integration.JoinNotification) error {
	return i.publish(payload.ApplicationID, payload.DevEUI, i.joinTemplate, i.joinRetained, payload)
}

// SendACKNotification sends an ACKNotification.
func (i *Integration) SendACKNotification(payload integration.ACKNotification) error {
	return i.publish(payload.ApplicationID, payload.DevEUI, i.ackTemplate, i.ackRetained, payload)
}

// SendErrorNotification sends an ErrorNotification.
func (i *Integration) SendErrorNotification(payload integration.ErrorNotification) error {
	return i.publish(payload.ApplicationID, payload.DevEUI, i.errorTemplate, i.errorRetained, payload)
}

// SendStatusNotification sends a StatusNotification.
func (i *Integration) SendStatusNotification(payload integration.StatusNotification) error {
	return i.publish(payload.ApplicationID, payload.DevEUI, i.statusTemplate, i.statusRetained, payload)
}

// SendLocationNotification sends a LocationNotification.
func (i *Integration) SendLocationNotification(payload integration.LocationNotification) error {
	return i.publish(payload.ApplicationID, payload.DevEUI, i.locationTemplate, i.locationRetained, payload)
}

func (i *Integration) publish(applicationID int64, devEUI lorawan.EUI64, topicTemplate *template.Template, retained bool, v interface{}) error {
	topic := bytes.NewBuffer(nil)
	err := topicTemplate.Execute(topic, struct {
		ApplicationID int64
		DevEUI        lorawan.EUI64
	}{applicationID, devEUI})
	if err != nil {
		return errors.Wrap(err, "execute template error")
	}

	jsonB, err := json.Marshal(v)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"topic": topic.String(),
		"qos":   i.config.QOS,
	}).Info("integration/mqtt: publishing message")
	if token := i.conn.Publish(topic.String(), i.config.QOS, retained, jsonB); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// DataDownChan returns the channel containing the received DataDownPayload.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
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

	log.WithField("topic", msg.Topic()).Info("integration/mqtt: data-down payload received")
	topicApplicationID, topicDevEUI, err := i.getTXTopicVariables(msg.Topic())
	if err != nil {
		log.WithError(err).Warning("integration/mqtt: get variables from topic error")
		return
	}

	var pl integration.DataDownPayload
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
	key := fmt.Sprintf("lora:as:downlink:lock:%d:%s", pl.ApplicationID, pl.DevEUI)
	redisConn := i.redisPool.Get()
	defer redisConn.Close()

	_, err = redis.String(redisConn.Do("SET", key, "lock", "PX", int64(downlinkLockTTL/time.Millisecond), "NX"))
	if err != nil {
		if err == redis.ErrNil {
			// the payload is already being processed by an other instance
			return
		}
		log.Errorf("integration/mqtt: acquire downlink payload lock error: %s", err)
		return
	}

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

package mqtthandler

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
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lorawan"
)

const downlinkLockTTL = time.Millisecond * 100

// Config holds the configuration for the MQTT handler.
type Config struct {
	Server                string
	Username              string
	Password              string
	QOS                   uint8  `mapstructure:"qos"`
	CleanSession          bool   `mapstructure:"clean_session"`
	ClientID              string `mapstructure:"client_id"`
	CACert                string `mapstructure:"ca_cert"`
	TLSCert               string `mapstructure:"tls_cert"`
	TLSKey                string `mapstructure:"tls_key"`
	UplinkTopicTemplate   string `mapstructure:"uplink_topic_template"`
	DownlinkTopicTemplate string `mapstructure:"downlink_topic_template"`
	JoinTopicTemplate     string `mapstructure:"join_topic_template"`
	AckTopicTemplate      string `mapstructure:"ack_topic_template"`
	ErrorTopicTemplate    string `mapstructure:"error_topic_template"`
	StatusTopicTemplate   string `mapstructure:"status_topic_template"`
}

// MQTTHandler implements a MQTT handler for sending and receiving data by
// an application.
type MQTTHandler struct {
	conn             mqtt.Client
	dataDownChan     chan handler.DataDownPayload
	wg               sync.WaitGroup
	redisPool        *redis.Pool
	config           Config
	uplinkTemplate   *template.Template
	downlinkTemplate *template.Template
	joinTemplate     *template.Template
	ackTemplate      *template.Template
	errorTemplate    *template.Template
	statusTemplate   *template.Template
	downlinkTopic    string
	downlinkRegexp   *regexp.Regexp
}

// NewHandler creates a new MQTT handler.
func NewHandler(p *redis.Pool, c Config) (handler.Handler, error) {
	var err error
	h := MQTTHandler{
		dataDownChan: make(chan handler.DataDownPayload),
		redisPool:    p,
		config:       c,
	}

	h.uplinkTemplate, err = template.New("uplink").Parse(h.config.UplinkTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse uplink template error")
	}
	h.downlinkTemplate, err = template.New("downlink").Parse(h.config.DownlinkTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse downlink template error")
	}
	h.joinTemplate, err = template.New("join").Parse(h.config.JoinTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse join template error")
	}
	h.ackTemplate, err = template.New("ack").Parse(h.config.AckTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse ack template error")
	}
	h.errorTemplate, err = template.New("error").Parse(h.config.ErrorTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse error template error")
	}
	h.statusTemplate, err = template.New("status").Parse(h.config.StatusTopicTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse status template error")
	}

	// generate downlink topic matching all applications and devices
	topic := bytes.NewBuffer(nil)
	err = h.downlinkTemplate.Execute(topic, struct {
		ApplicationID string
		DevEUI        string
	}{"+", "+"})
	if err != nil {
		return nil, errors.Wrap(err, "execute template error")
	}
	h.downlinkTopic = topic.String()

	// generate downlink topic regexp
	topic.Reset()
	err = h.downlinkTemplate.Execute(topic, struct {
		ApplicationID string
		DevEUI        string
	}{`(?P<application_id>\w+)`, `(?P<dev_eui>\w+)`})
	if err != nil {
		return nil, errors.Wrap(err, "execute template error")
	}
	h.downlinkRegexp, err = regexp.Compile(topic.String())
	if err != nil {
		return nil, errors.Wrap(err, "compile regexp error")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(h.config.Server)
	opts.SetUsername(h.config.Username)
	opts.SetPassword(h.config.Password)
	opts.SetCleanSession(h.config.CleanSession)
	opts.SetClientID(h.config.ClientID)
	opts.SetOnConnectHandler(h.onConnected)
	opts.SetConnectionLostHandler(h.onConnectionLost)

	tlsconfig, err := newTLSConfig(h.config.CACert, h.config.TLSCert, h.config.TLSKey)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"ca_cert":  h.config.CACert,
			"tls_cert": h.config.TLSCert,
			"tls_key":  h.config.TLSKey,
		}).Fatalf("error loading mqtt certificate files")
	}
	if tlsconfig != nil {
		opts.SetTLSConfig(tlsconfig)
	}

	log.WithField("server", h.config.Server).Info("handler/mqtt: connecting to mqtt broker")
	h.conn = mqtt.NewClient(opts)
	for {
		if token := h.conn.Connect(); token.Wait() && token.Error() != nil {
			log.Errorf("handler/mqtt: connecting to broker error, will retry in 2s: %s", token.Error())
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	return &h, nil
}

func newTLSConfig(cafile, certFile, certKeyFile string) (*tls.Config, error) {
	// Here are three valid options:
	//   - Only CA
	//   - TLS cert + key
	//   - CA, TLS cert + key

	if cafile == "" && certFile == "" && certKeyFile == "" {
		log.Info("handler/mqtt: TLS config is empty")
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	// Import trusted certificates from CAfile.pem.
	if cafile != "" {
		cacert, err := ioutil.ReadFile(cafile)
		if err != nil {
			log.Errorf("handler/mqtt: couldn't load cafile: %s", err)
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
			log.Errorf("handler/mqtt: couldn't load MQTT TLS key pair: %s", err)
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{kp}
	}

	return tlsConfig, nil
}

// Close stops the handler.
func (h *MQTTHandler) Close() error {
	log.Info("handler/mqtt: closing handler")
	log.WithField("topic", h.downlinkTopic).Info("handler/mqtt: unsubscribing from tx topic")
	if token := h.conn.Unsubscribe(h.downlinkTopic); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: unsubscribe from %s error: %s", h.downlinkTopic, token.Error())
	}
	log.Info("handler/mqtt: handling last items in queue")
	h.wg.Wait()
	close(h.dataDownChan)
	return nil
}

// SendDataUp sends a DataUpPayload.
func (h *MQTTHandler) SendDataUp(payload handler.DataUpPayload) error {
	return h.publish(payload.ApplicationID, payload.DevEUI, h.uplinkTemplate, payload)
}

// SendJoinNotification sends a JoinNotification.
func (h *MQTTHandler) SendJoinNotification(payload handler.JoinNotification) error {
	return h.publish(payload.ApplicationID, payload.DevEUI, h.joinTemplate, payload)
}

// SendACKNotification sends an ACKNotification.
func (h *MQTTHandler) SendACKNotification(payload handler.ACKNotification) error {
	return h.publish(payload.ApplicationID, payload.DevEUI, h.ackTemplate, payload)
}

// SendErrorNotification sends an ErrorNotification.
func (h *MQTTHandler) SendErrorNotification(payload handler.ErrorNotification) error {
	return h.publish(payload.ApplicationID, payload.DevEUI, h.errorTemplate, payload)
}

// SendStatusNotification sends a StatusNotification.
func (h *MQTTHandler) SendStatusNotification(payload handler.StatusNotification) error {
	return h.publish(payload.ApplicationID, payload.DevEUI, h.statusTemplate, payload)
}

func (h *MQTTHandler) publish(applicationID int64, devEUI lorawan.EUI64, topicTemplate *template.Template, v interface{}) error {
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
		"qos":   h.config.QOS,
	}).Info("handler/mqtt: publishing message")
	if token := h.conn.Publish(topic.String(), h.config.QOS, false, jsonB); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// DataDownChan returns the channel containing the received DataDownPayload.
func (h *MQTTHandler) DataDownChan() chan handler.DataDownPayload {
	return h.dataDownChan
}

func (h *MQTTHandler) getTXTopicVariables(topic string) (int64, lorawan.EUI64, error) {
	var applicationID int64
	var devEUI lorawan.EUI64
	var err error

	match := h.downlinkRegexp.FindStringSubmatch(topic)
	if len(match) != len(h.downlinkRegexp.SubexpNames()) {
		return applicationID, devEUI, errors.New("topic regex match error")
	}

	result := make(map[string]string)
	for i, name := range h.downlinkRegexp.SubexpNames() {
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

func (h *MQTTHandler) txPayloadHandler(c mqtt.Client, msg mqtt.Message) {
	h.wg.Add(1)
	defer h.wg.Done()

	log.WithField("topic", msg.Topic()).Info("handler/mqtt: data-down payload received")
	topicApplicationID, topicDevEUI, err := h.getTXTopicVariables(msg.Topic())
	if err != nil {
		log.WithError(err).Warning("handler/mqtt: get variables from topic error")
		return
	}

	var pl handler.DataDownPayload
	dec := json.NewDecoder(bytes.NewReader(msg.Payload()))
	if err := dec.Decode(&pl); err != nil {
		log.WithFields(log.Fields{
			"data_base64": base64.StdEncoding.EncodeToString(msg.Payload()),
		}).Errorf("handler/mqtt: tx payload unmarshal error: %s", err)
		return
	}

	pl.ApplicationID = topicApplicationID
	pl.DevEUI = topicDevEUI

	if pl.FPort == 0 || pl.FPort > 224 {
		log.WithFields(log.Fields{
			"topic":   msg.Topic(),
			"dev_eui": pl.DevEUI,
			"f_port":  pl.FPort,
		}).Error("handler/mqtt: fPort must be between 1 - 224")
		return
	}

	// Since with MQTT all subscribers will receive the downlink messages sent
	// by the application, the first instance receiving the message must lock it,
	// so that other instances can ignore the message.
	key := fmt.Sprintf("lora:as:downlink:lock:%d:%s", pl.ApplicationID, pl.DevEUI)
	redisConn := h.redisPool.Get()
	defer redisConn.Close()

	_, err = redis.String(redisConn.Do("SET", key, "lock", "PX", int64(downlinkLockTTL/time.Millisecond), "NX"))
	if err != nil {
		if err == redis.ErrNil {
			// the payload is already being processed by an other instance
			return
		}
		log.Errorf("handler/mqtt: acquire downlink payload lock error: %s", err)
		return
	}

	h.dataDownChan <- pl
}

func (h *MQTTHandler) onConnected(c mqtt.Client) {
	log.Info("handler/mqtt: connected to mqtt broker")
	for {
		log.WithFields(log.Fields{
			"topic": h.downlinkTopic,
			"qos":   h.config.QOS,
		}).Info("handler/mqtt: subscribing to tx topic")
		if token := h.conn.Subscribe(h.downlinkTopic, h.config.QOS, h.txPayloadHandler); token.Wait() && token.Error() != nil {
			log.WithField("topic", h.downlinkTopic).Errorf("handler/mqtt: subscribe error: %s", token.Error())
			time.Sleep(time.Second)
			continue
		}
		return
	}
}

func (h *MQTTHandler) onConnectionLost(c mqtt.Client, reason error) {
	log.Errorf("handler/mqtt: mqtt connection error: %s", reason)
}

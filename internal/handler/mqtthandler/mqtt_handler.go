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
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/handler"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/garyburd/redigo/redis"
)

const txTopic = "application/+/node/+/tx"
const downlinkLockTTL = time.Millisecond * 100

var txTopicRegex = regexp.MustCompile(`application/(\w+)/node/(\w+)/tx`)

// MQTTHandler implements a MQTT handler for sending and receiving data by
// an application.
type MQTTHandler struct {
	conn         mqtt.Client
	dataDownChan chan handler.DataDownPayload
	wg           sync.WaitGroup
	redisPool    *redis.Pool
}

// NewHandler creates a new MQTTHandler.
func NewHandler(server, username, password, cafile, certFile, certKeyFile string) (handler.Handler, error) {
	h := MQTTHandler{
		dataDownChan: make(chan handler.DataDownPayload),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetOnConnectHandler(h.onConnected)
	opts.SetConnectionLostHandler(h.onConnectionLost)

	tlsconfig, err := newTLSConfig(cafile, certFile, certKeyFile)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"ca_cert":  cafile,
			"tls_cert": certFile,
			"tls_key":  certKeyFile,
		}).Fatalf("error loading mqtt certificate files")
	}
	if tlsconfig != nil {
		opts.SetTLSConfig(tlsconfig)
	}

	log.WithField("server", server).Info("handler/mqtt: connecting to mqtt broker")
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
	log.WithField("topic", txTopic).Info("handler/mqtt: unsubscribing from tx topic")
	if token := h.conn.Unsubscribe(txTopic); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: unsubscribe from %s error: %s", txTopic, token.Error())
	}
	log.Info("handler/mqtt: handling last items in queue")
	h.wg.Wait()
	close(h.dataDownChan)
	return nil
}

// SendDataUp sends a DataUpPayload.
func (h *MQTTHandler) SendDataUp(payload handler.DataUpPayload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("handler/mqtt: data-up payload marshal error: %s", err)
	}

	topic := fmt.Sprintf("application/%d/node/%s/rx", payload.ApplicationID, payload.DevEUI)
	log.WithField("topic", topic).Info("handler/mqtt: publishing data-up payload")
	if token := h.conn.Publish(topic, 0, false, b); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: publish data-up payload error: %s", err)
	}
	return nil
}

// SendJoinNotification sends a JoinNotification.
func (h *MQTTHandler) SendJoinNotification(payload handler.JoinNotification) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("handler/mqtt: join notification marshal error: %s", err)
	}
	topic := fmt.Sprintf("application/%d/node/%s/join", payload.ApplicationID, payload.DevEUI)
	log.WithField("topic", topic).Info("handler/mqtt: publishing join notification")
	if token := h.conn.Publish(topic, 0, false, b); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: publish join notification error: %s", err)
	}
	return nil
}

// SendACKNotification sends an ACKNotification.
func (h *MQTTHandler) SendACKNotification(payload handler.ACKNotification) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("handler/mqtt: ack notification marshal error: %s", err)
	}
	topic := fmt.Sprintf("application/%d/node/%s/ack", payload.ApplicationID, payload.DevEUI)
	log.WithField("topic", topic).Info("handler/mqtt: publishing ack notification")
	if token := h.conn.Publish(topic, 0, false, b); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: publish ack notification error: %s", err)
	}
	return nil
}

// SendErrorNotification sends an ErrorNotification.
func (h *MQTTHandler) SendErrorNotification(payload handler.ErrorNotification) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("handler/mqtt: error notification marshal error: %s", err)
	}
	topic := fmt.Sprintf("application/%d/node/%s/error", payload.ApplicationID, payload.DevEUI)
	log.WithField("topic", topic).Info("handler/mqtt: publishing error notification")
	if token := h.conn.Publish(topic, 0, false, b); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: publish error notification error: %s", err)
	}
	return nil
}

// DataDownChan returns the channel containing the received DataDownPayload.
func (h *MQTTHandler) DataDownChan() chan handler.DataDownPayload {
	return h.dataDownChan
}

func (h *MQTTHandler) txPayloadHandler(c mqtt.Client, msg mqtt.Message) {
	h.wg.Add(1)
	defer h.wg.Done()

	log.WithField("topic", msg.Topic()).Info("handler/mqtt: data-down payload received")

	// get the name of the application and node from the topic
	match := txTopicRegex.FindStringSubmatch(msg.Topic())
	if len(match) != 3 {
		log.WithField("topic", msg.Topic()).Error("handler/mqtt: topic regex match error")
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

	// set ApplicationID and DevEUI from topic
	var err error
	pl.ApplicationID, err = strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		log.WithFields(log.Fields{
			"topic": msg.Topic(),
		}).Errorf("handler/mqtt: parse application id error: %s", err)
		return
	}

	if err = pl.DevEUI.UnmarshalText([]byte(match[2])); err != nil {
		log.WithFields(log.Fields{
			"topic": msg.Topic(),
		}).Errorf("handler/mqtt: parse dev_eui error: %s", err)
		return
	}

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
	// As an unique id, the Reference field is used.
	key := fmt.Sprintf("lora:as:downlink:lock:%d:%s:%s", pl.ApplicationID, pl.DevEUI, pl.Reference)
	redisConn := config.C.Redis.Pool.Get()
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
		log.WithField("topic", txTopic).Info("handler/mqtt: subscribling to tx topic")
		if token := h.conn.Subscribe(txTopic, 2, h.txPayloadHandler); token.Wait() && token.Error() != nil {
			log.WithField("topic", txTopic).Errorf("handler/mqtt: subscribe error: %s", token.Error())
			time.Sleep(time.Second)
			continue
		}
		return
	}
}

func (h *MQTTHandler) onConnectionLost(c mqtt.Client, reason error) {
	log.Errorf("handler/mqtt: mqtt connection error: %s", reason)
}

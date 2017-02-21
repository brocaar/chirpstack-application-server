package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/brocaar/lorawan"
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
	dataDownChan chan DataDownPayload
	wg           sync.WaitGroup
	redisPool    *redis.Pool
}

// ACKNotification defines the payload sent to the application
// on an ACK event.
type ACKNotification struct {
	ApplicationID   int64         `json:"applicationID,string"`
	ApplicationName string        `json:"applicationName"`
	NodeName        string        `json:"nodeName"`
	DevEUI          lorawan.EUI64 `json:"devEUI"`
	Reference       string        `json:"reference"`
}

// ErrorNotification defines the payload sent to the application
// on an error event.
type ErrorNotification struct {
	ApplicationID   int64         `json:"applicationID,string"`
	ApplicationName string        `json:"applicationName"`
	NodeName        string        `json:"nodeName"`
	DevEUI          lorawan.EUI64 `json:"devEUI"`
	Type            string        `json:"type"`
	Error           string        `json:"error"`
}

// NewMQTTHandler creates a new MQTTHandler.
func NewMQTTHandler(p *redis.Pool, server, username, password string) (Handler, error) {
	h := MQTTHandler{
		dataDownChan: make(chan DataDownPayload),
		redisPool:    p,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetOnConnectHandler(h.onConnected)
	opts.SetConnectionLostHandler(h.onConnectionLost)

	log.WithField("server", server).Info("handler/mqtt: connecting to mqtt broker")
	h.conn = mqtt.NewClient(opts)
	if token := h.conn.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("handler/mqtt: connecting to broker error: %s", token.Error())
	}
	return &h, nil
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
func (h *MQTTHandler) SendDataUp(payload DataUpPayload) error {
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
func (h *MQTTHandler) SendJoinNotification(payload JoinNotification) error {
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
func (h *MQTTHandler) SendACKNotification(payload ACKNotification) error {
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
func (h *MQTTHandler) SendErrorNotification(payload ErrorNotification) error {
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
func (h *MQTTHandler) DataDownChan() chan DataDownPayload {
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

	var pl DataDownPayload
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

	// Since with MQTT all subscribers will receive the downlink messages sent
	// by the application, the first instance receiving the message must lock it,
	// so that other instances can ignore the message.
	// As an unique id, the Reference field is used.
	key := fmt.Sprintf("lora:as:downlink:lock:%d:%s:%s", pl.ApplicationID, pl.DevEUI, pl.Reference)
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

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gusseleet/lora-app-server/internal/join"
	"github.com/brocaar/lorawan/backend"
)

// JoinServerAPI implements the join-server API as documented in the LoRaWAN
// backend interfaces specification.
type JoinServerAPI struct{}

// NewJoinServerAPI create a new JoinServerAPI.
func NewJoinServerAPI() http.Handler {
	return &JoinServerAPI{}
}

// ServeHTTP implements the http.Handler interface.
func (a *JoinServerAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var basePL backend.BasePayload

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		a.returnError(w, http.StatusInternalServerError, backend.Other, "read body error")
		return
	}

	err = json.Unmarshal(b, &basePL)
	if err != nil {
		a.returnError(w, http.StatusBadRequest, backend.Other, err.Error())
		return
	}

	log.WithFields(log.Fields{
		"message_type":   basePL.MessageType,
		"sender_id":      basePL.SenderID,
		"receiver_id":    basePL.ReceiverID,
		"transaction_id": basePL.TransactionID,
	}).Info("js: request received")

	switch basePL.MessageType {
	case backend.JoinReq:
		a.handleJoinReq(w, b)
	default:
		a.returnError(w, http.StatusBadRequest, backend.Other, fmt.Sprintf("invalid MessageType: %s", basePL.MessageType))
	}
}

func (a *JoinServerAPI) returnError(w http.ResponseWriter, code int, resultCode backend.ResultCode, msg string) {
	log.WithFields(log.Fields{
		"error": msg,
	}).Error("js: error handling request")

	w.WriteHeader(code)

	pl := backend.Result{
		ResultCode:  resultCode,
		Description: msg,
	}
	b, err := json.Marshal(pl)
	if err != nil {
		log.WithError(err).Error("marshal json error")
		return
	}

	w.Write(b)
}

func (a *JoinServerAPI) returnPayload(w http.ResponseWriter, code int, pl interface{}) {
	w.WriteHeader(code)

	b, err := json.Marshal(pl)
	if err != nil {
		log.WithError(err).Error("marshal json error")
		return
	}

	w.Write(b)
}

func (a *JoinServerAPI) handleJoinReq(w http.ResponseWriter, b []byte) {
	var joinReqPL backend.JoinReqPayload
	err := json.Unmarshal(b, &joinReqPL)
	if err != nil {
		a.returnError(w, http.StatusBadRequest, backend.Other, err.Error())
		return
	}

	ans := join.HandleJoinRequest(joinReqPL)

	log.WithFields(log.Fields{
		"message_type":   ans.BasePayload.MessageType,
		"sender_id":      ans.BasePayload.SenderID,
		"receiver_id":    ans.BasePayload.ReceiverID,
		"transaction_id": ans.BasePayload.TransactionID,
		"result_code":    ans.Result.ResultCode,
	}).Info("js: sending response")

	a.returnPayload(w, http.StatusOK, ans)
}

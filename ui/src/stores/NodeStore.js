import { EventEmitter } from "events";
import "whatwg-fetch";
import ReconnectingWebSocket from 'reconnecting-websocket';
import dispatcher from "../dispatcher";
import sessionStore from "./SessionStore";
import { checkStatus, errorHandler, errorHandlerIgnoreNotFound } from "./helpers";

class NodeStore extends EventEmitter {
  getAll(applicationID, pageSize, offset, search, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/devices?limit="+pageSize+"&offset="+offset+"&search="+search, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if(typeof(responseData.result) === "undefined") {
          callbackFunc(0, []);
        } else {
          callbackFunc(responseData.totalCount, responseData.result);
        }
      })
      .catch(errorHandler);
  }

  getNode(applicationID, name, callbackFunc) {
    fetch("/api/devices/"+name, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createNode(applicationID, node, callbackFunc) {
    fetch("/api/devices", {method: "POST", body: JSON.stringify(node), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateNode(applicationID, devEUI, node, callbackFunc) {
    fetch("/api/devices/"+devEUI, {method: "PUT", body: JSON.stringify(node), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteNode(applicationID, devEUI, callbackFunc) {
    fetch("/api/devices/"+devEUI, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createNodeKeys(devEUI, nodeKeys, callbackFunc) {
    fetch("/api/devices/"+devEUI+"/keys", {method: "POST", body: JSON.stringify(nodeKeys), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getNodeKeys(devEUI, callbackFunc) {
    fetch("/api/devices/"+devEUI+"/keys", {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandlerIgnoreNotFound);
  }
  updateNodeKeys(devEUI, nodeKeys, callbackFunc) {
    fetch("/api/devices/"+devEUI+"/keys", {method: "PUT", body: JSON.stringify(nodeKeys), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteNodeKeys(devEUI, callbackFunc) {
    fetch("/api/devices/"+devEUI+"/keys", {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  activateNode(devEUI, activation, callbackFunc) {
    fetch("/api/devices/"+devEUI+"/activate", {method: "POST", body: JSON.stringify(activation), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getActivation(devEUI, callbackFunc) {
    fetch("/api/devices/"+devEUI+"/activation", {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandlerIgnoreNotFound);
  }

  getRandomDevAddr(devEUI, callbackFunc) {
    fetch("/api/devices/"+devEUI+"/getRandomDevAddr", {method: "POST", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getFrameLogsConnection(devEUI, onOpen, onClose, onData) {
    const loc = window.location;
    const wsURL = (() => {
      if (loc.host === "localhost:3000") {
        return `wss://localhost:8080/api/devices/${devEUI}/frames`;
      }

      const wsProtocol = loc.protocol === "https:" ? "wss:" : "ws:";
      return `${wsProtocol}//${loc.host}/api/devices/${devEUI}/frames`;
    });

    const conn = new ReconnectingWebSocket(wsURL, ["Bearer", sessionStore.getToken()]);
    conn.onopen = () => {
      console.log('connected to', wsURL);
      onOpen();
    };

    conn.onclose = () => {
      console.log('closing', wsURL);
      onClose();
    };

    conn.onmessage = (e) => {
      const msg = JSON.parse(e.data);
      if (msg.error !== undefined) {
        dispatcher.dispatch({
          type: "CREATE_ERROR",
          error: {
            code: msg.error.grpcCode,
            error: msg.error.message,
          },
        });
      } else if (msg.result !== undefined) {
        onData(msg.result);
      }
    };

    return conn;
  }

  getEventLogsConnection(devEUI, onOpen, onClose, onData) {
    const loc = window.location;
    const wsURL = (() => {
      if (loc.host === "localhost:3000") {
        return `wss://localhost:8080/api/devices/${devEUI}/events`;
      }

      const wsProtocol = loc.protocol === "https:" ? "wss:" : "ws:";
      return `${wsProtocol}//${loc.host}/api/devices/${devEUI}/events`;
    });

    const conn = new ReconnectingWebSocket(wsURL, ["Bearer", sessionStore.getToken()]);
    conn.onopen = () => {
      console.log('connected to', wsURL);
      onOpen();
    };

    conn.onclose = () => {
      console.log('closing', wsURL);
      onClose();
    };

    conn.onmessage = (e) => {
      const msg = JSON.parse(e.data);
      if (msg.error !== undefined) {
        dispatcher.dispatch({
          type: "CREATE_ERROR",
          error: {
            code: msg.error.grpcCode,
            error: msg.error.message,
          },
        });
      } else if (msg.result !== undefined) {
        onData(msg.result);
      }
    };

    return conn;
  }
}

const nodeStore = new NodeStore();

export default nodeStore;

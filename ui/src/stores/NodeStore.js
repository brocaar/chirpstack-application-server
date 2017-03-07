import { EventEmitter } from "events";
import "whatwg-fetch";
import tokenStore from "./TokenStore";
import { checkStatus, errorHandler } from "./helpers";

var checkGetActivationStatus = (response) => {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    // dont return an error when there is no activation
  }
};

class NodeStore extends EventEmitter {
  getAll(applicationID, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/nodes?limit=999", {headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if(typeof(responseData.result) === "undefined") {
          callbackFunc([]);
        } else {
          callbackFunc(responseData.result);
        }
      })
      .catch(errorHandler);
  }

  getNode(applicationID, name, callbackFunc) {
    fetch("/api/nodes/"+name, {headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createNode(applicationID, node, callbackFunc) {
    fetch("/api/nodes", {method: "POST", body: JSON.stringify(node), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateNode(applicationID, devEUI, node, callbackFunc) {
    fetch("/api/nodes/"+devEUI, {method: "PUT", body: JSON.stringify(node), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteNode(applicationID, devEUI, callbackFunc) {
    fetch("/api/nodes/"+devEUI, {method: "DELETE", headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  activateNode(applicationID, devEUI, activation, callbackFunc) {
    fetch("/api/nodes/"+devEUI+"/activation", {method: "POST", body: JSON.stringify(activation), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getActivation(applicationID, devEUI, callbackFunc) {
    fetch("/api/nodes/"+devEUI+"/activation", {headers: tokenStore.getHeader()})
      .then(checkGetActivationStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const nodeStore = new NodeStore();

export default nodeStore;

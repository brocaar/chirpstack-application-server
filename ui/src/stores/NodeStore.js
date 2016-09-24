import { EventEmitter } from "events";
import "whatwg-fetch";
import dispatcher from "../dispatcher";

var checkStatus = (response) => {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    throw response.json();
  }
};

var errorHandler = (error) => {
  console.log(error);
  error.then((data) => dispatcher.dispatch({
    type: "CREATE_ERROR",
    error: data,
  }));
};

class NodeStore extends EventEmitter {
  getAll(callbackFunc) {
    fetch("/api/node?limit=999")
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData.result);
      })
      .catch(errorHandler);
  }

  getNode(devEUI, callbackFunc) {
    fetch("/api/node/"+devEUI)
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createNode(node, callbackFunc) {
    fetch("/api/node", {method: "POST", body: JSON.stringify(node)})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateNode(devEUI, node, callbackFunc) {
    fetch("/api/node/"+devEUI, {method: "PUT", body: JSON.stringify(node)})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);

  }
}

const nodeStore = new NodeStore();

export default nodeStore;

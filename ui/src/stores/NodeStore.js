import { EventEmitter } from "events";
import { hashHistory } from "react-router";
import "whatwg-fetch";
import dispatcher from "../dispatcher";
import tokenStore from "./TokenStore";

var checkStatus = (response) => {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    throw response.json();
  }
};

var errorHandler = (error) => {
  console.log(error);
  error.then((data) => {
    dispatcher.dispatch({
      type: "CREATE_ERROR",
      error: data,
    });

    if (data.Code === 16) {
      hashHistory.push("/jwt");
    }
  });
};

class NodeStore extends EventEmitter {
  getAll(applicationName, callbackFunc) {
    fetch("/api/applications/"+applicationName+"/nodes?limit=999", {headers: tokenStore.getHeader()})
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

  getNode(applicationName, name, callbackFunc) {
    fetch("/api/applications/"+applicationName+"/nodes/"+name, {headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createNode(node, callbackFunc) {
    fetch("/api/applications/"+node.applicationName+"/nodes", {method: "POST", body: JSON.stringify(node), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateNode(node, callbackFunc) {
    fetch("/api/applications/"+node.applicationName+"/nodes/"+node.name, {method: "PUT", body: JSON.stringify(node), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteNode(applicationName, name, callbackFunc) {
    fetch("/api/applications/"+applicationName+"/nodes/"+name, {method: "DELETE", headers: tokenStore.getHeader()})
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

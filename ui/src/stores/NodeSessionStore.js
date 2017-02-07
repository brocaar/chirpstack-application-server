import { EventEmitter } from "events";
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

var checkGetStatus = (response) => {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    //throw response.json(); TODO: fix node-session error code when it does not exist
  }
};


var errorHandler = (error) => {
  console.log(error);
  error.then((data) => dispatcher.dispatch({
    type: "CREATE_ERROR",
    error: data,
  }));
};

class NodeSessionStore extends EventEmitter {
  getNodeSession(applicationName, nodeName, callbackFunc) {
    fetch("/api/applications/"+applicationName+"/nodes/"+nodeName+"/session", {headers: tokenStore.getHeader()})
      .then(checkGetStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createNodeSession(applicationName, nodeName, session, callbackFunc) {
    fetch("/api/applications/"+applicationName+"/nodes/"+nodeName+"/session", {method: "POST", body: JSON.stringify(session), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateNodeSession(applicationName, nodeName, session, callbackFunc) {
    fetch("/api/applications/"+applicationName+"/nodes/"+nodeName+"/session", {method: "PUT", body: JSON.stringify(session), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteNodeSession(applicationName, nodeName, callbackFunc) {
    fetch("/api/applications/"+applicationName+"/nodes/"+nodeName+"/session", {method: "DELETE", headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getRandomDevAddr(applicationName, nodeName, callbackFunc) {
    fetch("/api/applications/"+applicationName+"/nodes/"+nodeName+"/getRandomDevAddr", {method: "POST", headers: tokenStore.getHeader()}) 
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const nodeSessionStore = new NodeSessionStore();

export default nodeSessionStore;

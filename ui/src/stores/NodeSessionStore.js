import { EventEmitter } from "events";
import "whatwg-fetch";
import sessionStore from "./SessionStore";
import { checkStatus, errorHandler } from "./helpers";

class NodeSessionStore extends EventEmitter {
  getRandomDevAddr(applicationID, devEUI, callbackFunc) {
    fetch("/api/nodes/"+devEUI+"/getRandomDevAddr", {method: "POST", headers: sessionStore.getHeader()}) 
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

import { EventEmitter } from "events";
import "whatwg-fetch";
import dispatcher from "../dispatcher";
import sessionStore from "./SessionStore";
import { checkStatus, errorHandler, errorHandlerIgnoreNotFound } from "./helpers";


class InternalStore extends EventEmitter {
    globalSearch(search, pageSize, offset, callbackFunc) {
        fetch(`/api/internal/search?limit=${pageSize}&offset=${offset}&search=${encodeURIComponent(search)}`, {headers: sessionStore.getHeader()})
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
}


const internalStore = new InternalStore();

export default internalStore;

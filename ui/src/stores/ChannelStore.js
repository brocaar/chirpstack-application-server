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

class ChannelStore extends EventEmitter {
  getAllChannelLists(callbackFunc) {
    fetch("/api/channelList?limit=999")
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData.result);
      })
      .catch(errorHandler);
  }

  getChannelList(id, callbackFunc) {
    fetch("/api/channelList/"+id)
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createChannelList(list, callbackFunc) {
    fetch("/api/channelList", {method: "POST", body: JSON.stringify(list)})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteChannelList(id, callbackFunc) {
    fetch("/api/channelList/"+id, {method: "DELETE"})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateChannelList(id, list, callbackFunc) {
    fetch("/api/channelList/"+id, {method: "PUT", body: JSON.stringify(list)})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const channelStore = new ChannelStore();

export default channelStore;

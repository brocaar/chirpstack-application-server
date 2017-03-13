import { EventEmitter } from "events";
import "whatwg-fetch";
import sessionStore from "./SessionStore";
import { checkStatus, errorHandler } from "./helpers";

class UserStore extends EventEmitter {
  getAll(search, callbackFunc) {
    fetch("/api/users?limit=999&search="+encodeURIComponent(search), {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if (typeof(responseData.result) === "undefined") {
          callbackFunc([]);
        } else {
          callbackFunc(responseData.result);
        }
      })
      .catch(errorHandler);
  } 

  getUser(userID, callbackFunc) {
    fetch("/api/users/"+userID, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createUser(user, callbackFunc) {
    fetch("/api/users", {method: "POST", body: JSON.stringify(user), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateUser(userID, user, callbackFunc) {
    fetch("/api/users/"+userID, {method: "PUT", body: JSON.stringify(user), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteUser(userID, callbackFunc) {
    fetch("/api/users/"+userID, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updatePassword(userID, password, callbackFunc) {
    fetch("/api/users/"+userID+"/password", {method: "PUT", body: JSON.stringify(password), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const userStore = new UserStore();

export default userStore;

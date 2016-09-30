import { EventEmitter } from "events";

class TokenStore extends EventEmitter {
  constructor() {
    super();
    this.token = "";
  }

  setToken(token) {
    this.token = token;
  }

  getToken() {
    return this.token;
  }

  getHeader() {
    if (this.token !== "") {
      return {
        "Grpc-Metadata-Authorization": this.token,
      }
    } else {
      return {}
    }
  }
}

const tokenStore = new TokenStore();

export default tokenStore;

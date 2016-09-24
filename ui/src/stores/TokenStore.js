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
    return {
      "Grpc-Metadata-Authorization": this.token,
    }
  }
}

const tokenStore = new TokenStore();

export default tokenStore;

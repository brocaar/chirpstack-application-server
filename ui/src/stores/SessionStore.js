import { EventEmitter } from "events";

import Swagger from "swagger-client";
import { checkStatus, errorHandler, errorHandlerLogin } from "./helpers";


class SessionStore extends EventEmitter {
  constructor() {
    super();
    this.client = null;
    this.user = null;
    this.organizations = [];
    this.settings = {};
    this.branding = {};

    this.swagger = Swagger("/swagger/internal.swagger.json", this.getClientOpts())
    
    this.swagger.then(client => {
      this.client = client;

      if (this.getToken() !== null) {
        this.fetchProfile(() => {});
      }
    });
  }

  getClientOpts() {
    return {
      requestInterceptor: (req) => {
        if (this.getToken() !== null) {
          req.headers["Grpc-Metadata-Authorization"] = "Bearer " + this.getToken();
        }
        return req;
      },
    }
  }

  setToken(token) {
    localStorage.setItem("jwt", token);
  }

  getToken() {
    return localStorage.getItem("jwt");
  }

  getOrganizationID() {
    const orgID = localStorage.getItem("organizationID");
    if (orgID === "") {
      return null;
    }

    return orgID;
  }

  setOrganizationID(id) {
    localStorage.setItem("organizationID", id);
    this.emit("organization.change");
  }

  getRowsPerPage() {
    const rowsPerPage = localStorage.getItem("rowsPerPage");
    if (rowsPerPage === "" || rowsPerPage === null) {
      return 10;
    }

    return JSON.parse(rowsPerPage);
  }

  setRowsPerPage(rowsPerPage) {
    localStorage.setItem("rowsPerPage", JSON.stringify(rowsPerPage));
  }

  getUser() {
    return this.user;
  }

  getSettings() {
    return this.settings;
  }

  isAdmin() {
    if (this.user === undefined || this.user === null) {
      return false;
    }
    return this.user.isAdmin;
  }

  isOrganizationAdmin(organizationID) {
    for (let i = 0; i < this.organizations.length; i++) {
      if (this.organizations[i].organizationID === organizationID) {
        return this.organizations[i].isAdmin;
      }
    }
  }

  isOrganizationDeviceAdmin(organizationID) {
    for (let i = 0; i < this.organizations.length; i++) {
      if (this.organizations[i].organizationID === organizationID) {
        return this.organizations[i].isAdmin || this.organizations[i].isDeviceAdmin;
      }
    }
  }

  isOrganizationGatewayAdmin(organizationID) {
    for (let i = 0; i < this.organizations.length; i++) {
      if (this.organizations[i].organizationID === organizationID) {
        return this.organizations[i].isAdmin || this.organizations[i].isGatewayAdmin;
      }
    }
  }

  login(login, callBackFunc) {
    this.swagger.then(client => {
      client.apis.InternalService.Login({body: login})
        .then(checkStatus)
        .then(resp => {
          this.setToken(resp.obj.jwt);
          this.fetchProfile(callBackFunc);
        })
        .catch(errorHandlerLogin);
    });
  }

  openidConnectLogin(code, state, callbackFunc) {
    this.swagger.then(client => {
      client.apis.InternalService.OpenIDConnectLogin({
        code: code,
        state: state,
      })
        .then(checkStatus)
        .then(resp => {
          this.setToken(resp.obj.jwtToken);
          this.fetchProfile(callbackFunc);
        })
        .catch(errorHandler);
    });
  }

  logout(emit, callBackFunc) {
    localStorage.clear();
    this.user = null;
    this.organizations = [];
    this.settings = {};

    if (emit === true) {
      this.emit("change");
    }

    callBackFunc();
  }

  fetchProfile(callBackFunc) {
    this.swagger.then(client => {
      client.apis.InternalService.Profile({})
        .then(checkStatus)
        .then(resp => {
          this.user = resp.obj.user;

          if(resp.obj.organizations !== undefined) {
            this.organizations = resp.obj.organizations;
          }

          if(resp.obj.settings !== undefined) {
            this.settings = resp.obj.settings;
          }

          this.emit("change");
          callBackFunc();
        })
        .catch(errorHandler);
    });
  }

  globalSearch(search, limit, offset, callbackFunc) {
    this.swagger.then(client => {
      client.apis.InternalService.GlobalSearch({
        search: search,
        limit: limit,
        offset: offset,
      })
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
      });
  }
}

const sessionStore = new SessionStore();
export default sessionStore;

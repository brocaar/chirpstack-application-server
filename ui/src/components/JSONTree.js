import React, { Component } from "react";

import { unparse } from "uuid-parse";
import JSONTreeOriginal from "react-json-tree";


class JSONTree extends Component {
  render() {
    const theme = {
      scheme: 'google',
      author: 'seth wright (http://sethawright.com)',
      base00: '#000000',
      base01: '#282a2e',
      base02: '#373b41',
      base03: '#969896',
      base04: '#b4b7b4',
      base05: '#c5c8c6',
      base06: '#e0e0e0',
      base07: '#ffffff',
      base08: '#CC342B',
      base09: '#F96A38',
      base0A: '#FBA922',
      base0B: '#198844',
      base0C: '#3971ED',
      base0D: '#3971ED',
      base0E: '#A36AC7',
      base0F: '#3971ED',
    }

    // :(
    let data = JSON.parse(JSON.stringify(this.props.data));

    if ("devEUI" in data) {
      data.devEUI = base64ToHex(data.devEUI);
    }

    if ("devAddr" in data) {
      data.devAddr = base64ToHex(data.devAddr);
    }

    if ("rxInfo" in data && data.rxInfo !== null) {
      for (let i = 0; i < data.rxInfo.length; i++) {
        if ("gatewayID" in data.rxInfo[i]) {
          data.rxInfo[i].gatewayID = base64ToHex(data.rxInfo[i].gatewayID);
        }

        if ("uplinkID" in data.rxInfo[i]) {
          const id = Buffer.from(data.rxInfo[i].uplinkID, 'base64');
          data.rxInfo[i].uplinkID = unparse(id);
        }
      }
    }

    if ("txInfo" in data && data.txInfo !== null) {
      if ("gatewayID" in data.txInfo) {
        data.txInfo.gatewayID = base64ToHex(data.txInfo.gatewayID);
      }
    }

    return(
      <JSONTreeOriginal
        data={data}
        theme={theme}
        hideRoot={true}
        shouldExpandNode={() => {return true}}
      />
    );
  }
}

function base64ToHex(str) {
  return Buffer.from(str, 'base64').toString('hex');
}

export default JSONTree;

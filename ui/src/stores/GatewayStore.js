import { EventEmitter } from "events";

import moment from "moment";

class GatewayStore extends EventEmitter {
  getAll(pageSize, offset, callbackFunc) {
    callbackFunc(3, [
      {mac: "0101010101010101", name: "gateway-1", description: "Description 1"},
      {mac: "0101010101010102", name: "gateway-2", description: "Description 2"},
      {mac: "0101010101010103", name: "gateway-2", description: "Description 3"},
    ]);
  }

  getGatewayStats(mac, interval, start, end, callbackFunc) {
    let stats = [];
    let d = moment(start);

    while (d.isBefore(moment(end))) {
      stats.push({
        timestamp: d.format(),
        rxPacketsReceived: (Math.random() * (30 - 20) + 20),
        rxPacketsReceivedOK: (Math.random() * (30 - 10) + 10),
        txPacketsReceived: (Math.random() * (30 - 20) + 20),
        txPacketsEmitted: (Math.random() * (30 - 10) + 10),
      });

      d = d.add(1, interval.toLowerCase() + 's');
    }

    callbackFunc(stats);
  }

  getGateway(mac, callbackFunc) {
    callbackFunc({
      mac: "010101010101010101",
      name: "test-gateway",
      description: "gateway located on the rooftop",
      createdAt: "2017-03-20T13:00:00+02:00",
      updatedAt: "2017-03-20T13:00:00+02:00",
      firstSeenAt: "2017-03-20T13:00:00+02:00",
      lastSeenAt: "2017-03-20T13:00:00+02:00",
      latitude: 52.3740693,
      longitude: 4.9121673,
      altitude: 10,
    });
  }

  createGateway(gateway, callbackFunc) {
    callbackFunc({});
  }

  deleteGateway(mac, callbackFunc) {
    callbackFunc({});
  }

  updateGateway(mac, gateway, callbackFunc) {
    callbackFunc({});
  }
}

const gatewayStore = new GatewayStore();

export default gatewayStore;
